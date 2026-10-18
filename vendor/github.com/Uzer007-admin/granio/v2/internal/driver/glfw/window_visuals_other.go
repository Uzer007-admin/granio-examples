//go:build !windows && !linux && !wasm && !test_web_driver

package glfw

import "github.com/Uzer007-admin/granio/v2/driver/desktop"

func windowVisualAlphaFormat() bool { return false }
func (w *window) applyNativeWindowBackdrop(o desktop.VisualOptions, caps *desktop.VisualCapabilities) (desktop.VisualFallbackReason, bool, uint32) {
	caps.OSKnown = true
	caps.OSBackdrops = 1 << desktop.BackdropOpaque
	if o.Backdrop == desktop.BackdropOpaque {
		return desktop.VisualReasonNone, false, 0
	}
	return desktop.VisualUnsupportedOS, false, 0
}
func (*window) clearNativeWindowBackdrop() {}
func (*window) refreshVisualPolicy()       {}
