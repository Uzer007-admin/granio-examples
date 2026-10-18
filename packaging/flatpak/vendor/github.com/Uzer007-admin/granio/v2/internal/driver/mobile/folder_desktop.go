//go:build !ios && !android

package mobile

import (
	"github.com/Uzer007-admin/granio/v2"
)

func canListURI(granio.URI) bool {
	// no-op as we use the internal FileRepository
	return false
}

func createListableURI(granio.URI) error {
	// no-op as we use the internal FileRepository
	return nil
}

func listURI(granio.URI) ([]granio.URI, error) {
	// no-op as we use the internal FileRepository
	return nil, nil
}
