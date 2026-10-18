package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
)

func main() {
	a := app.NewWithID("io.fyne.workshop.todo")
	w := a.NewWindow("TODO")

	data := a.Preferences().StringListWithFallback("todos",
		[]string{"Use this TODO list", "Build more Fyne apps"})
	w.SetContent(makeUI(data))

	w.Resize(granio.NewSize(180, 240))
	w.ShowAndRun()
}
