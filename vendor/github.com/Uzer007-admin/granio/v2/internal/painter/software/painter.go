package software

import (
	"image"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal/driver"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
)

// Painter is a simple software painter that can paint a canvas in memory.
type Painter struct{}

// NewPainter creates a new Painter.
func NewPainter() *Painter {
	return &Painter{}
}

// Paint is the main entry point for a simple software painter.
// The canvas to be drawn is passed in as a parameter and the return is an
// image containing the result of rendering.
func (*Painter) Paint(c granio.Canvas) image.Image {
	bounds := image.Rect(0, 0, scale.ToScreenCoordinate(c, c.Size().Width), scale.ToScreenCoordinate(c, c.Size().Height))
	base := image.NewNRGBA(bounds)

	paint := func(obj granio.CanvasObject, pos, clipPos granio.Position, clipSize granio.Size) bool {
		w := min(clipPos.X+clipSize.Width, c.Size().Width)
		h := min(clipPos.Y+clipSize.Height, c.Size().Height)
		clip := image.Rect(
			scale.ToScreenCoordinate(c, clipPos.X),
			scale.ToScreenCoordinate(c, clipPos.Y),
			scale.ToScreenCoordinate(c, w),
			scale.ToScreenCoordinate(c, h),
		)
		switch o := obj.(type) {
		case *canvas.Material:
			drawRectangle(c, o.FallbackRectangle(), pos, base, clip)
		case *canvas.Image:
			drawImage(c, o, pos, base, clip)
		case *canvas.Text:
			drawText(c, o, pos, base, clip)
		case gradient:
			drawGradient(c, o, pos, base, clip)
		case *canvas.Circle:
			drawCircle(c, o, pos, base, clip)
		case *canvas.Line:
			drawLine(c, o, pos, base, clip)
		case *canvas.RegularPolygon:
			drawPolygon(c, o, pos, base, clip)
		case *canvas.ArbitraryPolygon:
			drawArbitraryPolygon(c, o, pos, base, clip)
		case *canvas.Raster:
			drawRaster(c, o, pos, base, clip)
		case *canvas.Rectangle:
			drawRectangle(c, o, pos, base, clip)
		case *canvas.Blur:
			drawBlur(c, o, pos, base, clip)
		case *canvas.Arc:
			drawArc(c, o, pos, base, clip)
		case *canvas.BezierCurve:
			drawBezierCurve(c, o, pos, base, clip)
		case *canvas.Ellipse:
			drawEllipse(c, o, pos, base, clip)
		}

		return false
	}

	driver.WalkVisibleObjectTree(c.Content(), paint, nil)
	for _, o := range c.Overlays().List() {
		driver.WalkVisibleObjectTree(o, paint, nil)
	}

	return base
}
