//go:build !ci && !android && !ios && !mobile && !tamago && !noos && !tinygo

package app

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/driver/glfw"
)

// NewWithID returns a new app instance using the appropriate runtime driver.
// The ID string should be globally unique to this app.
func NewWithID(id string) granio.App {
	return newAppWithDriver(glfw.NewGLDriver(), glfw.NewClipboard(), id)
}
