//go:build ci

package app

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/painter/software"
	"github.com/Uzer007-admin/granio/v2/test"
)

// NewWithID returns a new app instance using the test (headless) driver.
// The ID string should be globally unique to this app.
func NewWithID(id string) granio.App {
	return newAppWithDriver(test.NewDriverWithPainter(software.NewPainter()), test.NewClipboard(), id)
}
