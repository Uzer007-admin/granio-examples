package ui

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type interactiveRaster struct {
	widget.BaseWidget

	edit *editor
	min  granio.Size
	img  *canvas.Raster
}

func (r *interactiveRaster) SetMinSize(size granio.Size) {
	pixWidth, _ := r.locationForPosition(granio.NewPos(size.Width, size.Height))
	scale := float32(1.0)
	c := granio.CurrentApp().Driver().CanvasForObject(r.img)
	if c != nil {
		scale = c.Scale()
	}

	texScale := float32(pixWidth) / size.Width * float32(r.edit.zoom) / scale
	size = granio.NewSize(size.Width/texScale, size.Height/texScale)
	r.min = size
	r.Resize(size)
}

func (r *interactiveRaster) MinSize() granio.Size {
	return r.min
}

func (r *interactiveRaster) CreateRenderer() granio.WidgetRenderer {
	return &rasterWidgetRender{raster: r, bg: canvas.NewRasterWithPixels(bgPattern)}
}

func (r *interactiveRaster) Tapped(ev *granio.PointEvent) {
	if r.edit.tool == nil || r.edit.img == nil {
		return
	}

	x, y := r.locationForPosition(ev.Position)
	if x >= r.edit.img.Bounds().Dx() || y >= r.edit.img.Bounds().Dy() {
		return
	}

	r.edit.tool.Clicked(x, y, r.edit)
}

func (r *interactiveRaster) TappedSecondary(*granio.PointEvent) {
}

func (r *interactiveRaster) locationForPosition(pos granio.Position) (int, int) {
	c := granio.CurrentApp().Driver().CanvasForObject(r.img)
	x, y := int(pos.X), int(pos.Y)
	if c != nil {
		x, y = c.PixelCoordinateForPosition(pos)
	}

	return x / r.edit.zoom, y / r.edit.zoom
}

func newInteractiveRaster(edit *editor) *interactiveRaster {
	r := &interactiveRaster{img: canvas.NewRaster(edit.draw), edit: edit}
	r.ExtendBaseWidget(r)
	return r
}

type rasterWidgetRender struct {
	raster *interactiveRaster
	bg     *canvas.Raster
}

func bgPattern(x, y, _, _ int) color.Color {
	const boxSize = 25

	if (x/boxSize)%2 == (y/boxSize)%2 {
		return color.Gray{Y: 58}
	}

	return color.Gray{Y: 84}
}

func (r *rasterWidgetRender) Layout(size granio.Size) {
	r.bg.Resize(size)
	r.raster.img.Resize(size)
}

func (r *rasterWidgetRender) MinSize() granio.Size {
	return r.MinSize()
}

func (r *rasterWidgetRender) Refresh() {
	canvas.Refresh(r.raster)
}

func (r *rasterWidgetRender) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *rasterWidgetRender) Objects() []granio.CanvasObject {
	return []granio.CanvasObject{r.bg, r.raster.img}
}

func (r *rasterWidgetRender) Destroy() {
}
