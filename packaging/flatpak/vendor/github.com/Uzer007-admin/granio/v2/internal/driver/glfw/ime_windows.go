//go:build windows && !wasm && !test_web_driver

package glfw

import (
	"sync"
	"syscall"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
	"golang.org/x/sys/windows"
)

var (
	imeUser32         = windows.NewLazySystemDLL("user32.dll")
	imeSetWindowProc  = imeUser32.NewProc("SetWindowLongPtrW")
	imeCallWindowProc = imeUser32.NewProc("CallWindowProcW")
	imeImm32          = windows.NewLazySystemDLL("imm32.dll")
	imeGetContext     = imeImm32.NewProc("ImmGetContext")
	imeReleaseContext = imeImm32.NewProc("ImmReleaseContext")
	imeGetComposition = imeImm32.NewProc("ImmGetCompositionStringW")
	imeSetCandidate   = imeImm32.NewProc("ImmSetCandidateWindow")
	imeSetComposition = imeImm32.NewProc("ImmSetCompositionWindow")
	imeNotify         = imeImm32.NewProc("ImmNotifyIME")
	imeAssociate      = imeImm32.NewProc("ImmAssociateContext")
	imeSetScope       = windows.NewLazySystemDLL("msctf.dll").NewProc("SetInputScope")
	imeWindowCallback = syscall.NewCallback(imeWindowProc)
	imeWindows        sync.Map // HWND -> *winIME; creation/destruction occur on the UI thread
)

type winIME struct {
	window          *window
	previous        uintptr
	receiver        desktop.TextInput
	active          bool
	passwordContext uintptr
	scope           uint32
	tsf             bool
	tsfReceiver     desktop.TextInput
	tsfComposing    bool
	preTranslate    bool
}
type imePoint struct{ X, Y int32 }
type imeRect struct{ Left, Top, Right, Bottom int32 }
type imeCandidateForm struct {
	Index, Style uint32
	Point        imePoint
	Area         imeRect
}
type imeCompositionForm struct {
	Style uint32
	Point imePoint
	Area  imeRect
}

func (w *window) initInputMethod() {
	if w.viewport == nil {
		return
	}
	if err := imeSetWindowProc.Find(); err != nil {
		imeSetWindowProc = imeUser32.NewProc("SetWindowLongW")
	}
	hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	state := &winIME{window: w}
	imeWindows.Store(hwnd, state)
	previous, _, _ := imeSetWindowProc.Call(hwnd, ^uintptr(3), imeWindowCallback)
	if previous == 0 {
		imeWindows.Delete(hwnd)
		return
	}
	state.previous = previous
	state.tsf = createTSF(hwnd)
	if state.tsf {
		state.preTranslate = true
		w.viewport.SetWin32MessageFilter(func(message uint32, key, flags uintptr) bool { return tsfKey(hwnd, message, key, flags) })
	}
}

func (w *window) resetInputMethod() {
	if w.viewport == nil {
		return
	}
	hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	if stored, ok := imeWindows.Load(hwnd); ok {
		state := stored.(*winIME)
		if state.receiver != nil {
			state.receiver.CancelComposition()
		}
		state.active, state.receiver = false, nil
		resetTSF(hwnd)
	}
	context, _, _ := imeGetContext.Call(hwnd)
	if context != 0 {
		imeNotify.Call(context, 0x0015, 4, 0)
		imeReleaseContext.Call(hwnd, context)
	}
}

func (w *window) updateInputMethod() {
	if w.viewport == nil {
		return
	}
	hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	if stored, ok := imeWindows.Load(hwnd); ok {
		state := stored.(*winIME)
		scope := uint32(0)
		var receiver desktop.TextInput
		if input, ok := w.canvas.Focused().(desktop.TextInput); ok {
			receiver = input
			if input.TextInputState().Sensitive {
				scope = 31
			}
			state.placeCandidate(hwnd, input)
		}
		state.updateTSF(hwnd, receiver)
		if scope != state.scope {
			state.scope = scope
			if imeSetScope.Find() == nil {
				imeSetScope.Call(hwnd, uintptr(scope))
			}
			if scope == 31 {
				state.passwordContext, _, _ = imeAssociate.Call(hwnd, 0)
			} else if state.passwordContext != 0 {
				imeAssociate.Call(hwnd, state.passwordContext)
				state.passwordContext = 0
			}
		}
	}
}

