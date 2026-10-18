package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// TreeTable displays a virtualized hierarchy with aligned columns. The first
// column contains the tree expander. Tree's selection, keyboard navigation and
// asynchronous bindings remain available through the embedded Tree.
type TreeTable struct {
	Tree
	Columns      int
	CreateCell   func(column int) granio.CanvasObject                      `json:"-"`
	UpdateCell   func(id TreeNodeID, column int, cell granio.CanvasObject) `json:"-"`
	columnWidths map[int]float32
	depths       map[TreeNodeID]int
}

// NewTreeTable creates a hierarchy with columns equal in width until changed by
// SetColumnWidth. Cells are created for viewport rows and recycled while scrolling.
func NewTreeTable(children func(TreeNodeID) []TreeNodeID, branch func(TreeNodeID) bool, columns int,
	create func(int) granio.CanvasObject, update func(TreeNodeID, int, granio.CanvasObject)) *TreeTable {
	t := &TreeTable{Columns: max(0, columns), CreateCell: create, UpdateCell: update,
		columnWidths: make(map[int]float32), depths: make(map[TreeNodeID]int)}
	t.ChildUIDs = func(id TreeNodeID) []TreeNodeID {
		if children == nil {
			return nil
		}
		ids := children(id)
		depth := t.depths[id]
		if id == t.Root && t.Root == "" {
			depth = -1
		}
		for _, child := range ids {
			t.depths[child] = depth + 1
		}
		return ids
	}
	t.IsBranch = branch
	t.CreateNode = func(bool) granio.CanvasObject { return newTreeTableRow(t) }
	t.UpdateNode = func(id TreeNodeID, _ bool, object granio.CanvasObject) {
		row := object.(*treeTableRow)
		row.depth = t.depths[id]
		for col, cell := range row.cells {
			if t.UpdateCell != nil {
				t.UpdateCell(id, col, cell)
			}
		}
		row.Refresh()
	}
	t.ExtendBaseWidget(t)
	return t
}

// SetColumnWidth changes a column's width in canvas units. The first column's
// width includes the expander and indentation. Negative widths are ignored.
func (t *TreeTable) SetColumnWidth(column int, width float32) {
	if column < 0 || column >= t.Columns || width < 0 {
		return
	}
	t.columnWidths[column] = width
	t.Refresh()
}

// ColumnWidth returns the configured width, or 120 canvas units by default.
func (t *TreeTable) ColumnWidth(column int) float32 {
	if width, ok := t.columnWidths[column]; ok {
		return width
	}
	return 120
}

type treeTableRow struct {
	BaseWidget
	table *TreeTable
	cells []granio.CanvasObject
	depth int
}

func newTreeTableRow(table *TreeTable) *treeTableRow {
	row := &treeTableRow{table: table}
	for col := 0; col < table.Columns; col++ {
		if table.CreateCell != nil {
			row.cells = append(row.cells, createItemAndApplyThemeScope(func() granio.CanvasObject { return table.CreateCell(col) }, table))
		}
	}
	row.ExtendBaseWidget(row)
	return row
}

func (row *treeTableRow) CreateRenderer() granio.WidgetRenderer {
	return &treeTableRowRenderer{row: row}
}

type treeTableRowRenderer struct{ row *treeTableRow }

func (r *treeTableRowRenderer) Destroy()                       {}
func (r *treeTableRowRenderer) Objects() []granio.CanvasObject { return r.row.cells }
func (r *treeTableRowRenderer) MinSize() granio.Size {
	var size granio.Size
	for col, cell := range r.row.cells {
		size.Width += r.row.table.ColumnWidth(col)
		size.Height = max(size.Height, cell.MinSize().Height)
	}
	// Tree's node renderer contributes expander/indentation separately.
	size.Width = max(0, size.Width-r.leading())
	return size
}

func (r *treeTableRowRenderer) leading() float32 {
	th := r.row.table.Theme()
	pad, icon := th.Size(theme.SizeNamePadding), th.Size(theme.SizeNameInlineIcon)
	return 2*pad + icon + float32(r.row.depth)*(icon+pad)
}

func (r *treeTableRowRenderer) Layout(size granio.Size) {
	x := float32(0)
	for col, cell := range r.row.cells {
		width := r.row.table.ColumnWidth(col)
		if col == 0 {
			width = max(0, width-r.leading())
		}
		cell.Move(granio.NewPos(x, 0))
		cell.Resize(granio.NewSize(width, size.Height))
		x += width
	}
}

func (r *treeTableRowRenderer) Refresh() {
	r.Layout(r.row.Size())
	for _, cell := range r.row.cells {
		cell.Refresh()
	}
	canvas.Refresh(r.row)
}
