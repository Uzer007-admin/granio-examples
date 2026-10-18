package widget

import "github.com/Uzer007-admin/granio/v2"

// BaseRenderer is a renderer base that provides part of the widget.Renderer interface.
type BaseRenderer struct {
	objects []granio.CanvasObject
}

// NewBaseRenderer creates a new BaseRenderer.
func NewBaseRenderer(objects []granio.CanvasObject) BaseRenderer {
	return BaseRenderer{objects}
}

// Destroy does nothing in the base implementation.
func (*BaseRenderer) Destroy() {
}

// Objects returns the objects that should be rendered.
func (r *BaseRenderer) Objects() []granio.CanvasObject {
	return r.objects
}

// SetObjects updates the objects of the renderer.
func (r *BaseRenderer) SetObjects(objects []granio.CanvasObject) {
	r.objects = objects
}
