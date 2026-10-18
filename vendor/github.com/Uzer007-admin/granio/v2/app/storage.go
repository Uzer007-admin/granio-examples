package app

import (
	"os"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/storage"
)

type store struct {
	*internal.Docs
	a *fyneApp
}

func (s *store) RootURI() granio.URI {
	if s.a.missingID {
		granio.LogError("Storage API requires a unique ID, use app.NewWithID()", nil)
		return storage.NewFileURI(os.TempDir())
	}

	u, err := storage.ParseURI(s.a.storageRoot())
	if err == nil {
		return u
	}
	return storage.NewFileURI(s.a.storageRoot())
}

func (s *store) docRootURI() (granio.URI, error) {
	return storage.Child(s.RootURI(), "Documents")
}
