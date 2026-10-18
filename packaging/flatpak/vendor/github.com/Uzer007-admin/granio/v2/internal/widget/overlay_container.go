package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal"
)

var (
	_ granio.Widget            = (*OverlayContainer)(nil)
	_ granio.Tappable          = (*OverlayContainer)(nil)
	_ granio.SecondaryTappable = (*OverlayContainer)(nil)
	_ desktop.Hoverable        = (*OverlayContainer)(nil)
)

// OverlayContainer is a transparent widget containing one granio.CanvasObject and meant to be used as overlay.
type OverlayContainer struct {
	Base
	Content, Background granio.CanvasObject

	canvas        granio.Canvas
	onDismiss     func()
	shown, manual bool
}

// NewOverlayContainer creates an OverlayContainer.
func NewOverlayContainer(c granio.CanvasObject, canvas granio.Canvas, onDismiss func()) *OverlayContainer {
	o := &OverlayContainer{canvas: canvas, Content: c, onDismiss: onDismiss}
	o.manual = c != nil && !c.Position().IsZero()
	o.ExtendBaseWidget(o)
	return o
}

// CreateRenderer returns a new renderer for the overlay container.
func (o *OverlayContainer) CreateRenderer() granio.WidgetRenderer {
	objs := []granio.CanvasObject{o.Content}
	if o.Background != nil {
		objs = []granio.CanvasObject{o.Background, o.Content}
	}

	o.manual = !o.Content.Position().IsZero()
	return &overlayRenderer{BaseRenderer{objs}, o}
}

func (o *OverlayContainer) Refresh() {
	o.Content.Refresh()
	o.Base.Refresh()
}

// Hide hides the overlay container.
func (o *OverlayContainer) Hide() {
	if o.shown {
		o.canvas.Overlays().Remove(o)
		o.shown = false
	}
	o.Base.Hide()
}

// MouseIn catches mouse-in events not handled by the container’s content. It does nothing.
func (*OverlayContainer) MouseIn(*desktop.MouseEvent) {
}

// MouseMoved catches mouse-moved events not handled by the container’s content. It does nothing.
func (*OverlayContainer) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut catches mouse-out events not handled by the container’s content. It does nothing.
func (*OverlayContainer) MouseOut() {
}

// SetCanvas allows an overlay container to be re-used on a different canvas.
//
// Since: 2.8
func (o *OverlayContainer) SetCanvas(c granio.Canvas) {
	o.canvas.Overlays().Remove(o)
	o.canvas = c
	o.canvas.Overlays().Add(o)
}

// Show makes the overlay container visible.
func (o *OverlayContainer) Show() {
	if !o.shown {
		o.canvas.Overlays().Add(o)
		o.shown = true
	}
	o.Base.Show()
}

// Tapped catches tap events not handled by the container’s content.
// It performs the overlay container’s dismiss action.
func (o *OverlayContainer) Tapped(*granio.PointEvent) {
	if o.onDismiss != nil {
		o.onDismiss()
	}
}

// TappedSecondary catches secondary tap events not handled by the container’s content.
// It performs the overlay container’s dismiss action.
func (o *OverlayContainer) TappedSecondary(*granio.PointEvent) {
	if o.onDismiss != nil {
		o.onDismiss()
	}
}

// HandleEscape delegates only to the top surface, preserving dialog callbacks.
func (o *OverlayContainer) HandleEscape() bool {
	if surface, ok := o.Content.(interface{ HandleEscape() bool }); ok {
		return surface.HandleEscape()
	}
	return false
}

type overlayRenderer struct {
	BaseRenderer
	o *OverlayContainer
}

var _ granio.WidgetRenderer = (*overlayRenderer)(nil)

func (r *overlayRenderer) Layout(s granio.Size) {
	if s.IsZero() {
		return
	}

	size := internal.MinSizes(internal.MaxSizes(r.o.Content.Size(), r.o.Content.MinSize()), s)
	r.o.Content.Resize(size)

	if r.o.Background != nil {
		r.o.Background.Resize(s)
	}

	if r.o.manual { // position is overridden
		return
	}

	midX := s.Width / 2
	midX -= size.Width / 2
	midY := s.Height / 2
	midY -= size.Height / 2
	r.o.Content.Move(granio.NewPos(midX, midY))
}

func (r *overlayRenderer) MinSize() granio.Size {
	return r.o.Content.MinSize()
}

func (r *overlayRenderer) Refresh() {
	r.Layout(r.o.Size()) // in case theme changed we might need to move
}
