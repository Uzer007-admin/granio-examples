//go:build ios || android

package dialog

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile"
	"github.com/Uzer007-admin/granio/v2/storage"
)

func (f *fileDialog) getPlaces() []favoriteItem {
	return []favoriteItem{}
}

func isHidden(file granio.URI) bool {
	if file.Scheme() != granio.URISchemeFile {
		granio.LogError("Cannot check if non file is hidden", nil)
		return false
	}
	return false
}

func hideFile(filename string) error {
	return nil
}

func fileOpenOSOverride(f *FileDialog) bool {
	if f.isDirectory() {
		mobile.ShowFolderOpenPicker(f.callback.(func(granio.ListableURI, error)))
	} else {
		mobile.ShowFileOpenPicker(f.callback.(func(granio.URIReadCloser, error)), f.filter)
	}
	return true
}

func fileSaveOSOverride(f *FileDialog) bool {
	mobile.ShowFileSavePicker(f.callback.(func(granio.URIWriteCloser, error)), f.filter, f.initialFileName)
	return true
}

func getFavoriteLocation(homeURI granio.URI, name string) (granio.URI, error) {
	return storage.Child(homeURI, name)
}
