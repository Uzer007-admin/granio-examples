package container

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	intTheme "github.com/Uzer007-admin/granio/v2/internal/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// ThemeOverride is a container where the child widgets are themed by the specified theme.
// Containers will be traversed and all child widgets will reflect the theme in this container.
// This should be used sparingly to avoid a jarring user experience.
//
// Since: 2.5
type ThemeOverride struct {
	widget.BaseWidget

	Content granio.CanvasObject
	Theme   granio.Theme

	holder *granio.Container

	mobile bool
}

// NewThemeOverride provides a container where the child widgets are themed by the specified theme.
// Containers will be traversed and all child widgets will reflect the theme in this container.
// This should be used sparingly to avoid a jarring user experience.
//
// If the content `obj` of this theme override is a container and items are later added to the container or any
// sub-containers ensure that you call `Refresh()` on this `ThemeOverride` to ensure the new items match the theme.
//
// Since: 2.5
func NewThemeOverride(obj granio.CanvasObject, th granio.Theme) *ThemeOverride {
	t := &ThemeOverride{Content: obj, Theme: th, holder: NewStack(obj)}
	t.ExtendBaseWidget(t)

	cache.OverrideTheme(obj, addFeatures(th, t))
	obj.Refresh() // required as the widgets passed in could have been initially rendered with default theme
	return t
}

// CreateRenderer implements the [granio.Widget] interface.
func (t *ThemeOverride) CreateRenderer() granio.WidgetRenderer {
	cache.OverrideTheme(t.Content, addFeatures(t.Theme, t))

	return &overrideRenderer{parent: t, objs: []granio.CanvasObject{t.holder}}
}

// Refresh implements the [granio.CanvasObject] interface.
func (t *ThemeOverride) Refresh() {
	if t.holder.Objects[0] != t.Content {
		t.holder.Objects[0] = t.Content
		t.holder.Refresh()
	}

	cache.OverrideTheme(t.Content, addFeatures(t.Theme, t))
	t.Content.Refresh()
	t.BaseWidget.Refresh()
}

// SetDeviceIsMobile allows a ThemeOverride container to shape the contained widgets as a mobile device.
// This will impact containers such as AppTabs and DocTabs, and more in the future, to display a layout
// that would automatically be used for a mobile device runtime.
//
// Since: 2.6
func (t *ThemeOverride) SetDeviceIsMobile(on bool) {
	t.mobile = on
	t.Refresh()
}

type featureTheme struct {
	granio.Theme

	over *ThemeOverride
}

func addFeatures(th granio.Theme, o *ThemeOverride) granio.Theme {
	return &featureTheme{Theme: th, over: o}
}

func (f *featureTheme) Feature(n intTheme.FeatureName) any {
	if n == intTheme.FeatureNameDeviceIsMobile {
		return f.over.mobile
	}

	return nil
}

type overrideRenderer struct {
	parent *ThemeOverride

	objs []granio.CanvasObject
}

func (*overrideRenderer) Destroy() {
}

func (r *overrideRenderer) Layout(s granio.Size) {
	intTheme.PushRenderingTheme(r.parent.Theme)
	defer intTheme.PopRenderingTheme()

	r.parent.holder.Resize(s)
}

func (r *overrideRenderer) MinSize() granio.Size {
	intTheme.PushRenderingTheme(r.parent.Theme)
	defer intTheme.PopRenderingTheme()

	return r.parent.Content.MinSize()
}

func (r *overrideRenderer) Objects() []granio.CanvasObject {
	return r.objs
}

func (*overrideRenderer) Refresh() {
}
