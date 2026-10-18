package wrapper

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/widget"
)

var _ granio.Widget = (*tappableObject)(nil)
var _ granio.Tappable = (*tappableObject)(nil)

// handles the tap event.
type tappableObject struct {
	widget.BaseWidget
	object   granio.CanvasObject
	onTapped func(*granio.PointEvent)
}

// Content returns the encapsulated widget.
func (t *tappableObject) Content() granio.CanvasObject {
	return t.object
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
//
// Implements: granio.Widget
func (t *tappableObject) CreateRenderer() granio.WidgetRenderer {
	if t.object == nil {
		return nil
	}
	if o, ok := t.object.(granio.Widget); ok {
		return o.CreateRenderer()
	}
	return widget.NewSimpleRenderer(t.object)
}

// Tapped reacts on tap (or click) events.
//
// Implements: granio.Tappable
func (t *tappableObject) Tapped(e *granio.PointEvent) {
	if t.object == nil {
		return
	}

	if o, ok := t.object.(granio.Tappable); ok {
		o.Tapped(e)
	}
	t.onTapped(e)
}

// MakeTappable set the object to be tappable.
func MakeTappable(object granio.CanvasObject, ontapped func(*granio.PointEvent)) granio.CanvasObject {
	tappable := &tappableObject{object: object, onTapped: ontapped}
	tappable.ExtendBaseWidget(tappable)
	return tappable
}
