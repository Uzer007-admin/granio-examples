package layout

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
)

// Declare conformity with Layout interface
var _ granio.Layout = (*centerLayout)(nil)

type centerLayout struct{}

// NewCenterLayout creates a new CenterLayout instance
func NewCenterLayout() granio.Layout {
	return &centerLayout{}
}

// Layout is called to pack all child objects into a specified size.
// For CenterLayout this sets all children to their minimum size, centered within the space.
func (*centerLayout) Layout(objects []granio.CanvasObject, size granio.Size) {
	for _, child := range objects {
		childMin := child.MinSize()
		child.Resize(childMin)
		child.Move(granio.NewPos((size.Width-childMin.Width)/2, (size.Height-childMin.Height)/2))
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For CenterLayout this is determined simply as the MinSize of the largest child.
func (*centerLayout) MinSize(objects []granio.CanvasObject) granio.Size {
	minSize := granio.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		minSize = internal.MaxSizes(minSize, child.MinSize())
	}

	return minSize
}
