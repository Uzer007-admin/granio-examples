package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"strconv"
	"unicode/utf8"
)

func (e *Entry) AccessibilityLabel() string             { return e.PlaceHolder }
func (*Entry) AccessibilityRole() granio.AccessibleRole { return granio.AccessibleRoleInput }
func (e *Entry) AccessibilityValue() string {
	if e.Password {
		return ""
	}
	return e.Text
}
func (e *Entry) AccessibilityState() granio.AccessibleState {
	return granio.AccessibleState{Disabled: e.Disabled(), Focused: e.focused, Sensitive: e.Password, ReadOnly: e.Disabled() || e.ReadOnly}
}

func (e *Entry) AccessibilityTextSelection() (cursor, start, end int) {
	state := e.TextInputState()
	return state.Cursor, state.SelectionStart, state.SelectionEnd
}
func (e *Entry) AccessibilitySetText(value string) bool {
	if e.Disabled() || e.ReadOnly {
		return false
	}
	e.CancelComposition()
	e.AccessibilitySetSelection(0, utf8.RuneCountInString(e.Text))
	if value == "" {
		e.TypedKey(&granio.KeyEvent{Name: granio.KeyBackspace})
	} else {
		e.TypedText(value)
	}
	return true
}
func (e *Entry) AccessibilitySetSelection(start, end int) bool {
	if e.Disabled() || start < 0 || end < start || end > utf8.RuneCountInString(e.Text) {
		return false
	}
	e.CancelComposition()
	e.syncSelectable()
	e.sel.selectRow, e.sel.selectColumn = e.rowColFromTextPos(start)
	e.CursorRow, e.CursorColumn = e.rowColFromTextPos(end)
	e.sel.selecting = start != end
	e.syncSelectable()
	e.Refresh()
	return true
}
func (*Entry) AccessibilityActions() []string { return []string{"focus"} }
func (e *Entry) AccessibilityPerformAction(action string) bool {
	if action != "focus" || e.Disabled() {
		return false
	}
	e.requestFocus()
	return true
}

func (b *Button) AccessibilityActions() []string { return []string{"click", "focus"} }
func (b *Button) AccessibilityPerformAction(action string) bool {
	if b.Disabled() {
		return false
	}
	if action == "click" {
		b.Tapped(&granio.PointEvent{})
		return true
	}
	if action == "focus" {
		if canvas := granio.CurrentApp().Driver().CanvasForObject(b); canvas != nil {
			canvas.Focus(b)
			return true
		}
	}
	return false
}
func (b *Button) AccessibilityState() granio.AccessibleState {
	return granio.AccessibleState{Disabled: b.Disabled(), Focused: b.focused}
}

func (c *Check) AccessibilityLabel() string             { return c.Text }
func (*Check) AccessibilityRole() granio.AccessibleRole { return granio.AccessibleRoleCheck }
func (c *Check) AccessibilityState() granio.AccessibleState {
	return granio.AccessibleState{Disabled: c.Disabled(), Checked: c.Checked}
}
func (*Check) AccessibilityActions() []string { return []string{"toggle", "focus"} }
func (c *Check) AccessibilityPerformAction(action string) bool {
	if c.Disabled() {
		return false
	}
	if action == "toggle" {
		c.SetChecked(!c.Checked)
		return true
	}
	if action == "focus" {
		if canvas := granio.CurrentApp().Driver().CanvasForObject(c); canvas != nil {
			canvas.Focus(c)
			return true
		}
	}
	return false
}

func (*List) AccessibilityLabel() string                    { return "" }
func (*List) AccessibilityRole() granio.AccessibleRole      { return granio.AccessibleRoleList }
func (*Tree) AccessibilityLabel() string                    { return "" }
func (*Tree) AccessibilityRole() granio.AccessibleRole      { return granio.AccessibleRoleTree }
func (*TreeTable) AccessibilityRole() granio.AccessibleRole { return granio.AccessibleRoleTable }
func (*Table) AccessibilityLabel() string                   { return "" }
func (*Table) AccessibilityRole() granio.AccessibleRole     { return granio.AccessibleRoleTable }

func (n *treeNode) AccessibilityLabel() string {
	if accessible, ok := n.content.(granio.Accessible); ok {
		return accessible.AccessibilityLabel()
	}
	return n.uid
}
func (*treeNode) AccessibilityRole() granio.AccessibleRole { return granio.AccessibleRoleTreeItem }
func (n *treeNode) AccessibilityState() granio.AccessibleState {
	return granio.AccessibleState{Selected: n.tree.isSelected(n.uid), Expanded: n.isBranch && n.tree.IsBranchOpen(n.uid)}
}
func (n *treeNode) AccessibilityActions() []string {
	if n.isBranch {
		return []string{"select", "add-selection", "remove-selection", "toggle"}
	}
	return []string{"select", "add-selection", "remove-selection"}
}
func (n *treeNode) AccessibilityPerformAction(action string) bool {
	if action == "select" {
		n.tree.SetSelection([]TreeNodeID{n.uid})
		return true
	}
	if action == "add-selection" {
		n.tree.Select(n.uid)
		return true
	}
	if action == "remove-selection" {
		n.tree.Unselect(n.uid)
		return true
	}
	if action == "toggle" && n.isBranch {
		n.tree.ToggleBranch(n.uid)
		return true
	}
	return false
}

func (l *listItem) AccessibilityLabel() string {
	if accessible, ok := l.child.(granio.Accessible); ok {
		return accessible.AccessibilityLabel()
	}
	return ""
}
func (*listItem) AccessibilityRole() granio.AccessibleRole { return granio.AccessibleRoleListItem }
func (l *listItem) AccessibilityState() granio.AccessibleState {
	return granio.AccessibleState{Selected: l.selected}
}
func (*listItem) AccessibilityActions() []string {
	return []string{"select", "add-selection", "remove-selection"}
}
func (l *listItem) AccessibilityPerformAction(action string) bool {
	if action != "select" && action != "add-selection" && action != "remove-selection" {
		return false
	}
	if l.collection != nil {
		return l.collection.AccessibilitySelectIndex(l.index, action != "remove-selection", action == "add-selection")
	}
	if action == "select" {
		l.Tapped(&granio.PointEvent{})
		return true
	}
	return false
}

func (*Slider) AccessibilityLabel() string                    { return "" }
func (*Slider) AccessibilityRole() granio.AccessibleRole      { return granio.AccessibleRoleSlider }
func (s *Slider) AccessibilityValue() string                  { return strconv.FormatFloat(s.Value, 'g', -1, 64) }
func (*ProgressBar) AccessibilityLabel() string               { return "" }
func (*ProgressBar) AccessibilityRole() granio.AccessibleRole { return granio.AccessibleRoleProgress }
func (p *ProgressBar) AccessibilityValue() string             { return strconv.FormatFloat(p.Value, 'g', -1, 64) }
