package layout

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// Declare conformity with Layout interface
var _ granio.Layout = (*borderLayout)(nil)

type borderLayout struct {
	top, bottom, left, right granio.CanvasObject
}

// NewBorderLayout creates a new BorderLayout instance with top, bottom, left
// and right objects set. All other items in the container will fill the remaining space in the middle.
// Multiple extra items will be stacked in the specified order as a Stack container.
func NewBorderLayout(top, bottom, left, right granio.CanvasObject) granio.Layout {
	return &borderLayout{top, bottom, left, right}
}

// Layout is called to pack all child objects into a specified size.
// For BorderLayout this arranges the top, bottom, left and right widgets at
// the sides and any remaining widgets are maximised in the middle space.
func (b *borderLayout) Layout(objects []granio.CanvasObject, size granio.Size) {
	padding := theme.Padding()
	var topSize, bottomSize, leftSize, rightSize granio.Size
	if b.top != nil && b.top.Visible() {
		topHeight := b.top.MinSize().Height
		b.top.Resize(granio.NewSize(size.Width, topHeight))
		b.top.Move(granio.NewPos(0, 0))
		topSize = granio.NewSize(size.Width, topHeight+padding)
	}
	if b.bottom != nil && b.bottom.Visible() {
		bottomHeight := b.bottom.MinSize().Height
		b.bottom.Resize(granio.NewSize(size.Width, bottomHeight))
		b.bottom.Move(granio.NewPos(0, size.Height-bottomHeight))
		bottomSize = granio.NewSize(size.Width, bottomHeight+padding)
	}
	if b.left != nil && b.left.Visible() {
		leftWidth := b.left.MinSize().Width
		b.left.Resize(granio.NewSize(leftWidth, size.Height-topSize.Height-bottomSize.Height))
		b.left.Move(granio.NewPos(0, topSize.Height))
		leftSize = granio.NewSize(leftWidth+padding, size.Height-topSize.Height-bottomSize.Height)
	}
	if b.right != nil && b.right.Visible() {
		rightWidth := b.right.MinSize().Width
		b.right.Resize(granio.NewSize(rightWidth, size.Height-topSize.Height-bottomSize.Height))
		b.right.Move(granio.NewPos(size.Width-rightWidth, topSize.Height))
		rightSize = granio.NewSize(rightWidth+padding, size.Height-topSize.Height-bottomSize.Height)
	}

	middleSize := granio.NewSize(size.Width-leftSize.Width-rightSize.Width, size.Height-topSize.Height-bottomSize.Height)
	middlePos := granio.NewPos(leftSize.Width, topSize.Height)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if child != b.top && child != b.bottom && child != b.left && child != b.right {
			child.Resize(middleSize)
			child.Move(middlePos)
		}
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For BorderLayout this is determined by the MinSize height of the top and
// plus the MinSize width of the left and right, plus any padding needed.
// This is then added to the union of the MinSize for any remaining content.
func (b *borderLayout) MinSize(objects []granio.CanvasObject) granio.Size {
	minSize := granio.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		if child != b.top && child != b.bottom && child != b.left && child != b.right {
			minSize = internal.MaxSizes(minSize, child.MinSize())
		}
	}

	padding := theme.Padding()

	if b.left != nil && b.left.Visible() {
		leftMin := b.left.MinSize()
		minHeight := max(minSize.Height, leftMin.Height)
		minSize = granio.NewSize(minSize.Width+leftMin.Width+padding, minHeight)
	}
	if b.right != nil && b.right.Visible() {
		rightMin := b.right.MinSize()
		minHeight := max(minSize.Height, rightMin.Height)
		minSize = granio.NewSize(minSize.Width+rightMin.Width+padding, minHeight)
	}

	if b.top != nil && b.top.Visible() {
		topMin := b.top.MinSize()
		minWidth := max(minSize.Width, topMin.Width)
		minSize = granio.NewSize(minWidth, minSize.Height+topMin.Height+padding)
	}
	if b.bottom != nil && b.bottom.Visible() {
		bottomMin := b.bottom.MinSize()
		minWidth := max(minSize.Width, bottomMin.Width)
		minSize = granio.NewSize(minWidth, minSize.Height+bottomMin.Height+padding)
	}

	return minSize
}
