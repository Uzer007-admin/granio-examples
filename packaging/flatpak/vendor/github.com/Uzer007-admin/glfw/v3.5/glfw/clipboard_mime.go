package glfw

/*
#include <stdlib.h>
int fyneClipboardBegin(void);
int fyneClipboardAdd(const char* mime, const void* bytes, int length);
int fyneClipboardPublish(void);
void fyneClipboardAbort(void);
void* fyneClipboardRead(const char* mime, int* size, int* status);
char* fyneClipboardFormats(void);
*/
import "C"

import (
	"errors"
	"sort"
	"strings"
	"unsafe"
)

// SetClipboardData publishes MIME representations through X11/Wayland. Values
// are copied before ownership changes; 64 MiB bounds the combined payload.
func SetClipboardData(data map[string][]byte) error {
	if len(data) > 128 {
		return errors.New("too many clipboard formats")
	}
	total := 0
	for format, value := range data {
		if len(format) > 255 || strings.ContainsAny(format, "\x00\r\n") {
			return errors.New("invalid clipboard format")
		}
		if len(value) > (64<<20)-total {
			return errors.New("clipboard exceeds 64 MiB")
		}
		total += len(value)
	}
	if C.fyneClipboardBegin() == 0 {
		return errors.New("MIME clipboard is unavailable")
	}
	published := false
	defer func() {
		if !published {
			C.fyneClipboardAbort()
		}
	}()
	formats := make([]string, 0, len(data))
	for format := range data {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	for _, format := range formats {
		mime := C.CString(format)
		value := C.CBytes(data[format])
		ok := C.fyneClipboardAdd(mime, value, C.int(len(data[format])))
		C.free(unsafe.Pointer(mime))
		C.free(value)
		if ok == 0 {
			return errors.New("clipboard payload exceeds native limits")
		}
	}
	if C.fyneClipboardPublish() == 0 {
		return errors.New("clipboard ownership was denied")
	}
	published = true
	return nil
}
func GetClipboardData(format string) ([]byte, error) {
	mime := C.CString(format)
	defer C.free(unsafe.Pointer(mime))
	var size, status C.int
	value := C.fyneClipboardRead(mime, &size, &status)
	defer C.free(value)
	if status != 0 {
		return nil, errors.New("clipboard MIME representation is unavailable or transfer timed out")
	}
	return C.GoBytes(value, size), nil
}
func GetClipboardFormats() ([]string, error) {
	value := C.fyneClipboardFormats()
	if value == nil {
		return nil, errors.New("clipboard formats are unavailable")
	}
	defer C.free(unsafe.Pointer(value))
	text := C.GoString(value)
	if text == "" {
		return []string{}, nil
	}
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n"), nil
}
