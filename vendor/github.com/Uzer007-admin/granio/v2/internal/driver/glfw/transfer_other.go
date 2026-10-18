//go:build !windows && !linux && !wasm && !test_web_driver

package glfw

func (*window) updateDataDrop()     {}
func (*window) cleanupDataDrop()    {}
func nativeDataDropSupported() bool { return false }
