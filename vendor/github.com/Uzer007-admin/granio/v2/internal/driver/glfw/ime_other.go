//go:build !windows && !linux && !wasm && !test_web_driver

package glfw

func (*window) initInputMethod()   {}
func (*window) resetInputMethod()  {}
func (*window) updateInputMethod() {}
