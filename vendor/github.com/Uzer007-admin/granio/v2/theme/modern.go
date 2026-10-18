package theme

import (
	"fmt"
	"image/color"
	"math"

	"github.com/Uzer007-admin/granio/v2"
)

// Density describes spacing rather than a platform-specific set of controls.
type Density uint8

const (
	DensityComfortable Density = iota
	DensityCompact
)

// ModernOptions configures a portable, opaque visual profile. TextScale changes
// typography independently of icon size and window/framebuffer scaling.
type ModernOptions struct {
	// TokenVersion zero selects the current token specification.
	TokenVersion uint16
	Density      Density
	TextScale    float32
	Accent       color.Color
	HighContrast bool
	// FollowSystem opts into OS accent and text scale.
	// A nil Accent and zero TextScale follow available OS values; explicit values
	// stay fixed. OS high contrast takes priority even when FollowSystem is false.
	FollowSystem bool
	// FollowSystemPalette включает полную доступную палитру рабочего стола.
	FollowSystemPalette bool
	DisableMotion       bool
	FontPolicy          FontPolicy
}
type modernTheme struct {
	base                 granio.Theme
	options              ModernOptions
	accent, onAccent     color.Color
	textScale            float32
	highContrast         bool
	colors               granio.SystemPreference[granio.SystemColors]
	fonts                [4]granio.Resource
	transparencyDisabled bool
}

// NewModern creates an opt-in theme for all Fyne renderers. It preserves the
// existing Theme contract and uses Fyne's font fallback and vector icon resources.
func NewModern(options ModernOptions) (granio.Theme, error) {
	if options.TokenVersion != 0 && options.TokenVersion != ModernTokenVersion {
		return nil, fmt.Errorf("unsupported modern token version %d", options.TokenVersion)
	}
	if options.FontPolicy > FontSystem {
		return nil, fmt.Errorf("invalid font policy %d", options.FontPolicy)
	}
	if options.Density > DensityCompact {
		return nil, fmt.Errorf("invalid density %d", options.Density)
	}
	if options.TextScale < 0 || options.TextScale > 4 || math.IsNaN(float64(options.TextScale)) || math.IsInf(float64(options.TextScale), 0) {
		return nil, fmt.Errorf("text scale must be finite and between 0 and 4")
	}
	t := &modernTheme{base: DefaultTheme(), options: options}
	if options.FontPolicy == FontSystem {
		t.fonts = loadSystemFonts()
	}
	return t.WithSystemAppearance(granio.SystemAppearance{}), nil
}

// WithSystemAppearance returns a theme value without mutating the original.
// Losing a provider or leaving high contrast restores the application's options.
func (t *modernTheme) WithSystemAppearance(snapshot granio.SystemAppearance) granio.Theme {
	next := *t
	next.textScale = t.options.TextScale
	if next.textScale == 0 {
		next.textScale = 1
		if t.options.FollowSystem && snapshot.Available && snapshot.TextScale.Available {
			v := snapshot.TextScale.Value
			if v > 0 && v <= 4 && !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) {
				next.textScale = v
			}
		}
	}
	accent := t.options.Accent
	if accent == nil {
		accent = color.NRGBA{R: 0x14, G: 0x57, B: 0xb7, A: 255}
		if t.options.FollowSystem && snapshot.Available && snapshot.Accent.Available {
			accent = snapshot.Accent.Value
		}
	}
	next.accent, next.onAccent = modernAccent(accent)
	next.highContrast = t.options.HighContrast
	next.transparencyDisabled = snapshot.Available && snapshot.Transparency.Available && !snapshot.Transparency.Value
	next.colors = granio.SystemPreference[granio.SystemColors]{}
	if snapshot.Available {
		next.highContrast = next.highContrast || snapshot.HighContrast.Available && snapshot.HighContrast.Value
		if t.options.FollowSystemPalette || snapshot.HighContrast.Available && snapshot.HighContrast.Value {
			next.colors = snapshot.Colors
			if next.colors.Available && !next.highContrast && t.options.Accent != nil {
				// Явный акцент приложения применяется также к выделению и фокусу.
				next.colors.Value.Highlight = color.NRGBAModel.Convert(next.accent).(color.NRGBA)
				next.colors.Value.HighlightText = color.NRGBAModel.Convert(next.onAccent).(color.NRGBA)
			}
		}
	}
	return &next
}
func (t *modernTheme) Font(style granio.TextStyle) granio.Resource {
	if !style.Monospace && !style.Symbol {
		i := 0
		if style.Bold {
			i++
		}
		if style.Italic {
			i += 2
		}
		if font := t.fonts[i]; font != nil {
			return font
		}
	}
	return t.base.Font(style)
}

