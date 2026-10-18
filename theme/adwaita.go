package theme

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

//go:generate go run ./adwaita_theme_generator.go

var _ granio.Theme = (*Adwaita)(nil)

// Adwaita is a theme that follows the Adwaita theme. It provides a light and dark theme + icons.
// See: https://gnome.pages.gitlab.gnome.org/libadwaita/doc/main/named-colors.html
type Adwaita struct{}

// AdwaitaTheme returns a new Adwaita theme.
func AdwaitaTheme() granio.Theme {
	return &Adwaita{}
}

// Color returns the named color for the current theme.
func (a *Adwaita) Color(name granio.ThemeColorName, variant granio.ThemeVariant) color.Color {
	switch variant {
	case theme.VariantLight:
		if c, ok := adwaitaLightScheme[name]; ok {
			return c
		}
	case theme.VariantDark:
		if c, ok := adwaitaDarkScheme[name]; ok {
			return c
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Font returns the named font for the current theme.
func (a *Adwaita) Font(style granio.TextStyle) granio.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the named resource for the current theme.
func (a *Adwaita) Icon(name granio.ThemeIconName) granio.Resource {
	if icon, ok := adwaitaIcons[name]; ok {
		return icon
	}
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size of the named resource for the current theme.
func (a *Adwaita) Size(name granio.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
