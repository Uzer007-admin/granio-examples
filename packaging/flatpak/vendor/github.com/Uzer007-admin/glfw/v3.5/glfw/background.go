package glfw

import "errors"

var (
	ErrBackgroundComposition = errors.New("glfw: background composition unavailable")
	ErrBackgroundBlur        = errors.New("glfw: compositor background blur unavailable")
	ErrBackgroundAlpha       = errors.New("glfw: background effect requires an alpha surface")
)

type BackgroundPlatform uint8

const (
	BackgroundUnsupported BackgroundPlatform = iota
	BackgroundX11
	BackgroundWayland
)

// BackgroundEffectState is a main-thread snapshot. AlphaSurface describes the
// created native format, TransparencyAvailable also requires a compositor.
// BlurAvailable follows the X11 root support property or Wayland capabilities,
// never a desktop-name/version heuristic. Revision/Compositor identify changes
// of the Wayland manager or X11 selection owner, including compositor restarts.
// BlurRequested records our own request, not proof of the compositor's output.
type BackgroundEffectState struct {
	Platform                                         BackgroundPlatform
	AlphaSurface, TransparencyAvailable, Composition bool
	BlurAvailable, BlurRequested                     bool
	Revision, Compositor                             uint64
}

func (w *Window) QueryBackgroundEffects() (BackgroundEffectState, error) {
	if w == nil || w.data == nil {
		return BackgroundEffectState{}, ErrWindowDestroyed
	}
	return w.queryBackgroundEffects()
}

// SetBackgroundBlur changes the full surface's blur request, preserving
// the frame, context and callbacks. Call only on the GLFW main thread. Enabling
// requires an alpha format and compositor capability. Disabling is allowed even
// after capability loss. On Wayland state is committed by the next buffer swap:
// this method does not commit a surface, pump events or perform a roundtrip.
// Always render premultiplied RGBA before presentation. Initialize the framebuffer
// before Show; background effects do not initialize graphics contents.
func (w *Window) SetBackgroundBlur(enabled bool) error {
	if w == nil || w.data == nil {
		return ErrWindowDestroyed
	}
	return w.setBackgroundBlur(enabled)
}
