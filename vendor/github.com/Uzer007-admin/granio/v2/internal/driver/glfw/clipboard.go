//go:build !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/granio/v2"
)

// Declare conformity with Clipboard interface
var _ granio.Clipboard = clipboard{}

func NewClipboard() granio.Clipboard {
	return clipboard{}
}

// clipboard represents the system clipboard
type clipboard struct{}

// Content returns the clipboard content
func (clipboard) Content() string {
	data, _ := readClipboardData(granio.MIMETextPlain)
	return string(data)
}

// SetContent sets the clipboard content
func (clipboard) SetContent(content string) {
	if err := writeClipboardData(map[string][]byte{granio.MIMETextPlain: []byte(content)}); err != nil {
		granio.LogError("Clipboard set failed", err)
	}
}
