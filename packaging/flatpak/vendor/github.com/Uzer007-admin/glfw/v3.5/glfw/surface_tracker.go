package glfw

import (
	"errors"
	"math"
)

var ErrSurfaceTracker = errors.New("glfw: nil surface tracker")

// SurfaceChanges identifies changes since the previous main-thread snapshot.
// Geometry/Scale/Context invalidate GPU material keys. Focus/Presentation alone
// do not invalidate lower-layer pixels, but Drawable gates frame scheduling.
type SurfaceChanges uint8

const (
	SurfaceGeometryChanged SurfaceChanges = 1 << iota
	SurfaceScaleChanged
	SurfacePresentationChanged
	SurfaceFocusChanged
	SurfaceContextChanged
)

type SurfaceState struct {
	Metrics            SurfaceMetrics
	Generation         uint64
	ResourceGeneration uint64
	Drawable           bool
}

// SurfaceTracker keeps per-window generations without installing/replacing
// GLFW callbacks. It performs no native calls, allocations or frame scheduling.
// The zero value is ready; use on the window owner thread after event processing.
// A tracker belongs to one window. Call InvalidateContext on context loss or
// replacement; unchanged geometry cannot detect that event automatically.
type SurfaceTracker struct{ state SurfaceState }

func (t *SurfaceTracker) State() SurfaceState { return t.state }

func (t *SurfaceTracker) Update(metrics SurfaceMetrics) (SurfaceState, SurfaceChanges) {
	old := t.state.Metrics
	var changes SurfaceChanges
	if t.state.Generation == 0 || old.Width != metrics.Width || old.Height != metrics.Height || old.FramebufferWidth != metrics.FramebufferWidth || old.FramebufferHeight != metrics.FramebufferHeight || old.TransparentFramebuffer != metrics.TransparentFramebuffer {
		changes |= SurfaceGeometryChanged
	}
	if t.state.Generation == 0 || !sameScale(old.ContentScaleX, metrics.ContentScaleX) || !sameScale(old.ContentScaleY, metrics.ContentScaleY) {
		changes |= SurfaceScaleChanged
	}
	if t.state.Generation == 0 || old.Visible != metrics.Visible || old.Iconified != metrics.Iconified || old.Maximized != metrics.Maximized {
		changes |= SurfacePresentationChanged
	}
	if t.state.Generation == 0 || old.Focused != metrics.Focused {
		changes |= SurfaceFocusChanged
	}
	if changes != 0 {
		t.state.Generation++
	}
	if changes&(SurfaceGeometryChanged|SurfaceScaleChanged) != 0 {
		t.state.ResourceGeneration++
	}
	t.state.Metrics = metrics
	_, _, valid := metrics.PixelScale()
	t.state.Drawable = valid && metrics.Visible && !metrics.Iconified
	return t.state, changes
}

func sameScale(a, b float32) bool { return a == b || math.IsNaN(float64(a)) && math.IsNaN(float64(b)) }

func (t *SurfaceTracker) InvalidateContext() SurfaceState {
	t.state.Generation++
	t.state.ResourceGeneration++
	return t.state
}

// UpdateSurface queries one native metrics bridge and updates the supplied
// tracker. It does not pump events. Destroyed windows return a typed error.
func (w *Window) UpdateSurface(tracker *SurfaceTracker) (SurfaceState, SurfaceChanges, error) {
	if w == nil || w.data == nil {
		return SurfaceState{}, 0, ErrWindowDestroyed
	}
	if tracker == nil {
		return SurfaceState{}, 0, ErrSurfaceTracker
	}
	state, changes := tracker.Update(w.GetSurfaceMetrics())
	return state, changes, nil
}

// FramebufferRegion is a clipped pixel rectangle with top-left origin.
type FramebufferRegion struct{ X, Y, Width, Height int }

// BottomLeftY converts a validated region to the GL bottom-left source origin.
func (r FramebufferRegion) BottomLeftY(framebufferHeight int) (int, bool) {
	if r.X < 0 || r.Y < 0 || r.Width <= 0 || r.Height <= 0 || r.Height > framebufferHeight || r.Y > framebufferHeight-r.Height {
		return 0, false
	}
	return framebufferHeight - r.Y - r.Height, true
}

// FramebufferRegion converts window coordinates to pixels, clipping first and
// rounding outwards. Each axis uses actual framebuffer/window density. Content
// scale must not be multiplied into it; it describes a separate UI policy.
func (s SurfaceMetrics) FramebufferRegion(x, y, width, height float64) (FramebufferRegion, bool) {
	sx, sy, valid := s.PixelScale()
	if !valid || width <= 0 || height <= 0 {
		return FramebufferRegion{}, false
	}
	for _, v := range [6]float64{x, y, width, height, x + width, y + height} {
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return FramebufferRegion{}, false
		}
	}
	left, top := max(0, x), max(0, y)
	right, bottom := min(float64(s.Width), x+width), min(float64(s.Height), y+height)
	if right <= left || bottom <= top {
		return FramebufferRegion{}, false
	}
	// Clamp in floating point before conversion, including at MaxInt boundaries.
	x0, y0 := math.Floor(left*sx), math.Floor(top*sy)
	x1, y1 := min(float64(s.FramebufferWidth), math.Ceil(right*sx)), min(float64(s.FramebufferHeight), math.Ceil(bottom*sy))
	// Native sizes are C int, so malformed caller-built snapshots above int32
	// are rejected rather than risking an overflowing conversion on 32-bit Go.
	if s.FramebufferWidth > math.MaxInt32 || s.FramebufferHeight > math.MaxInt32 || x1 <= x0 || y1 <= y0 {
		return FramebufferRegion{}, false
	}
	return FramebufferRegion{int(x0), int(y0), int(x1 - x0), int(y1 - y0)}, true
}
