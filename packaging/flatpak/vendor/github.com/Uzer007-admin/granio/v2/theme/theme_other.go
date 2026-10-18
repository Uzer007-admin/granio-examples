//go:build !hints

package theme

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
)

var (
	fallbackColor = color.Transparent
	fallbackIcon  = &granio.StaticResource{}
)
