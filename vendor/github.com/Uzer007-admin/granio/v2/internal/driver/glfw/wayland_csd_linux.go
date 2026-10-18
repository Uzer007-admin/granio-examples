//go:build linux && !wasm && !test_web_driver

package glfw

import "os"

// Автоматический режим оставляет GLFW согласование декораций и выбор протокола.
func forcePlatform() string {
	switch os.Getenv("FYNE_PLATFORM") {
	case platformX11:
		return platformX11
	case platformWayland:
		return platformWayland
	}
	return platformAuto
}