// HighContrast allows optional effects to follow an explicit application theme.
func (t *modernTheme) HighContrast() bool { return t.highContrast }
func (t *modernTheme) SystemColors() (granio.SystemColors, bool) {
	return t.colors.Value, (t.highContrast || t.options.FollowSystemPalette) && t.colors.Available
}
func (t *modernTheme) Icon(name granio.ThemeIconName) granio.Resource { return t.base.Icon(name) }
func (t *modernTheme) Size(name granio.ThemeSizeName) float32 {
	var size float32
	switch name {
	case SizeNameText:
		size = 14
	case SizeNameCaptionText:
		size = 12
	case SizeNameHeadingText:
		size = 28
	case SizeNameSubHeadingText:
		size = 20
	case SizeNameLineSpacing:
		size = 4
	}
	if size > 0 {
		return size * t.textScale
	}
	switch name {
	case SizeNameInnerPadding:
		if t.options.Density == DensityCompact {
			return 6
		}
		return 10
	case SizeNamePadding:
		return 8
	case SizeNameInputRadius, SizeNameButtonRadius:
		return 4
	case SizeNameInnerWindowRadius, SizeNameMenuRadius, SizeNamePopupRadius, SizeNameCardRadius:
		return 8
	case SizeNameInputBorder:
		if t.highContrast {
			return 2
		}
		return 1
	case SizeNameModalBlurRadius:
		if t.highContrast || t.transparencyDisabled {
			return 0
		}
		return t.base.Size(name)
	}
	if size := t.base.Size(name); size > 0 {
		return size
	}
	return 14 * t.textScale
}
func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	linear := func(v uint32) float64 {
		x := float64(v) / 65535
		if x <= 0.04045 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*linear(r) + 0.7152*linear(g) + 0.0722*linear(b)
}
func modernAccent(c color.Color) (color.Color, color.Color) {
	r, g, b, _ := c.RGBA()
	opaque := color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255}
	light := (luminance(opaque) + 0.05) / 0.05
	dark := 1.05 / (luminance(opaque) + 0.05)
	if light >= dark {
		return opaque, color.Black
	}
	return opaque, color.White
}
func (t *modernTheme) Color(name granio.ThemeColorName, variant granio.ThemeVariant) color.Color {
	dark := variant == VariantDark
	background, surface, foreground, secondary, border := color.NRGBA{R: 245, G: 246, B: 249, A: 255}, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, color.NRGBA{R: 28, G: 31, B: 38, A: 255}, color.NRGBA{R: 86, G: 91, B: 102, A: 255}, color.NRGBA{R: 121, G: 127, B: 140, A: 255}
	if dark {
		background = color.NRGBA{R: 26, G: 29, B: 36, A: 255}
		surface = color.NRGBA{R: 37, G: 41, B: 50, A: 255}
		foreground = color.NRGBA{R: 243, G: 245, B: 249, A: 255}
		secondary = color.NRGBA{R: 183, G: 189, B: 201, A: 255}
		border = color.NRGBA{R: 132, G: 142, B: 160, A: 255}
	}
	accent, onAccent := t.accent, t.onAccent
	if t.options.FollowSystemPalette && t.colors.Available {
		c := t.colors.Value
		background, surface, foreground = c.Window, c.ButtonFace, c.WindowText
		secondary = c.GrayText
		if t.options.Accent == nil {
			accent, onAccent = c.Highlight, c.HighlightText
		}
	}
	if t.highContrast {
		if dark {
			background = color.NRGBA{A: 255}
			surface = background
			foreground = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			accent = color.NRGBA{R: 255, G: 230, B: 0, A: 255}
			onAccent = color.Black
		} else {
			background = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
			surface = background
			foreground = color.NRGBA{A: 255}
			accent = color.NRGBA{R: 0, G: 45, B: 130, A: 255}
			onAccent = color.White
		}
		secondary, border = foreground, foreground
		if t.colors.Available {
			c := t.colors.Value
			background, surface, foreground = c.Window, c.ButtonFace, c.WindowText
			secondary, border = c.GrayText, c.WindowText
			accent, onAccent = c.Highlight, c.HighlightText
		}
	}
	switch name {
	case ColorNameBackground:
		return background
	case ColorNameButton, ColorNameHeaderBackground:
		return surface
	case ColorNameInputBackground, ColorNameMenuBackground:
		if t.highContrast && t.colors.Available {
			return background
		}
		return surface
	case ColorNameForeground:
		return foreground
	case ColorNameForegroundOnButton:
		if (t.highContrast || t.options.FollowSystemPalette) && t.colors.Available {
			return t.colors.Value.ButtonText
		}
		return foreground
	case ColorNameDisabledButton:
		return surface
	case ColorNamePlaceHolder, ColorNameDisabled:
		return secondary
	case ColorNameInputBorder, ColorNameSeparator, ColorNameInnerWindowBorder:
		return border
	case ColorNamePrimary:
		return accent
	case ColorNameForegroundOnPrimary:
		return onAccent
	case ColorNameError, ColorNameSuccess, ColorNameWarning:
		if t.highContrast {
			return accent
		}
		if dark {
			switch name {
			case ColorNameError:
				return color.NRGBA{R: 255, G: 180, B: 171, A: 255}
			case ColorNameSuccess:
				return color.NRGBA{R: 126, G: 224, B: 166, A: 255}
			default:
				return color.NRGBA{R: 247, G: 213, B: 106, A: 255}
			}
		}
		switch name {
		case ColorNameError:
			return color.NRGBA{R: 176, G: 29, B: 35, A: 255}
		case ColorNameSuccess:
			return color.NRGBA{R: 19, G: 106, B: 50, A: 255}
		default:
			return color.NRGBA{R: 133, G: 83, A: 255}
		}
	case ColorNameForegroundOnError, ColorNameForegroundOnSuccess, ColorNameForegroundOnWarning:
		if t.highContrast {
			return onAccent
		}
		if dark {
			return color.Black
		}
		return color.White
	case ColorNameFocus, ColorNameHyperlink:
		if t.highContrast {
			if t.colors.Available && name == ColorNameHyperlink {
				return t.colors.Value.Hotlight
			}
			return accent
		}
		if dark {
			return color.NRGBA{R: 125, G: 184, B: 255, A: 255}
		}
		return color.NRGBA{R: 20, G: 87, B: 183, A: 255}
	case ColorNameSelection:
		if t.highContrast {
			if t.colors.Available {
				return t.colors.Value.Highlight
			}
			if dark {
				return color.NRGBA{R: 100, G: 100, B: 100, A: 255}
			}
			return color.NRGBA{R: 128, G: 128, B: 128, A: 255}
		}
		if dark {
			return color.NRGBA{R: 54, G: 70, B: 95, A: 255}
		}
		return color.NRGBA{R: 213, G: 229, B: 252, A: 255}
	case ColorNameHover:
		if t.highContrast {
			return color.Transparent
		}
		if dark {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 20}
		}
		return color.NRGBA{A: 16}
	case ColorNamePressed:
		if t.highContrast {
			return color.Transparent
		}
		if dark {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 32}
		}
		return color.NRGBA{A: 25}
	case ColorNameShadow, ColorNameOverlayBackground:
		if t.highContrast {
			return color.Transparent
		}
	}
	if c := t.base.Color(name, variant); c != nil {
		return c
	}
	return foreground
}
