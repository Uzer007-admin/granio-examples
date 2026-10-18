package main

import (
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello Person")

	out := widget.NewLabel("Hello World!")
	in := widget.NewEntry()
	in.SetPlaceHolder("World")
	w.SetContent(container.NewVBox(
		out, in,
		widget.NewButton("Greet", func() {
			out.SetText("Hello " + in.Text + "!")
		}),
	))

	w.ShowAndRun()
}
