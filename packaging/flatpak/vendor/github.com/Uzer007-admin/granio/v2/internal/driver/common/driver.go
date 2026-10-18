package common

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
)

// CanvasForObject returns the canvas for the specified object.
func CanvasForObject(obj granio.CanvasObject) granio.Canvas {
	return cache.GetCanvasForObject(obj)
}
