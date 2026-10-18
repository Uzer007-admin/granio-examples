// Package software implements a more generic system for just rendering without a window.
// Primarily, it was used in tests but found to be useful elsewhere.
// Between embedded and test utilities this may not be needed any more.
package software

import (
	"image"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/app"
)

// RenderCanvas takes a canvas and renders it to a regular Go image using the provided Theme.
// This is the same as setting the application theme and then calling Canvas.Capture().
func RenderCanvas(c granio.Canvas, t granio.Theme) image.Image {
	granio.CurrentApp().Settings().SetTheme(t)
	app.ApplyThemeTo(c.Content(), c)

	return c.Capture()
}

// Render takes a canvas object and renders it to a regular Go image using the provided Theme.
// The returned image will be set to the object's minimum size.
// Use the theme.LightTheme() or theme.DarkTheme() to access the builtin themes.
func Render(obj granio.CanvasObject, t granio.Theme) image.Image {
	granio.CurrentApp().Settings().SetTheme(t)

	c := NewCanvas()
	c.SetPadded(false)
	c.SetContent(obj)

	app.ApplyThemeTo(obj, c)
	return c.Capture()
}
