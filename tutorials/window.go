package tutorials

import (
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func windowScreen(_ granio.Window) granio.CanvasObject {
	var visibilityWindow granio.Window = nil
	var visibilityState bool = false

	windowGroup := container.NewVBox(
		widget.NewButton("New window", func() {
			w := granio.CurrentApp().NewWindow("Hello")
			w.SetContent(widget.NewLabel("Hello World!"))
			w.Show()
		}),
		widget.NewButton("Fixed size window", func() {
			w := granio.CurrentApp().NewWindow("Fixed")
			w.SetContent(container.NewCenter(widget.NewLabel("Hello World!")))

			w.Resize(granio.NewSize(240, 180))
			w.SetFixedSize(true)
			w.Show()
		}),
		widget.NewButton("Toggle between fixed/not fixed window size", func() {
			w := granio.CurrentApp().NewWindow("Toggle fixed size")
			w.SetContent(container.NewCenter(widget.NewCheck("Fixed size", func(toggle bool) {
				if toggle {
					w.Resize(granio.NewSize(240, 180))
				}
				w.SetFixedSize(toggle)
			})))
			w.Show()
		}),
		widget.NewButton("Centered window", func() {
			w := granio.CurrentApp().NewWindow("Central")
			w.SetContent(container.NewCenter(widget.NewLabel("Hello World!")))

			w.CenterOnScreen()
			w.Show()
		}),
		widget.NewButton("Always on top window", func() {
			w := granio.CurrentApp().NewWindow("On Top")
			w.SetContent(container.NewCenter(widget.NewLabel("Hello World!")))

			deskWin, ok := w.(desktop.Window)
			if ok {
				deskWin.RequestAlwaysOnTop()
			}
			w.Show()
		}),
		widget.NewButton("Show/Hide window", func() {
			if visibilityWindow == nil {
				visibilityWindow = granio.CurrentApp().NewWindow("Hello")
				visibilityWindow.SetContent(widget.NewLabel("Hello World!"))
				visibilityWindow.Show()
				visibilityWindow.SetOnClosed(func() {
					visibilityWindow = nil
					visibilityState = !visibilityState
				})
			}
			if visibilityState {
				visibilityWindow.Hide()
			} else {
				visibilityWindow.Show()
			}
			visibilityState = !visibilityState
		}))

	drv := granio.CurrentApp().Driver()
	if drv, ok := drv.(desktop.Driver); ok {
		windowGroup.Objects = append(windowGroup.Objects,
			widget.NewButton("Splash Window (only use on start)", func() {
				w := drv.CreateSplashWindow()
				w.SetContent(widget.NewLabelWithStyle("Hello World!\n\nMake a splash!",
					granio.TextAlignCenter, granio.TextStyle{Bold: true}))
				w.Show()

				time.AfterFunc(3*time.Second, func() {
					granio.Do(w.Close)
				})
			}))
	}

	otherGroup := widget.NewCard("Other", "",
		container.NewVBox(
			widget.NewButton("Notification", func() {
				granio.CurrentApp().SendNotification(&granio.Notification{
					Title:   "Fyne Demo",
					Content: "This is a notification",
				})
			}),
			widget.NewButton("Notification in 30 sec", func() {
				_, _ = granio.CurrentApp().ScheduleNotification(&granio.Notification{
					Title:   "Fyne Demo",
					Content: "A scheduled notification",
				}, time.Now().Add(time.Second*30))
			}),
		))

	return container.NewVBox(widget.NewCard("Windows", "", windowGroup), otherGroup)
}
