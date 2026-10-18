//go:build !ios && !android

package mobile

import "github.com/Uzer007-admin/granio/v2"

// Content returns the clipboard content for mobile simulator runs
func (mobileClipboard) Content() string {
	granio.LogError("Clipboard is not supported in mobile simulation", nil)
	return ""
}

// SetContent sets the clipboard content for mobile simulator runs
func (mobileClipboard) SetContent(string) {
	granio.LogError("Clipboard is not supported in mobile simulation", nil)
}
