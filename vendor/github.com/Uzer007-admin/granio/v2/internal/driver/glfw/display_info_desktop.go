//go:build !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"os"
)

func (d *gLDriver) DisplayInfo() desktop.DisplayInfo {
	name := "unknown"
	switch int(d.displayPlatform.Load()) {
	case int(glfw.PlatformX11):
		name = "x11"
	case int(glfw.PlatformWayland):
		name = "wayland"
	case int(glfw.PlatformWin32):
		name = "win32"
	case int(glfw.PlatformCocoa):
		name = "cocoa"
	}
	requested := forcePlatform()
	if requested == "" {
		requested = "auto"
	}
	return desktop.DisplayInfo{Backend: name, Requested: requested, SessionType: os.Getenv("XDG_SESSION_TYPE")}
}
