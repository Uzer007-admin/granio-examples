//go:build windows && !wasm && !test_web_driver

package glfw

/*
#cgo LDFLAGS: -lole32 -luuid
#include "ime_tsf_windows.h"
*/
import "C"

import (
	"unicode/utf16"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
)

func windowsRealJapaneseStep() int { return int(C.FyneRealJapaneseStep()) }
func windowsRealJapaneseOpen(hwnd uintptr) int64 {
	return int64(C.FyneTSFOpenForTest(C.uintptr_t(hwnd)))
}
func windowsRealJapaneseBegin(hwnd uintptr) int64 {
	return int64(C.FyneRealJapaneseBegin(C.uintptr_t(hwnd)))
}
func windowsRealJapaneseEnd() { C.FyneRealJapaneseEnd() }
func windowsRealIMEKey(hwnd uintptr, key int) bool {
	return C.FyneRealIMEKey(C.uintptr_t(hwnd), C.int(key)) != 0
}

func createTSF(hwnd uintptr) bool { return C.FyneTSFCreate(C.uintptr_t(hwnd)) != 0 }
func resetTSF(hwnd uintptr)       { C.FyneTSFReset(C.uintptr_t(hwnd)) }
func destroyTSF(hwnd uintptr)     { C.FyneTSFDestroy(C.uintptr_t(hwnd)) }
func tsfKey(hwnd uintptr, message uint32, key, flags uintptr) bool {
	return C.FyneTSFKey(C.uintptr_t(hwnd), C.uint(message), C.uintptr_t(key), C.intptr_t(flags)) != 0
}

func (state *winIME) updateTSF(hwnd uintptr, input desktop.TextInput) {
	if !state.tsf {
		return
	}
	if state.tsfReceiver != input {
		C.FyneTSFReset(C.uintptr_t(hwnd))
		state.tsfReceiver = input
	}
	if input == nil {
		C.FyneTSFUpdate(C.uintptr_t(hwnd), nil, 0, 0, 0, 0, 0, 0, 0, 1, 1)
		return
	}
	text := input.TextInputState()
	if disabled, ok := input.(granio.Disableable); ok && disabled.Disabled() {
		C.FyneTSFUpdate(C.uintptr_t(hwnd), nil, 0, 0, 0, 0, 0, 0, 0, 1, 1)
		return
	}
	if text.Sensitive {
		C.FyneTSFUpdate(C.uintptr_t(hwnd), nil, 0, 0, 0, 0, 0, 0, 0, 1, 1)
		return
	}
	object, ok := input.(granio.CanvasObject)
	if !ok {
		return
	}
	units := utf16.Encode([]rune(text.Text))
	if len(units) > 1<<19 { // bound native copies; IMM remains available for very large documents
		C.FyneTSFUpdate(C.uintptr_t(hwnd), nil, 0, 0, 0, 0, 0, 0, 0, 1, 1)
		return
	}
	var pointer *C.ushort
	if len(units) > 0 {
		pointer = (*C.ushort)(unsafe.Pointer(&units[0]))
	}
	runes := []rune(text.Text)
	acp := func(offset int) C.int { return C.int(len(utf16.Encode(runes[:min(max(0, offset), len(runes))]))) }
	w := state.window
	position := w.driver.AbsolutePositionForObject(object).Add(text.CaretPosition)
	C.FyneTSFUpdate(C.uintptr_t(hwnd), pointer, C.int(len(units)), acp(text.Cursor), acp(text.SelectionStart), acp(text.SelectionEnd), 1,
		C.int(scale.ToScreenCoordinate(w.canvas, position.X)), C.int(scale.ToScreenCoordinate(w.canvas, position.Y)),
		C.int(max(1, scale.ToScreenCoordinate(w.canvas, text.CaretSize.Width))), C.int(max(1, scale.ToScreenCoordinate(w.canvas, text.CaretSize.Height))))
}

// TSF callbacks execute in the native window's STA while the text-store lock is
// held. They never queue a blocking call back to that same UI thread.
//
//export goWinTSFEdit
func goWinTSFEdit(hwnd C.uintptr_t, kind C.int, original *C.ushort, originalLength C.int,
	text *C.ushort, length, start, end, cursor C.int) {
	stored, ok := imeWindows.Load(uintptr(hwnd))
	if !ok || originalLength < 0 || originalLength > 1<<19 || length < 0 || length > 1<<19 {
		return
	}
	state := stored.(*winIME)
	input := state.tsfReceiver
	if input == nil {
		return
	}
	if kind == 2 {
		state.tsfComposing = false
		input.CancelComposition()
		return
	}
	old := unsafe.Slice((*uint16)(unsafe.Pointer(original)), int(originalLength))
	value := string(utf16.Decode(unsafe.Slice((*uint16)(unsafe.Pointer(text)), int(length))))
	first, last := utf16RuneOffset(old, int(start)), utf16RuneOffset(old, int(end))
	if kind == 1 {
		state.tsfComposing = true
		if !input.CompositionActive() {
			if editor, ok := input.(desktop.TextInputRangeEditor); ok && !editor.StartCompositionRange(first, last) {
				return
			}
		}
		input.UpdateComposition(desktop.CompositionEvent{Text: value, Cursor: utf16RuneOffset(unsafe.Slice((*uint16)(unsafe.Pointer(text)), int(length)), int(cursor))})
		return
	}
	state.tsfComposing = false
	if editor, ok := input.(desktop.TextInputRangeEditor); ok {
		editor.ReplaceTextRange(value, first, last)
	} else {
		input.CommitComposition(value)
	}
}

