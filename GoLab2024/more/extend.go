package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.NewWithID("io.fyne.workshop.extend")
	w := a.NewWindow("Extend Label")

	w.SetContent(newRightClickLabel("Right Click Me", w.Canvas()))

	w.ShowAndRun()
}

type rightClickLabel struct {
	widget.Label

	canvas granio.Canvas
}

func newRightClickLabel(text string, c granio.Canvas) granio.Widget {
	r := &rightClickLabel{canvas: c}
	r.Text = text
	r.ExtendBaseWidget(r)

	return r
}

func (r *rightClickLabel) TappedSecondary(e *granio.PointEvent) {
	m := granio.NewMenu("",
		granio.NewMenuItem("Item 1", func() {}),
		granio.NewMenuItem("Item 2", func() {}))
	widget.ShowPopUpMenuAtPosition(m, r.canvas, e.AbsolutePosition)
}
