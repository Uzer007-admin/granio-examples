package app

import (
	"context"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// At most one pending UI callback reads the latest value. A slow UI does not
// accumulate one callback per portal snapshot, and canceled work touches no UI.
func (s *settings) dispatchLatestAppearance(ctx context.Context, queue func(func())) {
	if ctx.Err() != nil || !s.appearanceQueued.CompareAndSwap(false, true) {
		return
	}
	queue(func() {
		s.appearanceQueued.Store(false)
		if ctx.Err() != nil {
			return
		}
		if snapshot := s.appearance.Load(); snapshot != nil {
			s.applySystemAppearancePreferences(*snapshot)
		}
	})
}

// applySystemAppearance runs on the UI thread. OS preference updates do not
// replace an explicit light/dark selection or a custom theme object. Other
// preferences remain accessible even when the theme variant is overridden.
func (s *settings) applySystemAppearance(snapshot granio.SystemAppearance) {
	s.appearance.Store(&snapshot)
	s.applySystemAppearancePreferences(snapshot)
}

func (s *settings) applySystemAppearancePreferences(snapshot granio.SystemAppearance) {
	if snapshot.Available && snapshot.ColorScheme.Available && s.explicitThemeVariantName() == "" {
		variant := theme.VariantLight // preserve the existing XDG no-preference fallback
		if snapshot.ColorScheme.Value == granio.SystemColorDark {
			variant = theme.VariantDark
		}
		s.variant = variant
		s.cacheSystemVariant(variant)
	}
	if themed, ok := s.theme.(granio.SystemAppearanceTheme); ok {
		s.theme = themed.WithSystemAppearance(snapshot)
	}
	s.apply()
}

func (s *settings) queueSystemAppearance(ctx context.Context, snapshot granio.SystemAppearance) {
	// Native callbacks publish values only. The UI reads the newest value once,
	// even if hundreds of OS events arrive while a layout is being performed.
	if previous := s.appearance.Load(); previous != nil && snapshot.Generation <= previous.Generation {
		snapshot.Generation = previous.Generation + 1
	}
	s.appearance.Store(&snapshot)
	app := granio.CurrentApp()
	if app == nil || app.Settings() != s {
		return
	}
	s.dispatchLatestAppearance(ctx, func(apply func()) {
		granio.Do(func() {
			app := granio.CurrentApp()
			if app == nil || app.Settings() != s {
				s.appearanceQueued.Store(false)
				return
			}
			apply()
		})
	})
}

func (s *settings) stopSystemAppearance() {
	s.appearanceMu.Lock()
	s.appearanceStopped = true
	cancel, done := s.appearanceCancel, s.appearanceDone
	s.appearanceMu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}
