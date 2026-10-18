//go:build android || ios || (mobile && !windows)

package app

import (
	"github.com/Uzer007-admin/granio/v2"
)

// SystemTheme contains the system’s theme variant.
// It is intended for internal use, only!
var SystemTheme granio.ThemeVariant

// DefaultVariant returns the systems default granio.ThemeVariant.
// Normally, you should not need this. It is extracted out of the root app package to give the
// settings app access to it.
func DefaultVariant() granio.ThemeVariant {
	return SystemTheme
}
