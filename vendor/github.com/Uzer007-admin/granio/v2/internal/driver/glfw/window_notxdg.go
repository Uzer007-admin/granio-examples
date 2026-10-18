//go:build !linux && !freebsd && !openbsd && !netbsd

package glfw

import "github.com/Uzer007-admin/granio/v2"

func (w *window) platformResize(canvasSize granio.Size) {
	d, ok := granio.CurrentApp().Driver().(*gLDriver)
	if !ok { // don't wait to redraw in this way if we are running on test
		w.canvas.Resize(canvasSize)
		return
	}

	w.canvas.Resize(canvasSize)
	d.repaintWindow(w)
}
