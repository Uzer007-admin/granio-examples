package bugs

import (
	"fmt"
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

var bug, code, flag *theme.ThemedResource

func init() {
	bug = theme.NewThemedResource(bugIcon)
	code = theme.NewThemedResource(codeIcon)
	flag = theme.NewThemedResource(flagIcon)
}

type gameRenderer struct {
	grid   *granio.Container
	header granio.CanvasObject

	game *game
}

func (g *gameRenderer) MinSize() granio.Size {
	return g.grid.MinSize().Add(granio.NewSize(0, g.header.MinSize().Height))
}

func (g *gameRenderer) Layout(size granio.Size) {
	headerHeight := g.header.MinSize().Height
	g.header.Resize(granio.NewSize(size.Width, headerHeight))
	g.grid.Move(granio.NewPos(0, headerHeight)) // TODO why ignored?
	gridSize := size.Subtract(granio.NewSize(0, headerHeight))
	g.grid.Layout.Layout(g.grid.Objects, gridSize)
}

func (g *gameRenderer) ApplyTheme() {
}

func (g *gameRenderer) BackgroundColor() color.Color {
	return theme.Color(theme.ColorNameBackground)
}

func (g *gameRenderer) Refresh() {
	canvas.Refresh(g.grid)
}

func (g *gameRenderer) Objects() []granio.CanvasObject {
	return []granio.CanvasObject{g.grid, g.header}
}

func (g *gameRenderer) Destroy() {
}

type game struct {
	widget.BaseWidget
	board  *board
	remain *widget.Label

	grid   *granio.Container
	window granio.Window
}

func (g *game) refreshSquare(x, y int) {
	if x < 0 || y < 0 || x >= g.board.width || y >= g.board.height {
		return
	}

	sq := g.board.bugs[y][x]
	i := y*g.board.width + x
	button := g.grid.Objects[i].(*bugButton)

	if sq.flagged {
		if button.icon == flag {
			return
		}
		button.icon = flag
		button.text = ""
	} else if !sq.shown {
		if button.icon == code {
			return
		}
		button.icon = code
		button.text = ""
	} else if sq.bug {
		if button.icon == bug {
			return
		}
		button.icon = bug
		button.text = ""
	} else if button.icon == nil {
		return
	} else {
		button.icon = nil
		button.text = squareString(sq)
	}

	button.Refresh()
}

func (g *game) refreshAround(xp, yp, d int) {
	x, y := xp-d, yp-d
	for ; x < xp+d; x++ {
		g.refreshSquare(x, y)
	}
	for ; y < yp+d; y++ {
		g.refreshSquare(x, y)
	}
	for ; x > xp-d; x-- {
		g.refreshSquare(x, y)
	}
	for ; y > yp-d; y-- {
		g.refreshSquare(x, y)
	}
}

func (g *game) refreshFrom(x, y int) {
	g.refreshSquare(x, y)

	for i := 1; i < int(granio.Max(float32(g.board.width), float32(g.board.height))); i++ {
		g.refreshAround(x, y, i)
	}
}

func (g *game) CreateRenderer() granio.WidgetRenderer {
	renderer := &gameRenderer{game: g}
	title := widget.NewLabel("Hunt bugs!")
	g.remain = widget.NewLabel("")
	g.updateRemain()
	renderer.header = container.NewBorder(nil, nil, title, g.remain)

	var buttons []granio.CanvasObject
	for y := 0; y < g.board.height; y++ {
		for x := 0; x < g.board.width; x++ {
			xx, yy := x, y

			buttons = append(buttons, newButton("", code, func(reveal bool) {
				if reveal {
					g.squareReveal(xx, yy)
				} else {
					g.squareFlagged(xx, yy)
				}

				g.updateRemain()
			}))
		}
	}

	renderer.grid = container.NewGridWithColumns(g.board.width, buttons...)
	g.grid = renderer.grid
	return renderer
}

func (g *game) squareReveal(x, y int) {
	if g.board.flagged(x, y) {
		return
	}

	g.board.reveal(x, y)
	g.refreshFrom(x, y)
}

func (g *game) squareFlagged(x, y int) {
	g.board.flag(x, y)
	g.refreshSquare(x, y)
}

func (g *game) loseCallback(yes bool) {
	if !yes {
		return
	}

	g.board.load(40)
	g.updateRemain()
	g.refreshFrom(g.board.width/2, g.board.height/2)
}

func (g *game) win() {
	dialog.ShowInformation("You won!", "Congratulations, you found all the bugs", g.window)
}

func (g *game) lose() {
	dialog.ShowConfirm("You lost!", "You hit a bug and lost the game, try again?", g.loseCallback, g.window)
}

func (g *game) updateRemain() {
	g.remain.SetText(fmt.Sprintf("remaining: %d", g.board.remaining()))
}

func newGame(f *board) *game {
	g := &game{board: f}
	g.ExtendBaseWidget(g)

	return g
}

// Show starts a new bugs game
func Show(win granio.Window) granio.CanvasObject {
	b := newBoard(20, 14)
	game := newGame(b)

	b.win = game.win
	b.lose = game.lose
	b.load(40)

	game.window = win
	return game
}
