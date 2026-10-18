//go:build !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/painter/gl"
)

// Prepare full frames before Show and on mode changes. Wayland suppresses swaps
// on hidden surfaces: its back buffer is painted here and the first visible swap
// follows a full repaint. PreparedFrames counts prepared draws, not scanouts.
func (w *window) paintVisualFrames() uint32 {
	width, height := w.viewport.GetFramebufferSize()
	if width <= 0 || height <= 0 {
		return 0
	}
	var prepared uint32
	w.RunWithContext(func() {
		for i := 0; i < 2; i++ {
			w.driver.repaintWindow(w)
			// Leave the last hidden draw in the back buffer. A final swap may
			// acquire an undefined buffer on drivers with more than two buffers.
			if !w.visible && i == 0 {
				w.viewport.SwapBuffers()
			}
			prepared++
		}
	})
	return prepared
}

func (w *window) applyWindowVisuals() bool {
	options, status, enabled := w.visualRequest()
	if !enabled {
		return true
	}
	if w.viewport == nil || w.closing {
		return false
	}
	caps := w.VisualCapabilities()
	caps.SurfaceKnown = true
	caps.Backend = desktop.VisualBackendOpaque
	mode := gl.WindowSurfaceOpaque
	var surface surfacePainter
	w.RunWithContext(func() {
		surface, _ = w.canvas.Painter().(surfacePainter)
		caps.RendererKnown = true
		caps.RendererBackdrops = 1 << desktop.BackdropOpaque
		if surface != nil {
			caps.AlphaBits = surface.WindowSurfaceAlphaBits()
			probeMode := gl.WindowSurfaceOpaque
			if options.Backdrop != desktop.BackdropOpaque {
				probeMode = gl.WindowSurfaceBackdrop
			}
			if surface.SetWindowSurface(probeMode) {
				if caps.AlphaBits > 0 {
					caps.RendererBackdrops = (1 << (desktop.BackdropSystemBlur + 1)) - 1
				}
			}
		}
	})
	status.Selected = desktop.BackdropOpaque
	status.Reason, status.NativeAccepted, status.NativeCode = w.applyNativeWindowBackdrop(options, &caps)
	if status.Reason == desktop.VisualReasonNone && options.Backdrop != desktop.BackdropOpaque {
		mode = gl.WindowSurfaceBackdrop
		status.Selected = options.Backdrop
	}
	status.Backend = caps.Backend
	if options.Backdrop != desktop.BackdropOpaque && status.Reason != desktop.VisualReasonNone {
		status.Phase = desktop.VisualFallback
		if options.Fallback == desktop.VisualFallbackFail {
			status.Phase = desktop.VisualFailed
		}
	} else {
		status.Phase = desktop.VisualReady
	}
	ready := false
	w.RunWithContext(func() {
		if surface != nil {
			ready = surface.SetWindowSurface(mode)
		}
	})
	if !ready && mode == gl.WindowSurfaceBackdrop {
		status.Selected = desktop.BackdropOpaque
		status.Reason = desktop.VisualUnsupportedBackend
		status.Phase = desktop.VisualFallback
		if options.Fallback == desktop.VisualFallbackFail {
			status.Phase = desktop.VisualFailed
		}
		w.RunWithContext(func() {
			if surface != nil {
				ready = surface.SetWindowSurface(gl.WindowSurfaceOpaque)
			}
		})
	}
	// On fallback/opaque, publish opaque pixels before disabling redirection
	// alpha. This avoids displaying an old transparent buffer as solid black.
	hidden := w.viewport.GetAttrib(glfw.Visible) == glfw.False
	preparedNow := uint32(0)
	if ready {
		// Wayland blur state is double-buffered. Queue its removal before the
		// opaque swap, so pixels and effect change in the same surface commit.
		if status.Selected == desktop.BackdropOpaque && caps.Backend == desktop.VisualBackendWayland {
			w.clearNativeWindowBackdrop()
		}
		preparedNow = w.paintVisualFrames()
		status.PreparedFrames += preparedNow
	}
	status.FirstFrameBeforeShow = status.FirstFrameBeforeShow || hidden && preparedNow > 0
	status.SurfaceReady = ready && preparedNow > 0
	if status.Selected == desktop.BackdropOpaque && status.SurfaceReady && caps.Backend != desktop.VisualBackendWayland {
		w.clearNativeWindowBackdrop()
	}
	if !status.SurfaceReady {
		status.Phase = desktop.VisualFailed
		status.Reason = desktop.VisualUnsupportedBackend
	}
	caps.CanChangeWithoutRecreate = caps.AlphaSurface && caps.RendererBackdrops.Supports(desktop.BackdropTransparent)
	w.publishVisualStatus(status, caps)
	return status.SurfaceReady
}

func (w *window) prepareVisualFirstFrame() bool { return w.applyWindowVisuals() }
