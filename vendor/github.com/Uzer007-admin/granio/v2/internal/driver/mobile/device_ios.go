//go:build ios

package mobile

import "github.com/Uzer007-admin/granio/v2"

const tapYOffset = -8.0 // to compensate for how we hold our fingers on the device

func (*device) SystemScaleForWindow(_ granio.Window) float32 {
	if currentDPI >= 450 {
		return 3
	} else if currentDPI >= 340 {
		return 2.5
	}
	return 2
}
