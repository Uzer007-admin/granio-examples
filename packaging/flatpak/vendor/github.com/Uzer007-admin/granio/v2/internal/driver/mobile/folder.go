package mobile

import (
	"errors"

	"github.com/Uzer007-admin/granio/v2"
)

type lister struct {
	granio.URI
}

func (l *lister) List() ([]granio.URI, error) {
	return listURI(l)
}

func listerForURI(uri granio.URI) (granio.ListableURI, error) {
	if !canListURI(uri) {
		return nil, errors.New("specified URI is not listable")
	}

	return &lister{uri}, nil
}
