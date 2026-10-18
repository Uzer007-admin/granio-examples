package gl

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2/theme"
)

type WindowSurfaceMode uint8

const (
	WindowSurfaceLegacy WindowSurfaceMode = iota
	WindowSurfaceOpaque
	WindowSurfaceBackdrop
)

type premultipliedContext interface{ setPremultipliedAlpha(bool) bool }

// SetWindowSurface runs with the owning context current. It never changes the
// global theme. Stock straight-color passes use ONE for source alpha; stock
// premultiplied texture passes retain their RGB/alpha factors. Custom shaders
// must obey the premultiplied framebuffer contract when using a backdrop.
func (p *painter) SetWindowSurface(mode WindowSurfaceMode) bool {
	if mode > WindowSurfaceBackdrop {
		return false
	}
	control, ok := p.ctx.(premultipliedContext)
	if !ok {
		return false
	}
	if mode == WindowSurfaceBackdrop && p.ctx.GetInteger(0x0D55) == 0 {
		return false
	}
	if !control.setPremultipliedAlpha(mode != WindowSurfaceLegacy) {
		return false
	}
	if p.windowSurface != mode {
		// The canvas clear colour is part of every material snapshot, even when
		// none of the lower objects changed. Do not reuse the old alpha/background.
		p.materialGeneration++
	}
	p.windowSurface = mode
	return true
}

func (p *painter) WindowSurfaceAlphaBits() int {
	if p.ctx == nil {
		return 0
	}
	return p.ctx.GetInteger(0x0D55)
}

func (p *painter) windowClearColor() (float32, float32, float32, float32) {
	if p.windowSurface == WindowSurfaceBackdrop {
		return 0, 0, 0, 0
	}
	background := theme.Color(theme.ColorNameBackground)
	if p.windowSurface == WindowSurfaceOpaque {
		straight := color.NRGBAModel.Convert(background).(color.NRGBA)
		return float32(straight.R) / 255, float32(straight.G) / 255, float32(straight.B) / 255, 1
	}
	r, g, b, a := background.RGBA()
	return float32(r) / max16bit, float32(g) / max16bit, float32(b) / max16bit, float32(a) / max16bit
}
