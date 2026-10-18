package glfw

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/painter/gl"
)

type surfacePainter interface {
	SetWindowSurface(gl.WindowSurfaceMode) bool
	WindowSurfaceAlphaBits() int
}

// All native state and options are protected by this mutex, independently of
// the legacy window fields which remain owned by the UI thread. No callbacks
// or GL/native operations execute while the coordinator mutex is held.
type windowVisualState struct {
	active                                                             atomic.Bool
	mu                                                                 sync.Mutex
	options                                                            desktop.VisualOptions
	status                                                             desktop.VisualStatus
	capabilities                                                       desktop.VisualCapabilities
	enabled, started, alphaReserved, closed, applyQueued, notifyQueued bool
	observers                                                          map[*visualObserver]struct{}
	policyNext                                                         time.Time          // UI thread only
	lastPolicy                                                         windowVisualPolicy // UI thread only
}

type visualObserver struct {
	mu       sync.Mutex
	active   bool
	callback func(desktop.VisualStatus)
}

func (o *visualObserver) deliver(status desktop.VisualStatus) {
	o.mu.Lock()
	if !o.active {
		o.mu.Unlock()
		return
	}
	callback := o.callback
	o.mu.Unlock()
	callback(status)
}

var _ desktop.WindowVisuals = (*window)(nil)

func (w *window) ConfigureVisuals(options desktop.VisualOptions) (uint64, error) {
	if err := desktop.ValidateVisualOptions(options); err != nil {
		return 0, err
	}
	v := &w.visuals
	v.mu.Lock()
	if v.closed {
		v.mu.Unlock()
		return 0, desktop.ErrVisualWindowClosed
	}
	if v.started && windowVisualAlphaFormat() && !v.alphaReserved && (options.PrepareAlpha || options.Backdrop != desktop.BackdropOpaque) {
		v.mu.Unlock()
		return 0, desktop.ErrVisualNeedsRecreate
	}
	v.enabled = true
	v.active.Store(true)
	v.options = options
	id := v.status.RequestID + 1
	v.status = desktop.VisualStatus{RequestID: id, Phase: desktop.VisualPending, Requested: options.Backdrop, Selected: desktop.BackdropOpaque}
	queue := !v.applyQueued
	v.applyQueued = true
	v.mu.Unlock()
	w.queueVisualNotification()
	if queue {
		runOnMainWithWait(w.applyQueuedVisuals, false)
	}
	return id, nil
}

func (w *window) VisualCapabilities() desktop.VisualCapabilities {
	w.visuals.mu.Lock()
	defer w.visuals.mu.Unlock()
	return w.visuals.capabilities
}
func (w *window) VisualStatus() desktop.VisualStatus {
	w.visuals.mu.Lock()
	defer w.visuals.mu.Unlock()
	return w.visuals.status
}

func (w *window) ObserveVisualStatus(callback func(desktop.VisualStatus)) func() {
	if callback == nil {
		return func() {}
	}
	o := &visualObserver{active: true, callback: callback}
	v := &w.visuals
	v.mu.Lock()
	if v.observers == nil {
		v.observers = make(map[*visualObserver]struct{})
	}
	v.observers[o] = struct{}{}
	v.enabled = true
	v.active.Store(!v.closed)
	v.mu.Unlock()
	w.queueVisualNotification()
	return func() {
		o.mu.Lock()
		o.active = false
		o.mu.Unlock()
		v.mu.Lock()
		delete(v.observers, o)
		v.mu.Unlock()
	}
}

func (w *window) queueVisualNotification() {
	v := &w.visuals
	v.mu.Lock()
	if len(v.observers) == 0 {
		v.mu.Unlock()
		return
	}
	// Deliver terminal callbacks before shutdown can discard queued work.
	if v.closed && async.IsMainGoroutine() {
		v.mu.Unlock()
		w.deliverVisualNotifications()
		return
	}
	if v.notifyQueued {
		v.mu.Unlock()
		return
	}
	v.notifyQueued = true
	v.mu.Unlock()
	runOnMainWithWait(w.deliverVisualNotifications, false)
}

func (w *window) deliverVisualNotifications() {
	v := &w.visuals
	v.mu.Lock()
	v.notifyQueued = false
	status := v.status
	observers := make([]*visualObserver, 0, len(v.observers))
	for observer := range v.observers {
		observers = append(observers, observer)
	}
	if v.closed {
		clear(v.observers)
	}
	v.mu.Unlock()
	for _, observer := range observers {
		observer.deliver(status)
	}
}

func (w *window) beginVisualCreation() bool {
	v := &w.visuals
	v.mu.Lock()
	defer v.mu.Unlock()
	v.started = true
	v.alphaReserved = windowVisualAlphaFormat() && (v.options.PrepareAlpha || v.options.Backdrop != desktop.BackdropOpaque)
	v.capabilities.Generation++
	return v.alphaReserved
}

func (w *window) visualRequest() (desktop.VisualOptions, desktop.VisualStatus, bool) {
	w.visuals.mu.Lock()
	defer w.visuals.mu.Unlock()
	return w.visuals.options, w.visuals.status, w.visuals.enabled && !w.visuals.closed
}

func (w *window) publishVisualStatus(status desktop.VisualStatus, capabilities desktop.VisualCapabilities) {
	v := &w.visuals
	v.mu.Lock()
	if v.closed || status.RequestID != v.status.RequestID {
		v.mu.Unlock()
		return
	}
	capabilities.Generation = v.capabilities.Generation
	v.capabilities = capabilities
	v.status = status
	v.mu.Unlock()
	w.queueVisualNotification()
}

func (w *window) failVisualCreation() {
	_, status, enabled := w.visualRequest()
	if !enabled {
		return
	}
	status.Phase = desktop.VisualFailed
	status.Reason = desktop.VisualNativeFailure
	w.publishVisualStatus(status, w.VisualCapabilities())
}

func (w *window) applyQueuedVisuals() {
	w.visuals.mu.Lock()
	w.visuals.applyQueued = false
	closed := w.visuals.closed
	w.visuals.mu.Unlock()
	if closed || w.viewport == nil || w.closing {
		return
	}
	w.applyWindowVisuals()
}

func (w *window) closeWindowVisuals() bool {
	v := &w.visuals
	v.mu.Lock()
	enabled := v.enabled
	if v.closed {
		v.mu.Unlock()
		return enabled
	}
	v.closed = true
	v.active.Store(false)
	v.status.Phase = desktop.VisualClosed
	v.status.SurfaceReady = false
	v.capabilities.SurfaceKnown = false
	v.capabilities.CanChangeWithoutRecreate = false
	v.mu.Unlock()
	w.queueVisualNotification()
	return enabled
}

func (w *window) visualsClosed() bool {
	w.visuals.mu.Lock()
	defer w.visuals.mu.Unlock()
	return w.visuals.enabled && w.visuals.closed
}
