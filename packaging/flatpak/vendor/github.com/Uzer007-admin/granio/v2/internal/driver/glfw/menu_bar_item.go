package glfw

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/theme"
	publicWidget "github.com/Uzer007-admin/granio/v2/widget"
)

var (
	_ granio.Widget     = (*menuBarItem)(nil)
	_ desktop.Hoverable = (*menuBarItem)(nil)
	_ granio.Focusable  = (*menuBarItem)(nil)
)

// menuBarItem is a widget for displaying an item for a granio.Menu in a MenuBar.
type menuBarItem struct {
	widget.Base
	Menu   *granio.Menu
	Parent *MenuBar

	active  bool
	child   *publicWidget.Menu
	hovered bool
}

func (i *menuBarItem) Child() *publicWidget.Menu {
	if i.child == nil {
		child := publicWidget.NewMenu(i.Menu)
		child.Hide()
		child.OnDismiss = i.Parent.deactivate
		i.child = child
	}
	return i.child
}

// CreateRenderer returns a new renderer for the menu bar item.
func (i *menuBarItem) CreateRenderer() granio.WidgetRenderer {
	background := canvas.NewRectangle(theme.Color(theme.ColorNameHover))
	background.CornerRadius = theme.SelectionRadiusSize()
	background.Hide()
	text := canvas.NewText(i.Menu.Label, theme.Color(theme.ColorNameForeground))
	objects := []granio.CanvasObject{background, text}

	return &menuBarItemRenderer{
		widget.NewBaseRenderer(objects),
		i,
		text,
		background,
	}
}

func (i *menuBarItem) FocusGained() {
	i.active = true
	if i.Parent.active {
		i.Parent.activateChild(i)
	}
	i.Refresh()
}

func (i *menuBarItem) FocusLost() {
	i.active = false
	i.Refresh()
}

func (i *menuBarItem) Focused() bool {
	return i.active
}

// MouseIn activates the item and shows the menu if the bar is active.
// The menu that was displayed before will be hidden.
//
// If the bar is not active, the item will be hovered.
func (i *menuBarItem) MouseIn(_ *desktop.MouseEvent) {
	i.hovered = true
	if i.Parent.active {
		i.Parent.canvas.Focus(i)
	}
	i.Refresh()
}

// MouseMoved activates the item and shows the menu if the bar is active.
// The menu that was displayed before will be hidden.
// This might have an effect when mouse and keyboard control are mixed.
// Changing the active menu with the keyboard will make the hovered menu bar item inactive.
// On the next mouse move the hovered item is activated again.
//
// If the bar is not active, this will do nothing.
func (i *menuBarItem) MouseMoved(_ *desktop.MouseEvent) {
	if i.Parent.active {
		i.Parent.canvas.Focus(i)
	}
}

// MouseOut does nothing if the bar is active.
//
// IF the bar is not active, it changes the item to not be hovered.
func (i *menuBarItem) MouseOut() {
	i.hovered = false
	i.Refresh()
}

// Tapped toggles the activation state of the menu bar.
// It shows the item’s menu if the bar is activated and hides it if the bar is deactivated.
func (i *menuBarItem) Tapped(*granio.PointEvent) {
	i.Parent.toggleItem(i)
}

func (i *menuBarItem) TypedKey(event *granio.KeyEvent) {
	switch event.Name {
	case granio.KeyLeft:
		if !i.Child().DeactivateLastSubmenu() {
			i.Parent.canvas.FocusPrevious()
		}
	case granio.KeyRight:
		if !i.Child().ActivateLastSubmenu() {
			i.Parent.canvas.FocusNext()
		}
	case granio.KeyDown:
		i.Child().ActivateNext()
	case granio.KeyUp:
		i.Child().ActivatePrevious()
	case granio.KeyEnter, granio.KeyReturn, granio.KeySpace:
		i.Child().TriggerLast()
	}
}

func (*menuBarItem) TypedRune(_ rune) {
}

type menuBarItemRenderer struct {
	widget.BaseRenderer
	i          *menuBarItem
	text       *canvas.Text
	background *canvas.Rectangle
}

func (r *menuBarItemRenderer) Layout(size granio.Size) {
	padding := r.padding()

	r.text.TextSize = theme.TextSize()
	r.text.Color = theme.Color(theme.ColorNameForeground)
	r.text.Resize(r.text.MinSize())
	r.text.Move(granio.NewPos(padding.Width/2, padding.Height/2))

	r.background.Resize(size)
}

func (r *menuBarItemRenderer) MinSize() granio.Size {
	return r.text.MinSize().Add(r.padding())
}

func (r *menuBarItemRenderer) Refresh() {
	r.background.CornerRadius = theme.SelectionRadiusSize()
	if r.i.active && r.i.Parent.active {
		r.background.FillColor = theme.Color(theme.ColorNameFocus)
		r.background.Show()
	} else if r.i.hovered && !r.i.Parent.active {
		r.background.FillColor = theme.Color(theme.ColorNameHover)
		r.background.Show()
	} else {
		r.background.Hide()
	}
	r.background.Refresh()
	canvas.Refresh(r.i)
}

func (*menuBarItemRenderer) padding() granio.Size {
	return granio.NewSize(theme.InnerPadding()*2, theme.InnerPadding())
}
