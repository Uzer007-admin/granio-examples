package test

import (
	"os"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/storage"
)

type testStorage struct {
	*internal.Docs
}

func (*testStorage) RootURI() granio.URI {
	return storage.NewFileURI(os.TempDir())
}

func (s *testStorage) docRootURI() (granio.URI, error) {
	return storage.Child(s.RootURI(), "Documents")
}
