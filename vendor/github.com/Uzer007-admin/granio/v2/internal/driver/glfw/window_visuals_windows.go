//go:build !wasm && !test_web_driver

package glfw

import (
	"time"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
	"golang.org/x/sys/windows/registry"
)

func windowVisualAlphaFormat() bool { return true }

type windowVisualPolicy struct {
	composition, highContrast, highContrastKnown, remote, transparent, transparencyKnown bool
	build                                                                                uint32
}

func (w *window) queryWindowVisualPolicy() (windowVisualPolicy, glfw.WindowVisualState, error) {
	s, err := w.viewport.QueryWindowVisuals()
	p := windowVisualPolicy{composition: s.Composition, highContrast: s.HighContrast, highContrastKnown: s.HighContrastKnown, remote: s.RemoteSession, build: s.OSBuild}
	key, keyErr := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if keyErr == nil {
		value, _, valueErr := key.GetIntegerValue("EnableTransparency")
		key.Close()
		if valueErr == nil {
			p.transparent = value != 0
			p.transparencyKnown = true
		}
	}
	if settings, ok := granio.CurrentApp().Settings().(granio.SystemAppearanceProvider); ok {
		snapshot := settings.SystemAppearance()
		if snapshot.HighContrast.Available {
			p.highContrast, p.highContrastKnown = p.highContrast || snapshot.HighContrast.Value, true
		}
		if snapshot.Transparency.Available {
			p.transparent = (!p.transparencyKnown || p.transparent) && snapshot.Transparency.Value
			p.transparencyKnown = true
		}
	}
	if theme, ok := granio.CurrentApp().Settings().Theme().(interface{ HighContrast() bool }); ok && theme.HighContrast() {
		p.highContrast = true
		p.highContrastKnown = true
	}
	return p, s, err
}

func (w *window) applyNativeWindowBackdrop(o desktop.VisualOptions, caps *desktop.VisualCapabilities) (desktop.VisualFallbackReason, bool, uint32) {
	policy, state, err := w.queryWindowVisualPolicy()
	w.visuals.lastPolicy = policy
	caps.OSKnown = err == nil
	caps.OSBuild = state.OSBuild
	caps.OSBackdrops = 1 << desktop.BackdropOpaque
	caps.Backend = desktop.VisualBackendWGL
	caps.TransparentSurface = state.WGLAlphaAvailable()
	caps.AlphaSurface = state.WGLAlphaAvailable() && caps.AlphaBits > 0
	caps.CompositorAvailable = state.Composition
	caps.OSBackdrops |= 1 << desktop.BackdropTransparent
	if state.OSAttributes&glfw.VisualBackdrop != 0 {
		caps.OSBackdrops |= 0xf | 1<<desktop.BackdropSystemBlur
	}
	if reason := windowBackdropRejection(o, policy, caps, err); reason != desktop.VisualReasonNone {
		return reason, false, 0
	}
	if o.Backdrop == desktop.BackdropOpaque {
		return desktop.VisualReasonNone, false, 0
	}
	backdrop := glfw.BackdropMainWindow
	switch o.Backdrop {
	case desktop.BackdropMicaAlt:
		backdrop = glfw.BackdropTabbedWindow
	case desktop.BackdropDesktopAcrylic:
		backdrop = glfw.BackdropTransientWindow
	case desktop.BackdropSystemBlur:
		backdrop = glfw.BackdropTransientWindow
	case desktop.BackdropTransparent:
		backdrop = glfw.BackdropNone
	}
	attributes := state.OSAttributes & (glfw.VisualBackdrop | glfw.VisualDarkCaption)
	// Older Windows uses GLFW's documented transparent framebuffer path. The
	// redirection alpha attribute is additional and requested only when supported.
	if state.OSAttributes&glfw.VisualPremultipliedAlpha != 0 {
		attributes |= glfw.VisualPremultipliedAlpha
	}
	result, applyErr := w.viewport.ApplyWindowVisuals(glfw.WindowVisualOptions{Attributes: attributes, Backdrop: backdrop, DarkCaption: granio.CurrentApp().Settings().ThemeVariant() == theme.VariantDark, PremultipliedAlpha: true})
	if applyErr != nil {
		code := uint32(0)
		for _, hr := range result.HRESULT {
			if hr != 0 {
				code = hr
				break
			}
		}
		return desktop.VisualNativeFailure, false, code
	}
	return desktop.VisualReasonNone, true, 0
}

// Resolve restrictions before issuing any native attributes. It is shared by
// runtime application and policy tests; no test needs to change OS settings.
func windowBackdropRejection(o desktop.VisualOptions, policy windowVisualPolicy, caps *desktop.VisualCapabilities, err error) desktop.VisualFallbackReason {
	if o.Backdrop == desktop.BackdropOpaque {
		return desktop.VisualReasonNone
	}
	if err != nil || !caps.OSBackdrops.Supports(o.Backdrop) {
		return desktop.VisualUnsupportedOS
	}
	if policy.highContrast {
		return desktop.VisualHighContrast
	}
	if policy.transparencyKnown && !policy.transparent {
		return desktop.VisualTransparencyDisabled
	}
	if policy.remote {
		return desktop.VisualRemoteSession
	}
	if !policy.composition {
		return desktop.VisualUnsupportedBackend
	}
	if !caps.RendererBackdrops.Supports(o.Backdrop) {
		return desktop.VisualUnsupportedBackend
	}
	if !caps.TransparentSurface {
		return desktop.VisualSurfaceNotTransparent
	}
	return desktop.VisualReasonNone
}

func (w *window) clearNativeWindowBackdrop() {
	state, err := w.viewport.QueryWindowVisuals()
	if err != nil {
		return
	}
	attributes := state.OSAttributes & (glfw.VisualBackdrop | glfw.VisualPremultipliedAlpha)
	if attributes == 0 {
		return
	}
	_, _ = w.viewport.ApplyWindowVisuals(glfw.WindowVisualOptions{Attributes: attributes, Backdrop: glfw.BackdropNone})
}

func (w *window) refreshVisualPolicy() {
	// Ordinary windows pay only an atomic load, not a mutex or system query.
	if !w.visuals.active.Load() || w.viewport == nil || w.closing || time.Now().Before(w.visuals.policyNext) {
		return
	}
	w.visuals.policyNext = time.Now().Add(500 * time.Millisecond)
	policy, _, err := w.queryWindowVisualPolicy()
	if err == nil && policy != w.visuals.lastPolicy {
		w.applyWindowVisuals()
	}
}
