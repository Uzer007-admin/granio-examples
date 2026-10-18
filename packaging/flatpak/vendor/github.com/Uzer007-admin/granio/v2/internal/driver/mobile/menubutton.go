package mobile

import (
	"github.com/Uzer007-admin/granio/v2"
	fynecanvas "github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type menuButton struct {
	widget.BaseWidget
	win  *window
	menu *granio.MainMenu
}

func (w *window) newMenuButton(menu *granio.MainMenu) *menuButton {
	b := &menuButton{win: w, menu: menu}
	b.ExtendBaseWidget(b)
	return b
}

func (m *menuButton) CreateRenderer() granio.WidgetRenderer {
	btn := widget.NewButtonWithIcon("", theme.MenuIcon(), func() {
		m.win.canvas.showMenu(m.menu)
	})
	bg := fynecanvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	return &menuButtonRenderer{btn: btn, bg: bg, objects: []granio.CanvasObject{bg, btn}}
}

type menuButtonRenderer struct {
	btn     *widget.Button
	bg      *fynecanvas.Rectangle
	objects []granio.CanvasObject
}

func (*menuButtonRenderer) Destroy() {
}

func (m *menuButtonRenderer) Layout(size granio.Size) {
	m.bg.Move(granio.NewPos(theme.Padding()/2, theme.Padding()/2))
	m.bg.Resize(size.Subtract(granio.NewSize(theme.Padding(), theme.Padding())))
	m.btn.Resize(size)
}

func (m *menuButtonRenderer) MinSize() granio.Size {
	return m.btn.MinSize()
}

func (m *menuButtonRenderer) Objects() []granio.CanvasObject {
	return m.objects
}

func (m *menuButtonRenderer) Refresh() {
	m.bg.FillColor = theme.Color(theme.ColorNameBackground)
}
