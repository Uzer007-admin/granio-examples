package widget

import "github.com/Uzer007-admin/granio/v2"

var _ granio.WidgetRenderer = (*SimpleRenderer)(nil)

// SimpleRenderer is a basic renderer that satisfies widget.Renderer interface by wrapping
// a single granio.CanvasObject.
//
// Since: 2.1
type SimpleRenderer struct {
	objects []granio.CanvasObject
}

// NewSimpleRenderer creates a new SimpleRenderer to render a widget using a
// single CanvasObject.
//
// Since: 2.1
func NewSimpleRenderer(object granio.CanvasObject) *SimpleRenderer {
	return &SimpleRenderer{[]granio.CanvasObject{object}}
}

// Destroy does nothing in this implementation.
//
// Since: 2.1
func (*SimpleRenderer) Destroy() {
}

// Layout updates the contained object to be the requested size.
//
// Since: 2.1
func (r *SimpleRenderer) Layout(s granio.Size) {
	r.objects[0].Resize(s)
}

// MinSize returns the smallest size that this render can use, returned from the underlying object.
//
// Since: 2.1
func (r *SimpleRenderer) MinSize() granio.Size {
	return r.objects[0].MinSize()
}

// Objects returns the objects that should be rendered.
//
// Since: 2.1
func (r *SimpleRenderer) Objects() []granio.CanvasObject {
	return r.objects
}

// Refresh requests the underlying object to redraw.
//
// Since: 2.1
func (r *SimpleRenderer) Refresh() {
	r.objects[0].Refresh()
}
