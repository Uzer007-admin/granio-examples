package desktop

import (
	"errors"

	"github.com/Uzer007-admin/granio/v2"
)

// WindowBackdrop describes the OS-composited window background. In-canvas
// materials have their own canvas.Material API and are not window backdrops.
type WindowBackdrop uint8

const (
	BackdropOpaque WindowBackdrop = iota
	BackdropMica
	BackdropMicaAlt
	BackdropDesktopAcrylic
	// BackdropTransparent reveals the desktop through unpainted canvas pixels.
	BackdropTransparent
	// BackdropSystemBlur asks the compositor to blur behind the window. The
	// compositor controls the algorithm and strength; availability can change.
	BackdropSystemBlur
)

func (b WindowBackdrop) String() string {
	values := [...]string{"Opaque", "Mica", "Mica Alt", "Desktop Acrylic", "Transparent", "System Blur"}
	if int(b) < len(values) {
		return values[b]
	}
	return "Unknown"
}

type BackdropSupport uint8

func (s BackdropSupport) Supports(backdrop WindowBackdrop) bool {
	return backdrop <= BackdropSystemBlur && s&(1<<backdrop) != 0
}

type VisualFallbackPolicy uint8

const (
	VisualFallbackOpaque VisualFallbackPolicy = iota
	VisualFallbackFail
)

// VisualFramePolicy preserves the existing native frame and caption buttons.
// Other frame policies must be explicitly implemented rather than replacing
// decorations when a material is requested.
type VisualFramePolicy uint8

const VisualFramePreserve VisualFramePolicy = 0

type VisualOptions struct {
	Backdrop WindowBackdrop
	// PrepareAlpha reserves the format before native surface creation. It does
	// not recreate an existing opaque surface. An opaque request may reserve
	// alpha so later backdrop switches retain the same native window. A native
	// backdrop request also reserves alpha before creation, without this flag.
	PrepareAlpha bool
	Fallback     VisualFallbackPolicy
	Frame        VisualFramePolicy
}

type VisualPhase uint8

const (
	VisualPending VisualPhase = iota
	VisualReady
	VisualFallback
	VisualFailed
	VisualClosed
)

func (p VisualPhase) String() string {
	values := [...]string{"Pending", "Ready", "Fallback", "Failed", "Closed"}
	if int(p) < len(values) {
		return values[p]
	}
	return "Unknown"
}

type VisualFallbackReason uint8

const (
	VisualReasonNone VisualFallbackReason = iota
	VisualUnsupportedOS
	VisualUnsupportedBackend
	VisualSurfaceNotTransparent
	VisualTransparencyDisabled
	VisualHighContrast
	VisualRemoteSession
	VisualResourceBudget
	VisualNativeFailure
	VisualCompositorUnavailable
	VisualBlurUnavailable
)

func (r VisualFallbackReason) String() string {
	values := [...]string{"None", "UnsupportedOS", "UnsupportedBackend", "SurfaceNotTransparent", "TransparencyDisabled", "HighContrast", "RemoteSession", "ResourceBudget", "NativeFailure", "CompositorUnavailable", "BlurUnavailable"}
	if int(r) < len(values) {
		return values[r]
	}
	return "Unknown"
}

type VisualBackend uint8

const (
	VisualBackendUnknown VisualBackend = iota
	VisualBackendOpaque
	VisualBackendWGL
	VisualBackendX11
	VisualBackendWayland
)

// Before native creation the Known fields may be false. Generation identifies
// a surface lifetime; OS support and actual created surface are independent.
type VisualCapabilities struct {
	OSBackdrops, RendererBackdrops       BackdropSupport
	OSKnown, RendererKnown, SurfaceKnown bool
	TransparentSurface                   bool
	// AlphaSurface describes the created format independently of whether a
	// compositor is running. TransparentSurface additionally requires composition.
	AlphaSurface, CompositorAvailable bool
	AlphaBits                         int
	CanChangeWithoutRecreate          bool
	Generation                        uint64
	OSBuild                           uint32
	Backend                           VisualBackend
}

// Completion reports our graphics/native request, not optical proof of a compositor.
// NativeAccepted is distinct from SurfaceReady. PreparedFrames counts frames
// submitted for this request; FirstFrameBeforeShow records hidden preparation.
type VisualStatus struct {
	RequestID                    uint64
	Phase                        VisualPhase
	Requested, Selected          WindowBackdrop
	Backend                      VisualBackend
	SurfaceReady, NativeAccepted bool
	Reason                       VisualFallbackReason
	NativeCode                   uint32
	PreparedFrames               uint32
	FirstFrameBeforeShow         bool
}

var (
	ErrVisualUnsupported   = errors.New("desktop: window visuals unsupported")
	ErrVisualOptions       = errors.New("desktop: invalid visual options")
	ErrVisualNeedsRecreate = errors.New("desktop: alpha surface must be configured before native creation")
	ErrVisualWindowClosed  = errors.New("desktop: window closed")
)

func ValidateVisualOptions(options VisualOptions) error {
	if options.Backdrop > BackdropSystemBlur || options.Fallback > VisualFallbackFail || options.Frame != VisualFramePreserve {
		return ErrVisualOptions
	}
	return nil
}

// WindowVisuals is optional and adds no methods to existing window interfaces.
// ConfigureVisuals is safe from any goroutine; format requirements commit
// synchronously and native work is queued to the GLFW owner thread. Latest
// accepted request wins. Observers run on the Fyne UI goroutine; cancellation
// drops subsequent notifications, including queued notifications. A callback
// which has already entered delivery may finish after cancellation.
type WindowVisuals interface {
	ConfigureVisuals(VisualOptions) (requestID uint64, err error)
	VisualCapabilities() VisualCapabilities
	VisualStatus() VisualStatus
	ObserveVisualStatus(func(VisualStatus)) (cancel func())
}

func ConfigureWindowVisuals(window granio.Window, options VisualOptions) (uint64, error) {
	if err := ValidateVisualOptions(options); err != nil {
		return 0, err
	}
	if visuals, ok := window.(WindowVisuals); ok {
		return visuals.ConfigureVisuals(options)
	}
	return 0, ErrVisualUnsupported
}
