//go:build windows && !wasm && !test_web_driver

package glfw

import (
	"encoding/binary"
	"unicode"
	"unicode/utf16"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2/driver/desktop"
)

// RECONVERTSTRING stores UTF-16 lengths in code units, but all three offsets
// are byte offsets. Bound the surrounding context independently of document size.
func reconversionPacket(state desktop.TextInputState, documentFeed bool) ([]byte, int) {
	if state.Sensitive {
		return nil, 0
	}
	runes := []rune(state.Text)
	a, b := min(max(0, state.SelectionStart), len(runes)), min(max(0, state.SelectionEnd), len(runes))
	if b < a {
		a, b = b, a
	}
	if a == b && !documentFeed {
		word := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) }
		for a > 0 && word(runes[a-1]) {
			a--
		}
		for b < len(runes) && word(runes[b]) {
			b++
		}
	}
	base, limit := max(0, a-512), min(len(runes), b+512)
	if limit-base > 4096 {
		return nil, 0
	}
	units := utf16.Encode(runes[base:limit])
	start := len(utf16.Encode(runes[base:a]))
	count := len(utf16.Encode(runes[a:b]))
	packet := make([]byte, 32+2*(len(units)+1))
	fields := []uint32{uint32(len(packet)), 0, uint32(len(units)), 32, uint32(count), uint32(start * 2), uint32(count), uint32(start * 2)}
	for i, value := range fields {
		binary.LittleEndian.PutUint32(packet[i*4:], value)
	}
	for i, unit := range units {
		binary.LittleEndian.PutUint16(packet[32+i*2:], unit)
	}
	return packet, base
}

func (state *winIME) reconversionRequest(input desktop.TextInput, request, pointer uintptr) uintptr {
	text := input.TextInputState()
	if text.Sensitive {
		return 0
	}
	if request == 4 || request == 7 { // IMR_RECONVERTSTRING / IMR_DOCUMENTFEED
		packet, _ := reconversionPacket(text, request == 7)
		if len(packet) == 0 {
			return 0
		}
		if pointer == 0 {
			return uintptr(len(packet))
		}
		capacity := *(*uint32)(unsafe.Pointer(pointer))
		if capacity < uint32(len(packet)) || capacity > 1<<20 {
			return 0
		}
		copy(unsafe.Slice((*byte)(unsafe.Pointer(pointer)), len(packet)), packet)
		return uintptr(len(packet))
	}
	if request != 5 || pointer == 0 {
		return 0
	} // IMR_CONFIRMRECONVERTSTRING
	capacity := *(*uint32)(unsafe.Pointer(pointer))
	if capacity < 32 || capacity > 1<<20 {
		return 0
	}
	packet := unsafe.Slice((*byte)(unsafe.Pointer(pointer)), int(capacity))
	length, offset := int(binary.LittleEndian.Uint32(packet[8:])), int(binary.LittleEndian.Uint32(packet[12:]))
	count, start := int(binary.LittleEndian.Uint32(packet[16:])), int(binary.LittleEndian.Uint32(packet[20:]))
	if offset < 32 || offset%2 != 0 || start%2 != 0 || length > (len(packet)-offset)/2 || start/2 > length || count > length-start/2 {
		return 0
	}
	if offset > len(packet) {
		return 0
	}
	units := make([]uint16, length)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(packet[offset+i*2:])
	}
	expected, base := reconversionPacket(text, false)
	if len(expected) == 0 || int(binary.LittleEndian.Uint32(expected[8:])) != length {
		return 0
	}
	for i, unit := range units {
		if binary.LittleEndian.Uint16(expected[32+i*2:]) != unit {
			return 0
		}
	}
	editor, ok := input.(desktop.TextInputRangeEditor)
	if !ok || !editor.StartCompositionRange(base+utf16RuneOffset(units, start/2), base+utf16RuneOffset(units, start/2+count)) {
		return 0
	}
	return 1
}
