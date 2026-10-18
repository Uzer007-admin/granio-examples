//go:build !darwin || no_native_menus

package glfw

import "github.com/Uzer007-admin/granio/v2"

func setupNativeMenu(_ *window, _ *granio.MainMenu) {
	// no-op
}
