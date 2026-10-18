//go:build accessibility && windows && !wasm && !test_web_driver

package glfw

/*
#cgo LDFLAGS: -lole32 -loleaut32 -lcomctl32
#include <stdlib.h>
#include <windows.h>
#include <uiautomationclient.h>
#include "accessibility_windows.h"
*/
import "C"

import (
	"fmt"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf16"
	"unsafe"
)

type winAccessibility struct {
	mu      sync.RWMutex
	objects map[uint64]granio.CanvasObject
	ids     map[accessibleKey]uint64
	nodes   map[uint64]accessibleNode
}

var winAccessibleWindows sync.Map // HWND -> *winAccessibility
var winAccessibleNext atomic.Uint64

func (w *window) updateAccessibility() {
	if w.viewport == nil {
		return
	}
	hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	stored, _ := winAccessibleWindows.LoadOrStore(hwnd, &winAccessibility{objects: make(map[uint64]granio.CanvasObject), ids: make(map[accessibleKey]uint64)})
	state := stored.(*winAccessibility)
	nodes := w.accessibilitySnapshot()
	ids := make([]uint64, len(nodes))
	state.mu.Lock()
	previousNodes := state.nodes
	newIDs, newObjects := make(map[accessibleKey]uint64, len(nodes)), make(map[uint64]granio.CanvasObject, len(nodes))
	newNodes := make(map[uint64]accessibleNode, len(nodes))
	for i, node := range nodes {
		key := nodeKey(node)
		ids[i] = state.ids[key]
		if ids[i] == 0 {
			ids[i] = winAccessibleNext.Add(1)
		}
		newIDs[key], newObjects[ids[i]], newNodes[ids[i]] = ids[i], node.object, node
	}
	state.ids, state.objects = newIDs, newObjects
	state.nodes = newNodes
	state.mu.Unlock()
	C.WinAccessibilityBegin(C.uintptr_t(hwnd))
	for i, node := range nodes {
		parent := uint64(0)
		if node.parent >= 0 {
			parent = ids[node.parent]
		}
		if old, ok := previousNodes[ids[i]]; ok && nativeNodeEqual(old, node) {
			C.WinAccessibilityKeep(C.uintptr_t(hwnd), C.uint64_t(ids[i]), C.uint64_t(parent))
			continue
		}
		name, value := C.CString(node.name), C.CString(node.value)
		flags := 0
		if node.state.Disabled {
			flags |= 1
		}
		if node.state.Focused {
			flags |= 2
		}
		if node.state.Checked {
			flags |= 4
		}
		if node.state.Selected {
			flags |= 8
		}
		if node.state.Expanded {
			flags |= 16
		}
		if node.state.Sensitive {
			flags |= 32
		}
		if node.state.ReadOnly {
			flags |= 64
		}
		if node.offscreen {
			flags |= 128
		}
		patterns := 0
		for _, action := range node.actions {
			switch action {
			case "click":
				patterns |= 1
			case "toggle":
				if node.role == granio.AccessibleRoleTreeItem {
					patterns |= 16
				} else {
					patterns |= 4
				}
			case "select":
				patterns |= 8
			case "focus":
				patterns |= 32
			}
		}
		if _, ok := node.object.(granio.AccessibleEditableText); ok {
			patterns |= 2
			if !node.state.Sensitive {
				patterns |= 64
			}
		}
		if node.numeric {
			patterns |= 128
		}
		if _, ok := node.object.(granio.AccessibleSelection); ok {
			patterns |= 256
		}
		if _, ok := node.object.(granio.AccessibleTable); ok {
			patterns |= 512
		}
		if node.tableOwner != nil {
			patterns |= 1024
		}
		if node.virtualOwner != nil {
			patterns |= 2048
			if node.virtualID != "" {
				patterns |= 4096
			}
		}
		C.WinAccessibilityAdd(C.uintptr_t(hwnd), C.uint64_t(ids[i]), C.uint64_t(parent), name, value, C.int(winAccessibilityRole(node.role)), C.int(flags), C.int(patterns),
			C.double(scale.ToScreenCoordinate(w.canvas, node.position.X)), C.double(scale.ToScreenCoordinate(w.canvas, node.position.Y)), C.double(scale.ToScreenCoordinate(w.canvas, node.size.Width)), C.double(scale.ToScreenCoordinate(w.canvas, node.size.Height)))
		C.free(unsafe.Pointer(name))
		C.free(unsafe.Pointer(value))
		runes := []rune(node.value)
		acp := func(offset int) C.int { return C.int(len(utf16.Encode(runes[:min(max(offset, 0), len(runes))]))) }
		C.WinAccessibilityDetails(C.uintptr_t(hwnd), C.uint64_t(ids[i]), acp(node.selectionStart), acp(node.selectionEnd), C.int(node.rows), C.int(node.columns), C.int(node.row), C.int(node.column), C.int(boolInt(node.multiple)), C.double(node.number), C.double(node.minimum), C.double(node.maximum), C.double(node.step))
	}
	C.WinAccessibilityEnd(C.uintptr_t(hwnd))
	if C.WinAccessibilityListening() != 0 {
		for i, node := range nodes {
			old, existed := previousNodes[ids[i]]
			if !existed {
				if node.state.Focused {
					C.WinAccessibilityEvent(C.uintptr_t(hwnd), C.uint64_t(ids[i]), C.UIA_AutomationFocusChangedEventId)
				}
				continue
			}
			raiseWindowsAccessibilityChanges(hwnd, ids[i], old, node)
		}
	}
}

