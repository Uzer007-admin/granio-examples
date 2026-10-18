package theme

import (
	"github.com/Uzer007-admin/granio/v2"
)

var themeStack []granio.Theme

// CurrentlyRenderingWithFallback returns the theme that is currently being used during rendering or layout
// calculations. If there is no override in effect then the fallback is returned.
func CurrentlyRenderingWithFallback(f granio.Theme) granio.Theme {
	if len(themeStack) == 0 {
		return f
	}

	return themeStack[len(themeStack)-1]
}

// PushRenderingTheme is used by the ThemeOverride container to stack the current theme during rendering
// and calculations.
func PushRenderingTheme(th granio.Theme) {
	themeStack = append(themeStack, th)
}

// PopRenderingTheme is used by the ThemeOverride container to remove an overridden theme during rendering
// and calculations.
func PopRenderingTheme() {
	themeStack[len(themeStack)-1] = nil
	themeStack = themeStack[:len(themeStack)-1]
}
