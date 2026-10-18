//go:build linux && !wasm && !test_web_driver

package glfw

import "github.com/Uzer007-admin/glfw/v3.5/glfw"

func nativeDataDropSupported() bool { return true }

func (w *window) updateDataDrop() {
	if w.viewport == nil {
		return
	}
	var callback func(*glfw.Window, string, []byte)
	if w.onDataDropped != nil {
		callback = func(_ *glfw.Window, format string, data []byte) {
			w.ensurePositionProcessed()
			w.deliverDataDrop(w.mousePos, format, data)
		}
	}
	w.viewport.SetDataDropCallback(w.dropFormats, callback)
}
func (*window) cleanupDataDrop() {}
