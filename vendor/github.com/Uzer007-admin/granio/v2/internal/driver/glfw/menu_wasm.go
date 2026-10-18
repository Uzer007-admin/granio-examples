//go:build wasm || test_web_driver

package glfw

import "github.com/Uzer007-admin/granio/v2"

func addMissingQuitForMainMenu(menus *granio.MainMenu, w *window) *granio.MainMenu {
	// no-op for a web browser
	return menus
}
