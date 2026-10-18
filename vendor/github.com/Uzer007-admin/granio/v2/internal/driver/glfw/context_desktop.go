//go:build !wasm && !test_web_driver

package glfw

import "github.com/Uzer007-admin/glfw/v3.5/glfw"

func (w *window) activateContext() func() {
	previous := glfw.GetCurrentContext()
	if previous == w.viewport {
		return func() {}
	}
	w.viewport.MakeContextCurrent()
	return func() {
		if previous != nil {
			previous.MakeContextCurrent()
		} else {
			glfw.DetachCurrentContext()
		}
	}
}
