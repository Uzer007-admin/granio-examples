package main

import (
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.NewWithID("io.fyne.workshop.extend")
	w := a.NewWindow("Bind Name")

	name := binding.BindPreferenceString("user.name", a.Preferences())
	w.SetContent(container.NewHBox(
		widget.NewLabel("Please enter your name:"),
		widget.NewEntryWithData(name),
	))

	w.ShowAndRun()
}
