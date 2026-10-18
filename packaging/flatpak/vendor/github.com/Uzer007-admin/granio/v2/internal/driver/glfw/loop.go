package glfw

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/app"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/internal/driver/common"
	"github.com/Uzer007-admin/granio/v2/internal/goos"
	"github.com/Uzer007-admin/granio/v2/internal/painter"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
)

type funcData struct {
	f    func()
	done chan struct{} // Zero allocation signalling channel
}

// channel for queuing functions on the main thread
var (
	funcQueue        = async.NewUnboundedChanWithNotify[funcData](wakeMainLoop)
	running, drained atomic.Bool
	funcQueueMu      sync.RWMutex
)

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
	async.SetMainGoroutine()
}

// force a function f to run on the main thread
func runOnMain(f func()) {
	runOnMainWithWait(f, true)
}

// force a function f to run on the main thread and specify if we should wait for it to return
func runOnMainWithWait(f func(), wait bool) {
	// If we are on main before app run just execute - otherwise add it to the main queue and wait.
	// Once the loop has stopped, UI callbacks cannot be executed on another thread.
	if drained.Load() {
		return
	}
	if !running.Load() && async.IsMainGoroutine() {
		f()
		return
	}

	if wait {
		done := common.DonePool.Get()
		defer common.DonePool.Put(done)

		funcQueueMu.RLock()
		if drained.Load() {
			funcQueueMu.RUnlock()
			return
		}
		funcQueue.In() <- funcData{f: f, done: done}
		funcQueueMu.RUnlock()
		<-done
	} else {
		funcQueueMu.RLock()
		if drained.Load() {
			funcQueueMu.RUnlock()
			return
		}
		funcQueue.In() <- funcData{f: f}
		funcQueueMu.RUnlock()
	}
}

func decideRepaint(visible, ready bool, checkDirtyAndClear func() bool) bool {
	return visible && ready && checkDirtyAndClear()
}

func (d *gLDriver) drawSingleFrame() {
	refreshed := false
	for _, win := range d.AllWindows() {
		w, _ := win.(*window)
		if w.closing {
			continue
		}

		if decideRepaint(w.visible, w.frame.ready(), w.canvas.CheckDirtyAndClear) {
			w.RunWithContext(func() {
				if w.driver.repaintWindow(w) {
					refreshed = true
				}
			})
			w.updateAccessibility()
			w.updateInputMethod()
		} else {
			w.markCacheAlive()
		}
	}
	cache.Clean(refreshed)
}

func (w *window) markCacheAlive() {
	threshold := time.Now().Add(10*time.Second - cache.ValidDuration)
	if w.lastWalkedTime.Before(threshold) {
		w.canvas.WalkTrees(nil, func(node *common.RenderCacheNode, _ granio.Position) {
			_ = cache.GetCanvasForObject(node.Obj())
			if wid, ok := node.Obj().(granio.Widget); ok {
				_, _ = cache.CachedRenderer(wid)
			}
		})
		w.lastWalkedTime = time.Now()
	}
}

func (*gLDriver) applyThemeToWindow(w granio.Window) {
	if win, ok := w.(*window); ok {
		win.setDarkMode()
		win.applyQueuedVisuals()
	}
}

func (d *gLDriver) runGL() {
	if !running.CompareAndSwap(false, true) {
		return // Run was called twice.
	}

	d.init()
	if d.trayStart != nil {
		d.trayStart()
	}

	applySettings := func(set granio.Settings) {
		painter.ClearFontCache()
		cache.ResetThemeCaches()
		app.ApplySettingsWithCallback(set, granio.CurrentApp(), func(w granio.Window) {
			d.applyThemeToWindow(w)
			c, ok := w.Canvas().(*glCanvas)
			if !ok {
				return
			}
			c.applyThemeOutOfTreeObjects()
			c.reloadScale()
		})
	}
	settings := granio.CurrentApp().Settings()
	settings.AddListener(applySettings)
	if provider, ok := settings.(granio.SystemAppearanceProvider); ok && provider.SystemAppearance().Generation != 0 {
		// A provider may have supplied its initial value before the GL loop was
		// ready. Invalidate already-created widget/layout caches before Show.
		applySettings(settings)
	}

	if f := granio.CurrentApp().Lifecycle().(*app.Lifecycle).OnStarted(); f != nil {
		f()
	}

	d.runLoop()
}

