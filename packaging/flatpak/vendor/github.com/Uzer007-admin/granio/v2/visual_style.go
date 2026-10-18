package granio

import (
	"image/color"
	"time"
)

// TextRole is a semantic size/style. Default preserves explicit SizeName and
// TextStyle; it does not reinterpret a low-level canvas.Text's explicit size.
type TextRole uint8

const (
	TextRoleDefault TextRole = iota
	TextRoleBody
	TextRoleCaption
	TextRoleEmphasis
	TextRoleSubtitle
	TextRoleTitle
	TextRoleCode
)

type TypographyStyle struct {
	Size, LineSpacing float32
	Style             TextStyle
	Font              Resource
}
type TypographyTheme interface {
	Theme
	Typography(TextRole) TypographyStyle
}

type MotionRole uint8

const (
	MotionHover MotionRole = iota
	MotionPressed
	MotionSurface
	MotionNavigation
)

type MotionSpec struct {
	Duration time.Duration
	Curve    AnimationCurve
	Disabled bool
}
type MotionTheme interface {
	Theme
	Motion(MotionRole) MotionSpec
}

// VisualState is composable: readonly, error and focus may coexist. Disabled
// suppresses hover/pressed; error and keyboard focus retain separate outlines.
type VisualState uint16

const (
	VisualHovered VisualState = 1 << iota
	VisualPressed
	VisualFocused
	VisualDisabled
	VisualReadOnly
	VisualSelected
	VisualError
	VisualInactive
)

type ComponentKind uint8

const (
	ComponentButton ComponentKind = iota
	ComponentInput
	ComponentChoice
	ComponentContent
	ComponentCard
	ComponentPopup
	ComponentSelection
)

type ComponentStyle struct {
	Background, Foreground, Border, Focus          color.Color
	Radius, BorderWidth, FocusWidth, MinimumHeight float32
}
type ComponentTheme interface {
	Theme
	ComponentStyle(ComponentKind, VisualState, ThemeVariant) ComponentStyle
}

type SurfaceRole uint8

const (
	SurfaceWindow SurfaceRole = iota
	SurfaceContent
	SurfaceCard
	SurfacePopup
)

// MaterialStyle describes in-canvas materials only. A native window backdrop
// continues to use desktop.WindowVisuals, independently of these tokens.
type MaterialStyle struct {
	BlurRadius, Noise, TintOpacity, CornerRadius float32
	Tint, Fallback                               color.Color
	Opaque                                       bool
}
type MaterialTheme interface {
	Theme
	MaterialStyle(SurfaceRole, ThemeVariant) MaterialStyle
}

// ThemeChangeAware is optional for widgets that preserve transient state while
// Fyne reflows their owned renderer objects. The returned completion runs after
// the child traversal, layout and Refresh, on the same UI goroutine.
type ThemeChangeAware interface {
	BeginThemeChange() func()
}
