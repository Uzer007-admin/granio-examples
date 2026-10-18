package glfw

/*
#include <stdint.h>
*/
import "C"

import (
	"sync"
	"unsafe"
)

var win32MessageFilters sync.Map     // native HWND -> func; UI-thread callbacks
func disposeMessageFilter(w *Window) { w.SetWin32MessageFilter(nil) }

// SetWin32MessageFilter handles native messages before TranslateMessage. TSF
// hosts need this ordering to prevent ordinary WM_CHAR generation for consumed
// composition keys. Nil removes the filter. Use on the window's owner thread.
func (w *Window) SetWin32MessageFilter(filter func(message uint32, key, flags uintptr) bool) {
	hwnd := uintptr(unsafe.Pointer(w.GetWin32Window()))
	if filter == nil {
		win32MessageFilters.Delete(hwnd)
	} else {
		win32MessageFilters.Store(hwnd, filter)
	}
}

//export goFynePreTranslateWin32
func goFynePreTranslateWin32(hwnd C.uintptr_t, message C.uint32_t, key, flags C.uintptr_t) C.int {
	if filter, ok := win32MessageFilters.Load(uintptr(hwnd)); ok && filter.(func(uint32, uintptr, uintptr) bool)(uint32(message), uintptr(key), uintptr(flags)) {
		return 1
	}
	return 0
}
