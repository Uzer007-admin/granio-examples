package test

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
)

func DarkTheme(fallback granio.Theme) granio.Theme {
	return &forcedVariant{Theme: fallback, variant: 0} // avoid import loops
}

func LightTheme(fallback granio.Theme) granio.Theme {
	return &forcedVariant{Theme: fallback, variant: 1} // avoid import loops
}

type forcedVariant struct {
	granio.Theme

	variant granio.ThemeVariant
}

func (f *forcedVariant) Color(name granio.ThemeColorName, _ granio.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.variant)
}
