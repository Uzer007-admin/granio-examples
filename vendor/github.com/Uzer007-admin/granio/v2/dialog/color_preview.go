package dialog

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal"
	internalwidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// colorPreview displays a 2 part rectangle showing the current and previous selected colours
type colorPreview struct {
	widget.BaseWidget

	previous, current color.Color
}

func newColorPreview(previousColor color.Color) *colorPreview {
	p := &colorPreview{previous: previousColor}

	p.ExtendBaseWidget(p)
	return p
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (p *colorPreview) CreateRenderer() granio.WidgetRenderer {
	oldC := canvas.NewRectangle(p.previous)
	newC := canvas.NewRectangle(p.current)
	background := newCheckeredBackground(false)
	return &colorPreviewRenderer{
		BaseRenderer: internalwidget.NewBaseRenderer([]granio.CanvasObject{background, oldC, newC}),
		preview:      p,
		background:   background,
		old:          oldC,
		new:          newC,
	}
}

func (p *colorPreview) SetColor(c color.Color) {
	p.current = c
	p.Refresh()
}

func (p *colorPreview) MinSize() granio.Size {
	p.ExtendBaseWidget(p)
	return p.BaseWidget.MinSize()
}

type colorPreviewRenderer struct {
	internalwidget.BaseRenderer
	preview    *colorPreview
	background *canvas.Raster
	old, new   *canvas.Rectangle
}

func (r *colorPreviewRenderer) Layout(size granio.Size) {
	s := granio.NewSize(size.Width/2, size.Height)
	r.background.Resize(size)
	r.old.Resize(s)
	r.new.Resize(s)
	r.new.Move(granio.NewPos(s.Width, 0))
}

func (r *colorPreviewRenderer) MinSize() granio.Size {
	s := r.old.MinSize()
	s.Width *= 2
	return internal.MaxSizes(s, granio.NewSize(16, 8)) //revive:disable-line:add-constant
}

func (r *colorPreviewRenderer) Refresh() {
	r.background.Refresh()

	r.old.FillColor = r.preview.previous
	r.old.Refresh()
	r.new.FillColor = r.preview.current
	r.new.Refresh()
}
