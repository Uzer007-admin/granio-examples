package theme

import (
	"image/color"
	"math"
	"time"

	"github.com/Uzer007-admin/granio/v2"
)

const ModernTokenVersion uint16 = 1

func validVisualSize(v float32) bool {
	return v > 0 && !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0)
}

// Typography resolves a role for both measuring and painting. Missing optional
// providers use existing Theme sizes/fonts. Invalid roles fall back to body.
func Typography(th granio.Theme, role granio.TextRole, explicit granio.TextStyle) granio.TypographyStyle {
	if th == nil {
		th = Current()
	}
	name, fallback := SizeNameText, float32(14)
	switch role {
	case granio.TextRoleCaption:
		name, fallback = SizeNameCaptionText, 12
	case granio.TextRoleEmphasis:
		fallback = 16
	case granio.TextRoleSubtitle:
		name, fallback = SizeNameSubHeadingText, 20
	case granio.TextRoleTitle:
		name, fallback = SizeNameHeadingText, 28
	}
	result := granio.TypographyStyle{Size: th.Size(name), LineSpacing: th.Size(SizeNameLineSpacing)}
	if provider, ok := th.(granio.TypographyTheme); ok {
		result = provider.Typography(role)
	} else {
		switch role {
		case granio.TextRoleEmphasis:
			result.Size, result.Style.Bold = max(result.Size, fallback), true
		case granio.TextRoleTitle, granio.TextRoleSubtitle:
			result.Style.Bold = true
		}
	}
	if !validVisualSize(result.Size) {
		result.Size = fallback
	}
	if result.LineSpacing < 0 || math.IsNaN(float64(result.LineSpacing)) || math.IsInf(float64(result.LineSpacing), 0) {
		result.LineSpacing = 4
	}
	providedStyle := result.Style
	result.Style.Bold = result.Style.Bold || explicit.Bold
	result.Style.Italic = result.Style.Italic || explicit.Italic
	result.Style.Monospace = result.Style.Monospace || explicit.Monospace || role == granio.TextRoleCode
	result.Style.Symbol = result.Style.Symbol || explicit.Symbol
	result.Style.Underline = result.Style.Underline || explicit.Underline
	result.Style.Strikethrough = result.Style.Strikethrough || explicit.Strikethrough
	if explicit.TabWidth != 0 {
		result.Style.TabWidth = explicit.TabWidth
	}
	// Font and style come from one resolution; a custom face supplied for an
	// unrelated weight cannot silently defeat an explicit Bold/Monospace request.
	if result.Font == nil || result.Style != providedStyle {
		result.Font = th.Font(result.Style)
	}
	if result.Font == nil {
		result.Font = DefaultTheme().Font(result.Style)
	}
	return result
}

func Motion(th granio.Theme, role granio.MotionRole) granio.MotionSpec {
	result := granio.MotionSpec{Duration: 100 * time.Millisecond, Curve: granio.AnimationEaseOut}
	if provider, ok := th.(granio.MotionTheme); ok {
		result = provider.Motion(role)
	}
	if result.Duration <= 0 || result.Duration > time.Second {
		result.Duration = 100 * time.Millisecond
	}
	if result.Curve == nil {
		result.Curve = granio.AnimationEaseOut
	}
	return result
}

