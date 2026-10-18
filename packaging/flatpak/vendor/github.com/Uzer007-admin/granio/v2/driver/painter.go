package driver

import (
	"image"

	"github.com/Uzer007-admin/granio/v2"
)

// Painter describes a simple type that can render canvases
//
// Since: 2.9
type Painter interface {
	Paint(granio.Canvas) image.Image
}
