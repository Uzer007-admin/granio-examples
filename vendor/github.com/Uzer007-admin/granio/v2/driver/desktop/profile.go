package desktop

import (
	"errors"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// ApplyModernProfile configures an existing window and application on the UI
// goroutine. Exact geometry scaling and independent typography opt in together.
// Visual backdrops remain an explicit ConfigureWindowVisuals choice. A software,
// mobile or web window without ScalableWindow still receives the portable theme.
func ApplyModernProfile(a granio.App, w granio.Window, options theme.ModernOptions) error {
	if a == nil || w == nil || a.Settings() == nil {
		return errors.New("modern profile requires an app and window")
	}
	th, err := theme.NewModern(options)
	if err != nil {
		return err
	}
	if scalable, ok := w.(ScalableWindow); ok {
		if err := scalable.SetScalePolicy(ScaleExact); err != nil {
			return err
		}
	}
	a.Settings().SetTheme(th)
	return nil
}
