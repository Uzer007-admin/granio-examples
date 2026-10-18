package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
)

// SelectionMode controls collection selection. The zero value preserves the
// default single selection. Multiple selection supports Ctrl toggles and Shift
// ranges on Windows and Linux, and Command toggles on macOS.
type SelectionMode int

const (
	SelectionSingle SelectionMode = iota
	SelectionMultiple
)

func selectionModifiers() granio.KeyModifier {
	if app := granio.CurrentApp(); app != nil {
		if driver, ok := app.Driver().(desktop.Driver); ok {
			return driver.CurrentKeyModifiers()
		}
	}
	return 0
}

func selectionDifference[T comparable](old, next []T, removed, added func(T)) {
	if removed == nil && added == nil {
		return
	}
	before, after := make(map[T]struct{}, len(old)), make(map[T]struct{}, len(next))
	for _, id := range old {
		before[id] = struct{}{}
	}
	for _, id := range next {
		after[id] = struct{}{}
	}
	if removed != nil {
		for _, id := range old {
			if _, ok := after[id]; !ok {
				removed(id)
			}
		}
	}
	if added != nil {
		for _, id := range next {
			if _, ok := before[id]; !ok {
				added(id)
			}
		}
	}
}

func selectPointer[T comparable](old []T, id T, mode SelectionMode, modifiers granio.KeyModifier, span func() []T) []T {
	if mode != SelectionMultiple {
		return []T{id}
	}
	toggle := modifiers&(granio.KeyModifierControl|granio.KeyModifierSuper) != 0
	if modifiers&granio.KeyModifierShift != 0 && span != nil {
		rangeIDs := span()
		if !toggle {
			return rangeIDs
		}
		result := append([]T(nil), old...)
		seen := make(map[T]struct{}, len(result)+len(rangeIDs))
		for _, entry := range result {
			seen[entry] = struct{}{}
		}
		for _, entry := range rangeIDs {
			if _, ok := seen[entry]; !ok {
				result = append(result, entry)
			}
		}
		return result
	}
	if !toggle {
		return []T{id}
	}
	result := append([]T(nil), old...)
	for i, entry := range result {
		if entry == id {
			return append(result[:i], result[i+1:]...)
		}
	}
	return append(result, id)
}
