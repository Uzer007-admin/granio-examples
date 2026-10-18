package app

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
)

// ApplyThemeTo ensures that the specified canvasobject and all widgets and themeable objects will
// be updated for the current theme.
func ApplyThemeTo(content granio.CanvasObject, canv granio.Canvas) {
	if content == nil {
		return
	}
	if stateful, ok := content.(granio.ThemeChangeAware); ok {
		if finish := stateful.BeginThemeChange(); finish != nil {
			defer finish()
		}
	}

	switch o := content.(type) {
	case granio.Widget:
		renderer := cache.Renderer(o)
		for _, co := range renderer.Objects() {
			ApplyThemeTo(co, canv)
		}
		renderer.Layout(content.Size()) // theme can cause sizing changes
	case *granio.Container:
		for _, co := range o.Objects {
			ApplyThemeTo(co, canv)
		}
		if l := o.Layout; l != nil {
			l.Layout(o.Objects, o.Size()) // theme can cause sizing changes
		}
	}
	content.Refresh()
}

// ApplySettings ensures that all widgets and themeable objects in an application will be updated for the current theme.
// It also checks that scale changes are reflected if required
func ApplySettings(set granio.Settings, app granio.App) {
	ApplySettingsWithCallback(set, app, nil)
}

// ApplySettingsWithCallback ensures that all widgets and themeable objects in an application will be updated for the current theme.
// It also checks that scale changes are reflected if required. Also it will call `onEveryWindow` on every window
// interaction
func ApplySettingsWithCallback(_ granio.Settings, app granio.App, onEveryWindow func(w granio.Window)) {
	for _, window := range app.Driver().AllWindows() {
		ApplyThemeTo(window.Content(), window.Canvas())
		for _, overlay := range window.Canvas().Overlays().List() {
			ApplyThemeTo(overlay, window.Canvas())
		}
		if onEveryWindow != nil {
			onEveryWindow(window)
		}
	}
}
