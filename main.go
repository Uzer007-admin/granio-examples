//go:generate fyne bundle -o data.go Icon.png

// Package main launches the life app
package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
)

const (
	minXCount = 50
	minYCount = 40
)

func show(app granio.App) {
	board := newBoard(minXCount, minYCount)
	board.load()
	game := newGame(board)

	window := app.NewWindow("Life")

	window.SetContent(game.buildUI())
	window.Canvas().SetOnTypedRune(game.typedRune)

	// start the board animation before we show the window - it will block
	game.animate()

	window.Resize(granio.NewSize(200, 220))
	window.Show()
}

func main() {
	app := app.New()
	app.SetIcon(resourceIconPng)

	show(app)
	app.Run()
}
