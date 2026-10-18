package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func makeUI(data binding.StringList) granio.CanvasObject {
	list := widget.NewListWithData(data,
		func() granio.CanvasObject {
			return widget.NewCheck("An item title", nil)
		},
		func(di binding.DataItem, co granio.CanvasObject) {
			str, _ := di.(binding.String).Get()
			ch := co.(*widget.Check)
			ch.OnChanged = func(done bool) {
				if !done {
					return
				}
				ch.SetChecked(false)

				// TODO StringList.Remove
				items, _ := data.Get()
				items = removeItem(items, ch.Text)
				data.Set(items)
			}
			ch.SetText(str)
		})

	input := widget.NewEntry()
	add := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		if input.Text == "" {
			return
		}

		data.Prepend(input.Text)
		input.SetText("")
	})

	top := container.NewBorder(nil, nil, nil, add, input)
	return container.NewBorder(top, nil, nil, nil, list)
}

func removeItem(items []string, rm string) []string {
	for i, s := range items {
		if s != rm {
			continue
		}

		if i == len(items)-1 {
			items = items[:i]
		} else {
			items = append(items[:i], items[i+1:]...)
		}
		break
	}

	return items
}
