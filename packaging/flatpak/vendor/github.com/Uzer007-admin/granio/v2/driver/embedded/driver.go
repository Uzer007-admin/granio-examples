// Package embedded implements the input/paint requirements for using in an embedded device
// (essentially replacement for the old v1 “set driver” concept).
package embedded

import (
	"image"

	"github.com/Uzer007-admin/granio/v2"
)

// Driver is an embedded driver designed for handling custom hardware.
// Various standard driver implementations are available in the fyne-x project.
//
// Since: 2.7
type Driver interface {
	Render(image.Image)
	Run(func())

	ScreenSize() granio.Size
	Queue() chan Event
}
