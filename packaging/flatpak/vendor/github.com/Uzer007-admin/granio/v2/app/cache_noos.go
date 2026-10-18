//go:build noos || tinygo

package app

import (
	"os"
	"path/filepath"

	"github.com/Uzer007-admin/granio/v2"
)

func rootCacheDir(_ granio.App) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fyne")
}
