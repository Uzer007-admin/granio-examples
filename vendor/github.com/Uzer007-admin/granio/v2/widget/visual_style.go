package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func componentStyle(th granio.Theme, kind granio.ComponentKind, state granio.VisualState) (granio.ComponentStyle, bool) {
	provider, ok := th.(granio.ComponentTheme)
	if !ok {
		return granio.ComponentStyle{}, false
	}
	return provider.ComponentStyle(kind, state, granio.CurrentApp().Settings().ThemeVariant()), true
}

func minimumComponentHeight(size granio.Size, th granio.Theme, kind granio.ComponentKind) granio.Size {
	if style, ok := componentStyle(th, kind, 0); ok {
		size.Height = max(size.Height, style.MinimumHeight)
	}
	return size
}

func animationsAllowed(th granio.Theme, role granio.MotionRole) bool {
	return granio.CurrentApp().Settings().ShowAnimations() && !theme.Motion(th, role).Disabled
}
