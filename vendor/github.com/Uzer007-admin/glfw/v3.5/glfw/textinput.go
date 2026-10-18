package glfw

/*
#define GLFW_INCLUDE_NONE
#include <stdlib.h>
#include "glfw/include/GLFW/glfw3.h"
int fyneEnableTextInput(GLFWwindow* window);
void fyneDisableTextInput(GLFWwindow* window);
void fyneResetTextInput(GLFWwindow* window);
void fyneTextInputCursor(GLFWwindow* window, int x, int y);
void fyneTextInputSurrounding(GLFWwindow* window, const char* text, int cursor, int anchor, int sensitive, int enabled);
*/
import "C"

import (
	"sync"
	"unicode/utf8"
	"unsafe"
)

// TextInputCallbacks separates X11 input-method preedit from committed text.
// Callbacks run on the event thread. A nil Commit preserves character callbacks.
type TextInputCallbacks struct {
	// Key filters X11 keysyms before ordinary key/text delivery, for native
	// input-method clients whose sessions do not supply an XIM server.
	Key              func(keysym, keycode, state uint32, release bool) bool
	Preedit          func(string, int)
	PreeditSelection func(string, int, int, int)
	Commit           func(string)
	Edit             func(string, int, int)
	Reset            func()
}

type textInputState struct {
	callbacks                          TextInputCallbacks
	text                               []rune
	cursor                             int
	active, pending                    bool
	surrounding                        string
	surroundingCursor                  int
	previousText                       string
	previousCursor, previousAnchor     int
	previousSensitive, previousEnabled bool
	selectionStart, selectionEnd       int
}

var textInputs sync.Map // *Window -> *textInputState; state is event-thread-owned

//export goFyneTextInputKey
func goFyneTextInputKey(window *C.GLFWwindow, sym, code, state C.uint, release C.int) C.int {
	input := lookupTextInput(window)
	if input != nil && input.callbacks.Key != nil && input.callbacks.Key(uint32(sym), uint32(code), uint32(state), release != 0) {
		return 1
	}
	return 0
}

// SetTextInputCallbacks enables inline XIM preedit when the server supports it.
// False retains GLFW's native input-method fallback and ordinary text input.
func (w *Window) SetTextInputCallbacks(callbacks TextInputCallbacks) bool {
	textInputs.Store(w, &textInputState{callbacks: callbacks})
	if C.fyneEnableTextInput(w.data) == 0 {
		return false
	}
	return true
}

// SetTextInputCursor positions the XIM candidate/preedit area in client pixels.
func (w *Window) SetTextInputCursor(x, y int) { C.fyneTextInputCursor(w.data, C.int(x), C.int(y)) }

// SetTextInputState supplies surrounding text and rune offsets to native IME.
// Password text is never transmitted. Call when widget focus or text changes.
func (w *Window) SetTextInputState(text string, cursor, anchor int, sensitive, enabled bool) {
	stored, ok := textInputs.Load(w)
	if !ok {
		return
	}
	input := stored.(*textInputState)
	if sensitive {
		text = ""
		cursor = 0
		anchor = 0
	}
	if text == input.previousText && cursor == input.previousCursor && anchor == input.previousAnchor && sensitive == input.previousSensitive && enabled == input.previousEnabled {
		return
	}
	input.previousText, input.previousCursor, input.previousAnchor, input.previousSensitive, input.previousEnabled = text, cursor, anchor, sensitive, enabled
	runes := []rune(text)
	cursor = min(max(cursor, 0), len(runes))
	anchor = min(max(anchor, 0), len(runes))
	// text-input-v3 permits at most 4000 bytes of surrounding text.
	start, end := max(0, cursor-500), min(len(runes), cursor+500)
	anchor = min(max(anchor, start), end)
	text = string(runes[start:end])
	cursor -= start
	anchor -= start
	input.surrounding, input.surroundingCursor = text, cursor
	bytesCursor, bytesAnchor := len(string([]rune(text)[:cursor])), len(string([]rune(text)[:anchor]))
	value := C.CString(text)
	defer C.free(unsafe.Pointer(value))
	secret, active := 0, 0
	if sensitive {
		secret = 1
	}
	if enabled {
		active = 1
	}
	C.fyneTextInputSurrounding(w.data, value, C.int(bytesCursor), C.int(bytesAnchor), C.int(secret), C.int(active))
}

// ResetTextInput cancels native preedit before focus moves to another widget.
func (w *Window) ResetTextInput() {
	C.fyneResetTextInput(w.data)
	if stored, ok := textInputs.Load(w); ok {
		state := stored.(*textInputState)
		state.text, state.active, state.pending = nil, false, false
	}
}

func disposeTextInput(w *Window) { textInputs.Delete(w); C.fyneDisableTextInput(w.data) }

