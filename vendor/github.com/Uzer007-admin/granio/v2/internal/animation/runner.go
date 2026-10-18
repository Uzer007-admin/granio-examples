package animation

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Uzer007-admin/granio/v2"
)

const initialAnimationsCapacity = 16

// Runner is the main driver for animations package
type Runner struct {
	// Visible is supplied by a driver on its UI thread, and can include hidden
	// ancestors, inactive tabs, and minimized windows. Nil checks Object.Visible.
	Visible func(granio.CanvasObject) bool
	// animationMutex synchronizes access to `animations` and `pendingAnimations`
	// between the runner goroutine and calls to Start and Stop
	animationMutex sync.RWMutex

	// animations is the list of animations that are being ticked in the current frame
	animations []*anim

	// pendingAnimations is animations that have been started but not yet picked up
	// by the runner goroutine to be ticked each frame
	pendingAnimations []*anim

	// nextFrameAnimations is the list of animations that will be ticked in the next frame.
	// It is accessed only by the runner goroutine and accumulates the continuing animations
	// during a tick that are not completed, plus the pendingAnimations picked up at the end of the frame.
	// At the end of a full frame of animations, the nextFrameAnimations slice is swapped with
	// the current `animations` slice which is then cleared out, while holding the mutex.
	nextFrameAnimations []*anim

	runnerStarted atomic.Bool
	runnable      atomic.Bool
}

// Start will register the passed application and initiate its ticking.
func (r *Runner) Start(a *granio.Animation) {
	if a == nil || a.Tick == nil {
		return
	}
	r.animationMutex.Lock()
	defer r.animationMutex.Unlock()
	r.runnable.Store(true)
	// Starting the same handle restarts it; pending copies must never accumulate.
	for _, previous := range r.animations {
		if previous.a == a {
			previous.setStopped()
		}
	}
	// Only the runner reads active slices outside the lock. Pending slices can
	// be compacted here, so repeated retargets before a paint use bounded space.
	pending := r.pendingAnimations[:0]
	for _, previous := range r.pendingAnimations {
		if previous.a == a || previous.isStopped() {
			previous.setStopped()
		} else {
			pending = append(pending, previous)
		}
	}
	clear(r.pendingAnimations[len(pending):])
	r.pendingAnimations = pending

	if !r.runnerStarted.Load() {
		r.runnerStarted.Store(true)
		if r.animations == nil {
			// initialize with excess capacity to avoid re-allocations
			// on subsequent Starts
			r.animations = make([]*anim, 0, initialAnimationsCapacity)
		}
		r.animations = append(r.animations, newAnim(a))
	} else {
		if r.pendingAnimations == nil {
			// initialize with excess capacity to avoid re-allocations
			// on subsequent Starts
			r.pendingAnimations = make([]*anim, 0, initialAnimationsCapacity)
		}
		r.pendingAnimations = append(r.pendingAnimations, newAnim(a))
	}
}

// Stop causes an animation to stop ticking (if it was still running) and removes it from the runner.
func (r *Runner) Stop(a *granio.Animation) {
	r.animationMutex.Lock()
	defer r.animationMutex.Unlock()

	newList := make([]*anim, 0, len(r.animations))
	for _, item := range r.animations {
		if item.a != a {
			newList = append(newList, item)
		} else {
			item.setStopped()
		}
	}
	r.animations = newList

	newList = make([]*anim, 0, len(r.pendingAnimations))
	for _, item := range r.pendingAnimations {
		if item.a != a {
			newList = append(newList, item)
		} else {
			item.setStopped()
		}
	}
	r.pendingAnimations = newList
}

// TickAnimations progresses all running animations by one tick.
// This will be called from the driver to update objects immediately before next paint.
func (r *Runner) TickAnimations() {
	if !r.runnerStarted.Load() {
		return
	}

	done := r.runOneFrame()

	if done {
		r.animationMutex.Lock()
		if len(r.animations) == 0 && len(r.pendingAnimations) == 0 {
			r.runnerStarted.Store(false)
		}
		r.animationMutex.Unlock()
	}
}

// HasAnimations is a cheap concurrent check for the event loop's idle policy.
func (r *Runner) HasAnimations() bool { return r.runnerStarted.Load() && r.runnable.Load() }

// VisitObjects visits scoped owners without copying the animation list.
func (r *Runner) VisitObjects(visit func(granio.CanvasObject)) {
	r.animationMutex.RLock()
	defer r.animationMutex.RUnlock()
	for _, list := range [][]*anim{r.animations, r.pendingAnimations} {
		for _, a := range list {
			if a.a.Object != nil && !a.isStopped() {
				visit(a.a.Object)
			}
		}
	}
}

