package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
)

var (
	_ granio.Widget    = (*PopUpMenu)(nil)
	_ granio.Focusable = (*PopUpMenu)(nil)
)

// PopUpMenu is a Menu which displays itself in an OverlayContainer.
type PopUpMenu struct {
	*Menu
	canvas        granio.Canvas
	overlay       *widget.OverlayContainer
	previousFocus granio.Focusable
}

// NewPopUpMenu creates a new, reusable popup menu. You can show it using ShowAtPosition.
//
// Since: 2.0
func NewPopUpMenu(menu *granio.Menu, c granio.Canvas) *PopUpMenu {
	if c == nil {
		return nil
	}
	m := &Menu{}
	m.setMenu(menu)
	p := &PopUpMenu{Menu: m, canvas: c}
	p.ExtendBaseWidget(p)
	p.Resize(p.MinSize())
	p.customSized = true

	p.Move(granio.NewPos(10, 10)) //revive:disable-line:add-constant // non-zero pos to get manual overlay, fixed on show
	o := widget.NewOverlayContainer(p, c, p.Dismiss)
	o.Resize(o.MinSize())
	p.overlay = o
	p.OnDismiss = func() {
		p.Hide()
	}
	return p
}

// ShowPopUpMenuAtPosition creates a PopUp menu populated with items from the passed menu structure.
// It will automatically be positioned at the provided location and shown as an overlay on the specified canvas.
func ShowPopUpMenuAtPosition(menu *granio.Menu, c granio.Canvas, pos granio.Position) {
	m := NewPopUpMenu(menu, c)
	if m == nil {
		return
	}
	m.ShowAtPosition(pos)
}

// ShowPopUpMenuAtRelativePosition creates a PopUp menu populated with menu items from the passed menu structure.
// It will automatically be positioned at the given position relative to stated object and shown as an overlay on the specified canvas.
//
// Since 2.4
func ShowPopUpMenuAtRelativePosition(menu *granio.Menu, c granio.Canvas, rel granio.Position, to granio.CanvasObject) {
	withRelativePosition(rel, to, func(pos granio.Position) {
		ShowPopUpMenuAtPosition(menu, c, pos)
	})
}

// FocusGained is triggered when the object gained focus. For the pop-up menu it does nothing.
func (*PopUpMenu) FocusGained() {}

// FocusLost is triggered when the object lost focus. For the pop-up menu it does nothing.
func (*PopUpMenu) FocusLost() {}

// Hide hides the pop-up menu.
func (p *PopUpMenu) Hide() {
	p.overlay.Hide()
	p.Menu.Hide()
	restorePopupFocus(p.canvas, p.previousFocus)
	p.previousFocus = nil
}

// Move moves the pop-up menu.
// The position is absolute because pop-up menus are shown in an overlay which covers the whole canvas.
func (p *PopUpMenu) Move(pos granio.Position) {
	p.BaseWidget.Move(p.adjustedPosition(pos, p.Size()))
}

// Resize changes the size of the pop-up menu.
func (p *PopUpMenu) Resize(size granio.Size) {
	p.BaseWidget.Move(p.adjustedPosition(p.Position(), size))
	p.Menu.Resize(size)
}

// SetCanvas allows a popup menu to be re-used on a different canvas.
//
// Since: 2.8
func (p *PopUpMenu) SetCanvas(c granio.Canvas) {
	p.canvas = c
	p.overlay.SetCanvas(c)
}

// Show makes the pop-up menu visible.
func (p *PopUpMenu) Show() {
	if _, modern := p.Theme().(granio.ComponentTheme); modern && p.canvas.Overlays().Top() != p.overlay {
		p.previousFocus = p.canvas.Focused()
		if p.previousFocus != nil {
			p.previousFocus.FocusLost()
		}
	}
	p.Refresh()
	p.overlay.Show()
	p.Menu.Show()
	if !granio.CurrentDevice().IsMobile() {
		p.canvas.Focus(p)
	}
}

func (p *PopUpMenu) HandleEscape() bool { p.Dismiss(); return true }

// ShowAtPosition shows the pop-up menu at the specified position.
func (p *PopUpMenu) ShowAtPosition(pos granio.Position) {
	p.Move(pos)
	p.Show()
}

// ShowAtRelativePosition shows the pop-up menu at the position relative to given object.
//
// Since 2.4
func (p *PopUpMenu) ShowAtRelativePosition(rel granio.Position, to granio.CanvasObject) {
	withRelativePosition(rel, to, p.ShowAtPosition)
}

// TypedKey handles key events. It allows keyboard control of the pop-up menu.
func (p *PopUpMenu) TypedKey(e *granio.KeyEvent) {
	switch e.Name {
	case granio.KeyDown:
		p.ActivateNext()
	case granio.KeyEnter, granio.KeyReturn, granio.KeySpace:
		p.TriggerLast()
	case granio.KeyEscape:
		p.Dismiss()
	case granio.KeyLeft:
		p.DeactivateLastSubmenu()
	case granio.KeyRight:
		p.ActivateLastSubmenu()
	case granio.KeyUp:
		p.ActivatePrevious()
	}
}

// TypedRune handles text events. For pop-up menus this does nothing.
func (*PopUpMenu) TypedRune(rune) {}

func (p *PopUpMenu) adjustedPosition(pos granio.Position, size granio.Size) granio.Position {
	x := pos.X
	y := pos.Y
	_, areaSize := p.canvas.InteractiveArea()
	if x+size.Width > areaSize.Width {
		x = areaSize.Width - size.Width
		if x < 0 {
			x = 0 // TODO here we may need a scroller as it's wider than our canvas
		}
	}
	if y+size.Height > areaSize.Height {
		y = areaSize.Height - size.Height
		if y < 0 {
			y = 0 // TODO here we may need a scroller as it's longer than our canvas
		}
	}
	return granio.NewPos(x, y)
}
