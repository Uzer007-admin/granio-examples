package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"image/color"
)

func activeSystemColors(th granio.Theme) (granio.SystemColors, bool) {
	if provider, ok := th.(granio.SystemColorTheme); ok {
		return provider.SystemColors()
	}
	return granio.SystemColors{}, false
}

func applySelectionStyle(rect *canvas.Rectangle, th granio.Theme, variant granio.ThemeVariant) {
	rect.FillColor = th.Color(theme.ColorNameSelection, variant)
	rect.StrokeColor, rect.StrokeWidth = nil, 0
	if colors, ok := activeSystemColors(th); ok {
		// User-supplied collection children can have their own text styling. A
		// system Highlight outline preserves that text and its shaping instead of
		// painting an opaque highlight behind an unrelated foreground color.
		rect.FillColor = color.Transparent
		rect.StrokeColor, rect.StrokeWidth = colors.Highlight, 2
	}
}

func applySystemFocusCircle(circle *canvas.Circle, th granio.Theme, focused bool) {
	circle.StrokeWidth = 0
	if colors, ok := activeSystemColors(th); ok {
		circle.FillColor = color.Transparent
		if focused {
			circle.StrokeColor, circle.StrokeWidth = colors.Highlight, 2
		}
	}
}
