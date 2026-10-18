//go:build !wasm && !test_web_driver

package glfw

import (
	"errors"
	"fmt"
	"mime"
	"sort"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
)

const maxClipboardBytes = 64 << 20

var errClipboardFormat = errors.New("clipboard representation is unavailable")
var _ granio.ClipboardData = clipboard{}

func (clipboard) Formats() ([]string, error) { return clipboardFormats() }
func (clipboard) ReadData(format string) ([]byte, error) {
	if err := validateClipboardFormat(format); err != nil {
		return nil, err
	}
	return readClipboardData(format)
}
func (clipboard) WriteData(data map[string][]byte) error {
	if len(data) > 128 {
		return errors.New("clipboard has more than 128 representations")
	}
	total := 0
	for format, value := range data {
		if err := validateClipboardFormat(format); err != nil {
			return err
		}
		if len(value) > maxClipboardBytes-total {
			return errors.New("clipboard exceeds 64 MiB")
		}
		total += len(value)
	}
	return writeClipboardData(data)
}
func validateClipboardFormat(format string) error {
	if len(format) > 255 || strings.ContainsAny(format, "\x00\r\n") {
		return errors.New("invalid clipboard MIME type")
	}
	media, _, err := mime.ParseMediaType(format)
	if err != nil || !strings.Contains(media, "/") {
		return fmt.Errorf("invalid clipboard MIME type %q", format)
	}
	return nil
}
func sortedClipboardFormats(data map[string][]byte) []string {
	formats := make([]string, 0, len(data))
	for format := range data {
		formats = append(formats, format)
	}
	sort.Strings(formats)
	return formats
}
