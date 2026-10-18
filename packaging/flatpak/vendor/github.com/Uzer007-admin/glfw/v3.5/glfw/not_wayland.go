//go:build (linux && x11 && !wayland) || ((freebsd || netbsd) && !wayland) || openbsd

package glfw

const WAYLAND = false
