package glfw

// #include "background_linux.h"
import "C"

func (w *Window) queryBackgroundEffects() (BackgroundEffectState, error) {
	s := C.glfwQueryBackgroundBridge(w.data)
	if s.valid == 0 {
		return BackgroundEffectState{}, ErrVisualUnsupported
	}
	if s.ownerThread == 0 {
		return BackgroundEffectState{}, ErrVisualThread
	}
	return BackgroundEffectState{Platform: BackgroundPlatform(s.platform),
		AlphaSurface: s.alphaSurface != 0, TransparencyAvailable: s.transparent != 0,
		Composition: s.composition != 0, BlurAvailable: s.blur != 0, BlurRequested: s.requested != 0,
		Revision: uint64(s.revision), Compositor: uint64(s.compositor)}, nil
}

func (w *Window) setBackgroundBlur(enabled bool) error {
	value := C.int(0)
	if enabled {
		value = 1
	}
	switch C.glfwSetBackgroundBlurBridge(w.data, value) {
	case 0:
		return nil
	case 1:
		return ErrVisualUnsupported
	case 2:
		return ErrVisualThread
	case 3:
		return ErrBackgroundAlpha
	case 4:
		return ErrBackgroundComposition
	case 5:
		return ErrBackgroundBlur
	default:
		return ErrVisualApply
	}
}
