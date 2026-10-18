//go:build ios

package mobile

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation

#import <stdlib.h>
#import <stdbool.h>

bool iosCanList(const char* url);
bool iosCreateListable(const char* url);
char* iosList(const char* url);
*/
import "C"

import (
	"errors"
	"strings"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
)

func canListURI(uri granio.URI) bool {
	uriStr := C.CString(uri.String())
	defer C.free(unsafe.Pointer(uriStr))

	return bool(C.iosCanList(uriStr))
}

func createListableURI(uri granio.URI) error {
	uriStr := C.CString(uri.String())
	defer C.free(unsafe.Pointer(uriStr))

	ok := bool(C.iosCreateListable(uriStr))
	if ok {
		return nil
	}
	return errors.New("failed to create directory")
}

func listURI(uri granio.URI) ([]granio.URI, error) {
	uriStr := C.CString(uri.String())
	defer C.free(unsafe.Pointer(uriStr))

	str := C.iosList(uriStr)
	parts := strings.Split(C.GoString(str), "|")
	var list []granio.URI
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		list = append(list, storage.NewURI(part))
	}
	return list, nil
}
