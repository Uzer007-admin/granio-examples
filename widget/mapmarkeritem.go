package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type mapMarkerItem struct {
	widget.BaseWidget
	img *canvas.Image
	fn  func()
}

func newMapMarkerItem(fn func()) *mapMarkerItem {
	item := &mapMarkerItem{fn: fn}
	item.ExtendBaseWidget(item)
	return item
}

func (m *mapMarkerItem) Tapped(ev *granio.PointEvent) {
	if m.fn == nil {
		return
	}
	m.fn()
}

func (m *mapMarkerItem) setup() {
	if m.img != nil {
		return
	}
	res := theme.NewColoredResource(resourceMapmarkerSvg, theme.ColorNamePrimary)
	m.img = canvas.NewImageFromResource(res)
	m.img.SetMinSize(granio.NewSize(50, 50))
}

func (m *mapMarkerItem) CreateRenderer() granio.WidgetRenderer {
	m.setup()
	return widget.NewSimpleRenderer(m.img)
}
