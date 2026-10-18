package mobile

import (
	"github.com/Uzer007-admin/granio/v2"
)

// Declare conformity with Clipboard interface
var _ granio.Clipboard = mobileClipboard{}

func NewClipboard() granio.Clipboard {
	return mobileClipboard{}
}

// mobileClipboard represents the system mobileClipboard
type mobileClipboard struct{}
