package ui

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestInteractiveRaster_MinSize(t *testing.T) {
	file := testFile("8x8")
	e := testEditorWithFile(file).(*editor)

	rast := newInteractiveRaster(e)
	e.drawSurface = rast
	e.setZoom(1)
	assert.Equal(t, granio.NewSize(8, 8), rast.MinSize())
	assert.Equal(t, granio.NewSize(8, 8), rast.Size())

	e.setZoom(2)
	assert.Equal(t, granio.NewSize(16, 16), rast.MinSize())
	assert.Equal(t, granio.NewSize(16, 16), rast.Size())
}

func TestInteractiveRaster_locationForPositon(t *testing.T) {
	file := testFile("8x8")
	e := testEditorWithFile(file).(*editor)

	r := newInteractiveRaster(e)
	x, y := r.locationForPosition(granio.NewPos(2, 2))
	assert.Equal(t, 2, x)
	assert.Equal(t, 2, y)

	testCanvas := granio.CurrentApp().Driver().CanvasForObject(r).(test.WindowlessCanvas)
	testCanvas.SetScale(2.0)
	x, y = r.locationForPosition(granio.NewPos(2, 2))
	assert.Equal(t, 4, x)
	assert.Equal(t, 4, y)
}
