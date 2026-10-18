package canvas

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
	"image/color"
	"math"
)

// NewThemedMaterial resolves an in-canvas surface token. Native window backdrops
// remain independent. Ordinary themes receive an opaque compatible surface.
func NewThemedMaterial(th granio.Theme, role granio.SurfaceRole, variant granio.ThemeVariant) *Material {
	if th == nil {
		th = theme.Current()
	}
	fallback := th.Color(theme.ColorNameMenuBackground, variant)
	if fallback == nil {
		fallback = color.White
	}
	m := NewMaterial(fallback)
	provider, ok := th.(granio.MaterialTheme)
	if !ok {
		m.Enabled = false
		return m
	}
	style := provider.MaterialStyle(role, variant)
	m.BlurRadius, m.CornerRadius, m.Noise = style.BlurRadius, style.CornerRadius, style.Noise
	if style.Tint != nil {
		m.Tint = color.NRGBAModel.Convert(style.Tint).(color.NRGBA)
	}
	opacity := style.TintOpacity
	invalidOpacity := math.IsNaN(float64(opacity)) || math.IsInf(float64(opacity), 0) || opacity < 0 || opacity > 1
	if invalidOpacity {
		opacity = 1
	}
	m.Tint.A = uint8(opacity * 255)
	m.Enabled = !style.Opaque && !invalidOpacity
	if style.Fallback != nil {
		m.Fallback = style.Fallback
	}
	if !m.Valid() {
		m.Enabled = false
	}
	return m
}
