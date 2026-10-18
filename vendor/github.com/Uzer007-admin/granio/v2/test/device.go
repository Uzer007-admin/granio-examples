package test

import (
	"runtime"

	"github.com/Uzer007-admin/granio/v2"
)

type device struct{}

// Declare conformity with Device
var _ granio.Device = (*device)(nil)

func (*device) Orientation() granio.DeviceOrientation {
	return granio.OrientationVertical
}

func (*device) HasKeyboard() bool {
	return false
}

func (d *device) SystemScale() float32 {
	return d.SystemScaleForWindow(nil)
}

func (*device) SystemScaleForWindow(granio.Window) float32 {
	return 1
}

func (*device) Locale() granio.Locale {
	return "en"
}

func (*device) IsBrowser() bool {
	return runtime.GOARCH == "js" || runtime.GOOS == "js"
}
