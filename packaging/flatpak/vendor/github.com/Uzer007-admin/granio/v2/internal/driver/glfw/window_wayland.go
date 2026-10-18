//go:build wayland && !x11 && (linux || freebsd || openbsd || netbsd)

package glfw

import (
	"unsafe"

	"github.com/Uzer007-admin/granio/v2/driver"
)

// assert we are implementing driver.NativeWindow
var _ driver.NativeWindow = (*window)(nil)

func (w *window) RunNative(f func(any)) {
	context := driver.WaylandWindowContext{}
	if v := w.view(); v != nil {
		context.WaylandSurface = uintptr(unsafe.Pointer(v.GetWaylandWindow()))
	}

	f(context)
}
