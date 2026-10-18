//go:build wasm && !test_web_driver

package glfw

import (
	"math"
	"syscall/js"
)

var browserFrameWake = make(chan struct{}, 1)

func wakeMainLoop() {
	select {
	case browserFrameWake <- struct{}{}:
	default:
	}
}

func browserFrameSource(d *gLDriver) (frames <-chan struct{}, request func(), stop func()) {
	ready := make(chan struct{}, 1)
	pending := false
	handle := 0
	var previous, nextDelivery float64
	var priorFPS int32
	var callback js.Func
	callback = js.FuncOf(func(this js.Value, args []js.Value) any {
		pending = false
		if len(args) == 0 {
			return nil
		}
		now := args[0].Float()
		if previous > 0 && now > previous && now-previous < 100 {
			rate := int32(1000/(now-previous) + 0.5)
			if rate > 0 && rate <= 1000 {
				if old := d.detectedFrameRate.Load(); old > 0 {
					rate = (old*3 + rate) / 4
				}
				d.detectedFrameRate.Store(rate)
			}
		}
		previous = now
		// Native RAF controls presentation. Explicit rates only lower the cap.
		fps := d.requestedFrameRate.Load()
		if fps != priorFPS {
			nextDelivery = now
			priorFPS = fps
		}
		if fps > 0 {
			if now+0.05 < nextDelivery {
				request()
				return nil
			}
			interval := 1000 / float64(fps)
			nextDelivery += (math.Floor((now-nextDelivery)/interval) + 1) * interval
		}
		select {
		case ready <- struct{}{}:
		default:
		}
		return nil
	})
	request = func() {
		if pending {
			return
		}
		pending = true
		handle = js.Global().Call("requestAnimationFrame", callback).Int()
	}
	stop = func() {
		if pending {
			js.Global().Call("cancelAnimationFrame", handle)
		}
		callback.Release()
	}
	return ready, request, stop
}
