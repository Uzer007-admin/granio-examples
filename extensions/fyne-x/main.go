package main

import (
	"net/url"

	"github.com/Uzer007-admin/granio-x/dialog"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Fyne-x demo")

	docURL, _ := url.Parse("https://docs.fyne.io")
	links := []*widget.Hyperlink{
		widget.NewHyperlink("Docs", docURL),
	}
	w.SetContent(container.NewGridWithColumns(1,
		widget.NewButton("About", func() {
			dialog.ShowAbout("Some **cool** stuff", links, a, w)
		}),
		widget.NewButton("About window", func() {
			dialog.ShowAboutWindow("Some **cool** stuff", links, a)
		}),
	))

	w.ShowAndRun()
}
