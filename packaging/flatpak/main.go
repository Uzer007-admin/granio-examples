package main

import (
	_ "embed"
	"io"
	"log"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

//go:embed Icon.png
var iconData []byte

func main() {
	a := app.NewWithID("io.fyne.flatpak_demo")
	a.SetIcon(&granio.StaticResource{StaticName: "Icon.png", StaticContent: iconData})

	w := a.NewWindow("Flatpak Demo")

	markdown := &widget.Entry{MultiLine: true, Wrapping: granio.TextWrapWord}
	preview := &widget.RichText{Wrapping: granio.TextWrapWord}
	markdown.OnChanged = preview.ParseMarkdown

	open := &widget.Button{Text: "Open file", Icon: theme.ContentAddIcon(), OnTapped: func() {
		files := dialog.NewFileOpen(func(file granio.URIReadCloser, err error) {
			if err != nil {
				log.Println(err)
				return
			} else if file == nil {
				return
			}

			text, err := io.ReadAll(file)
			if err != nil {
				log.Println(err)
				return
			}

			markdown.SetText(string(text))
			a.SendNotification(&granio.Notification{Title: "Opened a file", Content: file.URI().Name() + " was opened correctly."})
		}, w)

		files.SetFilter(storage.NewExtensionFileFilter([]string{".md"}))
		files.Show()
	}}

	w.SetContent(
		container.NewBorder(container.NewHBox(open), nil, nil, nil,
			container.NewHSplit(markdown, preview),
		),
	)

	if desk, ok := a.(desktop.App); ok {
		m := granio.NewMenu("Flatpak Demo", granio.NewMenuItem("Show", w.Show))
		desk.SetSystemTrayMenu(m)
		w.SetCloseIntercept(w.Hide)

	}

	w.Resize(granio.NewSize(600, 400))
	w.ShowAndRun()
}
