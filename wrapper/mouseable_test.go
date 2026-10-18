package wrapper

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func TestMouseable(t *testing.T) {
	app := test.NewApp()
	win := app.NewWindow("Test")

	label := widget.NewLabel("Label 1, not wrapped")
	label2 := widget.NewLabel("Label, mouseable")

	pos := granio.NewPos(0, 0)
	in := false
	out := false
	mouseable := MakeHoverable(label, func(e *desktop.MouseEvent) {
		in = true
	}, func(e *desktop.MouseEvent) {
		pos = e.Position
	}, func() {
		out = true
	})

	win.SetContent(container.NewHBox(label2, mouseable))

	moveTo := granio.NewPos(15, 15)
	expectedPos := mouseable.Position().Add(moveTo)
	test.MoveMouse(win.Canvas(), mouseable.Position().Add(granio.NewPos(5, 5))) // to place the mouse
	if !in {
		t.Error("MouseIn was not called")
	}
	test.MoveMouse(win.Canvas(), expectedPos) // to move the mouse
	if pos != moveTo.Subtract(granio.NewPos(theme.Padding(), theme.Padding())) {
		t.Error("MouseMoved was not called", pos, moveTo)
	}
	if out {
		t.Error("MouseOut was called")
	}

	test.MoveMouse(win.Canvas(), granio.NewPos(0, 0)) // go out now
	if !out {
		t.Error("MouseOut was not called")
	}

}
