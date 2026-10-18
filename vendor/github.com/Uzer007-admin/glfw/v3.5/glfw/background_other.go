//go:build !linux

package glfw

func (*Window) queryBackgroundEffects() (BackgroundEffectState, error) {
	return BackgroundEffectState{}, ErrVisualUnsupported
}

func (*Window) setBackgroundBlur(bool) error { return ErrVisualUnsupported }
