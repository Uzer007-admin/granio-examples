package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type mapButton struct {
	widget.Button
}

func newMapButton(icon granio.Resource, f func()) *mapButton {
	b := &mapButton{}
	b.ExtendBaseWidget(b)

	b.Icon = icon
	b.OnTapped = f
	return b
}

func (b *mapButton) CreateRenderer() granio.WidgetRenderer {
	return &mapButtonRenderer{WidgetRenderer: b.Button.CreateRenderer(),
		bg: canvas.NewRectangle(theme.Color(theme.ColorNameShadow))}
}

type mapButtonRenderer struct {
	granio.WidgetRenderer

	bg *canvas.Rectangle
}

func (r *mapButtonRenderer) Layout(s granio.Size) {
	halfPad := theme.Padding() / 2
	r.bg.Move(granio.NewPos(halfPad, halfPad))
	r.bg.Resize(s.Subtract(granio.NewSize(theme.Padding(), theme.Padding())))

	r.WidgetRenderer.Layout(s)
}

func (r *mapButtonRenderer) Objects() []granio.CanvasObject {
	return append([]granio.CanvasObject{r.bg}, r.WidgetRenderer.Objects()...)
}

func (r *mapButtonRenderer) Refresh() {
	r.bg.FillColor = theme.Color(theme.ColorNameShadow)
	r.WidgetRenderer.Refresh()
}
