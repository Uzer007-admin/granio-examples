//go:build wasm || test_web_driver

package glfw

// The browser binding has no GetCurrentContext. Track nesting on the UI
// goroutine, so an inner capture cannot detach its caller's active context.
var browserContextWindow *window

func (w *window) activateContext() func() {
	previous := browserContextWindow
	if previous == w {
		return func() {}
	}
	w.viewport.MakeContextCurrent()
	browserContextWindow = w
	return func() {
		browserContextWindow = previous
		if previous != nil {
			previous.viewport.MakeContextCurrent()
		} else {
			w.DetachCurrentContext()
		}
	}
}