func imeWindowProc(hwnd, message, wparam, lparam uintptr) uintptr {
	stored, found := imeWindows.Load(hwnd)
	if !found {
		return 0
	}
	state := stored.(*winIME)
	w := state.window
	input, accepts := w.canvas.Focused().(desktop.TextInput)
	if state.tsfComposing && (message == 0x010d || message == 0x010e || message == 0x010f || message == 0x0286) {
		return 0
	}
	keyMessage := message == 0x100 || message == 0x101 || message == 0x104 || message == 0x105
	if state.tsf && !(state.preTranslate && keyMessage) && tsfKey(hwnd, uint32(message), wparam, lparam) {
		return 0
	}
	switch message {
	case 0x0288: // WM_IME_REQUEST: legacy engines request committed reconversion/context
		if accepts {
			return state.reconversionRequest(input, wparam, lparam)
		}
	case 0x010d: // WM_IME_STARTCOMPOSITION
		if accepts {
			state.receiver, state.active = input, true
			state.placeCandidate(hwnd, input)
			return 0
		}
	case 0x010f: // WM_IME_COMPOSITION
		if accepts {
			state.receiver, state.active = input, true
			context, _, _ := imeGetContext.Call(hwnd)
			if context != 0 {
				defer imeReleaseContext.Call(hwnd, context)
				if lparam&0x0800 != 0 {
					input.CommitComposition(string(utf16.Decode(readIMEString(context, 0x0800))))
				}
				if lparam&0x0008 != 0 {
					text := readIMEString(context, 0x0008)
					cursor, _, _ := imeGetComposition.Call(context, 0x0080, 0, 0)
					offset := int(int32(cursor))
					attributes := readIMEBytes(context, 0x0010)
					start, end := 0, 0
					found := false
					for i, value := range attributes {
						if value == 1 || value == 3 {
							if !found {
								start = i
								found = true
							}
							end = i + 1
						}
					}
					input.UpdateComposition(desktop.CompositionEvent{Text: string(utf16.Decode(text)), Cursor: utf16RuneOffset(text, offset), SelectionStart: utf16RuneOffset(text, start), SelectionEnd: utf16RuneOffset(text, end)})
				} else if lparam == 0 {
					input.CancelComposition()
				}
			}
			state.placeCandidate(hwnd, input)
			return 0
		}
	case 0x010e: // WM_IME_ENDCOMPOSITION
		if state.active {
			if state.receiver != nil {
				state.receiver.CancelComposition()
			}
			state.active, state.receiver = false, nil
			return 0
		}
	case 0x0286: // WM_IME_CHAR: committed text is already handled above
		if accepts {
			return 0
		}
	case 0x0281: // WM_IME_SETCONTEXT: Fyne displays preedit, OS displays candidates
		if accepts {
			lparam &^= 0x80000000
		}
	case 0x0082: // WM_NCDESTROY
		w.viewport.SetWin32MessageFilter(nil)
		destroyTSF(hwnd)
		if state.passwordContext != 0 {
			imeAssociate.Call(hwnd, state.passwordContext)
		}
		imeWindows.Delete(hwnd)
	}
	result, _, _ := imeCallWindowProc.Call(state.previous, hwnd, message, wparam, lparam)
	return result
}

func readIMEString(context, field uintptr) []uint16 {
	bytes, _, _ := imeGetComposition.Call(context, field, 0, 0)
	if int32(bytes) <= 0 || bytes > 1<<20 || bytes%2 != 0 {
		return nil
	}
	text := make([]uint16, int(bytes)/2)
	read, _, _ := imeGetComposition.Call(context, field, uintptr(unsafe.Pointer(&text[0])), bytes)
	if int32(read) < 0 {
		return nil
	}
	return text[:min(len(text), int(read)/2)]
}

func readIMEBytes(context, field uintptr) []byte {
	size, _, _ := imeGetComposition.Call(context, field, 0, 0)
	if int32(size) <= 0 || size > 1<<20 {
		return nil
	}
	value := make([]byte, int(size))
	read, _, _ := imeGetComposition.Call(context, field, uintptr(unsafe.Pointer(&value[0])), size)
	if int32(read) < 0 {
		return nil
	}
	return value[:min(len(value), int(read))]
}

func utf16RuneOffset(text []uint16, offset int) int {
	offset = min(max(0, offset), len(text))
	if offset > 0 && offset < len(text) && text[offset-1] >= 0xd800 && text[offset-1] <= 0xdbff && text[offset] >= 0xdc00 && text[offset] <= 0xdfff {
		offset--
	}
	return utf8.RuneCountInString(string(utf16.Decode(text[:offset])))
}

func (state *winIME) placeCandidate(hwnd uintptr, input desktop.TextInput) {
	object, ok := input.(granio.CanvasObject)
	if !ok {
		return
	}
	text := input.TextInputState()
	w := state.window
	position := w.driver.AbsolutePositionForObject(object).Add(text.CaretPosition)
	x, y := scale.ToScreenCoordinate(w.canvas, position.X), scale.ToScreenCoordinate(w.canvas, position.Y)
	height := max(1, scale.ToScreenCoordinate(w.canvas, text.CaretSize.Height))
	context, _, _ := imeGetContext.Call(hwnd)
	if context == 0 {
		return
	}
	defer imeReleaseContext.Call(hwnd, context)
	candidate := imeCandidateForm{Style: 0x0080, Point: imePoint{int32(x), int32(y + height)}, Area: imeRect{int32(x), int32(y), int32(x + 1), int32(y + height)}}
	imeSetCandidate.Call(context, uintptr(unsafe.Pointer(&candidate)))
	composition := imeCompositionForm{Style: 0x0002, Point: imePoint{int32(x), int32(y)}}
	imeSetComposition.Call(context, uintptr(unsafe.Pointer(&composition)))
}
