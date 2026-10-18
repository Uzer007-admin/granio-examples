package glfw

import (
	"github.com/Uzer007-admin/granio/v2"
)

func buildMenuOverlay(menus *granio.MainMenu, w *window) granio.CanvasObject {
	if len(menus.Items) == 0 {
		granio.LogError("Main menu must have at least one child menu", nil)
		return nil
	}

	menus = addMissingQuitForMainMenu(menus, w)
	return NewMenuBar(menus, w.canvas)
}