func raiseWindowsAccessibilityChanges(hwnd uintptr, id uint64, old, node accessibleNode) {
	handle, element := C.uintptr_t(hwnd), C.uint64_t(id)
	for _, change := range []struct {
		property C.int
		old, new bool
		boolean  bool
	}{
		{C.UIA_HasKeyboardFocusPropertyId, old.state.Focused, node.state.Focused, true},
		{C.UIA_IsEnabledPropertyId, !old.state.Disabled, !node.state.Disabled, true},
		{C.UIA_ValueIsReadOnlyPropertyId, old.state.ReadOnly, node.state.ReadOnly, true},
		{C.UIA_SelectionItemIsSelectedPropertyId, old.state.Selected, node.state.Selected, true},
		{C.UIA_ToggleToggleStatePropertyId, old.state.Checked, node.state.Checked, false},
		{C.UIA_ExpandCollapseExpandCollapseStatePropertyId, old.state.Expanded, node.state.Expanded, false},
	} {
		if change.old != change.new {
			C.WinAccessibilityBoolProperty(handle, element, change.property, C.int(boolInt(change.old)), C.int(boolInt(change.new)), C.int(boolInt(change.boolean)))
		}
	}
	if !old.state.Focused && node.state.Focused {
		C.WinAccessibilityEvent(handle, element, C.UIA_AutomationFocusChangedEventId)
	}
	if old.name != node.name {
		before, after := C.CString(old.name), C.CString(node.name)
		C.WinAccessibilityTextProperty(handle, element, C.UIA_NamePropertyId, before, after)
		C.free(unsafe.Pointer(before))
		C.free(unsafe.Pointer(after))
	}
	if old.value != node.value && !node.state.Sensitive {
		before, after := C.CString(old.value), C.CString(node.value)
		C.WinAccessibilityTextProperty(handle, element, C.UIA_ValueValuePropertyId, before, after)
		C.free(unsafe.Pointer(before))
		C.free(unsafe.Pointer(after))
		if _, ok := node.object.(granio.AccessibleEditableText); ok {
			C.WinAccessibilityEvent(handle, element, C.UIA_Text_TextChangedEventId)
		}
	}
	if old.selectionStart != node.selectionStart || old.selectionEnd != node.selectionEnd {
		C.WinAccessibilityEvent(handle, element, C.UIA_Text_TextSelectionChangedEventId)
	}
	if old.number != node.number && node.numeric {
		C.WinAccessibilityNumberProperty(handle, element, C.UIA_RangeValueValuePropertyId, C.double(old.number), C.double(node.number))
	}
}
func nativeNodeEqual(a, b accessibleNode) bool {
	return a.name == b.name && a.value == b.value && a.state == b.state && a.position == b.position && a.size == b.size && a.role == b.role && a.selectionStart == b.selectionStart && a.selectionEnd == b.selectionEnd && a.rows == b.rows && a.columns == b.columns && a.row == b.row && a.column == b.column && a.multiple == b.multiple && a.number == b.number && a.minimum == b.minimum && a.maximum == b.maximum && a.step == b.step && a.offscreen == b.offscreen && slices.Equal(a.actions, b.actions)
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
func winAccessibilityRole(role granio.AccessibleRole) int {
	switch role {
	case granio.AccessibleRoleButton:
		return 50000
	case granio.AccessibleRoleCheck:
		return 50002
	case granio.AccessibleRoleSelect:
		return 50003
	case granio.AccessibleRoleInput:
		return 50004
	case granio.AccessibleRoleLink:
		return 50005
	case granio.AccessibleRoleImage:
		return 50006
	case granio.AccessibleRoleListItem:
		return 50007
	case granio.AccessibleRoleList:
		return 50008
	case granio.AccessibleRoleProgress:
		return 50012
	case granio.AccessibleRoleSlider:
		return 50015
	case granio.AccessibleRoleText:
		return 50020
	case granio.AccessibleRoleTree:
		return 50023
	case granio.AccessibleRoleTreeItem:
		return 50024
	case granio.AccessibleRoleTable:
		return 50036
	case granio.AccessibleRoleCell:
		return 50029
	case granio.AccessibleRoleWindow:
		return 50032
	default:
		return 50026
	}
}
func (*window) initAccessibilityForWindow() {}
func (w *window) cleanupAccessibilityForWindow() {
	if w.viewport == nil {
		return
	}
	hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	winAccessibleWindows.Delete(hwnd)
	C.WinAccessibilityCleanup(C.uintptr_t(hwnd))
	virtualAccessibilityWindows.Delete(w)
}

//export goWinAccessibilityAction
func goWinAccessibilityAction(hwnd C.uintptr_t, id C.uint64_t, operation C.int, text *C.ushort) {
	stored, ok := winAccessibleWindows.Load(uintptr(hwnd))
	if !ok {
		return
	}
	state := stored.(*winAccessibility)
	state.mu.RLock()
	object := state.objects[uint64(id)]
	state.mu.RUnlock()
	if object == nil {
		return
	}
	var value string
	if text != nil {
		units := unsafe.Slice((*uint16)(unsafe.Pointer(text)), 1<<20)
		n := 0
		for n < len(units) && units[n] != 0 {
			n++
		}
		value = string(utf16.Decode(units[:n]))
	}
	queueAccessibility(func() {
		state.mu.RLock()
		current := state.objects[uint64(id)]
		virtualNode := state.nodes[uint64(id)]
		state.mu.RUnlock()
		if virtualNode.virtualOwner != nil && virtualNode.virtualID != "" {
			action := "focus"
			switch operation {
			case 8:
				action = "select"
			case 9:
				action = "add-selection"
			case 10:
				action = "remove-selection"
			case 16:
				action = "toggle"
			case 512:
				action = "realize"
			}
			virtualNode.virtualOwner.(granio.AccessibleVirtualCollection).AccessibilityVirtualAction(virtualNode.virtualID, action)
			if canvas := stateCanvas(uintptr(hwnd)); canvas != nil {
				canvas.context.(*window).updateAccessibility()
			}
			return
		}
		if current != object {
			return
		}
		if operation == 2 {
			if editable, ok := object.(granio.AccessibleEditableText); ok {
				editable.AccessibilitySetText(value)
			}
			return
		}
		state.mu.RLock()
		node := state.nodes[uint64(id)]
		state.mu.RUnlock()
		if operation == 256 {
			if scroller, ok := object.(granio.AccessibleTextScroller); ok {
				parts := strings.Split(value, ":")
				if len(parts) == 3 {
					a, errA := strconv.Atoi(parts[0])
					b, errB := strconv.Atoi(parts[1])
					units := utf16.Encode([]rune(node.value))
					if errA == nil && errB == nil {
						scroller.AccessibilityScrollTextRange(utf16RuneOffset(units, a), utf16RuneOffset(units, b), parts[2] == "1")
					}
				}
			}
			return
		}
		if operation == 64 {
			if editable, ok := object.(granio.AccessibleEditableText); ok {
				parts := strings.Split(value, ":")
				if len(parts) == 2 {
					a, _ := strconv.Atoi(parts[0])
					b, _ := strconv.Atoi(parts[1])
					units := utf16.Encode([]rune(node.value))
					editable.AccessibilitySetSelection(utf16RuneOffset(units, a), utf16RuneOffset(units, b))
				}
			}
			return
		}
		if operation == 128 {
			if numeric, ok := object.(granio.AccessibleNumeric); ok {
				number, err := strconv.ParseFloat(value, 64)
				if err == nil {
					numeric.AccessibilitySetNumericValue(number)
				}
			}
			return
		}
		if node.tableOwner != nil {
			if selection, ok := node.tableOwner.(granio.AccessibleSelection); ok {
				_, cols := node.tableOwner.(granio.AccessibleTable).AccessibilityTableSize()
				selection.AccessibilitySelectIndex(node.row*cols+node.column, operation != 10, operation == 9)
			}
			return
		}
		action := "focus"
		switch operation {
		case 1:
			action = "click"
		case 4:
			action = "toggle"
		case 8:
			action = "select"
		case 9:
			action = "add-selection"
		case 10:
			action = "remove-selection"
		case 16:
			action = "toggle"
		}
		if accessible, ok := object.(granio.AccessibleAction); ok {
			accessible.AccessibilityPerformAction(action)
		}
	})
}

//export goWinTextBoundary
func goWinTextBoundary(text *C.ushort, length, unit, position, count C.int, moved *C.int) C.int {
	if length < 0 || length > 1<<20 {
		return 0
	}
	runes := utf16.Decode(unsafe.Slice((*uint16)(unsafe.Pointer(text)), int(length)))
	kind := 4
	switch unit {
	case 0:
		kind = 0
	case 1, 2:
		kind = 1
	case 3:
		kind = 2
	case 4:
		kind = 3
	}
	indices := accessibleTextBoundaries(string(runes), kind)
	acp := make([]int, len(indices))
	cursor := 0
	for i, offset := range indices {
		for cursor < offset {
			if runes[cursor] > 0xffff {
				acp[i] += 2
			} else {
				acp[i]++
			}
			cursor++
		}
		if i > 0 {
			acp[i] += acp[i-1]
		}
	}
	p := min(max(0, int(position)), int(length))
	index := sort.SearchInts(acp, p)
	if count == 0 {
		if index == len(acp) || acp[index] > p {
			index = max(0, index-1)
		}
		return C.int(acp[index])
	}
	if count < 0 && index < len(acp) && acp[index] > p {
		index--
	}
	next := min(max(0, index+int(count)), len(acp)-1)
	if moved != nil {
		*moved = C.int(next - index)
	}
	return C.int(acp[next])
}

//export goWinAccessibilityRectangles
func goWinAccessibilityRectangles(hwnd C.uintptr_t, id C.uint64_t, start, end C.int, out *C.double, capacity C.int) C.int {
	stored, ok := winAccessibleWindows.Load(uintptr(hwnd))
	if !ok || capacity <= 0 {
		return 0
	}
	state := stored.(*winAccessibility)
	count := 0
	accessibilityUI(func() {
		state.mu.RLock()
		node, ok := state.nodes[uint64(id)]
		state.mu.RUnlock()
		if !ok || node.state.Sensitive {
			return
		}
		geometry, ok := node.object.(granio.AccessibleTextGeometry)
		if !ok {
			return
		}
		units := utf16.Encode([]rune(node.value))
		rects := geometry.AccessibilityTextRectangles(utf16RuneOffset(units, int(start)), utf16RuneOffset(units, int(end)))
		buffer := unsafe.Slice(out, int(capacity)*4)
		count = min(len(rects), int(capacity))
		canvas := stateCanvas(uintptr(hwnd))
		if canvas == nil {
			return
		}
		for i, r := range rects[:count] {
			buffer[i*4] = C.double(scale.ToScreenCoordinate(canvas, node.position.X+r.Position.X))
			buffer[i*4+1] = C.double(scale.ToScreenCoordinate(canvas, node.position.Y+r.Position.Y))
			buffer[i*4+2] = C.double(scale.ToScreenCoordinate(canvas, r.Size.Width))
			buffer[i*4+3] = C.double(scale.ToScreenCoordinate(canvas, r.Size.Height))
		}
	})
	return C.int(count)
}
func stateCanvas(hwnd uintptr) *glCanvas {
	if stored, ok := imeWindows.Load(hwnd); ok {
		return stored.(*winIME).window.canvas
	}
	return nil
}

//export goWinAccessibilityPoint
func goWinAccessibilityPoint(hwnd C.uintptr_t, id C.uint64_t, x, y C.double) C.int {
	stored, ok := winAccessibleWindows.Load(uintptr(hwnd))
	if !ok {
		return 0
	}
	state := stored.(*winAccessibility)
	offset := 0
	accessibilityUI(func() {
		state.mu.RLock()
		node, ok := state.nodes[uint64(id)]
		state.mu.RUnlock()
		if !ok {
			return
		}
		geometry, ok := node.object.(granio.AccessibleTextGeometry)
		canvas := stateCanvas(uintptr(hwnd))
		if !ok || canvas == nil {
			return
		}
		pos := granio.NewPos(float32(x)/canvas.Scale()-node.position.X, float32(y)/canvas.Scale()-node.position.Y)
		runes := []rune(node.value)
		index := min(max(0, geometry.AccessibilityTextOffsetAtPosition(pos)), len(runes))
		offset = len(utf16.Encode(runes[:index]))
	})
	return C.int(offset)
}

//export goWinAccessibilityGridItem
func goWinAccessibilityGridItem(hwnd C.uintptr_t, id C.uint64_t, row, column C.int) C.uint64_t {
	stored, ok := winAccessibleWindows.Load(uintptr(hwnd))
	if !ok {
		return 0
	}
	state := stored.(*winAccessibility)
	result := uint64(0)
	accessibilityUI(func() {
		state.mu.RLock()
		node, ok := state.nodes[uint64(id)]
		state.mu.RUnlock()
		if !ok {
			return
		}
		table, ok := node.object.(granio.AccessibleTable)
		if !ok {
			return
		}
		if _, virtual := node.object.(granio.AccessibleVirtualCollection); virtual {
			rows, cols := table.AccessibilityTableSize()
			if row < 0 || column < 0 || int(row) >= rows || int(column) >= cols {
				return
			}
			canvas := stateCanvas(uintptr(hwnd))
			if canvas == nil {
				return
			}
			w := canvas.context.(*window)
			id := fmt.Sprintf("%d:%d", row, column)
			if !w.requestVirtual(node.object, id) {
				return
			}
			w.updateAccessibility()
			state.mu.RLock()
			result = state.ids[virtualKey(node.object, id)]
			state.mu.RUnlock()
			return
		}
		cell := table.AccessibilityTableCell(int(row), int(column))
		canvas := stateCanvas(uintptr(hwnd))
		if cell == nil || canvas == nil {
			return
		}
		window := canvas.context.(*window)
		window.updateAccessibility()
		state.mu.RLock()
		defer state.mu.RUnlock()
		for key, current := range state.nodes {
			if current.tableOwner == node.object && current.row == int(row) && current.column == int(column) {
				result = key
				break
			}
		}
	})
	return C.uint64_t(result)
}
func windowsAccessibilityContractTest() bool     { return C.WinAccessibilityContractTest() != 0 }
func windowsAccessibilityTextContractTest() bool { return C.WinAccessibilityTextContractTest() != 0 }
func windowsAccessibilityEventsContractTest() bool {
	return C.WinAccessibilityEventsContractTest() != 0
}
func windowsAccessibilityScrollRangeTest(hwnd uintptr, id uint64, start, end int) bool {
	return C.WinAccessibilityScrollTextRange(C.uintptr_t(hwnd), C.uint64_t(id), C.int(start), C.int(end), 1) != 0
}
func windowsAccessibilitySelectionContractTest(hwnd uintptr, id, child uint64) bool {
	return C.WinAccessibilitySelectionContractTest(C.uintptr_t(hwnd), C.uint64_t(id), C.uint64_t(child)) != 0
}
func windowsAccessibilityVirtualContractTest(hwnd uintptr, id uint64) bool {
	return C.WinAccessibilityVirtualContractTest(C.uintptr_t(hwnd), C.uint64_t(id)) != 0
}
