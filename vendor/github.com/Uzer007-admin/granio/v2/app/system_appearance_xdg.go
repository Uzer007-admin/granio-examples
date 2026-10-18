//go:build !ci && !wasm && !test_web_driver && !android && !ios && !mobile && (linux || openbsd || freebsd || netbsd) && !tinygo && !noos && !tamago

package app

import (
	"context"
	"image/color"
	"math"
	"time"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/dbus/v5/portal"
	"github.com/Uzer007-admin/granio/v2"
	internalapp "github.com/Uzer007-admin/granio/v2/internal/app"
)

func portalAppearance(snapshot portal.AppearanceSnapshot) granio.SystemAppearance {
	result := granio.SystemAppearance{Generation: snapshot.Generation, Available: snapshot.Available}
	if snapshot.Err != nil {
		result.UnavailableReason = snapshot.Err.Error()
	}
	if !snapshot.Available {
		return result
	}
	a := snapshot.Appearance
	result.ColorScheme = granio.SystemPreference[granio.SystemColorScheme]{Value: granio.SystemColorScheme(a.ColorScheme.Value), Available: a.ColorScheme.Available}
	result.HighContrast = granio.SystemPreference[bool]{Value: a.HighContrast.Value, Available: a.HighContrast.Available}
	result.ReducedMotion = granio.SystemPreference[bool]{Value: a.ReducedMotion.Value, Available: a.ReducedMotion.Available}
	if a.Accent.Available {
		byteValue := func(v float64) uint8 { return uint8(math.Round(min(1, max(0, v)) * 255)) }
		result.Accent = granio.SystemPreference[color.NRGBA]{Value: color.NRGBA{R: byteValue(a.Accent.Value.Red), G: byteValue(a.Accent.Value.Green), B: byteValue(a.Accent.Value.Blue), A: 255}, Available: true}
	}
	desktop := a.Desktop
	if desktop.Name.Available || desktop.WidgetStyle.Available || desktop.Font.Available || desktop.IconTheme.Available {
		result.DesktopTheme = granio.SystemPreference[granio.SystemThemeInfo]{Available: true, Value: granio.SystemThemeInfo{Name: desktop.Name.Value, WidgetStyle: desktop.WidgetStyle.Value, Font: desktop.Font.Value, IconTheme: desktop.IconTheme.Value}}
	}
	if desktop.Colors.Available {
		c := desktop.Colors.Value
		result.Colors = granio.SystemPreference[granio.SystemColors]{Available: true, Value: granio.SystemColors{Window: c.Window, WindowText: c.WindowText, ButtonFace: c.ButtonFace, ButtonText: c.ButtonText, Highlight: c.Highlight, HighlightText: c.HighlightText, GrayText: c.GrayText, Hotlight: c.Hotlight}}
	}
	return result // TextScale и разрешение прозрачности остаются недоступными.
}

func (s *settings) cacheSystemVariant(variant granio.ThemeVariant) {
	app := granio.CurrentApp()
	if app != nil && app.Settings() == s {
		internalapp.CurrentVariant.Store(uint64(variant))
	}
}

func (s *settings) startSystemAppearance() {
	s.appearanceMu.Lock()
	if s.appearanceCancel != nil || s.appearanceStopped {
		s.appearanceMu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	initial := make(chan granio.SystemAppearance, 1)
	ready := make(chan struct{})
	s.appearanceCancel, s.appearanceDone = cancel, done
	s.appearanceMu.Unlock()
	go func() {
		defer close(done)
		first := true
		publish := func(snapshot granio.SystemAppearance) {
			if first {
				first = false
				initial <- snapshot
				select {
				case <-ready:
				case <-ctx.Done():
					return
				}
				// Startup may already have published an explicit timeout fallback.
				if current := s.SystemAppearance(); current.UnavailableReason == "portal startup deadline exceeded" {
					s.queueSystemAppearance(ctx, snapshot)
				}
			} else {
				s.queueSystemAppearance(ctx, snapshot)
			}
		}
		// This one connection belongs to the app provider; notification/tray
		// connections and other consumers remain untouched when the app closes.
		conn, err := dbus.ConnectSessionBus(dbus.WithContext(ctx))
		if err != nil {
			publish(granio.SystemAppearance{Generation: 1, UnavailableReason: err.Error()})
			return
		}
		defer conn.Close()
		watcher, err := portal.WatchAppearance(ctx, conn, portal.WatchOptions{ReadTimeout: 500 * time.Millisecond, IncludeDesktopSettings: true})
		if err != nil {
			publish(granio.SystemAppearance{Generation: 1, UnavailableReason: err.Error()})
			return
		}
		defer watcher.Close()
		defer func() {
			previous := s.SystemAppearance()
			closed := granio.SystemAppearance{Generation: previous.Generation + 1, UnavailableReason: "system appearance provider closed"}
			if ctx.Err() != nil {
				s.appearance.Store(&closed)
			} else if previous.Available {
				s.queueSystemAppearance(ctx, closed)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case snapshot, open := <-watcher.Updates():
				if !open {
					return
				}
				publish(portalAppearance(snapshot))
			}
		}
	}()
	// A complete initial read (or explicit unavailable value) precedes Show.
	// A stalled portal cannot hold the application's startup indefinitely.
	timer := time.NewTimer(600 * time.Millisecond)
	defer timer.Stop()
	select {
	case snapshot := <-initial:
		if ctx.Err() == nil {
			s.applySystemAppearance(snapshot)
		}
	case <-timer.C:
		s.applySystemAppearance(granio.SystemAppearance{Generation: 1, UnavailableReason: "portal startup deadline exceeded"})
	case <-ctx.Done():
	}
	close(ready)
}
