package app

import (
	"encoding/base64"
	"io"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
)

type cache struct {
	a granio.App

	enc base64.Encoding
}

func makeCache(a granio.App) granio.Cache {
	c := &cache{a: a, enc: base64.Encoding{}}

	root := c.RootURI()
	exists, err := storage.Exists(root)
	if !exists || err != nil {
		parent, _ := storage.Parent(root)
		exists, err = storage.Exists(parent)
		if !exists || err != nil {
			err = storage.CreateListable(parent)
			if err != nil {
				granio.LogError("Failed to create fyne cache space", err)
			}
		}

		err = storage.CreateListable(root)
		if err != nil {
			granio.LogError("Failed to create app cache space", err)
		}
	}

	return c
}

func (c *cache) RootURI() granio.URI {
	return storage.NewFileURI(rootCacheDir(c.a))
}

func (c *cache) Exists(name string) bool {
	path := c.encodePath(name)

	ok, err := storage.Exists(path)
	return ok && err == nil
}

func (c *cache) Read(name string) (io.ReadCloser, error) {
	path := c.encodePath(name)

	return storage.Reader(path)
}

func (c *cache) Write(name string) (io.WriteCloser, error) {
	path := c.encodePath(name)

	return storage.Writer(path)
}

func (c *cache) Remove(name string) error {
	path := c.encodePath(name)

	return storage.Delete(path)
}

func (c *cache) encodePath(badName string) granio.URI {
	name := base64.StdEncoding.EncodeToString([]byte(badName))
	child, _ := storage.Child(c.a.Cache().RootURI(), name)
	return child
}
