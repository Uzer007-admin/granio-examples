//go:build wasm || test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/painter/gl"
)

func windowVisualAlphaFormat() bool             { return false }
func (w *window) prepareVisualFirstFrame() bool { return w.applyWindowVisuals() }
func (*window) refreshVisualPolicy()            {}
func (w *window) applyWindowVisuals() bool {
	o, status, enabled := w.visualRequest()
	if !enabled {
		return true
	}
	if w.viewport == nil || w.closing {
		return false
	}
	status.Phase = desktop.VisualReady
	status.Backend = desktop.VisualBackendOpaque
	w.RunWithContext(func() {
		if painter, ok := w.canvas.Painter().(surfacePainter); ok {
			status.SurfaceReady = painter.SetWindowSurface(gl.WindowSurfaceOpaque)
			if status.SurfaceReady {
				w.driver.repaintWindow(w)
				status.PreparedFrames++
			}
		}
	})
	if o.Backdrop != desktop.BackdropOpaque {
		status.Phase = desktop.VisualFallback
		if o.Fallback == desktop.VisualFallbackFail {
			status.Phase = desktop.VisualFailed
		}
		status.Reason = desktop.VisualUnsupportedBackend
	}
	if !status.SurfaceReady {
		status.Phase = desktop.VisualFailed
		status.Reason = desktop.VisualUnsupportedBackend
	}
	w.publishVisualStatus(status, desktop.VisualCapabilities{OSKnown: true, RendererKnown: true, SurfaceKnown: w.viewport != nil, OSBackdrops: 1, RendererBackdrops: 1, Backend: desktop.VisualBackendOpaque})
	return status.SurfaceReady
}
