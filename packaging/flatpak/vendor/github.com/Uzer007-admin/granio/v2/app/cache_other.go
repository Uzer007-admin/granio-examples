//go:build !noos && !tinygo && !android && !ios && !mobile

package app

import (
	"os"
	"path/filepath"

	"github.com/Uzer007-admin/granio/v2"
)

func rootCacheDir(a granio.App) string {
	desktopCache, _ := os.UserCacheDir()
	return filepath.Join(desktopCache, "fyne", a.UniqueID())
}
