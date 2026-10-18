package widget

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
)

// A data callback can populate a composite row. Read its semantic labels, not
// its renderer implementation, then release any renderer created by Refresh.
func virtualLabel(object granio.CanvasObject) string {
	if object == nil {
		return ""
	}
	if accessible, ok := object.(granio.Accessible); ok && accessible.AccessibilityRole() != granio.AccessibleRoleContainer {
		return accessible.AccessibilityLabel()
	}
	if container, ok := object.(*granio.Container); ok {
		labels := make([]string, 0, len(container.Objects))
		for _, child := range container.Objects {
			if label := virtualLabel(child); label != "" {
				labels = append(labels, label)
			}
		}
		return strings.Join(labels, " ")
	}
	return ""
}
func releaseVirtualTemplate(object granio.CanvasObject) {
	if container, ok := object.(*granio.Container); ok {
		for _, child := range container.Objects {
			releaseVirtualTemplate(child)
		}
	}
	if widget, ok := object.(granio.Widget); ok {
		cache.DestroyRenderer(widget)
	}
}
func virtualIndex(id string, count int) (int, bool) {
	index, err := strconv.Atoi(id)
	return index, err == nil && index >= 0 && index < count && strconv.Itoa(index) == id
}
func (l *List) AccessibilityVirtualCount(parent string) int {
	if parent != "" || l.Length == nil {
		return 0
	}
	return max(0, l.Length())
}
func (l *List) AccessibilityVirtualChild(parent string, index int) (string, bool) {
	return strconv.Itoa(index), index >= 0 && index < l.AccessibilityVirtualCount(parent)
}
func (l *List) AccessibilityVirtualItem(id string) (granio.AccessibleVirtualItem, bool) {
	index, ok := virtualIndex(id, l.AccessibilityVirtualCount(""))
	if !ok {
		return granio.AccessibleVirtualItem{}, false
	}
	label := ""
	if l.CreateItem != nil && l.UpdateItem != nil {
		object := l.CreateItem()
		defer releaseVirtualTemplate(object)
		l.UpdateItem(index, object)
		label = virtualLabel(object)
	}
	return granio.AccessibleVirtualItem{ID: id, Index: index, Label: label, Role: granio.AccessibleRoleListItem, State: granio.AccessibleState{Selected: slices.Contains(l.selected, index)}, Offscreen: true}, true
}
func (l *List) AccessibilityVirtualSelection() []string {
	selected := l.Selected()
	ids := make([]string, len(selected))
	for i, index := range selected {
		ids[i] = strconv.Itoa(index)
	}
	return ids
}
func (l *List) AccessibilityVirtualAction(id, action string) bool {
	index, ok := virtualIndex(id, l.AccessibilityVirtualCount(""))
	if !ok {
		return false
	}
	if action == "realize" || action == "focus" {
		l.ScrollTo(index)
		if action == "focus" {
			l.currentHighlight = index
			virtualFocus(l)
		}
		return true
	}
	return virtualSelectionAction(l, index, action)
}
func virtualSelectionAction(selection granio.AccessibleSelection, index int, action string) bool {
	switch action {
	case "select":
		return selection.AccessibilitySelectIndex(index, true, false)
	case "add-selection":
		return selection.AccessibilitySelectIndex(index, true, true)
	case "remove-selection":
		return selection.AccessibilitySelectIndex(index, false, false)
	}
	return false
}
func (t *Table) AccessibilityVirtualCount(parent string) int {
	if parent != "" {
		return 0
	}
	rows, cols := t.AccessibilityTableSize()
	if rows <= 0 || cols <= 0 || rows > int(^uint(0)>>1)/cols {
		return 0
	}
	return rows * cols
}
func (t *Table) AccessibilityVirtualChild(parent string, index int) (string, bool) {
	if index < 0 || index >= t.AccessibilityVirtualCount(parent) {
		return "", false
	}
	_, cols := t.AccessibilityTableSize()
	return fmt.Sprintf("%d:%d", index/cols, index%cols), true
}
func (t *Table) virtualCell(id string) (TableCellID, bool) {
	row, col, ok := strings.Cut(id, ":")
	if !ok {
		return TableCellID{}, false
	}
	rows, cols := t.AccessibilityTableSize()
	r, validR := virtualIndex(row, rows)
	c, validC := virtualIndex(col, cols)
	return TableCellID{Row: r, Col: c}, validR && validC
}
func (t *Table) AccessibilityVirtualItem(id string) (granio.AccessibleVirtualItem, bool) {
	cell, ok := t.virtualCell(id)
	if !ok {
		return granio.AccessibleVirtualItem{}, false
	}
	_, cols := t.AccessibilityTableSize()
	index := cell.Row*cols + cell.Col
	label := ""
	if t.CreateCell != nil && t.UpdateCell != nil {
		object := t.CreateCell()
		defer releaseVirtualTemplate(object)
		t.UpdateCell(cell, object)
		label = virtualLabel(object)
	}
	return granio.AccessibleVirtualItem{ID: id, Index: index, Label: label, Role: granio.AccessibleRoleCell, Row: cell.Row, Column: cell.Col, State: granio.AccessibleState{Selected: t.isSelected(cell)}, Offscreen: true}, true
}
func (t *Table) AccessibilityVirtualSelection() []string {
	selected := t.Selected()
	ids := make([]string, len(selected))
	for i, cell := range selected {
		ids[i] = fmt.Sprintf("%d:%d", cell.Row, cell.Col)
	}
	return ids
}
func (t *Table) AccessibilityVirtualAction(id, action string) bool {
	cell, ok := t.virtualCell(id)
	if !ok {
		return false
	}
	_, cols := t.AccessibilityTableSize()
	if action == "realize" || action == "focus" {
		t.ScrollTo(cell)
		if action == "focus" {
			t.currentHighlight = cell
			virtualFocus(t)
		}
		return true
	}
	return virtualSelectionAction(t, cell.Row*cols+cell.Col, action)
}
func (t *Tree) AccessibilityVirtualCount(parent string) int {
	if t.ChildUIDs == nil {
		return 0
	}
	if parent == "" {
		if t.Root != "" {
			return 1
		}
		parent = t.Root
	}
	return len(t.ChildUIDs(parent))
}
func (t *Tree) AccessibilityVirtualChild(parent string, index int) (string, bool) {
	if t.ChildUIDs == nil {
		return "", false
	}
	if parent == "" {
		if t.Root != "" {
			return t.Root, index == 0
		}
		parent = t.Root
	}
	children := t.ChildUIDs(parent)
	if index < 0 || index >= len(children) {
		return "", false
	}
	return children[index], children[index] != ""
}
func (t *Tree) AccessibilityVirtualItem(id string) (granio.AccessibleVirtualItem, bool) {
	if id == "" {
		return granio.AccessibleVirtualItem{}, false
	}
	found, parents := t.findPath(t.Root, id)
	if !found || len(parents) == 0 && id != t.Root {
		return granio.AccessibleVirtualItem{}, false
	}
	parent, index := "", 0
	if id != t.Root {
		parent = parents[len(parents)-1]
		index = -1
		for i, child := range t.ChildUIDs(parent) {
			if child == id {
				index = i
				break
			}
		}
	}
	if index < 0 {
		return granio.AccessibleVirtualItem{}, false
	}
	branch := t.IsBranch != nil && t.IsBranch(id)
	label := id
	if t.CreateNode != nil && t.UpdateNode != nil {
		object := t.CreateNode(branch)
		defer releaseVirtualTemplate(object)
		t.UpdateNode(id, branch, object)
		label = virtualLabel(object)
	}
	if parent == t.Root && t.Root == "" {
		parent = ""
	}
	count := 0
	if branch {
		count = t.AccessibilityVirtualCount(id)
	}
	return granio.AccessibleVirtualItem{ID: id, Parent: parent, Index: index, ChildCount: count, Label: label, Role: granio.AccessibleRoleTreeItem, State: granio.AccessibleState{Selected: t.isSelected(id), Expanded: branch && t.IsBranchOpen(id)}, Offscreen: true}, true
}
func (t *Tree) AccessibilityVirtualSelection() []string { return t.Selected() }
func (t *Tree) AccessibilityVirtualAction(id, action string) bool {
	if found, _ := t.findPath(t.Root, id); !found {
		return false
	}
	switch action {
	case "realize", "focus":
		t.ScrollTo(id)
		if action == "focus" {
			t.setItemFocus(id)
			virtualFocus(t)
		}
	case "select":
		t.SetSelection([]TreeNodeID{id})
	case "add-selection":
		t.Select(id)
	case "remove-selection":
		t.Unselect(id)
	case "toggle":
		if t.IsBranch == nil || !t.IsBranch(id) {
			return false
		}
		t.ToggleBranch(id)
	default:
		return false
	}
	return true
}
func virtualFocus(object granio.CanvasObject) {
	if canvas := granio.CurrentApp().Driver().CanvasForObject(object); canvas != nil {
		if focusable, ok := object.(granio.Focusable); ok {
			canvas.Focus(focusable)
		}
	}
}
