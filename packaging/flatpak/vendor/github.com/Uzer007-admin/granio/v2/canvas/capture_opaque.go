package canvas

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// CaptureOpaque captures the rendered Fyne canvas and composites its alpha over
// an opaque background. Nil background uses the current theme's window color.
// Call on the UI goroutine. The window, material policy and OS settings remain
// untouched. Native DWM/compositor effects are outside Canvas.Capture and must
// be captured separately for hardware acceptance.
func CaptureOpaque(c granio.Canvas, background color.Color) *image.NRGBA {
	if c == nil {
		return nil
	}
	source := c.Capture()
	if source == nil {
		return nil
	}
	if background == nil {
		background = theme.BackgroundColor()
	}
	if background == nil {
		background = color.White
	}
	bg := color.NRGBAModel.Convert(background).(color.NRGBA)
	bg.A = 255
	result := image.NewNRGBA(image.Rect(0, 0, source.Bounds().Dx(), source.Bounds().Dy()))
	draw.Draw(result, result.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	draw.Draw(result, result.Bounds(), source, source.Bounds().Min, draw.Over)
	return result
}
