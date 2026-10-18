package dialog

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal"
	internalwidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

var (
	_ granio.Widget     = (*colorButton)(nil)
	_ desktop.Hoverable = (*colorButton)(nil)
)

// colorButton displays a color and triggers the callback when tapped.
type colorButton struct {
	widget.BaseWidget
	color   color.Color
	onTap   func(color.Color)
	hovered bool
}

// newColorButton creates a colorButton with the given color and callback.
func newColorButton(c color.Color, onTap func(color.Color)) *colorButton {
	b := &colorButton{
		color: c,
		onTap: onTap,
	}
	b.ExtendBaseWidget(b)
	return b
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (b *colorButton) CreateRenderer() granio.WidgetRenderer {
	b.ExtendBaseWidget(b)
	background := newCheckeredBackground(false)
	rectangle := &canvas.Rectangle{
		FillColor: b.color,
	}
	return &colorButtonRenderer{
		BaseRenderer: internalwidget.NewBaseRenderer([]granio.CanvasObject{background, rectangle}),
		button:       b,
		background:   background,
		rectangle:    rectangle,
	}
}

// MouseIn is called when a desktop pointer enters the widget
func (b *colorButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

// MouseOut is called when a desktop pointer exits the widget
func (b *colorButton) MouseOut() {
	b.hovered = false
	b.Refresh()
}

// MouseMoved is called when a desktop pointer hovers over the widget
func (*colorButton) MouseMoved(*desktop.MouseEvent) {
}

// MinSize returns the size that this widget should not shrink below
func (b *colorButton) MinSize() granio.Size {
	return b.BaseWidget.MinSize()
}

// SetColor updates the color selected in this color widget
func (b *colorButton) SetColor(c color.Color) {
	if b.color == c {
		return
	}
	b.color = c
	b.Refresh()
}

// Tapped is called when a pointer tapped event is captured and triggers any change handler
func (b *colorButton) Tapped(*granio.PointEvent) {
	if f := b.onTap; f != nil {
		f(b.color)
	}
}

type colorButtonRenderer struct {
	internalwidget.BaseRenderer
	button     *colorButton
	background *canvas.Raster
	rectangle  *canvas.Rectangle
}

func (r *colorButtonRenderer) Layout(size granio.Size) {
	r.rectangle.Move(granio.NewPos(0, 0))
	r.rectangle.Resize(size)
	r.background.Resize(size)
}

func (r *colorButtonRenderer) MinSize() granio.Size {
	return internal.MaxSizes(r.rectangle.MinSize(), granio.NewSize(32, 32)) //revive:disable-line:add-constant
}

func (r *colorButtonRenderer) Refresh() {
	if r.button.hovered {
		r.rectangle.StrokeColor = theme.Color(theme.ColorNameHover)
		r.rectangle.StrokeWidth = theme.Padding()
	} else {
		r.rectangle.StrokeWidth = 0
	}
	r.rectangle.FillColor = r.button.color
	r.background.Refresh()
	canvas.Refresh(r.button)
}
