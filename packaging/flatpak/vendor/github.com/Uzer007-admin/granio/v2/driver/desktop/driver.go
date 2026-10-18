// Package desktop provides desktop specific driver functionality.
package desktop

import "github.com/Uzer007-admin/granio/v2"

// Driver represents the extended capabilities of a desktop driver
type Driver interface {
	// CreateSplashWindow creates a new borderless window that is centered on screen
	CreateSplashWindow() granio.Window

	// CurrentKeyModifiers returns the set of key modifiers that are currently active
	//
	// Since: 2.4
	CurrentKeyModifiers() granio.KeyModifier

	// HasSecondaryDisplay returns true if there are more than one screens available on this computer.
	// This is commonly used alongside the desktop Window.RequestFullScreenSecondary call.
	//
	// Since: 2.8
	HasSecondaryDisplay() bool
}