//export goWinTSFSelection
func goWinTSFSelection(hwnd C.uintptr_t, text *C.ushort, length, start, end C.int) {
	if length < 0 || length > 1<<19 {
		return
	}
	stored, ok := imeWindows.Load(uintptr(hwnd))
	if !ok {
		return
	}
	input := stored.(*winIME).tsfReceiver
	if editor, ok := input.(granio.AccessibleEditableText); ok {
		units := unsafe.Slice((*uint16)(unsafe.Pointer(text)), int(length))
		editor.AccessibilitySetSelection(utf16RuneOffset(units, int(start)), utf16RuneOffset(units, int(end)))
	}
}

func windowsTSFContractTest() bool { return C.FyneTSFContractTest() != 0 }

//export goWinTSFGeometry
func goWinTSFGeometry(hwnd C.uintptr_t, start, end C.int, left, top, right, bottom *C.int) C.int {
	stored, ok := imeWindows.Load(uintptr(hwnd))
	if !ok {
		return 0
	}
	state := stored.(*winIME)
	input := state.tsfReceiver
	if input == nil {
		return 0
	}
	text := input.TextInputState()
	if text.Sensitive {
		return 0
	}
	object, ok := input.(granio.CanvasObject)
	if !ok {
		return 0
	}
	position := state.window.driver.AbsolutePositionForObject(object)
	var bounds granio.AccessibleTextRectangle
	if input.CompositionActive() {
		bounds = granio.AccessibleTextRectangle{Position: text.CaretPosition, Size: text.CaretSize}
	} else if geometry, ok := input.(granio.AccessibleTextGeometry); ok {
		units := utf16.Encode([]rune(text.Text))
		rects := geometry.AccessibilityTextRectangles(utf16RuneOffset(units, int(start)), utf16RuneOffset(units, int(end)))
		if len(rects) == 0 {
			return 0
		}
		bounds = rects[0]
		for _, r := range rects[1:] {
			x, y := min(bounds.Position.X, r.Position.X), min(bounds.Position.Y, r.Position.Y)
			x2, y2 := max(bounds.Position.X+bounds.Size.Width, r.Position.X+r.Size.Width), max(bounds.Position.Y+bounds.Size.Height, r.Position.Y+r.Size.Height)
			bounds.Position, bounds.Size = granio.NewPos(x, y), granio.NewSize(x2-x, y2-y)
		}
	} else {
		bounds = granio.AccessibleTextRectangle{Position: text.CaretPosition, Size: text.CaretSize}
	}
	canvas := state.window.canvas
	position = position.Add(bounds.Position)
	*left = C.int(scale.ToScreenCoordinate(canvas, position.X))
	*top = C.int(scale.ToScreenCoordinate(canvas, position.Y))
	*right = *left + C.int(max(1, scale.ToScreenCoordinate(canvas, bounds.Size.Width)))
	*bottom = *top + C.int(max(1, scale.ToScreenCoordinate(canvas, bounds.Size.Height)))
	return 1
}

//export goWinTSFPoint
func goWinTSFPoint(hwnd C.uintptr_t, x, y C.int) C.int {
	stored, ok := imeWindows.Load(uintptr(hwnd))
	if !ok {
		return -1
	}
	state := stored.(*winIME)
	input := state.tsfReceiver
	geometry, ok := input.(granio.AccessibleTextGeometry)
	if !ok || input.TextInputState().Sensitive {
		return -1
	}
	object, ok := input.(granio.CanvasObject)
	if !ok {
		return -1
	}
	pos := granio.NewPos(scale.ToFyneCoordinate(state.window.canvas, int(x)), scale.ToFyneCoordinate(state.window.canvas, int(y))).Subtract(state.window.driver.AbsolutePositionForObject(object))
	text := []rune(input.TextInputState().Text)
	offset := min(max(0, geometry.AccessibilityTextOffsetAtPosition(pos)), len(text))
	return C.int(len(utf16.Encode(text[:offset])))
}
func windowsTSFEditContractTest(hwnd uintptr) bool {
	return C.FyneTSFEditContractTest(C.uintptr_t(hwnd)) != 0
}
