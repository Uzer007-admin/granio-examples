package glfw

//#define GLFW_INCLUDE_NONE
//#include "native_array.h"
import "C"

// SurfaceGeometry separates native window coordinates, framebuffer density and
// UI content scale. No focus, visibility or compositor attributes are queried.
type SurfaceGeometry struct {
	Width, Height                       int
	FramebufferWidth, FramebufferHeight int
	ContentScaleX, ContentScaleY        float32
}

// GetSurfaceGeometry returns three native geometry queries in one cgo call
// without Go output allocations or pumping events. Call only on the main thread.
// Queries are sequential, not an atomic compositor snapshot. Prefer collecting
// geometry after sizing/scale events and reusing it during rendering.
func (w *Window) GetSurfaceGeometry() SurfaceGeometry {
	value := C.glfwGetSurfaceGeometryBridge(w.data)
	panicError()
	return SurfaceGeometry{int(value.width), int(value.height), int(value.framebufferWidth), int(value.framebufferHeight), float32(value.scaleX), float32(value.scaleY)}
}

// PixelScale returns physical pixels per native window coordinate, independently
// of content scale. Zero/minimized geometry returns ok=false.
func (s SurfaceGeometry) PixelScale() (x, y float64, ok bool) {
	if s.Width <= 0 || s.Height <= 0 || s.FramebufferWidth <= 0 || s.FramebufferHeight <= 0 {
		return 0, 0, false
	}
	return float64(s.FramebufferWidth) / float64(s.Width), float64(s.FramebufferHeight) / float64(s.Height), true
}
