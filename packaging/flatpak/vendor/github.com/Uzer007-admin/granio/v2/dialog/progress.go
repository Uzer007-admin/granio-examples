package dialog

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// ProgressDialog is a simple dialog window that displays text and a progress bar.
//
// Deprecated: Use NewCustomWithoutButtons() and add a widget.ProgressBar() inside.
type ProgressDialog struct {
	*dialog

	bar *widget.ProgressBar
}

// SetValue updates the value of the progress bar - this should be between 0.0 and 1.0.
func (p *ProgressDialog) SetValue(v float64) {
	p.bar.SetValue(v)
}

// NewProgress creates a progress dialog and returns the handle.
// Using the returned type you should call Show() and then set its value through SetValue().
//
// Deprecated: Use NewCustomWithoutButtons() and add a widget.ProgressBar() inside.
func NewProgress(title, message string, parent granio.Window) *ProgressDialog {
	d := newTextDialog(title, message, theme.InfoIcon(), parent)
	bar := widget.NewProgressBar()
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(granio.NewSize(200, 0)) //revive:disable-line:add-constant

	d.create(container.NewStack(rect, bar))
	return &ProgressDialog{d, bar}
}
