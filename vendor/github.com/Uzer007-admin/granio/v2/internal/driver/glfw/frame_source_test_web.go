//go:build test_web_driver

package glfw

import (
	"sync/atomic"
	"time"
)

var browserFrameWake = make(chan struct{}, 1)

func wakeMainLoop() {
	select {
	case browserFrameWake <- struct{}{}:
	default:
	}
}
func browserFrameSource(d *gLDriver) (<-chan struct{}, func(), func()) {
	ready := make(chan struct{}, 1)
	var pending atomic.Bool
	var timer *time.Timer
	request := func() {
		if pending.Swap(true) {
			return
		}
		timer = time.AfterFunc(time.Second/time.Duration(d.FrameRate()), func() {
			pending.Store(false)
			select {
			case ready <- struct{}{}:
			default:
			}
		})
	}
	return ready, request, func() {
		if timer != nil {
			timer.Stop()
		}
	}
}
