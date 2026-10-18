//go:build linux && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2/internal/async"
)

func onClipboardThread(fn func()) {
	if async.IsMainGoroutine() {
		fn()
	} else {
		runOnMain(fn)
	}
}

func clipboardFormats() (formats []string, err error) {
	onClipboardThread(func() { formats, err = glfw.GetClipboardFormats() })
	return
}
func readClipboardData(format string) (data []byte, err error) {
	onClipboardThread(func() { data, err = glfw.GetClipboardData(format) })
	return
}
func writeClipboardData(data map[string][]byte) (err error) {
	onClipboardThread(func() { err = glfw.SetClipboardData(data) })
	return
}
