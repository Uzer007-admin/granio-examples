//go:build !wasm && !test_web_driver

package glfw

import (
	"errors"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"slices"
)

var _ desktop.DataDropWindow = (*window)(nil)

func queueNativeDrop(fn func()) { runOnMainWithWait(fn, false) }

func (w *window) SetOnDataDropped(formats []string, callback func(granio.Position, map[string][]byte)) error {
	if callback != nil && !nativeDataDropSupported() {
		return desktop.ErrDataDropUnsupported
	}
	if len(formats) > 128 {
		return errors.New("drop target has more than 128 MIME types")
	}
	for _, format := range formats {
		if err := validateClipboardFormat(format); err != nil {
			return err
		}
	}
	if callback != nil && len(formats) == 0 {
		return errors.New("drop target requires an accepted MIME type")
	}
	formats = append([]string(nil), formats...)
	w.runOnMainWhenCreated(func() { w.onDataDropped, w.dropFormats = callback, formats; w.updateDataDrop() })
	return nil
}
func (w *window) deliverDataDrop(position granio.Position, format string, data []byte) {
	if w.onDataDropped != nil && !w.closing && slices.Contains(w.dropFormats, format) {
		w.onDataDropped(position, map[string][]byte{format: data})
	}
}
