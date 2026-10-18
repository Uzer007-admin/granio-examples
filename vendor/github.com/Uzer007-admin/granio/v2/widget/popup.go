package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var _ granio.Widget = (*PopUp)(nil)

// PopUp is a widget that can float above the user interface.
// It wraps any standard elements with padding and a shadow.
// If it is modal then the shadow will cover the entire canvas it hovers over and block interactions.
type PopUp struct {
	BaseWidget

	Content granio.CanvasObject
	Canvas  granio.Canvas
	// Material optionally filters the lower canvas layers behind this popup.
	// Its opaque fallback follows the popup theme; unsupported painters retain
	// readable controls. Assign on the UI thread, then call Refresh.
	Material *canvas.Material

	// OnDismiss is a callback invoked when the PopUp is dismissed,
	// either by tapping outside of it if non-modal, or by calling Hide().
	//
	// Since: 2.9
	OnDismiss func() `json:"-"`
	// OnEscape handles Escape for a modal surface (for example a dialog's cancel
	// callback). Non-modal popups close by default. Nil keeps a modal popup open.
	OnEscape func() `json:"-"`

	overlay             *widget.OverlayContainer
	modal, manual       bool
	previousFocus       granio.Focusable
	appearanceAnimation *granio.Animation
}

// Hide this widget, if it was previously visible
func (p *PopUp) Hide() {
	if p.appearanceAnimation != nil {
		p.appearanceAnimation.Stop()
		p.appearanceAnimation = nil
	}
	shown := p.overlay != nil
	if p.overlay != nil {
		p.Canvas.Overlays().Remove(p.overlay)
		p.overlay = nil
	}

	p.BaseWidget.Hide()
	p.Move(granio.Position{}) // reset so that Show or ShowAtPosition next time is respected
	if shown {
		restorePopupFocus(p.Canvas, p.previousFocus)
		p.previousFocus = nil
	}

	if p.OnDismiss != nil {
		p.OnDismiss()
	}
}

// Refresh the background for a modal popup and the content of this popup.
func (p *PopUp) Refresh() {
	if p.appearanceAnimation != nil {
		p.appearanceAnimation.Stop()
		p.appearanceAnimation = nil
	}
	if p.modal && p.overlay != nil {
		th := p.Theme()
		v := granio.CurrentApp().Settings().ThemeVariant()

		bg, _ := p.overlay.Background.(*granio.Container).Objects[1].(*canvas.Rectangle)
		bg.FillColor = th.Color(theme.ColorNameShadow, v)
	}

	p.BaseWidget.Refresh()
}

// Show this pop-up as overlay if not already shown.
func (p *PopUp) Show() {
	newSurface := p.overlay == nil
	if p.overlay == nil {
		if _, modern := p.Theme().(granio.ComponentTheme); modern {
			p.previousFocus = p.Canvas.Focused()
			if p.previousFocus != nil {
				p.previousFocus.FocusLost()
			}
		}
		dismiss := p.Hide
		if p.modal {
			dismiss = nil
		}
		p.overlay = widget.NewOverlayContainer(p.super(), p.Canvas, dismiss)
		if p.modal {
			th := p.Theme()
			v := granio.CurrentApp().Settings().ThemeVariant()

			p.overlay.Background = &granio.Container{
				Layout: layout.NewStackLayout(),
				Objects: []granio.CanvasObject{
					canvas.NewBlur(th.Size(theme.SizeNameModalBlurRadius)),
					canvas.NewRectangle(th.Color(theme.ColorNameShadow, v)),
				},
			}
		}
		p.Canvas.Overlays().Add(p.overlay)
	}
	p.Refresh()
	p.BaseWidget.Show()
	if newSurface {
		p.animateAppearance()
	}
}

// HandleEscape is an optional overlay contract used by desktop key dispatch.
func (p *PopUp) HandleEscape() bool {
	if p.OnEscape != nil {
		p.OnEscape()
		return true
	}
	if p.modal {
		return false
	}
	p.Hide()
	return true
}

// ShowAtPosition shows this pop-up at the given position.
func (p *PopUp) ShowAtPosition(pos granio.Position) {
	p.manual = true
	p.Move(pos)
	p.Show()
}

// ShowAtRelativePosition shows this pop-up at the given position relative to stated object.
//
// Since 2.4
func (p *PopUp) ShowAtRelativePosition(rel granio.Position, to granio.CanvasObject) {
	withRelativePosition(rel, to, p.ShowAtPosition)
}

// Tapped is called when the user taps the popUp.
// This is not called when tapping the background, but non-modal popups will dismiss when tapped outside.
func (*PopUp) Tapped(*granio.PointEvent) {
}

// TappedSecondary is called when the user right/alt taps the popUp.
// This is not called when tapping the background, but non-modal popups will dismiss when tapped outside.
func (*PopUp) TappedSecondary(*granio.PointEvent) {
}