func (r *Runner) runOneFrame() (done bool) {
	r.animationMutex.Lock()
	oldList := r.animations
	r.animationMutex.Unlock()
	r.runnable.Store(false)
	for _, a := range oldList {
		if !a.isStopped() && r.tickAnimation(a) && !a.isStopped() {
			r.nextFrameAnimations = append(r.nextFrameAnimations, a)
		}
	}

	r.animationMutex.Lock()
	// Stop/Clear may have run from a later Tick callback after an earlier item
	// was scheduled for the next frame. Release those references immediately.
	next := r.nextFrameAnimations[:0]
	for _, item := range r.nextFrameAnimations {
		if !item.isStopped() {
			next = append(next, item)
		}
	}
	clear(r.nextFrameAnimations[len(next):])
	r.nextFrameAnimations = next
	// nil out old r.animations for re-use as next r.nextFrameAnimations
	tmp := r.animations
	clear(tmp)
	r.animations = append(r.nextFrameAnimations, r.pendingAnimations...)
	if len(r.pendingAnimations) > 0 {
		r.runnable.Store(true)
	}
	r.nextFrameAnimations = tmp[:0]
	clear(r.pendingAnimations)
	r.pendingAnimations = r.pendingAnimations[:0]
	done = len(r.animations) == 0
	r.animationMutex.Unlock()
	return done
}

// tickAnimation will process a frame of animation and return true if this should continue animating
func (r *Runner) tickAnimation(a *anim) bool {
	duration := a.a.Duration
	now := time.Now()
	disabled := a.a.Disabled
	if app := granio.CurrentApp(); app != nil && app.Settings() != nil {
		disabled = disabled || !app.Settings().ShowAnimations()
		if policy, ok := app.Settings().Theme().(granio.MotionTheme); ok {
			disabled = disabled || policy.Motion(granio.MotionHover).Disabled
		}
	}
	if obj, ok := a.a.Object.(interface{ Theme() granio.Theme }); ok {
		if policy, ok := obj.Theme().(granio.MotionTheme); ok {
			disabled = disabled || policy.Motion(granio.MotionHover).Disabled
		}
	}
	if disabled {
		end := float32(1)
		if a.a.AutoReverse && a.reverse {
			end = 0
		}
		a.a.Tick(end)
		a.finish()
		return false
	}
	if obj := a.a.Object; obj != nil {
		visible := obj.Visible()
		if visible && r.Visible != nil {
			visible = r.Visible(obj)
		}
		if !visible {
			if a.pausedAt.IsZero() {
				a.pausedAt = now
			}
			return true
		}
		if !a.pausedAt.IsZero() {
			elapsed := now.Sub(a.pausedAt)
			a.start = a.start.Add(elapsed)
			a.pinTime = a.pinTime.Add(elapsed)
			a.pausedAt = time.Time{}
		}
	}
	r.runnable.Store(true)

	if duration != a.lastDuration {
		a.pinProgress = a.progressFraction(now, a.lastDuration)
		a.pinTime = now
		a.lastDuration = duration
	}

	if !now.Before(a.start.Add(duration)) {
		if a.reverse {
			a.a.Tick(0.0)
			if a.repeatsLeft == 0 {
				a.finish()
				return false
			}
			a.reverse = false
		} else {
			a.a.Tick(1.0)
			if a.a.AutoReverse {
				a.reverse = true
			}
		}
		if !a.reverse {
			if a.repeatsLeft == 0 {
				a.finish()
				return false
			}
			if a.repeatsLeft > 0 {
				a.repeatsLeft--
			}
		}

		a.resetCycle(time.Now())
		return true
	}

	val := a.progressFraction(now, duration)
	curve := a.a.Curve
	if curve == nil {
		curve = granio.AnimationEaseInOut
	}
	if a.reverse {
		a.a.Tick(curve(1 - val))
	} else {
		a.a.Tick(curve(val))
	}

	return true
}

// Clear cancels handles on driver shutdown, including paused scopes.
func (r *Runner) Clear() {
	r.animationMutex.Lock()
	defer r.animationMutex.Unlock()
	for _, list := range [][]*anim{r.animations, r.pendingAnimations} {
		for _, a := range list {
			a.setStopped()
		}
	}
	// An in-flight tick may still read the old active slice. Mark its entries
	// stopped without mutating its storage; runOneFrame discards them on return.
	r.animations, r.pendingAnimations = nil, nil
	r.runnerStarted.Store(false)
	r.runnable.Store(false)
}
