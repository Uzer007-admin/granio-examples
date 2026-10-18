//go:build !linux && !wasm && !test_web_driver

package glfw

func forcePlatform() string { return platformAuto }
