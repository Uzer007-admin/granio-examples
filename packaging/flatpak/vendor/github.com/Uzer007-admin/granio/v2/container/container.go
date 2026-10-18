// Package container provides containers that are used to lay out and organise applications.
package container

import (
	"github.com/Uzer007-admin/granio/v2"
)

// New returns a new Container instance holding the specified CanvasObjects which will be laid out according to the specified Layout.
//
// Since: 2.0
func New(layout granio.Layout, objects ...granio.CanvasObject) *granio.Container {
	return &granio.Container{Layout: layout, Objects: objects}
}

// NewWithoutLayout returns a new Container instance holding the specified CanvasObjects that are manually arranged.
//
// Since: 2.0
func NewWithoutLayout(objects ...granio.CanvasObject) *granio.Container {
	return &granio.Container{Objects: objects}
}
