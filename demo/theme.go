package main

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
)

type forcedVariant struct {
	granio.Theme

	variant granio.ThemeVariant
}

func (f *forcedVariant) Color(name granio.ThemeColorName, _ granio.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}
