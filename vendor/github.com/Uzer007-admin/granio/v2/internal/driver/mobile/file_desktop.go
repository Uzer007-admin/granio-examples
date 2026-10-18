//go:build !ios && !android

package mobile

import (
	"io"

	"github.com/Uzer007-admin/granio/v2"
	intRepo "github.com/Uzer007-admin/granio/v2/internal/repository"
	"github.com/Uzer007-admin/granio/v2/storage/repository"
)

func deleteURI(_ granio.URI) error {
	// no-op as we use the internal FileRepository
	return nil
}

func existsURI(granio.URI) (bool, error) {
	// no-op as we use the internal FileRepository
	return false, nil
}

func nativeFileOpen(*fileOpen) (io.ReadCloser, error) {
	// no-op as we use the internal FileRepository
	return nil, nil
}

func nativeFileSave(*fileSave, bool) (io.WriteCloser, error) {
	// no-op as we use the internal FileRepository
	return nil, nil
}

func registerRepository(_ *driver) {
	repository.Register(granio.URISchemeFile, intRepo.NewFileRepository())
}
