package main

import (
	"slices"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func makeUI(data []string) granio.CanvasObject {
	var list *widget.List
	list = widget.NewList(
		func() int {
			return len(data)
		},
		func() granio.CanvasObject {
			return widget.NewCheck("An item title", nil)
		},
		func(i widget.ListItemID, co granio.CanvasObject) {
			ch := co.(*widget.Check)
			ch.SetText(data[i])

			ch.OnChanged = func(done bool) {
				if !done {
					return
				}
				ch.SetChecked(false)

				data = slices.Delete(data, i, i+1)
				list.Refresh()
				granio.CurrentApp().Preferences().SetStringList("todos", data)
			}
		})

	input := widget.NewEntry()
	add := widget.NewButtonWithIcon("",
		theme.ContentAddIcon(), func() {
			if input.Text == "" {
				return
			}

			data = append(data, input.Text)
			list.Refresh()
			input.SetText("")
			granio.CurrentApp().Preferences().SetStringList("todos", data)
		})

	top := container.NewBorder(nil, nil, nil, add, input)
	return container.NewBorder(top, nil, nil, nil, list)
}
