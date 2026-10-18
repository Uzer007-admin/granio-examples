package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var (
	_ granio.Widget      = (*SelectEntry)(nil)
	_ granio.Disableable = (*SelectEntry)(nil)
)

// SelectEntry is an input field which supports selecting from a fixed set of options.
type SelectEntry struct {
	Entry
	dropDown *granio.Menu
	popUp    *PopUpMenu
	options  []string
}

// NewSelectEntry creates a SelectEntry.
func NewSelectEntry(options []string) *SelectEntry {
	e := &SelectEntry{options: options}
	e.ExtendBaseWidget(e)
	e.Wrapping = granio.TextWrap(granio.TextTruncateClip)
	return e
}

// CreateRenderer returns a new renderer for this select entry.
func (e *SelectEntry) CreateRenderer() granio.WidgetRenderer {
	e.ExtendBaseWidget(e)
	e.SetOptions(e.options)
	return e.Entry.CreateRenderer()
}

// Enable this widget, updating any style or features appropriately.
func (e *SelectEntry) Enable() {
	if e.ActionItem != nil {
		e.ActionItem.(granio.Disableable).Enable()
	}
	e.Entry.Enable()
}

// Disable this widget so that it cannot be interacted with, updating any style appropriately.
func (e *SelectEntry) Disable() {
	if e.ActionItem != nil {
		e.ActionItem.(granio.Disableable).Disable()
	}
	e.Entry.Disable()
}

// MinSize returns the minimal size of the select entry.
func (e *SelectEntry) MinSize() granio.Size {
	e.ExtendBaseWidget(e)
	return e.Entry.MinSize()
}

// Move changes the relative position of the select entry.
func (e *SelectEntry) Move(pos granio.Position) {
	e.Entry.Move(pos)
	if e.popUp != nil {
		e.popUp.Move(e.popUpPos())
	}
}

// Resize changes the size of the select entry.
func (e *SelectEntry) Resize(size granio.Size) {
	e.Entry.Resize(size)
	if e.popUp != nil {
		e.popUp.Resize(granio.NewSize(size.Width, e.popUp.Size().Height))
	}
}

// SetOptions sets the options the user might select from.
func (e *SelectEntry) SetOptions(options []string) {
	e.options = options
	items := make([]*granio.MenuItem, len(options))
	for i, option := range options {
		option := option // capture
		items[i] = granio.NewMenuItem(option, func() { e.SetText(option) })
	}
	e.dropDown = granio.NewMenu("", items...)

	if e.ActionItem == nil {
		e.ActionItem = e.setupDropDown()
		if e.Disabled() {
			e.ActionItem.(granio.Disableable).Disable()
		}
	}
}

func (e *SelectEntry) popUpPos() granio.Position {
	entryPos := granio.CurrentApp().Driver().AbsolutePositionForObject(e.super())
	return entryPos.Add(granio.NewPos(0, e.Size().Height-e.Theme().Size(theme.SizeNameInputBorder)))
}

func (e *SelectEntry) setupDropDown() *Button {
	dropDownButton := NewButton("", func() {
		c := granio.CurrentApp().Driver().CanvasForObject(e.super())
		if c == nil {
			// SelectEntry detached from its canvas; nothing to host the
			// dropdown on (see fyne-io/fyne#5965).
			return
		}

		e.popUp = NewPopUpMenu(e.dropDown, c)
		e.popUp.ShowAtPosition(e.popUpPos())
		e.popUp.Resize(granio.NewSize(e.Size().Width, e.popUp.MinSize().Height))
	})
	dropDownButton.Importance = LowImportance
	dropDownButton.SetIcon(e.Theme().Icon(theme.IconNameArrowDropDown))
	return dropDownButton
}
