// Package dialog defines standard dialog windows for application GUIs.
package dialog // import "github.com/Uzer007-admin/granio/v2/dialog"

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/internal"
	col "github.com/Uzer007-admin/granio/v2/internal/color"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

const (
	padWidth  = 32
	padHeight = 16
)

// Dialog is the common API for any dialog window with a single dismiss button
type Dialog interface {
	Show()
	Hide()
	SetDismissText(label string)
	SetOnClosed(closed func())
	Refresh()
	Resize(size granio.Size)

	// MinSize returns the size that this dialog should not shrink below.
	//
	// Since: 2.1
	MinSize() granio.Size

	// Dismiss instructs the dialog to close without any affirmative action.
	//
	// Since: 2.6
	Dismiss()
}

// Declare conformity to Dialog interface
var _ Dialog = (*dialog)(nil)

type dialog struct {
	callback    func(bool)
	title       string
	icon        granio.Resource
	desiredSize granio.Size

	win     *widget.PopUp
	content granio.CanvasObject
	dismiss *widget.Button
	parent  granio.Window

	// allows derived dialogs to inject logic that runs before Show()
	beforeShowHook func()
}

func (d *dialog) Dismiss() {
	d.Hide()
}

func (d *dialog) Hide() {
	d.hideWithResponse(false)
}

// MinSize returns the size that this dialog should not shrink below.
//
// Since: 2.1
func (d *dialog) MinSize() granio.Size {
	if d.win == nil {
		return granio.NewSquareSize(1)
	}
	return d.win.MinSize()
}

func (d *dialog) Show() {
	if d.beforeShowHook != nil {
		d.beforeShowHook()
	}
	if !d.desiredSize.IsZero() {
		d.win.Resize(d.desiredSize)
	}
	d.win.Show()
}

func (d *dialog) Refresh() {
	d.win.Refresh()
}

// Resize dialog, call this function after dialog show
func (d *dialog) Resize(size granio.Size) {
	d.desiredSize = internal.MaxSizes(size, d.MinSize())
	if d.win != nil { // could be called before popup is created!
		d.win.Resize(internal.MaxSizes(size, d.MinSize()))
	}
}

// SetDismissText allows custom text to be set in the dismiss button
// This is a no-op for dialogs without dismiss buttons.
func (d *dialog) SetDismissText(label string) {
	if d.dismiss == nil {
		return
	}

	d.dismiss.SetText(label)
	d.win.Refresh()
}

// SetOnClosed allows to set a callback function that is called when
// the dialog is closed
func (d *dialog) SetOnClosed(closed func()) {
	// if there is already a callback set, remember it and call both
	originalCallback := d.callback

	d.callback = func(response bool) {
		if originalCallback != nil {
			originalCallback(response)
		}
		closed()
	}
}

func (d *dialog) hideWithResponse(resp bool) {
	d.win.Hide()
	if d.callback != nil {
		d.callback(resp)
	}
}

func (d *dialog) create(buttons granio.CanvasObject) {
	label := widget.NewLabelWithStyle(d.title, granio.TextAlignLeading, granio.TextStyle{Bold: true})

	var image granio.CanvasObject
	if d.icon != nil {
		image = &canvas.Image{Resource: d.icon}
	} else {
		image = &layout.Spacer{}
	}

	content := container.New(
		&dialogLayout{d: d},
		image,
		newThemedBackground(),
		d.content,
		buttons,
		label,
	)

	d.win = widget.NewModalPopUp(content, d.parent.Canvas())
	d.win.OnEscape = func() {
		if d.dismiss != nil {
			d.Dismiss()
		}
	}
}

func (d *dialog) setButtons(buttons granio.CanvasObject) {
	d.win.Content.(*granio.Container).Objects[3] = buttons
	d.win.Refresh()
}

func (d *dialog) setIcon(icon granio.Resource) {
	if icon == nil {
		d.win.Content.(*granio.Container).Objects[0] = &layout.Spacer{}
		d.win.Refresh()
		return
	}
	d.win.Content.(*granio.Container).Objects[0] = &canvas.Image{Resource: icon}
	d.win.Refresh()
}

// The method .create() needs to be called before the dialog can be shown.
func newDialog(title, message string, icon granio.Resource, callback func(bool), parent granio.Window) *dialog {
	d := &dialog{content: newCenterWrappedLabel(message), title: title, icon: icon, parent: parent}
	d.callback = callback

	return d
}

// ===============================================================
// ThemedBackground
// ===============================================================

type themedBackground struct {
	widget.BaseWidget
}

func newThemedBackground() *themedBackground {
	t := &themedBackground{}
	t.ExtendBaseWidget(t)
	return t
}

func (t *themedBackground) CreateRenderer() granio.WidgetRenderer {
	t.ExtendBaseWidget(t)
	rect := canvas.NewRectangle(theme.Color(theme.ColorNameOverlayBackground))
	rect.CornerRadius = theme.Size(theme.SizeNameDialogRadius)
	return &themedBackgroundRenderer{rect, []granio.CanvasObject{rect}}
}

type themedBackgroundRenderer struct {
	rect    *canvas.Rectangle
	objects []granio.CanvasObject
}

func (*themedBackgroundRenderer) Destroy() {
}

func (renderer *themedBackgroundRenderer) Layout(size granio.Size) {
	renderer.rect.Resize(size)
}

func (renderer *themedBackgroundRenderer) MinSize() granio.Size {
	return renderer.rect.MinSize()
}

func (renderer *themedBackgroundRenderer) Objects() []granio.CanvasObject {
	return renderer.objects
}

func (renderer *themedBackgroundRenderer) Refresh() {
	r, g, b, _ := col.ToNRGBA(theme.Color(theme.ColorNameOverlayBackground))
	bg := &color.NRGBA{R: r, G: g, B: b, A: 230} //revive:disable-line:add-constant
	renderer.rect.FillColor = bg
}

// ===============================================================
// DialogLayout
// ===============================================================

type dialogLayout struct {
	d *dialog
}

func (*dialogLayout) Layout(obj []granio.CanvasObject, size granio.Size) {
	btnMin := obj[3].MinSize()
	labelMin := obj[4].MinSize()

	// icon
	iconHeight := padHeight*2 + labelMin.Height*2 - theme.Padding()
	obj[0].Resize(granio.NewSize(iconHeight, iconHeight))
	obj[0].Move(granio.NewPos(size.Width-iconHeight+theme.Padding(), -theme.Padding()))

	// background
	obj[1].Move(granio.NewPos(0, 0))
	obj[1].Resize(size)

	// buttons
	obj[3].Resize(btnMin)
	obj[3].Move(granio.NewPos(size.Width/2-(btnMin.Width/2), size.Height-padHeight-btnMin.Height))

	// content
	contentStart := obj[4].Position().Y + labelMin.Height + padHeight
	contentEnd := obj[3].Position().Y - theme.Padding()
	obj[2].Move(granio.NewPos(padWidth/2, labelMin.Height+padHeight))
	obj[2].Resize(granio.NewSize(size.Width-padWidth, contentEnd-contentStart))
}

func (*dialogLayout) MinSize(obj []granio.CanvasObject) granio.Size {
	contentMin := obj[2].MinSize()
	btnMin := obj[3].MinSize()
	labelMin := obj[4].MinSize()

	width := max(contentMin.Width, btnMin.Width, labelMin.Width) + padWidth
	height := contentMin.Height + btnMin.Height + labelMin.Height + theme.Padding() + padHeight*2

	return granio.NewSize(width, height)
}
