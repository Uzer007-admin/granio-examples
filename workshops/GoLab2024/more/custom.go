package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.NewWithID("io.fyne.workshop.custom")
	w := a.NewWindow("IconSwapper")

	w.SetContent(newImageSwapper(theme.HomeIcon(), theme.ComputerIcon()))

	w.ShowAndRun()
}

type imageSwapper struct {
	widget.BaseWidget

	Reversed bool

	img1, img2 granio.Resource
}

func newImageSwapper(img1, img2 granio.Resource) granio.Widget {
	i := &imageSwapper{img1: img1, img2: img2}
	i.ExtendBaseWidget(i)

	return i
}

func (i *imageSwapper) CreateRenderer() granio.WidgetRenderer {
	return &imageSwapRenderer{img: i}
}

func (i *imageSwapper) Swap() {
	i.Reversed = !i.Reversed
	i.Refresh()
}

func (i *imageSwapper) Tapped(_ *granio.PointEvent) {
	i.Swap()
}

type imageSwapRenderer struct {
	img *imageSwapper

	left, right *canvas.Image
	content     *granio.Container
}

func (r *imageSwapRenderer) Destroy() {}

func (r *imageSwapRenderer) Layout(s granio.Size) {
	if r.content == nil {
		return
	}

	r.content.Resize(s)
}

func (r *imageSwapRenderer) MinSize() granio.Size {
	iconMin := theme.SizeForWidget(theme.SizeNameInlineIcon, r.img)
	pad := theme.SizeForWidget(theme.SizeNamePadding, r.img)

	return granio.NewSize(iconMin*2+pad, iconMin)
}

func (r *imageSwapRenderer) Objects() []granio.CanvasObject {
	if r.content == nil {
		r.left = canvas.NewImageFromResource(r.img.img1)
		r.right = canvas.NewImageFromResource(r.img.img2)

		r.content = container.NewGridWithColumns(2, r.left, r.right)
	}

	return []granio.CanvasObject{r.content}
}

func (r *imageSwapRenderer) Refresh() {
	if r.img.Reversed {
		r.left.Resource = r.img.img2
		r.right.Resource = r.img.img1
	} else {
		r.left.Resource = r.img.img1
		r.right.Resource = r.img.img2
	}

	r.left.Refresh()
	r.right.Refresh()
}
