package widget

func (l *List) AccessibilitySetSelection(indices []int) bool {
	if l.Length == nil {
		return false
	}
	count := l.Length()
	for _, index := range indices {
		if index < 0 || index >= count {
			return false
		}
	}
	l.SetSelection(indices)
	return true
}
func (t *Tree) AccessibilitySetSelection(indices []int) bool {
	rows := t.selectionRows()
	ids := make([]TreeNodeID, len(indices))
	for i, index := range indices {
		if index < 0 || index >= len(rows) {
			return false
		}
		ids[i] = rows[index].id
	}
	t.SetSelection(ids)
	return true
}
func (t *Table) AccessibilitySetSelection(indices []int) bool {
	rows, cols := t.AccessibilityTableSize()
	if cols <= 0 {
		return len(indices) == 0
	}
	ids := make([]TableCellID, len(indices))
	for i, index := range indices {
		if index < 0 || index/cols >= rows {
			return false
		}
		ids[i] = TableCellID{Row: index / cols, Col: index % cols}
	}
	t.SetSelection(ids)
	return true
}
