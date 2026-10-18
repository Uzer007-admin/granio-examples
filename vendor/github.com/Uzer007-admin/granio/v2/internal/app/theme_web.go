//go:build !wasm && test_web_driver

package app

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/theme"
)

// DefaultVariant returns the systems default granio.ThemeVariant.
// Normally, you should not need this. It is extracted out of the root app package to give the
// settings app access to it.
func DefaultVariant() granio.ThemeVariant {
	return theme.VariantDark
}
