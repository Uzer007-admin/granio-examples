package main

import (
	"strconv"

	"github.com/Uzer007-admin/granio-x/wrapper"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	app := app.New()
	win := app.NewWindow("Wrapper example")
	win.Resize(granio.NewSize(400, 400))

	label1 := widget.NewLabel("Label 1, not wrapped")
	label2 := widget.NewLabel("Label 2, click me")
	label3 := widget.NewLabel("Label 3, move mouse over me")
	positionLabel := widget.NewLabel("Informations will be displayed here")

	wrapped := wrapper.MakeTappable(label2, func(e *granio.PointEvent) {
		dialog.ShowInformation("Tapped", "Label 1 was tapped", win)
	})

	mousable := wrapper.MakeHoverable(label3,
		func(e *desktop.MouseEvent) {
			positionLabel.SetText("Mouse in")
		}, func(e *desktop.MouseEvent) {
			posx := strconv.FormatFloat(float64(e.Position.X), 'f', 2, 32)
			posy := strconv.FormatFloat(float64(e.Position.Y), 'f', 2, 32)
			positionLabel.SetText("Mouse moved at:" + posx + ", " + posy)
		}, func() {
			positionLabel.SetText("Mouse out")
		},
	)

	mainContainer := container.NewGridWithColumns(2, label1, wrapped, mousable, positionLabel)

	win.SetContent(mainContainer)
	win.ShowAndRun()
}
