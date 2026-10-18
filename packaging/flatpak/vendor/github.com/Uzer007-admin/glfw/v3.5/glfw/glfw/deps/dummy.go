//go:build required

// Package dummy keeps native source files in vendored modules.
package dummy

import (
	_ "github.com/Uzer007-admin/glfw/v3.5/glfw/glfw/deps/glad"
	_ "github.com/Uzer007-admin/glfw/v3.5/glfw/glfw/deps/wayland"
)
