//go:build linux && !wasm && !test_web_driver

package glfw

import (
	"errors"
	"time"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
)

func windowVisualAlphaFormat() bool { return true }

type windowVisualPolicy struct {
	platform                                     glfw.BackgroundPlatform
	composition, alpha, blur, queryFailed        bool
	highContrast, transparencyKnown, transparent bool
	revision, compositor                         uint64
}

func (w *window) queryWindowVisualPolicy() (windowVisualPolicy, glfw.BackgroundEffectState, error) {
	s, err := w.viewport.QueryBackgroundEffects()
	p := windowVisualPolicy{platform: s.Platform, composition: s.Composition,
		alpha: s.AlphaSurface, blur: s.BlurAvailable, revision: s.Revision,
		compositor: s.Compositor, queryFailed: err != nil}
	if settings, ok := granio.CurrentApp().Settings().(granio.SystemAppearanceProvider); ok {
		snapshot := settings.SystemAppearance()
		p.highContrast = snapshot.HighContrast.Available && snapshot.HighContrast.Value
		p.transparencyKnown = snapshot.Transparency.Available
		p.transparent = snapshot.Transparency.Value
	}
	if theme, ok := granio.CurrentApp().Settings().Theme().(interface{ HighContrast() bool }); ok && theme.HighContrast() {
		p.highContrast = true
	}
	return p, s, err
}

func (w *window) applyNativeWindowBackdrop(o desktop.VisualOptions, caps *desktop.VisualCapabilities) (desktop.VisualFallbackReason, bool, uint32) {
	policy, state, err := w.queryWindowVisualPolicy()
	w.visuals.lastPolicy = policy
	caps.OSKnown = err == nil
	caps.OSBackdrops = 1 << desktop.BackdropOpaque
	caps.AlphaSurface = state.AlphaSurface && caps.AlphaBits > 0
	caps.CompositorAvailable = state.Composition
	caps.TransparentSurface = caps.AlphaSurface && state.TransparencyAvailable
	switch state.Platform {
	case glfw.BackgroundX11:
		caps.Backend = desktop.VisualBackendX11
	case glfw.BackgroundWayland:
		caps.Backend = desktop.VisualBackendWayland
	}
	if state.Platform != glfw.BackgroundUnsupported {
		caps.OSBackdrops |= 1 << desktop.BackdropTransparent
	}
	if state.BlurAvailable {
		caps.OSBackdrops |= 1 << desktop.BackdropSystemBlur
	}
	if reason := windowBackdropRejection(o, policy, caps, err); reason != desktop.VisualReasonNone {
		return reason, false, 0
	}
	if o.Backdrop == desktop.BackdropOpaque {
		return desktop.VisualReasonNone, false, 0
	}
	if err := w.viewport.SetBackgroundBlur(o.Backdrop == desktop.BackdropSystemBlur); err != nil {
		switch {
		case errors.Is(err, glfw.ErrBackgroundComposition):
			return desktop.VisualCompositorUnavailable, false, 0
		case errors.Is(err, glfw.ErrBackgroundBlur):
			return desktop.VisualBlurUnavailable, false, 0
		case errors.Is(err, glfw.ErrBackgroundAlpha):
			return desktop.VisualSurfaceNotTransparent, false, 0
		default:
			return desktop.VisualNativeFailure, false, 0
		}
	}
	return desktop.VisualReasonNone, true, 0
}

func windowBackdropRejection(o desktop.VisualOptions, p windowVisualPolicy, caps *desktop.VisualCapabilities, err error) desktop.VisualFallbackReason {
	if o.Backdrop == desktop.BackdropOpaque {
		return desktop.VisualReasonNone
	}
	if o.Backdrop != desktop.BackdropTransparent && o.Backdrop != desktop.BackdropSystemBlur {
		return desktop.VisualUnsupportedOS
	}
	if err != nil || p.platform == glfw.BackgroundUnsupported {
		return desktop.VisualUnsupportedBackend
	}
	if p.highContrast {
		return desktop.VisualHighContrast
	}
	if p.transparencyKnown && !p.transparent {
		return desktop.VisualTransparencyDisabled
	}
	if !p.composition {
		return desktop.VisualCompositorUnavailable
	}
	if !caps.RendererBackdrops.Supports(o.Backdrop) {
		return desktop.VisualUnsupportedBackend
	}
	if !caps.TransparentSurface {
		return desktop.VisualSurfaceNotTransparent
	}
	if o.Backdrop == desktop.BackdropSystemBlur && !p.blur {
		return desktop.VisualBlurUnavailable
	}
	return desktop.VisualReasonNone
}

func (w *window) clearNativeWindowBackdrop() {
	state, err := w.viewport.QueryBackgroundEffects()
	if err == nil && state.BlurRequested {
		_ = w.viewport.SetBackgroundBlur(false)
	}
}

func (w *window) refreshVisualPolicy() {
	if !w.visuals.active.Load() || w.viewport == nil || w.closing || time.Now().Before(w.visuals.policyNext) {
		return
	}
	w.visuals.policyNext = time.Now().Add(500 * time.Millisecond)
	policy, _, _ := w.queryWindowVisualPolicy()
	if policy != w.visuals.lastPolicy {
		w.applyWindowVisuals()
	}
}
