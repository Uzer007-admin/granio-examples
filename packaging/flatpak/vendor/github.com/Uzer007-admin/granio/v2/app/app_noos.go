//go:build tamago || noos || tinygo

package app

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/embedded"
	intNoos "github.com/Uzer007-admin/granio/v2/internal/driver/embedded"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func NewWithID(id string) granio.App {
	return newAppWithDriver(nil, nil, id)
}

// SetDriverDetails provides the required information to our app to run without a standard
// driver. This is useful for embedded devices like GOOS=tamago, tinygo or noos.
//
// Since: 2.7
func SetDriverDetails(a granio.App, d embedded.Driver) {
	a.(*fyneApp).Settings().SetTheme(theme.DefaultTheme())
	a.(*fyneApp).driver = intNoos.NewNoOSDriver(d.Render, d.Run, d.Queue(), d.ScreenSize)
}
