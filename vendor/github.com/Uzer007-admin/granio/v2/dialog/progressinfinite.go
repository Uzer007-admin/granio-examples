package dialog

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// ProgressInfiniteDialog is a simple dialog window that displays text and a infinite progress bar.
//
// Deprecated: Use NewCustomWithoutButtons() and add a widget.ProgressBarInfinite() inside.
type ProgressInfiniteDialog struct {
	*dialog

	bar *widget.ProgressBarInfinite
}

// NewProgressInfinite creates a infinite progress dialog and returns the handle.
// Using the returned type you should call Show().
//
// Deprecated: Use NewCustomWithoutButtons() and add a widget.ProgressBarInfinite() inside.
func NewProgressInfinite(title, message string, parent granio.Window) *ProgressInfiniteDialog {
	d := newTextDialog(title, message, theme.InfoIcon(), parent)
	bar := widget.NewProgressBarInfinite()
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(granio.NewSize(200, 0)) //revive:disable-line:add-constant

	d.create(container.NewStack(rect, bar))
	return &ProgressInfiniteDialog{d, bar}
}

// Hide this dialog and stop the infinite progress goroutine
func (d *ProgressInfiniteDialog) Hide() {
	d.bar.Hide()
	d.dialog.Hide()
}
