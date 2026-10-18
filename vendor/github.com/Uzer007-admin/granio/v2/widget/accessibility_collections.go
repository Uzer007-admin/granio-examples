package widget

import (
	"fmt"
	"math"
	"slices"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
)

func (s *Slider) AccessibilityNumericValue() (float64, float64, float64, float64) {
	return s.Value, s.Min, s.Max, s.Step
}
func (s *Slider) AccessibilitySetNumericValue(value float64) bool {
	if s.Disabled() || math.IsNaN(value) || math.IsInf(value, 0) || value < s.Min || value > s.Max {
		return false
	}
	s.SetValue(value)
	return true
}
func (s *Slider) AccessibilityState() granio.AccessibleState {
	return granio.AccessibleState{Disabled: s.Disabled(), Focused: s.focused, ReadOnly: s.Disabled()}
}
func (p *ProgressBar) AccessibilityNumericValue() (float64, float64, float64, float64) {
	return p.Value, p.Min, p.Max, 0
}
func (*ProgressBar) AccessibilitySetNumericValue(float64) bool { return false }
func (*ProgressBar) AccessibilityState() granio.AccessibleState {
	return granio.AccessibleState{ReadOnly: true}
}

func (l *List) AccessibilityCanSelectMultiple() bool  { return l.SelectionMode == SelectionMultiple }
func (t *Tree) AccessibilityCanSelectMultiple() bool  { return t.SelectionMode == SelectionMultiple }
func (t *Table) AccessibilityCanSelectMultiple() bool { return t.SelectionMode == SelectionMultiple }
func (l *List) AccessibilitySelection() ([]int, bool) {
	return l.Selected(), l.SelectionMode == SelectionMultiple
}
func (l *List) AccessibilitySelectIndex(index int, selected, extend bool) bool {
	if l.Length == nil || index < 0 || index >= l.Length() {
		return false
	}
	if !selected {
		l.Unselect(index)
	} else if extend {
		l.Select(index)
	} else {
		l.SetSelection([]ListItemID{index})
	}
	return true
}
func (li *listItem) AccessibilityIdentity() string {
	return fmt.Sprintf("list:%p:%d", li.collection, li.index)
}
func (li *listItem) AccessibilityCollectionIndex() (granio.CanvasObject, int) {
	return li.collection, li.index
}
func (t *Tree) AccessibilitySelection() ([]int, bool) {
	rows, selected := t.selectionRows(), make([]int, 0, len(t.selected))
	for i, row := range rows {
		if t.isSelected(row.id) {
			selected = append(selected, i)
		}
	}
	return selected, t.SelectionMode == SelectionMultiple
}
func (t *Tree) AccessibilitySelectIndex(index int, selected, extend bool) bool {
	rows := t.selectionRows()
	if index < 0 || index >= len(rows) {
		return false
	}
	id := rows[index].id
	if !selected {
		t.Unselect(id)
	} else if extend {
		t.Select(id)
	} else {
		t.SetSelection([]TreeNodeID{id})
	}
	return true
}
func (n *treeNode) AccessibilityIdentity() string { return fmt.Sprintf("tree:%p:%s", n.tree, n.uid) }
func (n *treeNode) AccessibilityCollectionIndex() (granio.CanvasObject, int) {
	r := cache.Renderer(n.tree).(*treeRenderer)
	content := cache.Renderer(r.content).(*treeContentRenderer)
	content.prepareGeometry()
	index, ok := content.rowIndex[n.uid]
	if !ok {
		return n.tree, -1
	}
	return n.tree, index
}
func (t *Table) AccessibilityTableSize() (int, int) {
	if t.Length == nil {
		return 0, 0
	}
	return t.Length()
}
func (t *Table) AccessibilityVisibleCells() []granio.AccessibleTableCell {
	if t.cells == nil {
		return nil
	}
	r, ok := cache.Renderer(t.cells).(*tableCellsRenderer)
	if !ok {
		return nil
	}
	result := make([]granio.AccessibleTableCell, 0, len(r.visible))
	for id, object := range r.visible {
		result = append(result, granio.AccessibleTableCell{Object: object, Row: id.Row, Column: id.Col, Selected: t.isSelected(id)})
	}
	slices.SortFunc(result, func(a, b granio.AccessibleTableCell) int {
		if a.Row != b.Row {
			return a.Row - b.Row
		}
		return a.Column - b.Column
	})
	return result
}
func (t *Table) AccessibilityTableCell(row, column int) granio.CanvasObject {
	rows, cols := t.AccessibilityTableSize()
	if row < 0 || column < 0 || row >= rows || column >= cols {
		return nil
	}
	id := TableCellID{Row: row, Col: column}
	t.ScrollTo(id)
	if t.cells == nil {
		return nil
	}
	r := cache.Renderer(t.cells).(*tableCellsRenderer)
	return r.visible[id]
}
func (t *Table) AccessibilitySelection() ([]int, bool) {
	_, cols := t.AccessibilityTableSize()
	selected := t.Selected()
	result := make([]int, 0, len(selected))
	for _, id := range selected {
		result = append(result, id.Row*cols+id.Col)
	}
	return result, t.SelectionMode == SelectionMultiple
}
func (t *Table) AccessibilitySelectIndex(index int, selected, extend bool) bool {
	rows, cols := t.AccessibilityTableSize()
	if cols <= 0 || index < 0 || index/cols >= rows {
		return false
	}
	id := TableCellID{Row: index / cols, Col: index % cols}
	if !selected {
		t.Unselect(id)
	} else if extend {
		t.Select(id)
	} else {
		t.SetSelection([]TableCellID{id})
	}
	return true
}
