//go:build !darwin

package widget

import (
	"github.com/Uzer007-admin/granio/v2"
)

const (
	textModifierAlt     = "Alt+"
	textModifierControl = "Ctrl+"
	textModifierShift   = "Shift+"
	textModifierSuper   = "Super+"
)

var (
	styleModifiers   = granio.TextStyle{}
	defaultStyleKeys = granio.TextStyle{}
)

var keyTexts = map[granio.KeyName]string{
	granio.KeyBackspace: "Backspace",
	granio.KeyDelete:    "Del",
	granio.KeyDown:      "↓",
	granio.KeyEnd:       "End",
	granio.KeyEnter:     "Enter",
	granio.KeyEscape:    "Esc",
	granio.KeyHome:      "Home",
	granio.KeyLeft:      "←",
	granio.KeyPageDown:  "PgDn",
	granio.KeyPageUp:    "PgUp",
	granio.KeyReturn:    "Return",
	granio.KeyRight:     "→",
	granio.KeySpace:     "Space",
	granio.KeyTab:       "Tab",
	granio.KeyUp:        "↑",
}
