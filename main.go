//go:generate granio bundle -o data.go Icon.png

package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/Uzer007-admin/granio-examples/bugs"
	"github.com/Uzer007-admin/granio-examples/clock"
	"github.com/Uzer007-admin/granio-examples/fractal"
	"github.com/Uzer007-admin/granio-examples/img/icon"
	"github.com/Uzer007-admin/granio-examples/tictactoe"
	"github.com/Uzer007-admin/granio-examples/xkcd"
)

type appInfo struct {
	name string
	icon granio.Resource
	canv bool
	run  func(granio.Window) granio.CanvasObject
}

var apps = []appInfo{
	{"Bugs", icon.BugBitmap, false, bugs.Show},
	{"XKCD", icon.XKCDBitmap, false, xkcd.Show},
	{"Clock", icon.ClockBitmap, true, clock.Show},
	{"Fractal", icon.FractalBitmap, true, fractal.Show},
	{"Tic Tac Toe", nil, true, tictactoe.Show},
}

func main() {
	a := app.New()
	a.SetIcon(resourceIconPng)

	content := container.NewStack()
	w := a.NewWindow("Примеры Granio")

	apps[4].icon = theme.RadioButtonIcon() // lazy load Fyne resource to avoid error
	appList := widget.NewList(
		func() int {
			return len(apps)
		},
		func() granio.CanvasObject {
			icon := &canvas.Image{}
			label := widget.NewLabel("Text Editor")
			labelHeight := label.MinSize().Height
			icon.SetMinSize(granio.NewSize(labelHeight, labelHeight))
			return container.NewBorder(nil, nil, icon, nil,
				label)
		},
		func(id widget.ListItemID, obj granio.CanvasObject) {
			img := obj.(*granio.Container).Objects[1].(*canvas.Image)
			text := obj.(*granio.Container).Objects[0].(*widget.Label)
			img.Resource = apps[id].icon
			img.Refresh()
			text.SetText(apps[id].name)
		})
	appList.OnSelected = func(id widget.ListItemID) {
		content.Objects = []granio.CanvasObject{apps[id].run(w)}
	}

	split := container.NewHSplit(appList, content)
	split.Offset = 0.1
	w.SetContent(split)
	w.Resize(granio.NewSize(480, 360))
	w.ShowAndRun()
}
