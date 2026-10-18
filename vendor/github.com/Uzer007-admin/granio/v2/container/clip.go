package container

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// Declare conformity with Widget interface
var _ granio.Widget = (*Clip)(nil)

// Clip describes a rectangular region  that will clip anything outside its bounds.
//
// Since: 2.7
type Clip struct {
	widget.BaseWidget
	Content granio.CanvasObject
}

// NewClip returns a new rectangular clipping object.
//
// Since: 2.7
func NewClip(content granio.CanvasObject) *Clip {
	return &Clip{Content: content}
}

// CreateRenderer implements the [granio.Widget] interface.
func (c *Clip) CreateRenderer() granio.WidgetRenderer {
	c.ExtendBaseWidget(c)
	return newClipRenderer(c)
}

// MinSize for a Clip simply returns Size{1, 1} as there is no
// explicit content
func (c *Clip) MinSize() granio.Size {
	c.ExtendBaseWidget(c)
	return granio.NewSize(1, 1)
}

type clipRenderer struct {
	c       *Clip
	objects []granio.CanvasObject
}

func newClipRenderer(c *Clip) *clipRenderer {
	return &clipRenderer{c: c, objects: []granio.CanvasObject{c.Content}}
}

func (*clipRenderer) Destroy() {
}

func (r *clipRenderer) Layout(s granio.Size) {
	o := r.objects[0]
	o.Resize(internal.MaxSizes(s, o.MinSize()))
}

func (r *clipRenderer) MinSize() granio.Size {
	return r.objects[0].MinSize()
}

func (r *clipRenderer) Objects() []granio.CanvasObject {
	return r.objects
}

func (r *clipRenderer) Refresh() {
	r.objects[0] = r.c.Content
	r.Layout(r.c.Size())
	r.objects[0].Refresh()
}

// IsClip marks this widget as clipping. It is on the renderer to avoid a public API addition.
func (*clipRenderer) IsClip() {}
