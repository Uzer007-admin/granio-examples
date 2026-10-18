package embedded

import (
	"github.com/Uzer007-admin/granio/v2"
)

type noosDevice struct{}

func (*noosDevice) Orientation() granio.DeviceOrientation {
	return granio.OrientationVertical
}

func (*noosDevice) IsMobile() bool {
	return false
}

func (*noosDevice) IsBrowser() bool {
	return false
}

func (*noosDevice) HasKeyboard() bool {
	return true
}

func (*noosDevice) SystemScaleForWindow(granio.Window) float32 {
	return 1.0
}

func (*noosDevice) Locale() granio.Locale {
	return "en"
}
