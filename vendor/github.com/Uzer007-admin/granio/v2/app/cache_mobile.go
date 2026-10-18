//go:build android || ios || mobile

package app

import "github.com/Uzer007-admin/granio/v2"

func rootCacheDir(a granio.App) string {
	return a.(*fyneApp).storageRoot()
}
