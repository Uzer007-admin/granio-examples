//go:build !ios

package dialog

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
)

const folderVideos = "Movies"

func getFavoriteLocation(homeURI granio.URI, name string) (granio.URI, error) {
	return storage.Child(homeURI, name)
}
