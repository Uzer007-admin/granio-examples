//go:build (gles || arm || arm64) && !darwin && !wasm && !test_web_driver

package glfw

import "github.com/Uzer007-admin/glfw/v3.5/glfw"

func initWindowHints() {
	glfw.WindowHint(glfw.ClientAPI, glfw.OpenGLESAPI)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 0)
}