func (t *modernTheme) TokenVersion() uint16 { return ModernTokenVersion }
func (t *modernTheme) Typography(role granio.TextRole) granio.TypographyStyle {
	size, style := float32(14), granio.TextStyle{}
	switch role {
	case granio.TextRoleCaption:
		size = 12
	case granio.TextRoleEmphasis:
		size, style.Bold = 16, true
	case granio.TextRoleSubtitle:
		size, style.Bold = 20, true
	case granio.TextRoleTitle:
		size, style.Bold = 28, true
	case granio.TextRoleCode:
		style.Monospace = true
	}
	return granio.TypographyStyle{Size: size * t.textScale, LineSpacing: 4 * t.textScale, Style: style, Font: t.Font(style)}
}
func (t *modernTheme) Motion(role granio.MotionRole) granio.MotionSpec {
	duration := 100 * time.Millisecond
	switch role {
	case granio.MotionPressed:
		duration = 80 * time.Millisecond
	case granio.MotionSurface:
		duration = 180 * time.Millisecond
	case granio.MotionNavigation:
		duration = 220 * time.Millisecond
	}
	return granio.MotionSpec{Duration: duration, Curve: granio.AnimationEaseOut, Disabled: t.options.DisableMotion || t.highContrast}
}
func (t *modernTheme) ComponentStyle(kind granio.ComponentKind, state granio.VisualState, variant granio.ThemeVariant) granio.ComponentStyle {
	if state&granio.VisualDisabled != 0 {
		state &^= granio.VisualHovered | granio.VisualPressed | granio.VisualFocused
	}
	s := granio.ComponentStyle{Background: t.Color(ColorNameBackground, variant), Foreground: t.Color(ColorNameForeground, variant), Border: t.Color(ColorNameInputBorder, variant), Focus: t.Color(ColorNameFocus, variant), Radius: 4, BorderWidth: 1, FocusWidth: 2, MinimumHeight: 40}
	if t.options.Density == DensityCompact {
		s.MinimumHeight = 32
	}
	switch kind {
	case granio.ComponentButton:
		s.Background = t.Color(ColorNameButton, variant)
		s.Foreground = t.Color(ColorNameForegroundOnButton, variant)
	case granio.ComponentInput, granio.ComponentChoice:
		s.Background = t.Color(ColorNameInputBackground, variant)
	case granio.ComponentCard, granio.ComponentPopup:
		s.Background = t.Color(ColorNameMenuBackground, variant)
		s.Radius = 8
	case granio.ComponentSelection:
		s.Background = t.Color(ColorNameSelection, variant)
	}
	if state&granio.VisualReadOnly != 0 && kind == granio.ComponentInput {
		s.Background = t.Color(ColorNameBackground, variant)
	}
	if state&granio.VisualSelected != 0 && kind == granio.ComponentContent {
		s.Background = t.Color(ColorNameSelection, variant)
	}
	if state&granio.VisualPressed != 0 {
		s.Background = overlayVisualColor(s.Background, t.Color(ColorNamePressed, variant))
	} else if state&granio.VisualHovered != 0 {
		s.Background = overlayVisualColor(s.Background, t.Color(ColorNameHover, variant))
	}
	if state&granio.VisualDisabled != 0 {
		s.Foreground = t.Color(ColorNameDisabled, variant)
	}
	if state&granio.VisualError != 0 {
		s.Border = t.Color(ColorNameError, variant)
		s.BorderWidth = 2
	}
	if t.highContrast {
		s.BorderWidth = 2
	}
	if state&granio.VisualFocused == 0 || state&granio.VisualDisabled != 0 {
		s.FocusWidth = 0
	}
	return s
}

func overlayVisualColor(base, overlay color.Color) color.Color {
	br, bg, bb, _ := base.RGBA()
	r, g, b, a := overlay.RGBA()
	return color.RGBA64{R: uint16(r + br*(65535-a)/65535), G: uint16(g + bg*(65535-a)/65535), B: uint16(b + bb*(65535-a)/65535), A: 65535}
}
func (t *modernTheme) MaterialStyle(role granio.SurfaceRole, variant granio.ThemeVariant) granio.MaterialStyle {
	r := granio.MaterialStyle{BlurRadius: 16, Noise: 0.025, TintOpacity: 0.82, CornerRadius: 8, Tint: t.Color(ColorNameMenuBackground, variant), Fallback: t.Color(ColorNameMenuBackground, variant), Opaque: t.highContrast || t.transparencyDisabled}
	if role == granio.SurfaceWindow {
		r.CornerRadius = 0
		r.Fallback = t.Color(ColorNameBackground, variant)
	}
	if r.Fallback == nil {
		r.Fallback = color.Black
	}
	return r
}
