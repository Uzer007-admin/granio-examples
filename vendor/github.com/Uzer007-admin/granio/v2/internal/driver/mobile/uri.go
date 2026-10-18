//go:build !android

package mobile

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
)

func nativeURI(path string) granio.URI {
	uri, err := storage.ParseURI(path)
	if err != nil {
		granio.LogError("Error on parsing uri", err)
	}
	return uri
}
