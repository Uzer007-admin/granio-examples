//go:build linux || freebsd || openbsd || netbsd

package glfw

import "github.com/Uzer007-admin/granio/v2"

func (w *window) platformResize(canvasSize granio.Size) {
	w.canvas.Resize(canvasSize)
}
