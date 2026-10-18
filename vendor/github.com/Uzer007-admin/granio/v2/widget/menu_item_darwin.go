package widget

import (
	"github.com/Uzer007-admin/granio/v2"
)

const (
	textModifierAlt     = "⌥"
	textModifierControl = "⌃"
	textModifierShift   = "⇧"
	textModifierSuper   = "⌘"
)

var (
	styleModifiers   = granio.TextStyle{Symbol: true}
	defaultStyleKeys = granio.TextStyle{Monospace: true}
)

var keyTexts = map[granio.KeyName]string{
	granio.KeyBackspace: "⌫",
	granio.KeyDelete:    "⌦",
	granio.KeyDown:      "↓",
	granio.KeyEnd:       "↘",
	granio.KeyEnter:     "↩",
	granio.KeyEscape:    "⎋",
	granio.KeyHome:      "↖",
	granio.KeyLeft:      "←",
	granio.KeyPageDown:  "⇟",
	granio.KeyPageUp:    "⇞",
	granio.KeyReturn:    "↩",
	granio.KeyRight:     "→",
	granio.KeySpace:     "␣",
	granio.KeyTab:       "⇥",
	granio.KeyUp:        "↑",
}
