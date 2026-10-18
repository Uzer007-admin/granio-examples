//go:build darwin && !cgo && !ios && !mobile && !wasm && !test_web_driver

package app

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/theme"
)

// The software/CI build has no AppKit adapter. Use the same explicit fallback
// as unsupported desktop platforms instead of claiming to read macOS settings.
func DefaultVariant() granio.ThemeVariant { return theme.VariantDark }
