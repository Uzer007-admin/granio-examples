package main

import (
	"os"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/storage"

	"github.com/Uzer007-admin/granio-pixeledit/internal/api"
	"github.com/Uzer007-admin/granio-pixeledit/internal/ui"
)

func loadFileArgs(e api.Editor) {
	if len(os.Args) < 2 {
		return
	}

	time.Sleep(time.Second / 3) // wait for us to be shown before loading so scales are correct
	read, err := storage.Reader(storage.NewFileURI(os.Args[1]))
	if err != nil {
		granio.LogError("Unable to open file \""+os.Args[1]+"\"", err)
		return
	}
	e.LoadFile(read)
}

func main() {
	e := ui.NewEditor()

	a := app.NewWithID("io.fyne.pixeledit")
	w := a.NewWindow("Pixel Editor")
	e.BuildUI(w)
	w.Resize(granio.NewSize(520, 320))

	go loadFileArgs(e)
	w.ShowAndRun()
}
