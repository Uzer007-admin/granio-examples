package storage

import (
	"io"

	"github.com/Uzer007-admin/granio/v2"
)

// LoadResourceFromURI creates a new StaticResource in memory using the contents of the specified URI.
// The URI will be opened using the current driver, so valid schemas will vary from platform to platform.
// The file:// schema will always work.
func LoadResourceFromURI(u granio.URI) (granio.Resource, error) {
	read, err := Reader(u)
	if err != nil {
		return nil, err
	}

	defer read.Close()
	bytes, err := io.ReadAll(read)
	if err != nil {
		return nil, err
	}

	return granio.NewStaticResource(u.Name(), bytes), nil
}
