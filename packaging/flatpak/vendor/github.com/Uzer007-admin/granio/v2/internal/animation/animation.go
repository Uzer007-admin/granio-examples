package animation

import (
	"sync/atomic"
	"time"

	"github.com/Uzer007-admin/granio/v2"
)

type anim struct {
	a            *granio.Animation
	repeatsLeft  int
	reverse      bool
	start        time.Time
	lastDuration time.Duration
	pinTime      time.Time
	pinProgress  float32
	stopped      atomic.Bool
	finished     atomic.Bool
	pausedAt     time.Time
}

func (a *anim) finish() {
	if !a.isStopped() && a.finished.CompareAndSwap(false, true) && a.a.OnFinished != nil {
		a.a.OnFinished()
	}
}

func newAnim(a *granio.Animation) *anim {
	now := time.Now()
	animate := &anim{
		a:            a,
		start:        now,
		lastDuration: a.Duration,
		pinTime:      now,
	}
	animate.repeatsLeft = a.RepeatCount
	return animate
}

func (a *anim) progressFraction(now time.Time, duration time.Duration) float32 {
	remaining := a.start.Add(duration).Sub(a.pinTime)
	if remaining <= 0 {
		return 1
	}
	val := a.pinProgress + (1-a.pinProgress)*float32(now.Sub(a.pinTime))/float32(remaining)
	if val > 1 {
		return 1
	}
	if val < 0 {
		return 0
	}
	return val
}

func (a *anim) resetCycle(now time.Time) {
	a.start = now
	a.pinTime = now
	a.pinProgress = 0
	a.lastDuration = a.a.Duration
}

func (a *anim) setStopped() {
	a.stopped.Store(true)
}

func (a *anim) isStopped() bool {
	return a.stopped.Load()
}
