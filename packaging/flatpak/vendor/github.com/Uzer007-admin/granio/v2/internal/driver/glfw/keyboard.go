//go:build wasm

package glfw

import (
	"github.com/Uzer007-admin/granio/v2"
)

func hideVirtualKeyboard() {
	if d, ok := granio.CurrentDevice().(*glDevice); ok {
		d.hideVirtualKeyboard()
	}
}

func handleKeyboard(obj granio.Focusable) {
	isDisabled := false
	if disWid, ok := obj.(granio.Disableable); ok {
		isDisabled = disWid.Disabled()
	}
	if obj != nil && !isDisabled {
		showVirtualKeyboard()
	} else {
		hideVirtualKeyboard()
	}
}

func showVirtualKeyboard() {
	if d, ok := granio.CurrentDevice().(*glDevice); ok {
		d.showVirtualKeyboard()
	}
}
