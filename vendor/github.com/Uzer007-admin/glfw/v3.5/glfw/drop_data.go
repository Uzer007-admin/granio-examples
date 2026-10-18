package glfw

/*
#define GLFW_INCLUDE_NONE
#include <stdlib.h>
#include "glfw/include/GLFW/glfw3.h"
void fyneDropFormats(GLFWwindow* window,const char* formats);
void fyneDropDisposeWindow(GLFWwindow* window);
*/
import "C"

import (
	"strings"
	"sync"
	"unsafe"
)

var dataDropCallbacks sync.Map // *Window -> callback; invoked on the event thread

// SetDataDropCallback accepts a preferred MIME representation from an external
// X11/Wayland drag source. File-drop callbacks remain available.
func (w *Window) SetDataDropCallback(formats []string, callback func(*Window, string, []byte)) {
	if callback == nil {
		dataDropCallbacks.Delete(w)
		formats = nil
	} else {
		dataDropCallbacks.Store(w, callback)
	}
	text := C.CString(strings.Join(formats, "\n"))
	defer C.free(unsafe.Pointer(text))
	C.fyneDropFormats(w.data, text)
}
func disposeDataDrop(w *Window) { C.fyneDropDisposeWindow(w.data); w.SetDataDropCallback(nil, nil) }

//export goFyneDataDrop
func goFyneDataDrop(window *C.GLFWwindow, mime *C.char, bytes unsafe.Pointer, size C.int) {
	w := windows.get(window)
	if w == nil || size < 0 || size > 64<<20 {
		return
	}
	if callback, ok := dataDropCallbacks.Load(w); ok {
		callback.(func(*Window, string, []byte))(w, C.GoString(mime), C.GoBytes(bytes, size))
	}
}
