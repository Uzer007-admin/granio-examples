package glfw

//#define GLFW_INCLUDE_NONE
//#include "native_array.h"
import "C"

// SurfaceMetrics describes window geometry and compositor-related attributes.
// ContentScale describes UI/DPI scaling; PixelScale describes framebuffer density.
// Framebuffer transparency does not prove support for any OS backdrop material.
type SurfaceMetrics struct {
	Width, Height                          int
	FramebufferWidth, FramebufferHeight    int
	ContentScaleX, ContentScaleY           float32
	Focused, Iconified, Visible, Maximized bool
	TransparentFramebuffer                 bool
}

// GetSurfaceMetrics queries the window in one cgo transition without Go output
// allocations or pumping events. Call only from the main thread. This is a set
// of sequential native queries, not an atomic transaction with the compositor.
func (w *Window) GetSurfaceMetrics() SurfaceMetrics {
	state := C.glfwGetSurfaceMetricsBridge(w.data)
	panicError()
	return SurfaceMetrics{
		Width: int(state.width), Height: int(state.height),
		FramebufferWidth: int(state.framebufferWidth), FramebufferHeight: int(state.framebufferHeight),
		ContentScaleX: float32(state.scaleX), ContentScaleY: float32(state.scaleY),
		Focused: state.focused != 0, Iconified: state.iconified != 0,
		Visible: state.visible != 0, Maximized: state.maximized != 0,
		TransparentFramebuffer: state.transparent != 0,
	}
}

// PixelScale returns framebuffer pixels per window coordinate. It is distinct
// from the content scale and may differ by platform. Zero/minimized geometry
// returns ok=false instead of an invalid ratio.
func (s SurfaceMetrics) PixelScale() (x, y float64, ok bool) {
	if s.Width <= 0 || s.Height <= 0 || s.FramebufferWidth <= 0 || s.FramebufferHeight <= 0 {
		return 0, 0, false
	}
	return float64(s.FramebufferWidth) / float64(s.Width), float64(s.FramebufferHeight) / float64(s.Height), true
}
