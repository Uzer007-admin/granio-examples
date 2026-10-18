package tool

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"

	"github.com/Uzer007-admin/granio-pixeledit/internal/api"
	"github.com/Uzer007-admin/granio-pixeledit/internal/data"
)

// Pencil is a pixel setting tool
type Pencil struct {
	Out *canvas.Rectangle
}

// Name returns the name of this tool
func (p *Pencil) Name() string {
	return "Pencil"
}

// Icon returns the icon of this tool
func (p *Pencil) Icon() granio.Resource {
	return data.PencilIcon
}

// Clicked will set the pixel under the cursor to the current editor foreground color
func (p *Pencil) Clicked(x, y int, edit api.Editor) {
	edit.SetPixelColor(x, y, edit.FGColor())
}
