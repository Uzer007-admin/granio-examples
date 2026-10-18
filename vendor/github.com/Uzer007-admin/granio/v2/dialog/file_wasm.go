//go:build wasm || js

package dialog

import (
	"github.com/Uzer007-admin/granio/v2"
)

func (f *fileDialog) loadPlaces() []granio.CanvasObject {
	return nil
}

func isHidden(file granio.URI) bool {
	return false
}

func fileOpenOSOverride(f *FileDialog) bool {
	// TODO #2737
	return true
}

func fileSaveOSOverride(f *FileDialog) bool {
	// TODO #2738
	return true
}

func (f *fileDialog) getPlaces() []favoriteItem {
	return []favoriteItem{}
}

func getFavoriteLocation(homeURI granio.URI, name string) (granio.URI, error) {
	return nil, nil
}
