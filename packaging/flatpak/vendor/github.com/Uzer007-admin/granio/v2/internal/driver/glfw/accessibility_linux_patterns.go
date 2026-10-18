//go:build accessibility && linux && !wasm && !test_web_driver

package glfw

import (
	"fmt"
	"slices"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2"
)

type atspiValueObject struct{ *atspiObject }
type atspiSelectionObject struct{ *atspiObject }
type atspiTableObject struct{ *atspiObject }

func (o *atspiValueObject) SetCurrentValue(value float64) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	applied := false
	accessibilityUI(func() {
		if target, ok := e.node.object.(granio.AccessibleNumeric); ok {
			applied = target.AccessibilitySetNumericValue(value)
			if applied {
				e.window.updateAccessibility()
			}
		}
	})
	return applied, nil
}
func (o *atspiSelectionObject) selected() ([]int, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return nil, err
	}
	var selected []int
	accessibilityUI(func() {
		if target, ok := e.node.object.(granio.AccessibleSelection); ok {
			selected, _ = target.AccessibilitySelection()
		}
	})
	return selected, nil
}
func (o *atspiSelectionObject) GetSelectedChild(index int32) (atspiReference, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return o.manager.reference(""), err
	}
	if source, ok := e.node.object.(granio.AccessibleVirtualCollection); ok {
		ids, err := o.selectedIDs()
		if err != nil {
			return o.manager.reference(""), err
		}
		if index < 0 || int(index) >= len(ids) {
			return o.manager.reference(""), dbus.NewError("org.a11y.atspi.Error.InvalidIndex", nil)
		}
		path := dbus.ObjectPath("")
		accessibilityUI(func() {
			if e.window.requestVirtual(e.node.object, ids[index]) {
				e.window.updateAccessibility()
				path = o.manager.pathForKey(e.window, virtualKey(e.node.object, ids[index]))
			}
		})
		_ = source
		return o.manager.reference(path), nil
	}
	selected := o.selectedChildren(e)
	if index < 0 || int(index) >= len(selected) {
		return o.manager.reference(""), dbus.NewError("org.a11y.atspi.Error.InvalidIndex", nil)
	}
	return o.manager.reference(e.children[selected[index]]), nil
}

// AT-SPI child indices are offsets in the exposed children, not source row IDs.
// Snapshotting remains bounded by the viewport; offscreen proxies are a separate
// virtual-hierarchy capability rather than incorrectly returning another row.
func (o *atspiSelectionObject) selectedChildren(e atspiElement) []int {
	o.manager.mu.RLock()
	defer o.manager.mu.RUnlock()
	var selected []int
	for index, path := range e.children {
		node := o.manager.elements[path].node
		if node.selectionOwner == e.node.object && node.state.Selected {
			selected = append(selected, index)
		}
	}
	return selected
}
func (o *atspiSelectionObject) selectIndex(index int32, selected, extend bool) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	if source, ok := e.node.object.(granio.AccessibleVirtualCollection); ok {
		applied := false
		accessibilityUI(func() {
			id, valid := source.AccessibilityVirtualChild("", int(index))
			if !valid {
				return
			}
			action := "remove-selection"
			if selected {
				action = "select"
				if extend {
					action = "add-selection"
				}
			}
			applied = source.AccessibilityVirtualAction(id, action)
			if applied {
				e.window.updateAccessibility()
			}
		})
		return applied, nil
	}
	if index < 0 || int(index) >= len(e.children) {
		return false, nil
	}
	o.manager.mu.RLock()
	child := o.manager.elements[e.children[index]].node
	o.manager.mu.RUnlock()
	if child.selectionOwner != e.node.object || child.selectionIndex < 0 {
		return false, nil
	}
	applied := false
	accessibilityUI(func() {
		if target, ok := e.node.object.(granio.AccessibleSelection); ok {
			applied = target.AccessibilitySelectIndex(child.selectionIndex, selected, extend)
			if applied {
				e.window.updateAccessibility()
			}
		}
	})
	return applied, nil
}
func (o *atspiSelectionObject) SelectChild(index int32) (bool, *dbus.Error) {
	return o.selectIndex(index, true, true)
}
func (o *atspiSelectionObject) DeselectChild(index int32) (bool, *dbus.Error) {
	return o.selectIndex(index, false, false)
}
func (o *atspiSelectionObject) DeselectSelectedChild(index int32) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	if source, ok := e.node.object.(granio.AccessibleVirtualCollection); ok {
		ids, err := o.selectedIDs()
		if err != nil {
			return false, err
		}
		if index < 0 || int(index) >= len(ids) {
			return false, nil
		}
		applied := false
		accessibilityUI(func() {
			applied = source.AccessibilityVirtualAction(ids[index], "remove-selection")
			e.window.updateAccessibility()
		})
		return applied, nil
	}
	selected := o.selectedChildren(e)
	if index < 0 || int(index) >= len(selected) {
		return false, nil
	}
	return o.DeselectChild(int32(selected[index]))
}
func (o *atspiSelectionObject) IsChildSelected(index int32) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	if source, ok := e.node.object.(granio.AccessibleVirtualCollection); ok {
		ids, err := o.selectedIDs()
		id := ""
		valid := false
		accessibilityUI(func() { id, valid = source.AccessibilityVirtualChild("", int(index)) })
		return valid && slices.Contains(ids, id), err
	}
	selected := o.selectedChildren(e)
	return slices.Contains(selected, int(index)), err
}
func (o *atspiSelectionObject) ClearSelection() (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	applied := false
	accessibilityUI(func() {
		if target, ok := e.node.object.(granio.AccessibleSelectionBatch); ok {
			applied = target.AccessibilitySetSelection(nil)
			if applied {
				e.window.updateAccessibility()
			}
		}
	})
	return applied, nil
}
func (o *atspiSelectionObject) SelectAll() (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	applied := false
	accessibilityUI(func() {
		if target, ok := e.node.object.(interface{ SelectAll() }); ok {
			target.SelectAll()
			applied = true
			e.window.updateAccessibility()
		}
	})
	return applied, nil
}

