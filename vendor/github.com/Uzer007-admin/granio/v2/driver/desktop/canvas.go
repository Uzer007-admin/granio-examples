package desktop

import "github.com/Uzer007-admin/granio/v2"

// Canvas defines the desktop specific extensions to a granio.Canvas.
type Canvas interface {
	OnKeyDown() func(*granio.KeyEvent)
	SetOnKeyDown(func(*granio.KeyEvent))
	OnKeyUp() func(*granio.KeyEvent)
	SetOnKeyUp(func(*granio.KeyEvent))
}