func executeMainFunc(f funcData) {
	f.f()
	if f.done != nil {
		f.done <- struct{}{}
	}
}

func (d *gLDriver) finishLoop() {
	d.animation.Clear()
	clear(d.animationOwners)
	l := granio.CurrentApp().Lifecycle().(*app.Lifecycle)
	if f := l.OnStopped(); f != nil {
		l.QueueEvent(f)
	}
	l.DestroyEventQueue()
	// Lifecycle callbacks still need the UI thread. Quiesce their dispatcher
	// before closing this queue; stopping first would discard its barrier and
	// deadlock Run's final WaitForEvents.
	if !l.EventQueueStarted() {
		l.RunEventQueue(func(fn func(), _ bool) { fn() })
	}
	lifecycleDone := false
	for !lifecycleDone {
		select {
		case <-l.EventQueueDone():
			lifecycleDone = true
		case f := <-funcQueue.Out():
			executeMainFunc(f)
		}
	}
	// Serialize closing against submissions, then acknowledge every accepted
	// waiter. Sampling len(Out) misses work still in the unbounded input buffer.
	funcQueueMu.Lock()
	drained.Store(true)
	funcQueue.Close()
	funcQueueMu.Unlock()
	for f := range funcQueue.Out() {
		if f.done != nil {
			f.done <- struct{}{}
		}
	}
	d.Terminate()
}

func (d *gLDriver) processWindowEvents() {
	for i := 0; i < len(d.windows); i++ {
		w := d.windows[i].(*window)
		w.ensurePositionProcessed()
		w.refreshVisualPolicy()
		if w.viewport == nil {
			if w.closing && w.visualsClosed() {
				d.destroyWindow(w, i)
				i--
			}
			continue
		}
		if w.viewport.ShouldClose() {
			d.destroyWindow(w, i)
			i--
			continue
		}
		if w.shouldExpand && !w.fullScreen {
			w.fitContent()
			shouldExpand := w.shouldExpand
			w.shouldExpand = false
			if shouldExpand && runtime.GOOS != goos.JavaScript {
				w.viewport.SetSize(w.shouldWidth, w.shouldHeight)
			}
		}
	}
}

func (d *gLDriver) destroyWindow(w *window, index int) {
	w.visible = false
	if w.viewport != nil {
		w.viewport.Destroy()
	}
	w.destroy(d)

	if index < len(d.windows)-1 {
		copy(d.windows[index:], d.windows[index+1:])
	}
	d.windows[len(d.windows)-1] = nil
	d.windows = d.windows[:len(d.windows)-1]

	if len(d.windows) == 0 {
		d.Quit()
	}
}

func (*gLDriver) repaintWindow(w *window) bool {
	canvas := w.canvas
	freed := false
	if canvas.EnsureMinSize() {
		w.shouldExpand = true
	}
	freed = canvas.FreeDirtyTextures() > 0

	updateGLContext(w)
	canvas.paint(canvas.Size())

	view := w.viewport
	visible := w.visible

	if view != nil && visible {
		w.frame.requestFrame()
		view.SwapBuffers()
	}

	// mark that we have walked the window and don't
	// need to walk it again to mark caches alive
	w.lastWalkedTime = time.Now()
	return freed
}

// refreshWindow requests that the specified window be redrawn
func refreshWindow(w *window) {
	w.canvas.SetDirty()
}

func updateGLContext(w *window) {
	canvas := w.canvas
	size := canvas.Size()

	// w.width and w.height are not correct if we are maximised, so figure from canvas
	winWidth := float32(scale.ToScreenCoordinate(canvas, size.Width)) * canvas.texScale
	winHeight := float32(scale.ToScreenCoordinate(canvas, size.Height)) * canvas.texScale

	canvas.Painter().SetFrameBufferScale(canvas.texScale)
	canvas.Painter().SetOutputSize(int(winWidth), int(winHeight))
}
