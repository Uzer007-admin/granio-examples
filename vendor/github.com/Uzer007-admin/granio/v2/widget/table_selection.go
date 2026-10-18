package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"sort"
)

// Selected returns a copy of selected cells in row, column order.
func (t *Table) Selected() []TableCellID {
	if t.SelectionMode != SelectionMultiple || t.selectedCells == nil {
		if t.selectedCell == nil {
			return nil
		}
		return []TableCellID{*t.selectedCell}
	}
	ids := make([]TableCellID, 0, len(t.selectedCells))
	for id := range t.selectedCells {
		ids = append(ids, id)
	}
	sortCells(ids)
	return ids
}

func sortCells(ids []TableCellID) {
	sort.Slice(ids, func(i, j int) bool {
		if ids[i].Row == ids[j].Row {
			return ids[i].Col < ids[j].Col
		}
		return ids[i].Row < ids[j].Row
	})
}

func (t *Table) isSelected(id TableCellID) bool {
	if t.SelectionMode != SelectionMultiple || t.selectedCells == nil {
		return t.selectedCell != nil && *t.selectedCell == id
	}
	_, ok := t.selectedCells[id]
	return ok
}

// SetSelection changes selection once, ignoring invalid and duplicate cells.
// Callbacks report only cells whose selected state changed.
func (t *Table) SetSelection(ids []TableCellID) {
	if t.Length == nil {
		return
	}
	rows, cols := t.Length()
	old := t.Selected()
	next := make([]TableCellID, 0, len(ids))
	seen := make(map[TableCellID]struct{}, len(ids))
	for _, id := range ids {
		if id.Row < 0 || id.Row >= rows || id.Col < 0 || id.Col >= cols {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			next = append(next, id)
		}
	}
	if t.SelectionMode != SelectionMultiple && len(next) > 1 {
		next = next[len(next)-1:]
	}
	t.selectedCells = seen
	t.selectedCell = nil
	if len(next) > 0 {
		last := next[len(next)-1]
		t.selectedCell = &last
		t.currentHighlight = last
	}
	t.Refresh()
	selectionDifference(old, next, t.OnUnselected, t.OnSelected)
}

// SelectRange selects the inclusive rectangle between two cells.
func (t *Table) SelectRange(first, last TableCellID) {
	if t.SelectionMode != SelectionMultiple {
		t.Select(last)
		return
	}
	t.SetSelection(t.selectionSpan(first, last))
}

// SelectAll selects every data cell when multiple selection is enabled.
func (t *Table) SelectAll() {
	if t.SelectionMode != SelectionMultiple || t.Length == nil {
		return
	}
	rows, cols := t.Length()
	if rows > 0 && cols > 0 {
		t.SelectRange(TableCellID{}, TableCellID{rows - 1, cols - 1})
	}
}

func (t *Table) selectionSpan(first, last TableCellID) []TableCellID {
	if t.Length == nil {
		return nil
	}
	rows, cols := t.Length()
	startRow, endRow := max(0, min(first.Row, last.Row)), min(rows-1, max(first.Row, last.Row))
	startCol, endCol := max(0, min(first.Col, last.Col)), min(cols-1, max(first.Col, last.Col))
	if startRow > endRow || startCol > endCol {
		return nil
	}
	ids := make([]TableCellID, 0, (endRow-startRow+1)*(endCol-startCol+1))
	for row := startRow; row <= endRow; row++ {
		for col := startCol; col <= endCol; col++ {
			ids = append(ids, TableCellID{row, col})
		}
	}
	return ids
}

func (t *Table) pointerSelect(id TableCellID, modifiers granio.KeyModifier) {
	if !t.selectionAnchored {
		t.selectionAnchor, t.selectionAnchored = id, true
		if t.selectedCell != nil {
			t.selectionAnchor = *t.selectedCell
		}
	}
	t.SetSelection(selectPointer(t.Selected(), id, t.SelectionMode, modifiers, func() []TableCellID { return t.selectionSpan(t.selectionAnchor, id) }))
	if modifiers&granio.KeyModifierShift == 0 {
		t.selectionAnchor = id
	}
	t.currentHighlight = id
	t.ScrollTo(id)
}

// TypedShortcut handles the standard Select All shortcut.
func (t *Table) TypedShortcut(shortcut granio.Shortcut) {
	if _, ok := shortcut.(*granio.ShortcutSelectAll); ok {
		t.SelectAll()
	}
}

func (r *tableCellsRenderer) moveSelectionMarkers(offX, offY float32, minCol, minRow int, widths, heights map[int]float32) {
	t := r.cells.t
	ids := make([]TableCellID, 0)
	if t.SelectionMode == SelectionMultiple {
		for id := range r.visible {
			if t.isSelected(id) && (t.selectedCell == nil || *t.selectedCell != id) {
				ids = append(ids, id)
			}
		}
	}
	sortCells(ids)
	grew := false
	th, variant := t.Theme(), granio.CurrentApp().Settings().ThemeVariant()
	for i, id := range ids {
		if i >= len(r.selectionMarkers) {
			r.selectionMarkers = append(r.selectionMarkers, canvas.NewRectangle(nil))
			grew = true
		}
		marker := r.selectionMarkers[i]
		applySelectionStyle(marker, th, variant)
		marker.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
		r.moveMarker(marker, id.Row, id.Col, offX, offY, minCol, minRow, widths, heights)
	}
	for _, marker := range r.selectionMarkers[len(ids):] {
		marker.Hide()
	}
	if grew {
		objects := []granio.CanvasObject{r.marker, r.hover}
		for _, marker := range r.selectionMarkers {
			objects = append(objects, marker)
		}
		t.dividerLayer.Content.(*granio.Container).Objects = append(objects, r.dividers...)
		t.dividerLayer.Content.Refresh()
	}
}
