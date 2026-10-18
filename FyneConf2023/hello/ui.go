package main

import (
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello World")

	hello := widget.NewLabel("Hello World!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Welcome", func() {
			hello.SetText("Fyne Conf!")
		}),
	))

	w.ShowAndRun()
}
