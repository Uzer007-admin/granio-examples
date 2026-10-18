package mobile

import (
	"github.com/Uzer007-admin/granio/v2"

	"github.com/Uzer007-admin/granio/v2/storage/repository"
)

// declare conformance with repository types
var (
	_ repository.Repository             = (*mobileFileRepo)(nil)
	_ repository.HierarchicalRepository = (*mobileFileRepo)(nil)
	_ repository.ListableRepository     = (*mobileFileRepo)(nil)
	_ repository.WritableRepository     = (*mobileFileRepo)(nil)
	_ repository.AppendableRepository   = (*mobileFileRepo)(nil)
)

type mobileFileRepo struct{}

func (*mobileFileRepo) CanList(u granio.URI) (bool, error) {
	return canListURI(u), nil
}

func (*mobileFileRepo) CanRead(granio.URI) (bool, error) {
	return true, nil // TODO check a file can be read
}

func (*mobileFileRepo) CanWrite(granio.URI) (bool, error) {
	return true, nil // TODO check a file can be written
}

func (*mobileFileRepo) Child(u granio.URI, name string) (granio.URI, error) {
	if u == nil || u.Scheme() != granio.URISchemeFile {
		return nil, repository.ErrOperationNotSupported
	}

	return repository.GenericChild(u, name)
}

func (*mobileFileRepo) CreateListable(u granio.URI) error {
	return createListableURI(u)
}

func (*mobileFileRepo) Delete(u granio.URI) error {
	return deleteURI(u)
}

func (*mobileFileRepo) Destroy(string) {
}

func (*mobileFileRepo) Exists(u granio.URI) (bool, error) {
	return existsURI(u)
}

func (*mobileFileRepo) List(u granio.URI) ([]granio.URI, error) {
	return listURI(u)
}

func (*mobileFileRepo) Parent(u granio.URI) (granio.URI, error) {
	if u == nil || u.Scheme() != granio.URISchemeFile {
		return nil, repository.ErrOperationNotSupported
	}

	return repository.GenericParent(u)
}

func (*mobileFileRepo) Reader(u granio.URI) (granio.URIReadCloser, error) {
	return fileReaderForURI(u)
}

func (*mobileFileRepo) Writer(u granio.URI) (granio.URIWriteCloser, error) {
	return fileWriterForURI(u, true)
}

func (*mobileFileRepo) Appender(u granio.URI) (granio.URIWriteCloser, error) {
	return fileWriterForURI(u, false)
}
