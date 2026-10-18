package canvas

import (
	"image/color"
	"math"

	"github.com/Uzer007-admin/granio/v2"
)

// Material filters the already painted lower canvas layers. It never captures
// other windows or the desktop. Native GL/GLES painters share a bounded pool of
// two material layers per window. Other painters, disabled effects and exhausted
// budgets paint Fallback instead. A material does not request animation frames.
type Material struct {
	baseObject
	Enabled                         bool
	BlurRadius, CornerRadius, Noise float32
	Tint                            color.NRGBA
	Fallback                        color.Color
	Downsample                      uint8 // 1, 2, 4 or 8; zero selects 2
	Linear                          bool
	fallbackRectangle               Rectangle
}

func NewMaterial(fallback color.Color) *Material {
	return &Material{Enabled: true, BlurRadius: 12, CornerRadius: 8, Tint: color.NRGBA{R: 255, G: 255, B: 255, A: 160}, Fallback: fallback, Downsample: 2}
}
func (m *Material) Refresh() { Refresh(m) }
func (m *Material) Hide()    { m.baseObject.Hide(); repaint(m) }
func (m *Material) Move(pos granio.Position) {
	if pos != m.Position() {
		m.baseObject.Move(pos)
		repaint(m)
	}
}
func (m *Material) Resize(size granio.Size) {
	if size != m.Size() {
		m.baseObject.Resize(size)
		m.Refresh()
	}
}

// OpaqueFallback is defined even for nil or translucent caller colors.
func (m *Material) OpaqueFallback() color.Color {
	if m.Fallback == nil {
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	}
	c := color.NRGBAModel.Convert(m.Fallback).(color.NRGBA)
	c.A = 255
	return c
}
func (m *Material) Valid() bool {
	finite := func(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
	return finite(m.BlurRadius) && finite(m.CornerRadius) && finite(m.Noise) && m.BlurRadius >= 0 && m.BlurRadius <= 50 && m.CornerRadius >= 0 && m.Noise >= 0 && m.Noise <= 1 && (m.Downsample == 0 || m.Downsample == 1 || m.Downsample == 2 || m.Downsample == 4 || m.Downsample == 8)
}

// FallbackRectangle returns a borrowed opaque primitive with matching geometry.
// Painters use it without registering another canvas object or refreshing a tree.
func (m *Material) FallbackRectangle() *Rectangle {
	r := &m.fallbackRectangle
	r.FillColor = m.OpaqueFallback()
	r.CornerRadius = m.CornerRadius
	if math.IsNaN(float64(r.CornerRadius)) || math.IsInf(float64(r.CornerRadius), 0) || r.CornerRadius < 0 {
		r.CornerRadius = 0
	}
	r.baseObject.Resize(m.Size())
	return r
}
