package desktop

import (
	"github.com/Uzer007-admin/granio/v2"
)

const (
	// KeyNone represents no key
	KeyNone granio.KeyName = ""
	// KeyShiftLeft represents the left shift key
	KeyShiftLeft granio.KeyName = "LeftShift"
	// KeyShiftRight represents the right shift key
	KeyShiftRight granio.KeyName = "RightShift"
	// KeyControlLeft represents the left control key
	KeyControlLeft granio.KeyName = "LeftControl"
	// KeyControlRight represents the right control key
	KeyControlRight granio.KeyName = "RightControl"
	// KeyAltLeft represents the left alt key
	KeyAltLeft granio.KeyName = "LeftAlt"
	// KeyAltRight represents the right alt key
	KeyAltRight granio.KeyName = "RightAlt"
	// KeySuperLeft represents the left "Windows" key (or "Command" key on macOS)
	KeySuperLeft granio.KeyName = "LeftSuper"
	// KeySuperRight represents the right "Windows" key (or "Command" key on macOS)
	KeySuperRight granio.KeyName = "RightSuper"
	// KeyMenu represents the left or right menu / application key
	KeyMenu granio.KeyName = "Menu"
	// KeyPrintScreen represents the key used to cause a screen capture
	KeyPrintScreen granio.KeyName = "PrintScreen"

	// KeyCapsLock represents the caps lock key, tapping once is the down event then again is the up
	KeyCapsLock granio.KeyName = "CapsLock"
)

// Modifier captures any key modifiers (shift etc.) pressed during a key event
//
// Deprecated: Use granio.KeyModifier instead.
type Modifier = granio.KeyModifier

const (
	// ShiftModifier represents a shift key being held
	//
	// Deprecated: Use granio.KeyModifierShift instead.
	ShiftModifier = granio.KeyModifierShift
	// ControlModifier represents the ctrl key being held
	//
	// Deprecated: Use granio.KeyModifierControl instead.
	ControlModifier = granio.KeyModifierControl
	// AltModifier represents either alt keys being held
	//
	// Deprecated: Use granio.KeyModifierAlt instead.
	AltModifier = granio.KeyModifierAlt
	// SuperModifier represents either super keys being held
	//
	// Deprecated: Use granio.KeyModifierSuper instead.
	SuperModifier = granio.KeyModifierSuper
)

// Keyable describes any focusable canvas object that can accept desktop key events.
// This is the traditional key down and up event that is not applicable to all devices.
type Keyable interface {
	granio.Focusable

	KeyDown(*granio.KeyEvent)
	KeyUp(*granio.KeyEvent)
}
