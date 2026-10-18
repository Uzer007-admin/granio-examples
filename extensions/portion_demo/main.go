package main

import (
	"github.com/Uzer007-admin/granio-x/layout"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Portions")

	long := widget.NewButton("A very long", nil)
	shorter := widget.NewButton("Short", nil)
	long2 := widget.NewButton("I am also long", nil)
	btn := widget.NewButton("123", nil)
	w.SetContent(container.New(layout.NewHPortion([]float64{30, 20, 30, 10}), long, shorter, long2, btn))
	w.ShowAndRun()
}
