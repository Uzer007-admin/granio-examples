package ui

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestDefaultZoom(t *testing.T) {
	file := testFile("8x8")
	e := testEditorWithFile(file)

	p := newPalette(e.(*editor))
	zoom := p.(*granio.Container).Objects[2].(*granio.Container).Objects[1].(*widget.Label)
	assert.Equal(t, "100%", zoom.Text)
}

func TestZoomIn(t *testing.T) {
	file := testFile("8x8")
	e := testEditorWithFile(file)
	assert.Equal(t, 1, e.(*editor).zoom)

	p := newPalette(e.(*editor))
	zoomItems := p.(*granio.Container).Objects[2].(*granio.Container).Objects
	zoom := zoomItems[1].(*widget.Label)
	zoomIn := zoomItems[2].(*widget.Button)
	test.Tap(zoomIn)

	assert.Equal(t, 2, e.(*editor).zoom)
	assert.Equal(t, "200%", zoom.Text)
}
