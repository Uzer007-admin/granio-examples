package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
)

// Selected returns a copy of the selected node IDs.
func (t *Tree) Selected() []TreeNodeID { return append([]TreeNodeID(nil), t.selected...) }

// SetSelection changes selected node IDs in one refresh. IDs must belong to the
// data source; closed descendants can stay selected while their branch is closed.
func (t *Tree) SetSelection(ids []TreeNodeID) {
	next := make([]TreeNodeID, 0, len(ids))
	seen := make(map[TreeNodeID]bool, len(ids))
	for _, id := range ids {
		if !seen[id] {
			next = append(next, id)
			seen[id] = true
		}
	}
	if t.SelectionMode != SelectionMultiple && len(next) > 1 {
		next = next[len(next)-1:]
	}
	t.setSelection(next)
}

// SelectAll selects all currently expanded nodes.
func (t *Tree) SelectAll() {
	if t.SelectionMode != SelectionMultiple {
		return
	}
	rows := t.selectionRows()
	ids := make([]TreeNodeID, len(rows))
	for i, row := range rows {
		ids[i] = row.id
	}
	t.setSelection(ids)
}

// SelectRange selects an inclusive range in the visible tree order.
func (t *Tree) SelectRange(first, last TreeNodeID) {
	if t.SelectionMode != SelectionMultiple {
		t.Select(last)
		return
	}
	t.setSelection(t.selectionSpan(first, last))
}

func (t *Tree) selectionRows() []treeRow {
	r := cache.Renderer(t).(*treeRenderer)
	content := cache.Renderer(r.content).(*treeContentRenderer)
	content.prepareGeometry()
	return content.rows
}

func (t *Tree) selectionSpan(first, last TreeNodeID) []TreeNodeID {
	rows := t.selectionRows()
	a, b := -1, -1
	for i, row := range rows {
		if row.id == first {
			a = i
		}
		if row.id == last {
			b = i
		}
	}
	if a < 0 || b < 0 {
		return []TreeNodeID{last}
	}
	if a > b {
		a, b = b, a
	}
	ids := make([]TreeNodeID, b-a+1)
	for i := range ids {
		ids[i] = rows[a+i].id
	}
	return ids
}

func (t *Tree) isSelected(id TreeNodeID) bool {
	if t.SelectionMode != SelectionMultiple {
		return len(t.selected) > 0 && t.selected[0] == id
	}
	_, ok := t.selectionSet[id]
	return ok
}

func (t *Tree) setSelection(next []TreeNodeID) {
	old := t.selected
	t.selected = next
	t.selectionSet = make(map[TreeNodeID]struct{}, len(next))
	for _, id := range next {
		t.selectionSet[id] = struct{}{}
	}
	if t.scroller != nil {
		t.scroller.Content.(*treeContent).refreshForID(allTreeNodesID)
	}
	selectionDifference(old, next, t.OnUnselected, t.OnSelected)
}

func (t *Tree) pointerSelect(id TreeNodeID, modifiers granio.KeyModifier) {
	if !t.selectionAnchored {
		t.selectionAnchor = id
		if len(t.selected) > 0 {
			t.selectionAnchor = t.selected[0]
		}
		t.selectionAnchored = true
	}
	t.setSelection(selectPointer(t.selected, id, t.SelectionMode, modifiers, func() []TreeNodeID { return t.selectionSpan(t.selectionAnchor, id) }))
	if modifiers&granio.KeyModifierShift == 0 {
		t.selectionAnchor = id
	}
	t.setItemFocus(id)
}

// TypedShortcut handles Select All for the expanded tree.
func (t *Tree) TypedShortcut(shortcut granio.Shortcut) {
	if _, ok := shortcut.(*granio.ShortcutSelectAll); ok {
		t.SelectAll()
	}
}
