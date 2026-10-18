package glfw

import (
	"runtime"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/goos"
	"github.com/Uzer007-admin/granio/v2/lang"
)

type glDevice struct{}

// Declare conformity with Device
var _ granio.Device = (*glDevice)(nil)

func (*glDevice) Locale() granio.Locale {
	return lang.SystemLocale()
}

func (*glDevice) Orientation() granio.DeviceOrientation {
	return granio.OrientationHorizontalLeft // TODO should we consider the monitor orientation or topmost window?
}

func (*glDevice) HasKeyboard() bool {
	return true // TODO actually check - we could be in tablet mode
}

func (*glDevice) IsBrowser() bool {
	return runtime.GOARCH == "js" || runtime.GOOS == goos.JavaScript
}