func (o *atspiTableObject) GetAccessibleAt(row, column int32) (atspiReference, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return o.manager.reference(""), err
	}
	path := dbus.ObjectPath("")
	accessibilityUI(func() {
		if _, ok := e.node.object.(granio.AccessibleVirtualCollection); ok {
			_, cols := e.node.object.(granio.AccessibleTable).AccessibilityTableSize()
			if row < 0 || column < 0 || int(row) >= e.node.rows || int(column) >= cols {
				return
			}
			key := virtualKey(e.node.object, fmt.Sprintf("%d:%d", row, column))
			if e.window.requestVirtual(e.node.object, fmt.Sprintf("%d:%d", row, column)) {
				e.window.updateAccessibility()
				path = o.manager.pathForKey(e.window, key)
			}
			return
		}
		target, ok := e.node.object.(granio.AccessibleTable)
		if !ok {
			return
		}
		if target.AccessibilityTableCell(int(row), int(column)) == nil {
			return
		}
		e.window.updateAccessibility()
		o.manager.mu.RLock()
		defer o.manager.mu.RUnlock()
		for _, candidate := range o.manager.windows[e.window] {
			node := o.manager.elements[candidate].node
			if node.tableOwner == e.node.object && node.row == int(row) && node.column == int(column) {
				path = candidate
				break
			}
		}
	})
	if path == "" {
		return o.manager.reference(""), dbus.NewError("org.a11y.atspi.Error.InvalidIndex", nil)
	}
	return o.manager.reference(path), nil
}
func (o *atspiTableObject) GetIndexAt(row, column int32) (int32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return -1, err
	}
	if row < 0 || column < 0 || int(row) >= e.node.rows || int(column) >= e.node.columns {
		return -1, nil
	}
	return row*int32(e.node.columns) + column, nil
}
func (o *atspiTableObject) GetRowAtIndex(index int32) (int32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return -1, err
	}
	if index < 0 || e.node.columns <= 0 || int(index)/e.node.columns >= e.node.rows {
		return -1, nil
	}
	return index / int32(e.node.columns), nil
}
func (o *atspiTableObject) GetColumnAtIndex(index int32) (int32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return -1, err
	}
	if index < 0 || e.node.columns <= 0 || int(index)/e.node.columns >= e.node.rows {
		return -1, nil
	}
	return index % int32(e.node.columns), nil
}
func (o *atspiTableObject) GetRowDescription(row int32) (string, *dbus.Error)       { return "", nil }
func (o *atspiTableObject) GetColumnDescription(column int32) (string, *dbus.Error) { return "", nil }
func (o *atspiTableObject) GetRowExtentAt(row, column int32) (int32, *dbus.Error) {
	index, err := o.GetIndexAt(row, column)
	if index < 0 {
		return 0, err
	}
	return 1, err
}
func (o *atspiTableObject) GetColumnExtentAt(row, column int32) (int32, *dbus.Error) {
	return o.GetRowExtentAt(row, column)
}
func (o *atspiTableObject) GetRowHeader(row int32) (atspiReference, *dbus.Error) {
	return o.manager.reference(""), nil
}
func (o *atspiTableObject) GetColumnHeader(column int32) (atspiReference, *dbus.Error) {
	return o.manager.reference(""), nil
}
func (o *atspiTableObject) IsSelected(row, column int32) (bool, *dbus.Error) {
	index, err := o.GetIndexAt(row, column)
	if err != nil || index < 0 {
		return false, err
	}
	selected, err := (&atspiSelectionObject{o.atspiObject}).selected()
	return slices.Contains(selected, int(index)), err
}
func (o *atspiTableObject) selectedLines(rows bool) ([]int32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return nil, err
	}
	selected, err := (&atspiSelectionObject{o.atspiObject}).selected()
	if err != nil {
		return nil, err
	}
	counts := map[int]int{}
	span := e.node.columns
	if !rows {
		span = e.node.rows
	}
	for _, id := range selected {
		if e.node.columns <= 0 {
			continue
		}
		line := id / e.node.columns
		if !rows {
			line = id % e.node.columns
		}
		counts[line]++
	}
	result := []int32{}
	for line, count := range counts {
		if count == span {
			result = append(result, int32(line))
		}
	}
	slices.Sort(result)
	return result, nil
}
func (o *atspiTableObject) GetSelectedRows() ([]int32, *dbus.Error)    { return o.selectedLines(true) }
func (o *atspiTableObject) GetSelectedColumns() ([]int32, *dbus.Error) { return o.selectedLines(false) }
func (o *atspiTableObject) IsRowSelected(row int32) (bool, *dbus.Error) {
	selected, err := o.GetSelectedRows()
	return slices.Contains(selected, row), err
}
func (o *atspiTableObject) IsColumnSelected(column int32) (bool, *dbus.Error) {
	selected, err := o.GetSelectedColumns()
	return slices.Contains(selected, column), err
}
func (o *atspiTableObject) selectLine(index int32, rows, selected bool) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	if index < 0 || (rows && int(index) >= e.node.rows) || (!rows && int(index) >= e.node.columns) {
		return false, nil
	}
	size := e.node.columns
	if !rows {
		size = e.node.rows
	}
	applied := false
	accessibilityUI(func() {
		target, ok := e.node.object.(granio.AccessibleSelection)
		if !ok || !target.AccessibilityCanSelectMultiple() {
			return
		}
		prior, _ := target.AccessibilitySelection()
		set := make(map[int]bool, len(prior)+size)
		for _, id := range prior {
			set[id] = true
		}
		for i := 0; i < size; i++ {
			id := int(index)*e.node.columns + i
			if !rows {
				id = i*e.node.columns + int(index)
			}
			if selected {
				set[id] = true
			} else {
				delete(set, id)
			}
		}
		indices := make([]int, 0, len(set))
		for id := range set {
			indices = append(indices, id)
		}
		slices.Sort(indices)
		if batch, ok := e.node.object.(granio.AccessibleSelectionBatch); ok {
			applied = batch.AccessibilitySetSelection(indices)
		}
		e.window.updateAccessibility()
	})
	return applied, nil
}
func (o *atspiTableObject) AddRowSelection(index int32) (bool, *dbus.Error) {
	return o.selectLine(index, true, true)
}
func (o *atspiTableObject) AddColumnSelection(index int32) (bool, *dbus.Error) {
	return o.selectLine(index, false, true)
}
func (o *atspiTableObject) RemoveRowSelection(index int32) (bool, *dbus.Error) {
	return o.selectLine(index, true, false)
}
func (o *atspiTableObject) RemoveColumnSelection(index int32) (bool, *dbus.Error) {
	return o.selectLine(index, false, false)
}
func (o *atspiTableObject) GetRowColumnExtentsAtIndex(index int32) (bool, int32, int32, int32, int32, bool, *dbus.Error) {
	row, err := o.GetRowAtIndex(index)
	column, _ := o.GetColumnAtIndex(index)
	if err != nil || row < 0 {
		return false, 0, 0, 0, 0, false, err
	}
	selected, err := o.IsSelected(row, column)
	return true, row, column, 1, 1, selected, err
}
