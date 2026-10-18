package container

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
)

// Scroll defines a container that is smaller than the Content.
// The Offset is used to determine the position of the child widgets within the container.
//
// Since: 1.4
type Scroll = widget.Scroll

// ScrollDirection represents the directions in which a Scroll container can scroll its child content.
//
// Since: 1.4
type ScrollDirection = granio.ScrollDirection

// Constants for valid values of ScrollDirection.
const (
	// ScrollBoth supports horizontal and vertical scrolling.
	ScrollBoth ScrollDirection = granio.ScrollBoth
	// ScrollHorizontalOnly specifies the scrolling should only happen left to right.
	ScrollHorizontalOnly = granio.ScrollHorizontalOnly
	// ScrollVerticalOnly specifies the scrolling should only happen top to bottom.
	ScrollVerticalOnly = granio.ScrollVerticalOnly
	// ScrollNone turns off scrolling for this container.
	//
	// Since: 2.1
	ScrollNone = granio.ScrollNone
)

// NewScroll creates a scrollable parent wrapping the specified content.
// Note that this may cause the MinSize to be smaller than that of the passed object.
//
// Since: 1.4
func NewScroll(content granio.CanvasObject) *Scroll {
	return widget.NewScroll(content)
}

// NewHScroll create a scrollable parent wrapping the specified content.
// Note that this may cause the MinSize.Width to be smaller than that of the passed object.
//
// Since: 1.4
func NewHScroll(content granio.CanvasObject) *Scroll {
	return widget.NewHScroll(content)
}

// NewVScroll a scrollable parent wrapping the specified content.
// Note that this may cause the MinSize.Height to be smaller than that of the passed object.
//
// Since: 1.4
func NewVScroll(content granio.CanvasObject) *Scroll {
	return widget.NewVScroll(content)
}
