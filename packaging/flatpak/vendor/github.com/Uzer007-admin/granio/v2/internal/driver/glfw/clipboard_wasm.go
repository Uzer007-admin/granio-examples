//go:build wasm || test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/granio-glfw-js"
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
func (c clipboard) Content() string {
	return glfw.GetClipboardString()
}

// SetContent sets the clipboard content
func (c clipboard) SetContent(content string) {
	glfw.SetClipboardString(content)
}
