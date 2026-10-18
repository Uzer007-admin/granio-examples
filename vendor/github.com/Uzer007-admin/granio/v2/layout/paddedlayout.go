package layout

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// Declare conformity with Layout interface
var _ granio.Layout = (*paddedLayout)(nil)

type paddedLayout struct{}

// Layout is called to pack all child objects into a specified size.
// For PaddedLayout this sets all children to the full size passed minus padding all around.
func (paddedLayout) Layout(objects []granio.CanvasObject, size granio.Size) {
	padding := theme.Padding()
	pos := granio.NewSquareOffsetPos(padding)
	siz := granio.NewSize(size.Width-2*padding, size.Height-2*padding)
	for _, child := range objects {
		child.Resize(siz)
		child.Move(pos)
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For PaddedLayout this is determined simply as the MinSize of the largest child plus padding all around.
func (paddedLayout) MinSize(objects []granio.CanvasObject) (minSize granio.Size) {
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		minSize = internal.MaxSizes(minSize, child.MinSize())
	}
	minSize = minSize.Add(granio.NewSquareSize(2 * theme.Padding()))
	return minSize
}

// NewPaddedLayout creates a new PaddedLayout instance
//
// Since: 1.4
func NewPaddedLayout() granio.Layout {
	return paddedLayout{}
}
