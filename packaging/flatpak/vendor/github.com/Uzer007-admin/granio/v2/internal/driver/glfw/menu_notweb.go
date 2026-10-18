//go:build !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/lang"
)

func addMissingQuitForMainMenu(menus *granio.MainMenu, w *window) *granio.MainMenu {
	localQuit := lang.L("Quit")
	var lastItem *granio.MenuItem
	if len(menus.Items[0].Items) > 0 {
		lastItem = menus.Items[0].Items[len(menus.Items[0].Items)-1]
		if lastItem.Label == localQuit {
			lastItem.IsQuit = true
		}
	}
	if lastItem == nil || !lastItem.IsQuit { // make sure the first menu always has a quit option
		quitItem := granio.NewMenuItem(localQuit, nil)
		quitItem.IsQuit = true
		menus.Items[0].Items = append(menus.Items[0].Items, granio.NewMenuItemSeparator(), quitItem)
	}
	for _, item := range menus.Items[0].Items {
		if item.IsQuit && item.Action == nil {
			item.Action = func() {
				for _, win := range w.driver.AllWindows() {
					if glWin, ok := win.(*window); ok {
						glWin.closed(glWin.view())
					} else {
						win.Close() // for test windows
					}
				}
			}
		}
	}
	return menus
}
