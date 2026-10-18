package main

import (
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Counter")

	count := binding.NewInt()
	button := widget.NewButton("Count", func() {
		i, _ := count.Get()
		_ = count.Set(i + 1)
	})

	w.SetContent(container.NewGridWithColumns(2,
		widget.NewLabelWithData(binding.IntToString(count)), button))
	w.ShowAndRun()
}
