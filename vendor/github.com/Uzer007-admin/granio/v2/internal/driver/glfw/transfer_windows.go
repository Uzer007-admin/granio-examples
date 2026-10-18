//go:build windows && !wasm && !test_web_driver

package glfw

/*
#cgo LDFLAGS: -lole32
#include <stdlib.h>
#include "transfer_windows.h"
*/
import "C"

import (
	"encoding/binary"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf16"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
	"github.com/Uzer007-admin/granio/v2/storage"
)

func nativeDataDropSupported() bool { return true }

var dropWindows sync.Map // HWND -> *window; accessed by native OLE callbacks
func (w *window) updateDataDrop() {
	if w.viewport == nil {
		return
	}
	hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	if w.onDataDropped == nil {
		w.cleanupDataDrop()
		return
	}
	dropWindows.Store(hwnd, w)
	if C.FyneDropBegin(C.uintptr_t(hwnd)) == 0 {
		granio.LogError("Unable to register native OLE drop target", nil)
		return
	}
	for _, format := range w.dropFormats {
		id, err := winClipboardFormat(format)
		if err != nil {
			granio.LogError("Invalid native drop representation", err)
			continue
		}
		mime := C.CString(format)
		C.FyneDropAdd(C.uintptr_t(hwnd), C.uint(id), mime)
		C.free(unsafe.Pointer(mime))
		if format == granio.MIMETextURIList {
			mime = C.CString(format)
			C.FyneDropAdd(C.uintptr_t(hwnd), 15, mime)
			C.free(unsafe.Pointer(mime))
		}
	}
	// OLE consumes external file offers before WM_DROPFILES. Preserve the legacy
	// file callback even when the MIME target accepts only other representations.
	if w.onFilesDropped != nil {
		id, err := winClipboardFormat(granio.MIMETextURIList)
		if err == nil {
			mime := C.CString(granio.MIMETextURIList)
			C.FyneDropAdd(C.uintptr_t(hwnd), C.uint(id), mime)
			C.FyneDropAdd(C.uintptr_t(hwnd), 15, mime)
			C.free(unsafe.Pointer(mime))
		}
	}
}
func (w *window) cleanupDataDrop() {
	if w.viewport == nil {
		return
	}
	hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	dropWindows.Delete(hwnd)
	C.FyneDropRemove(C.uintptr_t(hwnd))
}

//export goWinDataDrop
func goWinDataDrop(hwnd C.uintptr_t, format C.uint, mime *C.char, pointer unsafe.Pointer, size, x, y C.int) {
	if size < 0 || size > maxClipboardBytes {
		return
	}
	stored, ok := dropWindows.Load(uintptr(hwnd))
	if !ok {
		return
	}
	w := stored.(*window)
	data := C.GoBytes(pointer, size)
	media := C.GoString(mime)
	if format == 13 {
		units := make([]uint16, 0, len(data)/2)
		for i := 0; i+1 < len(data); i += 2 {
			unit := binary.LittleEndian.Uint16(data[i:])
			if unit == 0 {
				break
			}
			units = append(units, unit)
		}
		data = []byte(string(utf16.Decode(units)))
	} else if format == 15 {
		data = decodeDropFiles(data)
	} else if media == granio.MIMETextHTML {
		data = decodeClipboardHTML(data)
	}
	queueNativeDrop(func() {
		if current, ok := dropWindows.Load(uintptr(hwnd)); ok && current == w {
			position := granio.NewPos(scale.ToFyneCoordinate(w.canvas, int(x)), scale.ToFyneCoordinate(w.canvas, int(y)))
			w.deliverDataDrop(position, media, data)
			if media == granio.MIMETextURIList && w.onFilesDropped != nil && !w.closing {
				uris := make([]granio.URI, 0)
				for _, line := range strings.Split(string(data), "\n") {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					uri, err := storage.ParseURI(line)
					if err == nil && uri.Scheme() == "file" {
						uris = append(uris, uri)
					}
				}
				if len(uris) > 0 {
					w.onFilesDropped(position, uris)
				}
			}
		}
	})
}
func decodeDropFiles(data []byte) []byte {
	if len(data) < 20 || binary.LittleEndian.Uint32(data[16:]) == 0 {
		return nil
	}
	start := int(binary.LittleEndian.Uint32(data))
	if start < 20 || start > len(data) {
		return nil
	}
	var result strings.Builder
	units := make([]uint16, 0, 32768)
	for i := start; i+1 < len(data); i += 2 {
		unit := binary.LittleEndian.Uint16(data[i:])
		if unit != 0 {
			if len(units) >= 32768 {
				return nil
			}
			units = append(units, unit)
			continue
		}
		if len(units) == 0 {
			break
		}
		path := filepath.ToSlash(string(utf16.Decode(units)))
		uri := url.URL{Scheme: "file", Path: "/" + path}
		if strings.HasPrefix(path, "//") {
			parts := strings.SplitN(path[2:], "/", 2)
			uri.Host = parts[0]
			uri.Path = "/"
			if len(parts) > 1 {
				uri.Path += parts[1]
			}
		}
		result.WriteString(uri.String() + "\r\n")
		units = units[:0]
	}
	return []byte(result.String())
}

func windowsExternalDropTest(hwnd uintptr, format uint32, data []byte) bool {
	if len(data) == 0 {
		return false
	}
	return C.FyneExternalDropTest(C.uintptr_t(hwnd), C.uint(format), unsafe.Pointer(&data[0]), C.int(len(data))) != 0
}
