//go:build !windows

package glfw

func (w *Window) queryWindowVisuals() (WindowVisualState, error) {
	return WindowVisualState{}, ErrVisualUnsupported
}
func (w *Window) applyWindowVisuals(WindowVisualOptions) ([4]uint32, error) {
	return [4]uint32{}, ErrVisualUnsupported
}
