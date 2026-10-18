package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"sort"
)

// Selected returns a copy of the selected row IDs in ascending order.
func (l *List) Selected() []ListItemID { return append([]ListItemID(nil), l.selected...) }

// SetSelection changes the selection in one refresh and reports only changed
// items through OnSelected and OnUnselected. Invalid and duplicate IDs are ignored.
func (l *List) SetSelection(ids []ListItemID) {
	if l.Length == nil {
		return
	}
	length := l.Length()
	next := make([]ListItemID, 0, len(ids))
	seen := make(map[ListItemID]bool, len(ids))
	for _, id := range ids {
		if id >= 0 && id < length && !seen[id] {
			next = append(next, id)
			seen[id] = true
		}
	}
	if l.SelectionMode != SelectionMultiple && len(next) > 1 {
		next = next[len(next)-1:]
	}
	l.setSelection(next)
}

// SelectAll selects every row when multiple selection is enabled.
func (l *List) SelectAll() {
	if l.SelectionMode != SelectionMultiple || l.Length == nil {
		return
	}
	ids := make([]ListItemID, l.Length())
	for i := range ids {
		ids[i] = i
	}
	l.setSelection(ids)
}

// SelectRange selects an inclusive range, replacing the previous selection.
func (l *List) SelectRange(first, last ListItemID) {
	if l.Length == nil || l.Length() == 0 {
		return
	}
	if l.SelectionMode != SelectionMultiple {
		l.Select(last)
		return
	}
	first, last = max(0, min(first, last)), min(l.Length()-1, max(first, last))
	if first > last {
		return
	}
	ids := make([]ListItemID, last-first+1)
	for i := range ids {
		ids[i] = first + i
	}
	l.setSelection(ids)
}

func (l *List) setSelection(next []ListItemID) {
	sort.Ints(next)
	old := l.selected
	l.selected = next
	l.Refresh()
	selectionDifference(old, next, l.OnUnselected, l.OnSelected)
}

func (l *List) pointerSelect(id ListItemID, modifiers granio.KeyModifier) {
	if l.Length == nil || id < 0 || id >= l.Length() {
		return
	}
	if !l.selectionAnchored {
		l.selectionAnchor = id
		if len(l.selected) > 0 {
			l.selectionAnchor = l.selected[0]
		}
		l.selectionAnchored = true
	}
	span := func() []ListItemID {
		first, last := max(0, min(l.selectionAnchor, id)), min(l.Length()-1, max(l.selectionAnchor, id))
		ids := make([]ListItemID, last-first+1)
		for i := range ids {
			ids[i] = first + i
		}
		return ids
	}
	l.setSelection(selectPointer(l.selected, id, l.SelectionMode, modifiers, span))
	if modifiers&granio.KeyModifierShift == 0 {
		l.selectionAnchor = id
	}
	l.scrollWithoutItemCheckTo(id)
}

// TypedShortcut handles the standard Select All shortcut for multiple selection.
func (l *List) TypedShortcut(shortcut granio.Shortcut) {
	if _, ok := shortcut.(*granio.ShortcutSelectAll); ok {
		l.SelectAll()
	}
}
