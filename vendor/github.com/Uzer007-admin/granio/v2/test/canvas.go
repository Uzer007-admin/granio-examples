package test

import (
	"github.com/Uzer007-admin/granio/v2"
	fynedriver "github.com/Uzer007-admin/granio/v2/driver"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/driver/software"
)

// WindowlessCanvas provides functionality for a canvas to operate without a window
//
// Deprecated: Use driver/software.WindowlessCanvas instead.
type WindowlessCanvas = software.WindowlessCanvas

var dummyCanvas software.WindowlessCanvas

// Canvas returns a reusable in-memory canvas used for testing
func Canvas() granio.Canvas {
	if dummyCanvas == nil {
		dummyCanvas = NewCanvas()
	}

	return dummyCanvas
}

// NewCanvas returns a single use in-memory canvas used for testing.
// This canvas has no painter so calls to Capture() will return a blank image.
func NewCanvas() software.WindowlessCanvas {
	return wrapCanvas(software.NewCanvasWithPainter(nil))
}

// NewCanvasWithPainter allows creation of an in-memory canvas with a specific painter.
// The painter will be used to render in the Capture() call.
func NewCanvasWithPainter(painter fynedriver.Painter) software.WindowlessCanvas {
	return wrapCanvas(software.NewCanvasWithPainter(painter))
}

// NewTransparentCanvasWithPainter allows creation of an in-memory canvas with a specific painter without a background color.
// The painter will be used to render in the Capture() call.
//
// Since: 2.2
func NewTransparentCanvasWithPainter(painter fynedriver.Painter) software.WindowlessCanvas {
	return wrapCanvas(software.NewTransparentCanvasWithPainter(painter))
}

func wrapCanvas(c software.WindowlessCanvas) *canvas {
	return &canvas{WindowlessCanvas: c}
}

type canvas struct {
	software.WindowlessCanvas
	hovered desktop.Hoverable
}