func lookupTextInput(window *C.GLFWwindow) *textInputState {
	w := windows.get(window)
	if w == nil {
		return nil
	}
	state, found := textInputs.Load(w)
	if !found {
		return nil
	}
	return state.(*textInputState)
}

//export goFyneTextInputActive
func goFyneTextInputActive(window *C.GLFWwindow) C.int {
	if state := lookupTextInput(window); state != nil && (state.active || state.pending) {
		return 1
	}
	return 0
}

//export goFyneWaylandBatch
func goFyneWaylandBatch(window *C.GLFWwindow, preedit *C.char, begin, end C.int, commit *C.char, before, after C.uint) {
	state := lookupTextInput(window)
	if state == nil {
		return
	}
	if commit != nil || before != 0 || after != 0 {
		value := ""
		if commit != nil {
			value = C.GoString(commit)
		}
		cursor := len(string([]rune(state.surrounding)[:state.surroundingCursor]))
		left := max(0, cursor-int(before))
		right := min(len(state.surrounding), cursor+int(after))
		removeBefore, removeAfter := utf8.RuneCountInString(state.surrounding[left:cursor]), utf8.RuneCountInString(state.surrounding[cursor:right])
		if state.callbacks.Edit != nil {
			state.callbacks.Edit(value, removeBefore, removeAfter)
		} else if state.callbacks.Commit != nil {
			state.callbacks.Commit(value)
		}
		state.text, state.active, state.pending = nil, false, false
	}
	if preedit != nil {
		text := C.GoString(preedit)
		state.text = []rune(text)
		state.cursor = utf8.RuneCountInString(text[:min(max(0, int(begin)), len(text))])
		state.selectionStart = state.cursor
		state.selectionEnd = utf8.RuneCountInString(text[:min(max(0, int(end)), len(text))])
		if state.selectionEnd < state.selectionStart {
			state.selectionStart, state.selectionEnd = state.selectionEnd, state.selectionStart
		}
		state.active = len(text) > 0
		if state.active {
			state.emitPreedit()
		} else if state.callbacks.Reset != nil {
			state.callbacks.Reset()
		}
	}
}

func (state *textInputState) emitPreedit() {
	if state.callbacks.PreeditSelection != nil {
		state.callbacks.PreeditSelection(string(state.text), state.cursor, state.selectionStart, state.selectionEnd)
	} else if state.callbacks.Preedit != nil {
		state.callbacks.Preedit(string(state.text), state.cursor)
	}
}

//export goFyneTextInputRegistered
func goFyneTextInputRegistered(window *C.GLFWwindow) C.int {
	if lookupTextInput(window) != nil {
		return 1
	}
	return 0
}

//export goFynePreeditStart
func goFynePreeditStart(window *C.GLFWwindow) {
	if state := lookupTextInput(window); state != nil {
		state.text, state.cursor, state.active, state.pending = nil, 0, true, false
	}
}

//export goFynePreeditDone
func goFynePreeditDone(window *C.GLFWwindow) {
	if state := lookupTextInput(window); state != nil {
		state.active, state.pending = false, state.active || len(state.text) > 0
		if state.callbacks.Reset != nil {
			state.callbacks.Reset()
		}
	}
}

//export goFynePreeditDraw
func goFynePreeditDraw(window *C.GLFWwindow, first, removed C.int, text *C.char, size, caret, keep C.int) {
	state := lookupTextInput(window)
	if state == nil {
		return
	}
	if keep == 0 {
		start := min(max(0, int(first)), len(state.text))
		end := min(start+max(0, int(removed)), len(state.text))
		inserted := []rune(C.GoStringN(text, size))
		result := make([]rune, 0, len(state.text)-(end-start)+len(inserted))
		result = append(result, state.text[:start]...)
		result = append(result, inserted...)
		result = append(result, state.text[end:]...)
		state.text = result
	}
	state.cursor = min(max(0, int(caret)), len(state.text))
	state.active = true
	state.emitPreedit()
}

//export goFynePreeditCaret
func goFynePreeditCaret(window *C.GLFWwindow, position, direction C.int) C.int {
	state := lookupTextInput(window)
	if state == nil {
		return position
	}
	state.cursor = ximCaret(state.text, state.cursor, int(position), int(direction))
	state.emitPreedit()
	return C.int(state.cursor)
}

//export goFyneCommitText
func goFyneCommitText(window *C.GLFWwindow, text *C.char, size C.int) C.int {
	state := lookupTextInput(window)
	if state == nil || state.callbacks.Commit == nil {
		return 0
	}
	value := C.GoStringN(text, size)
	if !state.active && !state.pending && utf8.RuneCountInString(value) <= 1 {
		return 0
	}
	state.callbacks.Commit(value)
	state.text, state.active, state.pending = nil, false, false
	return 1
}
