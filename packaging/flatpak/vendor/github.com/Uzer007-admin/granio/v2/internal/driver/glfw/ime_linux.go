//go:build linux && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
)

func (w *window) initInputMethod() {
	if w.viewport == nil {
		return
	}
	ibus := w.initIBus()
	w.viewport.SetTextInputCallbacks(glfw.TextInputCallbacks{
		Key: func(sym, code, state uint32, release bool) bool {
			if ibus == nil {
				return false
			}
			ibus.update()
			return ibus.key(sym, code, state, release)
		},
		PreeditSelection: func(text string, cursor, start, end int) {
			if input, ok := w.canvas.Focused().(desktop.TextInput); ok {
				input.UpdateComposition(desktop.CompositionEvent{Text: text, Cursor: cursor, SelectionStart: start, SelectionEnd: end})
			}
		},
		Preedit: func(text string, cursor int) {
			if input, ok := w.canvas.Focused().(desktop.TextInput); ok {
				input.UpdateComposition(desktop.CompositionEvent{Text: text, Cursor: cursor})
				state := input.TextInputState()
				if object, ok := input.(granio.CanvasObject); ok {
					position := w.driver.AbsolutePositionForObject(object).Add(state.CaretPosition)
					w.viewport.SetTextInputCursor(scale.ToScreenCoordinate(w.canvas, position.X), scale.ToScreenCoordinate(w.canvas, position.Y+state.CaretSize.Height))
				}
			}
		},
		Commit: func(text string) {
			if input, ok := w.canvas.Focused().(desktop.TextInput); ok {
				input.CommitComposition(text)
			} else {
				for _, char := range text {
					w.processCharInput(char)
				}
			}
		},
		Edit: func(text string, before, after int) {
			if input, ok := w.canvas.Focused().(desktop.TextInputEditor); ok {
				input.CommitTextEdit(text, before, after)
			} else if input, ok := w.canvas.Focused().(desktop.TextInput); ok {
				input.CommitComposition(text)
			}
		},
		Reset: func() {
			if input, ok := w.canvas.Focused().(desktop.TextInput); ok {
				input.CancelComposition()
			}
		},
	})
}

func (w *window) resetInputMethod() {
	if stored, ok := ibusWindows.Load(w); ok {
		input := stored.(*ibusInput)
		_ = input.call("Reset")
	}
	if w.viewport != nil {
		w.viewport.ResetTextInput()
	}
}

func (w *window) updateInputMethod() {
	if stored, ok := ibusWindows.Load(w); ok {
		stored.(*ibusInput).update()
	}
	if w.viewport == nil {
		return
	}
	input, ok := w.canvas.Focused().(desktop.TextInput)
	if !ok {
		w.viewport.SetTextInputState("", 0, 0, false, false)
		return
	}
	state := input.TextInputState()
	w.viewport.SetTextInputState(state.Text, state.Cursor, state.SelectionStart, state.Sensitive, true)
	if object, ok := input.(granio.CanvasObject); ok {
		position := w.driver.AbsolutePositionForObject(object).Add(state.CaretPosition)
		w.viewport.SetTextInputCursor(scale.ToScreenCoordinate(w.canvas, position.X), scale.ToScreenCoordinate(w.canvas, position.Y+state.CaretSize.Height))
	}
}
