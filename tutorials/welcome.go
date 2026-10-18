package tutorials

import (
	"image/color"
	"net/url"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/Uzer007-admin/granio-demo/data"
)

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		granio.LogError("Could not parse URL", err)
	}

	return link
}

func welcomeScreen(_ granio.Window) granio.CanvasObject {
	logo := canvas.NewImageFromResource(data.FyneLogoTransparent)
	logo.FillMode = canvas.ImageFillContain
	if granio.CurrentDevice().IsMobile() {
		logo.SetMinSize(granio.NewSize(192, 192))
	} else {
		logo.SetMinSize(granio.NewSize(256, 256))
	}

	footer := container.NewHBox(
		layout.NewSpacer(),
		widget.NewHyperlink("fyne.io", parseURL("https://fyne.io/")),
		widget.NewLabel("-"),
		widget.NewHyperlink("documentation", parseURL("https://docs.fyne.io/")),
		widget.NewLabel("-"),
		widget.NewHyperlink("sponsor", parseURL("https://fyne.io/sponsor/")),
		layout.NewSpacer(),
	)

	authors := widget.NewRichTextFromMarkdown(formatAuthors(string(data.Authors.Content())))
	content := container.NewVBox(
		widget.NewLabelWithStyle("\n\nWelcome to the Fyne toolkit demo app", granio.TextAlignCenter, granio.TextStyle{Bold: true}),
		logo,
		container.NewCenter(authors),
		widget.NewLabelWithStyle("\nWith great thanks to our many kind sponsors\n", granio.TextAlignCenter, granio.TextStyle{Italic: true}))
	scroll := container.NewScroll(content)

	bgColor := withAlpha(theme.Color(theme.ColorNameBackground), 0xe0)
	shadowColor := withAlpha(theme.Color(theme.ColorNameBackground), 0x33)

	underlay := canvas.NewImageFromResource(data.FyneLogo)
	bg := canvas.NewRectangle(bgColor)
	underlayer := underLayout{}
	slideBG := container.New(underlayer, underlay)
	footerBG := canvas.NewRectangle(shadowColor)

	granio.CurrentApp().Settings().AddListener(func(granio.Settings) {
		bgColor = withAlpha(theme.Color(theme.ColorNameBackground), 0xe0)
		bg.FillColor = bgColor
		bg.Refresh()

		shadowColor = withAlpha(theme.Color(theme.ColorNameBackground), 0x33)
		footerBG.FillColor = bgColor
		footer.Refresh()
	})

	underlay.Resize(granio.NewSize(1024, 1024))
	scroll.OnScrolled = func(p granio.Position) {
		underlayer.offset = -p.Y / 3
		underlayer.Layout(slideBG.Objects, slideBG.Size())
	}

	bgClip := container.NewScroll(slideBG)
	bgClip.Direction = container.ScrollNone
	return container.NewStack(container.New(unpad{top: true}, bgClip, bg),
		container.NewBorder(nil,
			container.NewStack(footerBG, footer), nil, nil,
			container.New(unpad{top: true, bottom: true}, scroll)))
}

func withAlpha(c color.Color, alpha uint8) color.Color {
	r, g, b, _ := c.RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: alpha}
}

type underLayout struct {
	offset float32
}

func (u underLayout) Layout(objs []granio.CanvasObject, size granio.Size) {
	under := objs[0]
	left := size.Width/2 - under.Size().Width/2
	under.Move(granio.NewPos(left, u.offset-50))
}

func (u underLayout) MinSize(_ []granio.CanvasObject) granio.Size {
	return granio.Size{}
}

type unpad struct {
	top, bottom bool
}

func (u unpad) Layout(objs []granio.CanvasObject, s granio.Size) {
	pad := theme.Padding()
	var pos granio.Position
	if u.top {
		pos = granio.NewPos(0, -pad)
	}
	size := s
	if u.top {
		size = size.AddWidthHeight(0, pad)
	}
	if u.bottom {
		size = size.AddWidthHeight(0, pad)
	}
	for _, o := range objs {
		o.Move(pos)
		o.Resize(size)
	}
}

func (u unpad) MinSize(_ []granio.CanvasObject) granio.Size {
	return granio.NewSize(100, 100)
}

func formatAuthors(lines string) string {
	markdown := &strings.Builder{}
	markdown.WriteString("### Authors\n\n")

	for _, line := range strings.Split(lines, "\n") {
		if len(line) == 0 {
			continue
		}

		markdown.WriteString("* ")
		markdown.WriteString(line)
		markdown.WriteByte('\n')
	}

	return markdown.String()
}
