package layout

import (
	"math"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// Declare conformity with Layout interface
var _ granio.Layout = (*gridWrapLayout)(nil)

type gridWrapLayout struct {
	CellSize granio.Size
	colCount int
	rowCount int
}

// NewGridWrapLayout returns a new GridWrapLayout instance
func NewGridWrapLayout(size granio.Size) granio.Layout {
	return &gridWrapLayout{size, 1, 1}
}

// Layout is called to pack all child objects into a specified size.
// For a GridWrapLayout this will attempt to lay all the child objects in a row
// and wrap to a new row if the size is not large enough.
func (g *gridWrapLayout) Layout(objects []granio.CanvasObject, size granio.Size) {
	padding := theme.Padding()
	g.colCount = 1
	g.rowCount = 0

	if size.Width > g.CellSize.Width {
		g.colCount = int(math.Floor(float64(size.Width+padding) / float64(g.CellSize.Width+padding)))
	}

	i, x, y := 0, float32(0), float32(0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if i%g.colCount == 0 {
			g.rowCount++
		}

		child.Move(granio.NewPos(x, y))
		child.Resize(g.CellSize)

		if (i+1)%g.colCount == 0 {
			x = 0
			y += g.CellSize.Height + padding
		} else {
			x += g.CellSize.Width + padding
		}
		i++
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For a GridWrapLayout this is simply the specified cellsize as a single column
// layout has no padding. The returned size does not take into account the number
// of columns as this layout re-flows dynamically.
func (g *gridWrapLayout) MinSize([]granio.CanvasObject) granio.Size {
	rows := g.rowCount
	if rows < 1 {
		rows = 1
	}
	return granio.NewSize(g.CellSize.Width,
		(g.CellSize.Height*float32(rows))+(float32(rows-1)*theme.Padding()))
}
