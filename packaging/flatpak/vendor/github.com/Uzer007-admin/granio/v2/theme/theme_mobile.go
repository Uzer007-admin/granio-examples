//go:build android || ios || mobile

package theme

import "github.com/Uzer007-admin/granio/v2"

func setupSystemTheme(fallback granio.Theme) granio.Theme {
	return fallback
}
