//go:build !wasm && !test_web_driver && !android && !ios && !mobile && (linux || openbsd || freebsd || netbsd)

package app

import (
	"sync/atomic"

	"github.com/Uzer007-admin/granio/v2"
)

// CurrentVariant contains the system’s theme variant.
// It is intended for internal use, only!
var CurrentVariant atomic.Uint64

// DefaultVariant returns the systems default granio.ThemeVariant.
// Normally, you should not need this. It is extracted out of the root app package to give the
// settings app access to it.
func DefaultVariant() granio.ThemeVariant {
	return granio.ThemeVariant(CurrentVariant.Load())
}
