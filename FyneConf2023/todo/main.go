package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/data/binding"
)

func main() {
	a := app.NewWithID("io.fyne.workshop.todo")
	w := a.NewWindow("TODO")

	data := binding.NewStringList()
	data.Set(a.Preferences().StringListWithFallback("todos",
		[]string{"Use this TODO list", "Build more Fyne apps"}))
	w.SetContent(makeUI(data))

	// work around missing BindPreferenceString
	data.AddListener(binding.NewDataListener(func() {
		vals, _ := data.Get()
		a.Preferences().SetStringList("todos", vals)
	}))

	w.Resize(granio.NewSize(180, 240))
	w.ShowAndRun()
}
