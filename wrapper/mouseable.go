package wrapper

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/widget"
)

var _ granio.Widget = (*mouseableObject)(nil)
var _ desktop.Hoverable = (*mouseableObject)(nil)

// handles the canvas object and mouse events
type mouseableObject struct {
	widget.BaseWidget
	object     granio.CanvasObject
	mouseIn    func(*desktop.MouseEvent)
	mouseMoved func(*desktop.MouseEvent)
	mouseOut   func()
}

// Content returns the originl object that was set to be hoverable.
func (m *mouseableObject) Content() granio.CanvasObject {
	return m.object
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
//
// Implements: granio.Widget
func (m *mouseableObject) CreateRenderer() granio.WidgetRenderer {
	if m.object == nil {
		return nil
	}
	if o, ok := m.object.(granio.Widget); ok {
		return o.CreateRenderer()
	}
	return widget.NewSimpleRenderer(m.object)
}

// MouseIn is called when the mouse enters the widget.
//
// Implements: desktop.Hoverable
func (m *mouseableObject) MouseIn(e *desktop.MouseEvent) {
	if m.mouseIn == nil {
		return
	}

	if o, ok := m.object.(desktop.Hoverable); ok {
		o.MouseIn(e)
	}

	m.mouseIn(e)
}

// MouseMoved is called when the mouse moves over the widget.
//
// Implements: desktop.Hoverable
func (m *mouseableObject) MouseMoved(e *desktop.MouseEvent) {
	if m.mouseMoved == nil {
		return
	}

	if o, ok := m.object.(desktop.Hoverable); ok {
		o.MouseMoved(e)
	}

	m.mouseMoved(e)
}

// MouseOut is called when the mouse exits the widget.
//
// Implements: desktop.Hoverable
func (m *mouseableObject) MouseOut() {
	if m.mouseOut == nil {
		return
	}

	if o, ok := m.object.(desktop.Hoverable); ok {
		o.MouseOut()
	}

	m.mouseOut()
}

// MakeHoverable sets the object to be hoverable.
func MakeHoverable(object granio.CanvasObject, mouseIn, mouseMoved func(*desktop.MouseEvent), mouseout func()) granio.CanvasObject {
	m := &mouseableObject{
		object:     object,
		mouseIn:    mouseIn,
		mouseMoved: mouseMoved,
		mouseOut:   mouseout,
	}
	m.ExtendBaseWidget(m)
	return m
}
