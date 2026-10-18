package cache

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/async"
)

var renderers async.Map[granio.Widget, *rendererInfo]

type isBaseWidget interface {
	ExtendBaseWidget(granio.Widget)
	super() granio.Widget
}

// Renderer looks up the render implementation for a widget
// If one does not exist, it creates and caches a renderer for the widget.
func Renderer(wid granio.Widget) granio.WidgetRenderer {
	renderer, ok := CachedRenderer(wid)
	if !ok && wid != nil {
		renderer = wid.CreateRenderer()
		rinfo := &rendererInfo{renderer: renderer}
		rinfo.setAlive()
		renderers.Store(wid, rinfo)
		// Store before Layout: layout may look up this renderer recursively.
		// A renderer recreated after cache expiry must recover the widget's size.
		if restorer, ok := wid.(interface{ RestoreRendererLayout(granio.WidgetRenderer) }); ok {
			restorer.RestoreRendererLayout(renderer)
		} else {
			renderer.Layout(wid.Size())
		}
	}

	return renderer
}

// CachedRenderer looks up the cached render implementation for a widget
// If a renderer does not exist in the cache, it returns nil, false.
func CachedRenderer(wid granio.Widget) (granio.WidgetRenderer, bool) {
	if wid == nil {
		return nil, false
	}

	if wd, ok := wid.(isBaseWidget); ok {
		if wd.super() != nil {
			wid = wd.super()
		}
	}

	rinfo, ok := renderers.Load(wid)
	if !ok {
		return nil, false
	}

	rinfo.setAlive()
	return rinfo.renderer, true
}

// DestroyRenderer frees a render implementation for a widget.
// This is typically for internal use only.
func DestroyRenderer(wid granio.Widget) {
	rinfo, ok := renderers.LoadAndDelete(wid)
	if !ok {
		return
	}
	if rinfo != nil {
		rinfo.renderer.Destroy()
	}
	overrides.Delete(wid)
}

// IsRendered returns true of the widget currently has a renderer.
// One will be created the first time a widget is shown but may be removed after it is hidden.
func IsRendered(wid granio.Widget) bool {
	_, found := renderers.Load(wid)
	return found
}

type rendererInfo struct {
	expiringCache
	renderer granio.WidgetRenderer
}
