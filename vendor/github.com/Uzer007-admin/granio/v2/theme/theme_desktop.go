//go:build !android && !ios && !mobile && !wasm && !test_web_driver

package theme

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/Uzer007-admin/granio/v2"
	internalApp "github.com/Uzer007-admin/granio/v2/internal/app"
)

func setupSystemTheme(fallback granio.Theme) granio.Theme {
	path := filepath.Join(internalApp.RootConfigDir(), "theme.json")
	f, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			granio.LogError("Failed to load user theme file: "+path, err)
		}
		return nil
	}
	defer f.Close()

	th, err := FromJSONReaderWithFallback(bufio.NewReader(f), fallback)
	if err != nil {
		granio.LogError("Failed to parse user theme file: "+path, err)
		return nil
	}

	return th
}
