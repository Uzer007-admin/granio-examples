//go:build !windows && !android && !ios && !wasm && !js

package dialog

import (
	"path/filepath"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func (*fileDialog) getPlaces() []favoriteItem {
	lister, err := storage.ListerForURI(storage.NewFileURI("/"))
	if err != nil {
		granio.LogError("could not create lister for /", err)
		return []favoriteItem{}
	}
	return []favoriteItem{{
		"Computer",
		theme.ComputerIcon(),
		lister,
	}}
}

func isHidden(file granio.URI) bool {
	if file.Scheme() != granio.URISchemeFile {
		granio.LogError("Cannot check if non file is hidden", nil)
		return false
	}

	name := filepath.Base(file.Path())
	return name == "" || name[0] == '.'
}

func hideFile(_ string) error {
	return nil
}
