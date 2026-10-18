//go:build accessibility && windows && !wasm && !test_web_driver

package glfw

/*
#include <stdint.h>
*/
import "C"

import (
	"github.com/Uzer007-admin/granio/v2"
	"unsafe"
)

//export goWinAccessibilityRealize
func goWinAccessibilityRealize(hwnd C.uintptr_t, id C.uint64_t) C.int {
	stored, ok := winAccessibleWindows.Load(uintptr(hwnd))
	if !ok {
		return 0
	}
	state := stored.(*winAccessibility)
	applied := false
	accessibilityUI(func() {
		state.mu.RLock()
		node, ok := state.nodes[uint64(id)]
		state.mu.RUnlock()
		if !ok || node.virtualOwner == nil || node.virtualID == "" {
			return
		}
		applied = node.virtualOwner.(granio.AccessibleVirtualCollection).AccessibilityVirtualAction(node.virtualID, "realize")
		if canvas := stateCanvas(uintptr(hwnd)); canvas != nil {
			canvas.context.(*window).updateAccessibility()
		}
	})
	if applied {
		return 1
	}
	return 0
}

//export goWinAccessibilityVirtualNavigate
func goWinAccessibilityVirtualNavigate(hwnd C.uintptr_t, id C.uint64_t, direction, index C.int) C.uint64_t {
	stored, ok := winAccessibleWindows.Load(uintptr(hwnd))
	if !ok {
		return 0
	}
	state := stored.(*winAccessibility)
	var result uint64
	accessibilityUI(func() {
		state.mu.RLock()
		node, ok := state.nodes[uint64(id)]
		state.mu.RUnlock()
		if !ok || node.virtualOwner == nil {
			return
		}
		canvas := stateCanvas(uintptr(hwnd))
		if canvas == nil {
			return
		}
		w := canvas.context.(*window)
		source := node.virtualOwner.(granio.AccessibleVirtualCollection)
		parent := node.virtualID
		childIndex := int(index)
		switch int(direction) {
		case 3:
			childIndex = 0
		case 4:
			childIndex = source.AccessibilityVirtualCount(parent) - 1
		case 1, 2, 0:
			if node.virtualID == "" {
				return
			}
			item, valid := source.AccessibilityVirtualItem(node.virtualID)
			if !valid {
				return
			}
			parent = item.Parent
			if direction == 0 {
				key := accessibleKey{object: node.virtualOwner}
				if parent != "" {
					if !w.requestVirtual(node.virtualOwner, parent) {
						return
					}
					key = virtualKey(node.virtualOwner, parent)
				}
				w.updateAccessibility()
				state.mu.RLock()
				result = state.ids[key]
				state.mu.RUnlock()
				return
			}
			childIndex = item.Index + 1
			if direction == 2 {
				childIndex = item.Index - 1
			}
		}
		child, valid := source.AccessibilityVirtualChild(parent, childIndex)
		if !valid || !w.requestVirtual(node.virtualOwner, child) {
			return
		}
		w.updateAccessibility()
		state.mu.RLock()
		result = state.ids[virtualKey(node.virtualOwner, child)]
		state.mu.RUnlock()
	})
	return C.uint64_t(result)
}

//export goWinAccessibilitySelected
func goWinAccessibilitySelected(hwnd C.uintptr_t, id C.uint64_t, out *C.uint64_t, capacity C.int) C.int {
	stored, ok := winAccessibleWindows.Load(uintptr(hwnd))
	if !ok {
		return 0
	}
	state := stored.(*winAccessibility)
	count := 0
	accessibilityUI(func() {
		state.mu.RLock()
		node, ok := state.nodes[uint64(id)]
		state.mu.RUnlock()
		if !ok {
			return
		}
		source, ok := node.object.(granio.AccessibleVirtualCollection)
		if !ok {
			return
		}
		ids := source.AccessibilityVirtualSelection()
		count = len(ids)
		if out == nil || capacity <= 0 {
			return
		}
		canvas := stateCanvas(uintptr(hwnd))
		if canvas == nil {
			count = 0
			return
		}
		w := canvas.context.(*window)
		for _, id := range ids {
			w.requestVirtual(node.object, id)
		}
		w.updateAccessibility()
		buffer := unsafe.Slice(out, int(capacity))
		state.mu.RLock()
		defer state.mu.RUnlock()
		count = min(len(ids), len(buffer))
		for i, id := range ids[:count] {
			buffer[i] = C.uint64_t(state.ids[virtualKey(node.object, id)])
		}
	})
	return C.int(count)
}
