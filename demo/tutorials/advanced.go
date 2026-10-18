package tutorials

import (
	"strconv"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func scaleToString(c granio.Canvas) string {
	return strconv.FormatFloat(float64(c.Scale()), 'f', 2, 32)
}

func textureScaleToString(c granio.Canvas) string {
	pixels, _ := c.PixelCoordinateForPosition(granio.NewPos(1, 1))
	texScale := float32(pixels) / c.Scale()
	return strconv.FormatFloat(float64(texScale), 'f', 2, 32)
}

func prependTo(g *granio.Container, s string) {
	g.Objects = append([]granio.CanvasObject{widget.NewLabel(s)}, g.Objects...)
	g.Refresh()
}

// advancedScreen loads a panel that shows details and settings that are a bit
// more detailed than normally needed.
func advancedScreen(win granio.Window) granio.CanvasObject {
	scale := widget.NewLabel("")
	tex := widget.NewLabel("")

	screen := widget.NewCard("Screen info", "", widget.NewForm(
		&widget.FormItem{Text: "Scale", Widget: scale},
		&widget.FormItem{Text: "Texture Scale", Widget: tex},
	))

	ticker := time.NewTicker(time.Second)
	OnChangeFuncs = append(OnChangeFuncs, ticker.Stop)

	go func(canvas granio.Canvas) {
		for range ticker.C {
			granio.Do(func() {
				scale.SetText(scaleToString(canvas))
				tex.SetText(textureScaleToString(canvas))
			})
		}
	}(win.Canvas())

	label := widget.NewLabel("Just type...")
	generic := container.NewVBox()
	desk := container.NewVBox()

	genericCard := widget.NewCard("", "Generic", container.NewVScroll(generic))
	deskCard := widget.NewCard("", "Desktop", container.NewVScroll(desk))

	win.Canvas().SetOnTypedRune(func(r rune) {
		prependTo(generic, "Rune: "+string(r))
	})
	win.Canvas().SetOnTypedKey(func(ev *granio.KeyEvent) {
		prependTo(generic, "Key : "+string(ev.Name))
	})
	if deskCanvas, ok := win.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(ev *granio.KeyEvent) {
			prependTo(desk, "KeyDown: "+string(ev.Name))
		})
		deskCanvas.SetOnKeyUp(func(ev *granio.KeyEvent) {
			prependTo(desk, "KeyUp  : "+string(ev.Name))
		})
	}

	buildStatus, ok := granio.CurrentApp().Metadata().Custom["HelperText"]
	if !ok {
		buildStatus = "No helper text provided at compile time."
	}
	labelBuildStatus := widget.NewLabel(buildStatus)

	deskWin, ok := win.(desktop.Window)
	secondary := widget.NewButton("Fullscreen Secondary", func() {
		if win.FullScreen() {
			win.SetFullScreen(false)
		} else if ok {
			deskWin.RequestFullScreenSecondary()
		}
	})
	if !ok || !granio.CurrentApp().Driver().(desktop.Driver).HasSecondaryDisplay() {
		secondary.Disable()
	}

	return container.NewHBox(
		container.NewVBox(
			screen,
			widget.NewButton("Custom Theme", func() {
				granio.CurrentApp().Settings().SetTheme(newCustomTheme())
			}),
			widget.NewButton("Fullscreen", func() {
				win.SetFullScreen(!win.FullScreen())
			}),
			secondary,
		),
		container.NewBorder(
			label, labelBuildStatus, nil, nil,
			container.NewGridWithColumns(2, genericCard, deskCard),
		),
	)
}
