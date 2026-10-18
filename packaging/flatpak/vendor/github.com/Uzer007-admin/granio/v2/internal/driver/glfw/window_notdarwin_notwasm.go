//go:build !darwin && !wasm && !test_web_driver

package glfw

import "github.com/Uzer007-admin/glfw/v3.5/glfw"

type monitor = glfw.Monitor

func (*window) getSecondaryMonitor() *monitor {
	primary := glfw.GetPrimaryMonitor()
	for _, m := range glfw.GetMonitors() {
		if m.GetName() != primary.GetName() {
			return m
		}
	}

	return primary
}
