package desktop

import (
	"runtime"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/goos"
)

// Declare conformity with Shortcut interface
var (
	_ granio.Shortcut         = (*CustomShortcut)(nil)
	_ granio.KeyboardShortcut = (*CustomShortcut)(nil)
)

// CustomShortcut describes a shortcut desktop event.
type CustomShortcut struct {
	granio.KeyName
	Modifier granio.KeyModifier
}

// Key returns the key name of this shortcut.
func (cs *CustomShortcut) Key() granio.KeyName {
	return cs.KeyName
}

// Mod returns the modifier of this shortcut.
func (cs *CustomShortcut) Mod() granio.KeyModifier {
	return cs.Modifier
}

// ShortcutName returns the shortcut name associated to the event.
func (cs *CustomShortcut) ShortcutName() string {
	id := &strings.Builder{}
	id.WriteString("CustomDesktop:")
	writeModifiers(id, cs.Modifier)
	id.WriteString(string(cs.KeyName))
	return id.String()
}

func writeModifiers(w *strings.Builder, mods granio.KeyModifier) {
	if (mods & granio.KeyModifierShift) != 0 {
		w.WriteString("Shift+")
	}
	if (mods & granio.KeyModifierControl) != 0 {
		w.WriteString("Control+")
	}
	if (mods & granio.KeyModifierAlt) != 0 {
		w.WriteString("Alt+")
	}
	if (mods & granio.KeyModifierSuper) != 0 {
		if runtime.GOOS == goos.Darwin {
			w.WriteString("Command+")
		} else {
			w.WriteString("Super+")
		}
	}
}
