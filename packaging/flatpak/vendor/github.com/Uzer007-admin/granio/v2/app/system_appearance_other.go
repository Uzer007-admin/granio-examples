//go:build ci || wasm || test_web_driver || android || ios || mobile || (!linux && !openbsd && !freebsd && !netbsd && !windows) || (windows && !cgo) || tinygo || noos || tamago

package app

import "github.com/Uzer007-admin/granio/v2"

func (s *settings) startSystemAppearance() {}

func (s *settings) cacheSystemVariant(granio.ThemeVariant) {}
