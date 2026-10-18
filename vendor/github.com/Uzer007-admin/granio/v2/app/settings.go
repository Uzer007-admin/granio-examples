package app

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/app"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// SettingsSchema is used for loading and storing global settings
type SettingsSchema struct {
	// these items are used for global settings load
	ThemeName         string  `json:"theme"`
	Scale             float32 `json:"scale"`
	PrimaryColor      string  `json:"primary_color"`
	CloudName         string  `json:"cloud_name"`
	CloudConfig       string  `json:"cloud_config"`
	DisableAnimations bool    `json:"no_animations"`
}

// StoragePath returns the location of the settings storage
func (*SettingsSchema) StoragePath() string {
	return filepath.Join(app.RootConfigDir(), "settings.json")
}

// Declare conformity with Settings interface
var _ granio.Settings = (*settings)(nil)

const (
	themeVariantNameDark  = "dark"
	themeVariantNameLight = "light"
)

type settings struct {
	theme          granio.Theme
	themeSpecified bool
	variant        granio.ThemeVariant

	listeners         []func(granio.Settings)
	changeListeners   async.Map[chan granio.Settings, bool]
	watcher           any // normally *fsnotify.Watcher or nil - avoid import in this file
	appearance        atomic.Pointer[granio.SystemAppearance]
	appearanceQueued  atomic.Bool
	appearanceMu      sync.Mutex
	appearanceCancel  func()
	appearanceDone    <-chan struct{}
	appearanceStopped bool

	schema SettingsSchema
}

func (*settings) BuildType() granio.BuildType {
	return build.Mode
}

func (s *settings) PrimaryColor() string {
	return s.schema.PrimaryColor
}

// OverrideTheme allows the settings app to temporarily preview different theme details.
// Please make sure that you remember the original settings and call this again to revert the change.
//
// Deprecated: Use container.NewThemeOverride to change the appearance of part of your application.
func (s *settings) OverrideTheme(t granio.Theme, name string) {
	s.schema.PrimaryColor = name
	s.theme = t
}

func (s *settings) Theme() granio.Theme {
	if s == nil {
		granio.LogError("Attempt to access current Fyne theme when no app is started", nil)
		return nil
	}
	return s.theme
}

func (s *settings) SetTheme(t granio.Theme) {
	s.themeSpecified = true
	s.applyTheme(t, s.variant)
}

func (s *settings) ShowAnimations() bool {
	appearance := s.SystemAppearance()
	local := false
	if policy, ok := s.theme.(granio.MotionTheme); ok {
		local = policy.Motion(granio.MotionHover).Disabled
	}
	return !local && !s.schema.DisableAnimations && !build.NoAnimations && !(appearance.Available && appearance.ReducedMotion.Available && appearance.ReducedMotion.Value)
}

func (s *settings) SystemAppearance() granio.SystemAppearance {
	if snapshot := s.appearance.Load(); snapshot != nil {
		return *snapshot
	}
	return granio.SystemAppearance{}
}

func (s *settings) ThemeVariant() granio.ThemeVariant {
	return s.variant
}

func (s *settings) applyTheme(t granio.Theme, variant granio.ThemeVariant) {
	s.variant = variant
	if themed, ok := t.(granio.SystemAppearanceTheme); ok {
		t = themed.WithSystemAppearance(s.SystemAppearance())
	}
	s.theme = t
	s.apply()
}

func (s *settings) Scale() float32 {
	if s.schema.Scale < 0.0 {
		return 1.0 // catching any really old data still using the `-1`  value for "auto" scale
	}
	return s.schema.Scale
}

func (s *settings) AddChangeListener(listener chan granio.Settings) {
	s.changeListeners.Store(listener, true) // the boolean is just a dummy value here.
}

func (s *settings) AddListener(listener func(granio.Settings)) {
	s.listeners = append(s.listeners, listener)
}

func (s *settings) apply() {
	s.changeListeners.Range(func(listener chan granio.Settings, _ bool) bool {
		select {
		case listener <- s:
		default:
			l := listener
			go func() { l <- s }()
		}
		return true
	})

	for _, l := range s.listeners {
		l(s)
	}
}

func (s *settings) fileChanged() {
	s.load()
	s.apply()
}

func (s *settings) setupTheme() {
	name := s.explicitThemeVariantName()
	variant := app.DefaultVariant()
	effectiveTheme := s.theme
	if !s.themeSpecified {
		effectiveTheme = theme.DefaultTheme()
	}
	switch name {
	case themeVariantNameLight:
		variant = theme.VariantLight
	case themeVariantNameDark:
		variant = theme.VariantDark
	}

	s.applyTheme(effectiveTheme, variant)
}

func (s *settings) explicitThemeVariantName() string {
	if name := os.Getenv("FYNE_THEME"); name != "" {
		return name
	}

	return s.schema.ThemeName
}

func loadSettings() *settings {
	s := &settings{}
	s.load()

	return s
}
