package granio

import "image/color"

// SystemPreference distinguishes a missing OS key from an explicit zero/false.
type SystemPreference[T any] struct {
	Value     T
	Available bool
}

type SystemColorScheme uint8

const (
	SystemColorDefault SystemColorScheme = iota
	SystemColorDark
	SystemColorLight
)

// SystemAppearance is an immutable value snapshot. Availability describes the
// source; each preference has independent availability. Providers must not
// invent text scale or transparency preferences for systems lacking these keys.
// Accent is an opaque sRGB color. Generation is local to the settings provider.
// SystemThemeInfo содержит метаданные темы; стиль стороннего toolkit не исполняется.
type SystemThemeInfo struct{ Name, WidgetStyle, Font, IconTheme string }

type SystemAppearance struct {
	DesktopTheme                SystemPreference[SystemThemeInfo]
	Generation                  uint64
	Available                   bool
	ColorScheme                 SystemPreference[SystemColorScheme]
	HighContrast, ReducedMotion SystemPreference[bool]
	Accent                      SystemPreference[color.NRGBA]
	TextScale                   SystemPreference[float32]
	Transparency                SystemPreference[bool]
	Colors                      SystemPreference[SystemColors]
	UnavailableReason           string
}

// SystemColors contains the user's semantic accessibility palette. Providers
// publish the palette as a whole; colors are opaque sRGB values. In high contrast
// a theme must preserve these roles instead of replacing a user's chosen colors.
type SystemColors struct {
	Window, WindowText, ButtonFace, ButtonText   color.NRGBA
	Highlight, HighlightText, GrayText, Hotlight color.NRGBA
}

// SystemAppearanceTheme is an optional, immutable adapter for a theme that
// follows OS preferences. It retains explicit application overrides when making
// a new value. The existing Theme interface does not change.
type SystemAppearanceTheme interface {
	Theme
	WithSystemAppearance(SystemAppearance) Theme
}

// SystemColorTheme exposes an active accessibility palette to renderers that
// need paired colors or selection outlines. Ordinary themes need not implement it.
type SystemColorTheme interface {
	Theme
	SystemColors() (SystemColors, bool)
}

// SystemAppearanceProvider is optional; Settings implementations need not change
// their existing interface. Settings change listeners observe new snapshots.
// Consumers merge explicit application/user overrides with available OS keys.
type SystemAppearanceProvider interface{ SystemAppearance() SystemAppearance }
