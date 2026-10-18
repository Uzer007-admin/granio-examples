package wrapper

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func TestTappable(t *testing.T) {
	app := test.NewApp()
	win := app.NewWindow("Test")
	win.Resize(granio.NewSize(400, 400))
	defer win.Close()

	label1 := widget.NewLabel("Label 1, not wrapped")
	label2 := widget.NewLabel("Label 2, click me")

	tapped := false
	wrapped := MakeTappable(label2, func(e *granio.PointEvent) {
		tapped = true
	})

	mainContainer := container.NewGridWithColumns(2, label1, wrapped)
	win.SetContent(mainContainer)

	// Click on wrapped label
	test.Tap(wrapped.(granio.Tappable))
	if !tapped {
		t.Error("Tapped event not fired")
	}
}
