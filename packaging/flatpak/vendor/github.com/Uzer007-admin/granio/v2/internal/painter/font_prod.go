//go:build !ci && !test

package painter

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-text/typesetting/fontscan"

	"github.com/Uzer007-admin/granio/v2/internal/goos"
)

func loadSystemFonts(fm *fontscan.FontMap) error {
	cacheDir := ""
	if runtime.GOOS == goos.Android {
		parent := os.Getenv("FILESDIR")
		cacheDir = filepath.Join(parent, "fontcache")
	}

	return fm.UseSystemFonts(cacheDir)
}
