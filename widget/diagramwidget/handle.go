package diagramwidget

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// Validate implementation of Draggable
var _ granio.Draggable = (*Handle)(nil)

var defaultHandleSize float32 = 10.0

// Handle is a widget used to manipulate the size or shape of its owning DiagramElement
type Handle struct {
	widget.BaseWidget
	handleSize float32
	de         DiagramElement
}

// NewHandle creates a handle for the specified DiagramElement
func NewHandle(diagramElement DiagramElement) *Handle {
	handle := &Handle{
		de:         diagramElement,
		handleSize: defaultHandleSize,
	}
	handle.BaseWidget.ExtendBaseWidget(handle)
	return handle
}

// CreateRenderer is the required method for the Handle widget
func (h *Handle) CreateRenderer() granio.WidgetRenderer {
	hr := &handleRenderer{
		handle: h,
		rect:   canvas.NewRectangle(h.getStrokeColor()),
	}
	hr.rect.FillColor = color.Transparent
	hr.Refresh()
	return hr
}

// Dragged respondss to drag events, passing them on to the owning DiagramElement. It is the
// DiagramElement that determines what to do as a result of the drag.
func (h *Handle) Dragged(event *granio.DragEvent) {
	h.de.handleDragged(h, event)
}

// DragEnd passes the event on to the owning DiagramElement
func (h *Handle) DragEnd() {
	h.de.handleDragEnd(h)
}

func (h *Handle) getStrokeColor() color.Color {
	return h.de.GetDiagram().GetForegroundColor()
}

func (h *Handle) getStrokeWidth() float32 {
	return 1.0
}

// Move changes the position of the handle
func (h *Handle) Move(position granio.Position) {
	delta := granio.Position{X: -h.handleSize / 2, Y: -h.handleSize / 2}
	h.BaseWidget.Move(position.Add(delta))
}

// handleRenderer
type handleRenderer struct {
	handle *Handle
	rect   *canvas.Rectangle
}

func (hr *handleRenderer) Destroy() {

}

// Layout sets both the handle and its rectangle to the minimum size
func (hr *handleRenderer) Layout(size granio.Size) {
	hr.rect.Resize(hr.MinSize())
	hr.handle.Resize(hr.MinSize())
}

// MinSize returns the minimum size of the Handle widget
func (hr *handleRenderer) MinSize() granio.Size {
	return granio.Size{Height: hr.handle.handleSize, Width: hr.handle.handleSize}
}

// Objects returns the objects that comprise the Handel
func (hr *handleRenderer) Objects() []granio.CanvasObject {
	obj := []granio.CanvasObject{
		hr.rect,
	}
	return obj
}

// Refresh re-renders the Handle after rendering properties have been changed
func (hr *handleRenderer) Refresh() {
	hr.rect.StrokeColor = hr.handle.getStrokeColor()
	hr.rect.FillColor = color.Transparent
	hr.rect.StrokeWidth = hr.handle.getStrokeWidth()
	hr.rect.Refresh()
}
