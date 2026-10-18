package storage

import (
	"github.com/Uzer007-admin/granio/v2/storage/repository"
)

// URIRootError is a wrapper for repository.URIRootError
//
// Deprecated - use repository.ErrURIRoot instead
var URIRootError = repository.ErrURIRoot
