package glfw

import "errors"

var (
	ErrVisualUnsupported = errors.New("glfw: native window visuals unsupported")
	ErrWindowDestroyed   = errors.New("glfw: window destroyed")
	ErrVisualThread      = errors.New("glfw: window visuals require the window owner thread")
	ErrVisualOptions     = errors.New("glfw: invalid window visual options")
	ErrSurfaceAlpha      = errors.New("glfw: premultiplied redirection requires a verified transparent WGL surface")
	ErrVisualApply       = errors.New("glfw: native visual attribute rejected")
)

// SystemBackdrop uses the documented DWM roles. On current Windows 11 these
// correspond to Mica, Desktop Acrylic and Mica Alt; the OS controls their visual
// appearance and fallback. A successful attribute call does not prove visibility.
type SystemBackdrop int32

const (
	BackdropAuto SystemBackdrop = iota
	BackdropNone
	BackdropMainWindow
	BackdropTransientWindow
	BackdropTabbedWindow
)

type WindowCorners int32

const (
	CornersDefault WindowCorners = iota
	CornersSquare
	CornersRound
	CornersSmallRound
)

type WindowVisualAttributes uint8

const (
	VisualBackdrop WindowVisualAttributes = 1 << iota
	VisualDarkCaption
	VisualCorners
	VisualPremultipliedAlpha
)
const allWindowVisualAttributes = VisualBackdrop | VisualDarkCaption | VisualCorners | VisualPremultipliedAlpha

// WindowVisualOptions changes only Attributes explicitly selected by the caller.
// TransparentFramebuffer must be hinted before CreateWindow. Enabling alpha
// requires premultiplied RGB in every framebuffer pixel, including RGB=0 at A=0.
// This API changes no window decoration, native buttons, size, or context.
type WindowVisualOptions struct {
	Attributes                      WindowVisualAttributes
	Backdrop                        SystemBackdrop
	Corners                         WindowCorners
	DarkCaption, PremultipliedAlpha bool
}

// WindowVisualState separates OS API support from the created surface. Values
// are a sequential snapshot on the window owner thread, not compositor evidence.
// PixelFormat/AlphaBits describe the actual Win32 pixel format, not an alpha hint.
// No WGL pixel format (for example EGL/NoAPI) leaves PixelFormat/AlphaBits zero.
// Other platforms return ErrVisualUnsupported; they can still use in-app materials.
type WindowVisualState struct {
	OSBuild                                                     uint32
	OSAttributes                                                WindowVisualAttributes
	Composition, RemoteSession, HighContrast, HighContrastKnown bool
	TransparentFramebuffer                                      bool
	PixelFormat, AlphaBits                                      int
	Backdrop                                                    SystemBackdrop
	BackdropKnown                                               bool
}

func (s WindowVisualState) WGLAlphaAvailable() bool {
	return s.TransparentFramebuffer && s.PixelFormat != 0 && s.AlphaBits > 0
}

// WindowVisualResult reports accepted native requests, not a rendered material.
// HRESULT order is backdrop, dark caption, corners, premultiplied alpha. Native
// failures may leave a subset applied; Accepted identifies that subset. State
// was captured during preflight. Only backdrop is documented as queryable.
type WindowVisualResult struct {
	Requested, Accepted WindowVisualAttributes
	HRESULT             [4]uint32
	State               WindowVisualState
}

func (w *Window) QueryWindowVisuals() (WindowVisualState, error) {
	if w == nil || w.data == nil {
		return WindowVisualState{}, ErrWindowDestroyed
	}
	return w.queryWindowVisuals()
}

// ApplyWindowVisuals performs capability/options preflight before changing any
// attribute. Unsupported platforms/OS attributes return a reasoned error. This
// is main-thread, explicit opt-in; it does not decide application accessibility
// policy, recreate an opaque surface, or install per-window callback hooks.
func (w *Window) ApplyWindowVisuals(options WindowVisualOptions) (WindowVisualResult, error) {
	result := WindowVisualResult{Requested: options.Attributes}
	if w == nil || w.data == nil {
		return result, ErrWindowDestroyed
	}
	if options.Attributes & ^allWindowVisualAttributes != 0 || options.Attributes&VisualBackdrop != 0 && (options.Backdrop < BackdropAuto || options.Backdrop > BackdropTabbedWindow) || options.Attributes&VisualCorners != 0 && (options.Corners < CornersDefault || options.Corners > CornersSmallRound) {
		return result, ErrVisualOptions
	}
	state, err := w.queryWindowVisuals()
	result.State = state
	if err != nil {
		return result, err
	}
	if options.Attributes & ^state.OSAttributes != 0 {
		return result, ErrVisualUnsupported
	}
	if options.Attributes&VisualPremultipliedAlpha != 0 && options.PremultipliedAlpha && !state.WGLAlphaAvailable() {
		return result, ErrSurfaceAlpha
	}
	if options.Attributes == 0 {
		return result, nil
	}
	result.HRESULT, err = w.applyWindowVisuals(options)
	if err != nil {
		return result, err
	}
	for i := range 4 {
		attribute := WindowVisualAttributes(1 << i)
		if options.Attributes&attribute != 0 && result.HRESULT[i] == 0 {
			result.Accepted |= attribute
		}
	}
	if result.Accepted != result.Requested {
		return result, ErrVisualApply
	}
	return result, nil
}
