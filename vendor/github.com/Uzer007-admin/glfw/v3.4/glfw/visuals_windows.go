package glfw

// #include "visuals_windows.h"
import "C"

func (w *Window) queryWindowVisuals() (WindowVisualState, error) {
	s := C.glfwQueryVisualsBridge(w.data)
	if err := acceptError(ErrorCode(0x0001000E)); err != nil {
		return WindowVisualState{}, ErrVisualUnsupported
	}
	if s.valid == 0 {
		return WindowVisualState{}, ErrVisualUnsupported
	}
	if s.ownerThread == 0 {
		return WindowVisualState{}, ErrVisualThread
	}
	return WindowVisualState{OSBuild: uint32(s.build), OSAttributes: WindowVisualAttributes(s.supported),
		Composition: s.composition != 0, RemoteSession: s.remote != 0, HighContrast: s.contrast != 0, HighContrastKnown: s.contrastKnown != 0,
		TransparentFramebuffer: s.transparent != 0, PixelFormat: int(s.pixelFormat), AlphaBits: int(s.alphaBits),
		Backdrop: SystemBackdrop(s.backdrop), BackdropKnown: s.backdropKnown != 0}, nil
}

func (w *Window) applyWindowVisuals(o WindowVisualOptions) ([4]uint32, error) {
	result := C.glfwApplyVisualsBridge(w.data, C.uint32_t(o.Attributes), C.int(o.Backdrop), C.int(boolToVisualInt(o.DarkCaption)), C.int(o.Corners), C.int(boolToVisualInt(o.PremultipliedAlpha)))
	if err := acceptError(ErrorCode(0x0001000E)); err != nil {
		return [4]uint32{}, ErrVisualUnsupported
	}
	if result.valid == 0 {
		return [4]uint32{}, ErrVisualUnsupported
	}
	if result.ownerThread == 0 {
		return [4]uint32{}, ErrVisualThread
	}
	return [4]uint32{uint32(result.hr[0]), uint32(result.hr[1]), uint32(result.hr[2]), uint32(result.hr[3])}, nil
}

func boolToVisualInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
