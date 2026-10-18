// Package binding provides extended sources of data binding.
package binding

import (
	"io"

	"github.com/Uzer007-admin/granio/v2/data/binding"
)

// StringCloser is an extension of the String interface that allows resources to be freed
// using the standard `Close()` method.
type StringCloser interface {
	binding.String
	io.Closer
}
