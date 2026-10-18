// Package canvas contains all of the primitive CanvasObjects that make up a Fyne GUI.
//
// The types implemented in this package are used as building blocks in order
// to build higher order functionality. These types are designed to be
// non-interactive, by design. If additional functionality is required,
// it's usually a sign that this type should be used as part of a custom
// widget.
package canvas // import "github.com/Uzer007-admin/granio/v2/canvas"

import (
	"github.com/Uzer007-admin/granio/v2"
)

type baseObject struct {
	size     granio.Size     // The current size of the canvas object
	position granio.Position // The current position of the object
	Hidden   bool            // Is this object currently hidden

	min granio.Size // The minimum size this object can be
}

// Hide will set this object to not be visible.
func (o *baseObject) Hide() {
	o.Hidden = true
}

// MinSize returns the specified minimum size, if set, or {1, 1} otherwise.
func (o *baseObject) MinSize() granio.Size {
	if o.min.IsZero() {
		return granio.Size{Width: 1, Height: 1}
	}

	return o.min
}

// Move the object to a new position, relative to its parent.
func (o *baseObject) Move(pos granio.Position) {
	o.position = pos
}

// Position gets the current position of this canvas object, relative to its parent.
func (o *baseObject) Position() granio.Position {
	return o.position
}

// Resize sets a new size for the canvas object.
func (o *baseObject) Resize(size granio.Size) {
	o.size = size
}

// SetMinSize specifies the smallest size this object should be.
func (o *baseObject) SetMinSize(size granio.Size) {
	o.min = size
}

// Show will set this object to be visible.
func (o *baseObject) Show() {
	o.Hidden = false
}

// Size returns the current size of this canvas object.
func (o *baseObject) Size() granio.Size {
	return o.size
}

// Visible returns true if this object is visible, false otherwise.
func (o *baseObject) Visible() bool {
	return !o.Hidden
}
