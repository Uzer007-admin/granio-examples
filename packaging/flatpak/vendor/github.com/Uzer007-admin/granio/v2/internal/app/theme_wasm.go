//go:build wasm

package app

import (
	"syscall/js"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/theme"
)

// DefaultVariant returns the systems default granio.ThemeVariant.
// Normally, you should not need this. It is extracted out of the root app package to give the
// settings app access to it.
func DefaultVariant() granio.ThemeVariant {
	matches := js.Global().Call("matchMedia", "(prefers-color-scheme: dark)")
	if matches.Truthy() {
		if matches.Get("matches").Bool() {
			return theme.VariantDark
		}
		return theme.VariantLight
	}
	return theme.VariantDark
}
