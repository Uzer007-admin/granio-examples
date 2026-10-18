package granio

// AccessibleRole describes the different roles an accessible element can take.
//
// Since: 2.8
type AccessibleRole string

// Known values for [AccessibleRole].
const (
	AccessibleRoleButton    AccessibleRole = "button"
	AccessibleRoleContainer AccessibleRole = "container"
	AccessibleRoleLink      AccessibleRole = "link"
	AccessibleRoleText      AccessibleRole = "text"
	AccessibleRoleInput     AccessibleRole = "input"
	AccessibleRoleCheck     AccessibleRole = "check"
	AccessibleRoleList      AccessibleRole = "list"
	AccessibleRoleTable     AccessibleRole = "table"
	AccessibleRoleTree      AccessibleRole = "tree"
	AccessibleRoleTreeItem  AccessibleRole = "treeitem"
	AccessibleRoleListItem  AccessibleRole = "listitem"
	AccessibleRoleCell      AccessibleRole = "cell"
	AccessibleRoleSlider    AccessibleRole = "slider"
	AccessibleRoleProgress  AccessibleRole = "progress"
	AccessibleRoleSelect    AccessibleRole = "select"
	AccessibleRoleImage     AccessibleRole = "image"
	AccessibleRoleWindow    AccessibleRole = "window"
)

// Accessible interface should be implemented for a widget that should be accessible
//
// Since: 2.8
type Accessible interface {
	AccessibilityLabel() string
	AccessibilityRole() AccessibleRole
}

// AccessibleValue exposes a control's current value to assistive technology.
// Password controls must return an empty value.
type AccessibleValue interface{ AccessibilityValue() string }

// AccessibleAction exposes actions that assistive technology can request.
// Actions run on the UI thread and return false when unavailable.
type AccessibleAction interface {
	AccessibilityActions() []string
	AccessibilityPerformAction(string) bool
}

// AccessibleState describes control states independently of native protocols.
type AccessibleState struct {
	Disabled, Focused, Selected, Checked, Expanded, Sensitive bool
	ReadOnly                                                  bool
	Active                                                    bool // active top-level window
}

type AccessibleStateProvider interface{ AccessibilityState() AccessibleState }

// AccessibleEditableText exposes committed text and rune offsets. Native adapters
// dispatch mutations on the UI thread. Password controls must conceal their text.
type AccessibleEditableText interface {
	AccessibleValue
	AccessibilityTextSelection() (cursor, start, end int)
	AccessibilitySetText(string) bool
	AccessibilitySetSelection(start, end int) bool
}

// AccessibleTextRectangle describes a visible piece of a text range, relative
// to its control. Implementations clip rectangles to their text viewport.
type AccessibleTextRectangle struct {
	Position Position
	Size     Size
}

// AccessibleTextGeometry exposes real layout geometry instead of estimating a
// character position from the control's complete bounding box. Offsets are runes.
type AccessibleTextGeometry interface {
	AccessibilityTextRectangles(start, end int) []AccessibleTextRectangle
	AccessibilityTextOffsetAtPosition(Position) int
}

// AccessibleTextScroller reveals a rune range without changing its selection,
// caret, committed text, or undo history. alignTop chooses the viewport edge.
type AccessibleTextScroller interface {
	AccessibilityScrollTextRange(start, end int, alignTop bool) bool
}

// AccessibleNumeric exposes numeric controls to native range/value patterns.
type AccessibleNumeric interface {
	AccessibilityNumericValue() (value, minimum, maximum, step float64)
	AccessibilitySetNumericValue(float64) bool
}

// AccessibleSelection exposes complete logical indices, including virtual items.
// extend=false replaces the selection; selected=false removes an item.
type AccessibleSelection interface {
	AccessibilityCanSelectMultiple() bool
	AccessibilitySelection() (selected []int, multiple bool)
	AccessibilitySelectIndex(index int, selected, extend bool) bool
}

// AccessibleSelectionBatch changes a logical selection with one refresh.
type AccessibleSelectionBatch interface{ AccessibilitySetSelection([]int) bool }

// AccessibleTableCell connects a visible recycled renderer to a logical cell.
type AccessibleTableCell struct {
	Object      CanvasObject
	Row, Column int
	Selected    bool
}

// AccessibleTable preserves virtualization: the adapter asks for one cell and
// may scroll it into view instead of allocating all cells for a native client.
type AccessibleTable interface {
	AccessibilityTableSize() (rows, columns int)
	AccessibilityVisibleCells() []AccessibleTableCell
	AccessibilityTableCell(row, column int) CanvasObject
}

// AccessibleIdentity identifies a logical item across renderer reuse. The name
// is scoped to its collection, and must change when a renderer displays another item.
type AccessibleIdentity interface{ AccessibilityIdentity() string }

// AccessibleCollectionItem maps a recycled child to its collection's logical
// selection index. Native child indices still refer to the exposed child list.
type AccessibleCollectionItem interface {
	AccessibilityCollectionIndex() (collection CanvasObject, index int)
}

// AccessibleVirtualItem describes a logical collection item without retaining a
// renderer. IDs are stable within the collection. Parent is empty for a direct
// child, and Index is the offset among that parent's children. Offscreen also
// covers descendants of collapsed branches; querying them must not open a branch.
type AccessibleVirtualItem struct {
	ID, Parent, Label, Value       string
	Role                           AccessibleRole
	State                          AccessibleState
	Index, ChildCount, Row, Column int
	Offscreen                      bool
}

// AccessibleVirtualCollection exposes the complete data hierarchy on demand.
// Methods run on the UI thread. They must not scroll, focus or change selection
// when reading metadata. Empty parent addresses the collection's children.
type AccessibleVirtualCollection interface {
	AccessibilityVirtualCount(parent string) int
	AccessibilityVirtualChild(parent string, index int) (string, bool)
	AccessibilityVirtualItem(id string) (AccessibleVirtualItem, bool)
	AccessibilityVirtualSelection() []string
	AccessibilityVirtualAction(id, action string) bool
}
