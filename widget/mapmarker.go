package widget

import (
	_ "embed"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type MapMarker interface {
	Lat() float64
	Lon() float64
	Title() string
}

type genericMapMarker struct {
	lat   float64
	lon   float64
	title string
}

func NewMapMarker(lat, lon float64, title string) MapMarker {
	return &genericMapMarker{lat, lon, title}
}

func (m *genericMapMarker) Lat() float64 {
	return m.lat
}

func (m *genericMapMarker) Lon() float64 {
	return m.lon
}

func (m *genericMapMarker) Title() string {
	return m.title
}

type mapMarker struct {
	widget.BaseWidget
	container *granio.Container
	title     *canvas.Text
	obj       MapMarker
	item      *mapMarkerItem
}

//go:embed mapmarker.svg
var resourceMapmarkerSvgData []byte

var resourceMapmarkerSvg = &granio.StaticResource{
	StaticName:    "mapmarker.svg",
	StaticContent: resourceMapmarkerSvgData,
}

func newMapMarker(obj MapMarker) *mapMarker {
	m := &mapMarker{
		obj: obj,
	}
	m.ExtendBaseWidget(m)
	return m
}

func (m *mapMarker) setup() {
	if m.container != nil {
		return
	}

	m.title = canvas.NewText(m.obj.Title(), theme.ColorForWidget(theme.ColorNamePrimary, m))
	m.title.Hide()

	m.item = newMapMarkerItem(func() {
		if m.title.Hidden {
			m.title.Show()
		} else {
			m.title.Hide()
		}
		m.container.Refresh()
	})

	m.container = container.NewVBox(
		m.title,
		container.NewHBox(layout.NewSpacer(), m.item, layout.NewSpacer()),
	)
}

func (m *mapMarker) pinOffset() granio.Position {
	m.setup()
	imgSize := m.item.Size()
	if m.title.Hidden {
		return granio.NewPos(imgSize.Width/2, imgSize.Height+(imgSize.Height*.05))
	}
	conSize := m.container.Size()
	return granio.NewPos(conSize.Width/2, conSize.Height+(imgSize.Height*.05))
}

func (m *mapMarker) MinSize() granio.Size {
	m.setup()
	return m.container.MinSize()
}

func (m *mapMarker) CreateRenderer() granio.WidgetRenderer {
	m.setup()
	return widget.NewSimpleRenderer(m.container)
}
