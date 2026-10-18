package glfw

import (
	"fmt"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
)

var _ desktop.FrameRateDriver = (*gLDriver)(nil)

func (d *gLDriver) SetFrameRate(fps int) error {
	if fps < 0 || fps > 1000 {
		return fmt.Errorf("frame rate must be 0 (automatic) or between 1 and 1000, got %d", fps)
	}
	d.requestedFrameRate.Store(int32(fps))
	wakeMainLoop()
	return nil
}
func (d *gLDriver) FrameRate() int {
	if fps := d.requestedFrameRate.Load(); fps > 0 {
		return int(fps)
	}
	if fps := d.detectedFrameRate.Load(); fps > 0 {
		return int(fps)
	}
	return 60
}
func (d *gLDriver) hasFrameWork() bool {
	if d.animation.HasAnimations() {
		return true
	}
	for _, win := range d.windows {
		w := win.(*window)
		if !w.closing && w.visible && w.frame.ready() && w.canvas.HasDirty() {
			return true
		}
	}
	return false
}
