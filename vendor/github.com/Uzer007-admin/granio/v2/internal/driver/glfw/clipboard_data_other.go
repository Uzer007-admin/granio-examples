//go:build !windows && !linux && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
)

func clipboardFormats() ([]string, error) { return []string{granio.MIMETextPlain}, nil }
func readClipboardData(format string) ([]byte, error) {
	if format != granio.MIMETextPlain {
		return nil, errClipboardFormat
	}
	return []byte(glfw.GetClipboardString()), nil
}
func writeClipboardData(data map[string][]byte) error {
	if len(data) != 1 {
		return errClipboardFormat
	}
	text, ok := data[granio.MIMETextPlain]
	if !ok {
		return errClipboardFormat
	}
	glfw.SetClipboardString(string(text))
	return nil
}
