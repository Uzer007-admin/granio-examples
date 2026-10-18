package cache

import (
	"strconv"
	"sync/atomic"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/async"
)

var (
	overrides     async.Map[granio.CanvasObject, *overrideScope]
	overrideCount atomic.Uint32
)

type overrideScope struct {
	th      granio.Theme
	cacheID string
}

// OverrideTheme allows an app to specify that a single object should use a different theme to the app.
// This should be used sparingly to avoid a jarring user experience.
// If the object is a container it will theme the children, if it is a canvas primitive it will do nothing.
//
// Since: 2.5
func OverrideTheme(o granio.CanvasObject, th granio.Theme) {
	id := overrideCount.Add(1)
	s := &overrideScope{th: th, cacheID: strconv.Itoa(int(id))}
	overrideThemeWithScope(o, s)
}

func OverrideThemeMatchingScope(o, parent granio.CanvasObject) bool {
	scope, ok := overrides.Load(parent)
	if !ok { // not overridden in parent
		return false
	}

	overrideThemeWithScope(o, scope)
	return true
}

func WidgetScopeID(o granio.CanvasObject) string {
	scope, ok := overrides.Load(o)
	if !ok {
		return ""
	}

	return scope.cacheID
}

func WidgetTheme(o granio.CanvasObject) granio.Theme {
	scope, ok := overrides.Load(o)
	if !ok {
		return nil
	}

	return scope.th
}

func overrideContainer(c *granio.Container, s *overrideScope) {
	for _, o := range c.Objects {
		overrideThemeWithScope(o, s)
	}
}

func overrideThemeWithScope(o granio.CanvasObject, s *overrideScope) {
	if _, ok := o.(interface{ SetDeviceIsMobile(bool) }); ok { // ThemeOverride without the import loop
		return // do not apply this theme over a new scope
	}

	switch c := o.(type) {
	case granio.Widget:
		overrideWidget(c, s)
	case *granio.Container:
		overrideContainer(c, s)
	default:
		overrides.Store(c, s)
	}
}

func overrideWidget(w granio.Widget, s *overrideScope) {
	ResetThemeCaches()
	overrides.Store(w, s)

	r := Renderer(w)
	if r == nil {
		return
	}

	for _, o := range r.Objects() {
		overrideThemeWithScope(o, s)
	}
}
