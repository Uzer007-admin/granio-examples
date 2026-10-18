//go:generate granio bundle --package=main -o data.go Icon.png

// Package main launches the solitaire app
package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// show creates a new game and loads a table rendered in a new window.
func show(app granio.App) {
	game := NewGame()
	table := NewTable(game)

	w := app.NewWindow("Solitaire")
	shuffle := widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
		table.game.Hand.Shuffle()
	})
	table.shuffle = shuffle
	bar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			checkRestart(table, w)
		}),
		shuffle,
	)
	w.SetContent(container.NewBorder(bar, nil, nil, nil, table))
	w.Resize(granio.NewSize(minWidth, minHeight))

	game.OnWin = func() {
		table.finishAnimation()
		d := dialog.NewInformation("You Win!", "Congratulations", w)
		d.SetOnClosed(table.Restart)
		d.Show()
	}
	w.Show()
}

func checkRestart(t *Table, w granio.Window) {
	dialog.ShowConfirm("New Game", "Start a new game?", func(ok bool) {
		if !ok {
			return
		}

		t.Restart()
	}, w)
}

func main() {
	a := app.New()
	a.SetIcon(resourceIconPng)
	a.Settings().SetTheme(newGameTheme())

	show(a)
	a.Run()
}