// MinSize returns the size that this widget should not shrink below
func (p *PopUp) MinSize() granio.Size {
	p.ExtendBaseWidget(p)
	return p.BaseWidget.MinSize()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (p *PopUp) CreateRenderer() granio.WidgetRenderer {
	th := p.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	p.ExtendBaseWidget(p)
	background := canvas.NewRectangle(th.Color(theme.ColorNameOverlayBackground, v))
	widget.ApplyShadowForLevel(&background.Shadow, widget.PopUpLevel, th.Color(theme.ColorNameShadow, v))
	background.CornerRadius = th.Size(theme.SizeNamePopupRadius)
	objects := []granio.CanvasObject{background, p.Content}
	if p.Material != nil {
		background.FillColor = nil
		objects = []granio.CanvasObject{background, p.Material, p.Content}
	}
	return &popUpRenderer{
		widget.NewBaseRenderer(objects),
		popUpBaseRenderer{popUp: p, background: background},
		p.Material,
	}
}

// ShowPopUpAtPosition creates a new popUp for the specified content at the specified absolute position.
// It will then display the popup on the passed canvas.
func ShowPopUpAtPosition(content granio.CanvasObject, c granio.Canvas, pos granio.Position) {
	NewPopUp(content, c).ShowAtPosition(pos)
}

// ShowPopUpAtRelativePosition shows a new popUp for the specified content at the given position relative to stated object.
// It will then display the popup on the passed canvas.
//
// Since 2.4
func ShowPopUpAtRelativePosition(content granio.CanvasObject, c granio.Canvas, rel granio.Position, to granio.CanvasObject) {
	withRelativePosition(rel, to, func(pos granio.Position) {
		ShowPopUpAtPosition(content, c, pos)
	})
}

// NewPopUp creates a new popUp for the specified content and displays it on the passed canvas.
func NewPopUp(content granio.CanvasObject, c granio.Canvas) *PopUp {
	ret := &PopUp{Content: content, Canvas: c, modal: false}
	ret.ExtendBaseWidget(ret)
	return ret
}

// ShowPopUp creates a new popUp for the specified content and displays it on the passed canvas.
func ShowPopUp(content granio.CanvasObject, c granio.Canvas) {
	NewPopUp(content, c).Show()
}

// NewModalPopUp creates a new popUp for the specified content and displays it on the passed canvas.
// A modal PopUp blocks interactions with underlying elements, covered with a semi-transparent overlay.
func NewModalPopUp(content granio.CanvasObject, c granio.Canvas) *PopUp {
	p := &PopUp{Content: content, Canvas: c, modal: true}
	p.ExtendBaseWidget(p)
	return p
}

// ShowModalPopUp creates a new popUp for the specified content and displays it on the passed canvas.
// A modal PopUp blocks interactions with underlying elements, covered with a semi-transparent overlay.
func ShowModalPopUp(content granio.CanvasObject, c granio.Canvas) {
	p := NewModalPopUp(content, c)
	p.Show()
}

type popUpBaseRenderer struct {
	popUp      *PopUp
	background *canvas.Rectangle
}

func (r *popUpBaseRenderer) padding() granio.Size {
	th := r.popUp.Theme()
	return granio.NewSquareSize(th.Size(theme.SizeNameInnerPadding))
}

type popUpRenderer struct {
	widget.BaseRenderer
	popUpBaseRenderer
	material *canvas.Material
}

func (r *popUpRenderer) Destroy() {
	if r.popUp.appearanceAnimation != nil {
		r.popUp.appearanceAnimation.Stop()
		r.popUp.appearanceAnimation = nil
	}
}

func (r *popUpRenderer) Layout(s granio.Size) {
	size := internal.MaxSizes(s, r.popUp.Content.MinSize())
	if r.popUp.Canvas != nil {
		canvasSize := r.popUp.Canvas.Size()
		if !canvasSize.IsZero() {
			size = internal.MinSizes(size, r.popUp.Canvas.Size())
		}
	}
	r.popUp.Content.Resize(size)

	r.background.Resize(size)
	if r.material != nil {
		r.material.Resize(size)
		r.material.Move(r.background.Position())
	}
}

func (r *popUpRenderer) MinSize() granio.Size {
	return r.popUp.Content.MinSize()
}

func (r *popUpRenderer) Refresh() {
	innerPos := r.popUp.Content.Position()
	innerSize := r.popUp.Content.Size()

	th := r.popUp.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()
	r.background.FillColor = th.Color(theme.ColorNameOverlayBackground, v)
	if r.material != r.popUp.Material {
		r.material = r.popUp.Material
		objects := []granio.CanvasObject{r.background, r.popUp.Content}
		if r.material != nil {
			objects = []granio.CanvasObject{r.background, r.material, r.popUp.Content}
		}
		r.SetObjects(objects)
		r.Layout(r.popUp.Size())
	}
	if r.material != nil {
		r.material.Fallback = r.background.FillColor
		r.background.FillColor = nil
		r.material.Refresh()
	}
	r.background.Shadow.Color = th.Color(theme.ColorNameShadow, v)
	r.background.CornerRadius = th.Size(theme.SizeNamePopupRadius)
	expectedContentSize := internal.MaxSizes(innerSize, r.popUp.MinSize()).Subtract(r.padding())
	shouldRelayout := r.popUp.Content.Size() != expectedContentSize

	if r.background.Size() != innerSize || r.background.Position() != innerPos || shouldRelayout {
		r.Layout(r.popUp.Size())
	}
	r.popUp.Content.Refresh()
	r.background.Refresh()
}

func withRelativePosition(rel granio.Position, to granio.CanvasObject, f func(position granio.Position)) {
	d := granio.CurrentApp().Driver()
	c := d.CanvasForObject(to)
	if c == nil {
		granio.LogError("Could not locate parent object to display relative to", nil)
		f(rel)
		return
	}

	pos := d.AbsolutePositionForObject(to).Add(rel)
	f(pos)
}
