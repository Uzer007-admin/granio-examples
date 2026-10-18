package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// Card widget groups title, subtitle with content and a header image
//
// Since: 1.4
type Card struct {
	BaseWidget
	Title, Subtitle string
	Image           *canvas.Image
	Content         granio.CanvasObject
}

// NewCard creates a new card widget with the specified title, subtitle and content (all optional).
//
// Since: 1.4
func NewCard(title, subtitle string, content granio.CanvasObject) *Card {
	card := &Card{
		Title:    title,
		Subtitle: subtitle,
		Content:  content,
	}

	card.ExtendBaseWidget(card)
	return card
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (c *Card) CreateRenderer() granio.WidgetRenderer {
	c.ExtendBaseWidget(c)
	th := c.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	header := canvas.NewText(c.Title, th.Color(theme.ColorNameForeground, v))
	header.TextStyle.Bold = true
	subHeader := canvas.NewText(c.Subtitle, header.Color)

	background := canvas.NewRectangle(th.Color(theme.ColorNameBackground, v))
	background.CornerRadius = th.Size(theme.SizeNameCardRadius)
	widget.ApplyShadowForLevel(&background.Shadow, widget.CardLevel, th.Color(theme.ColorNameShadow, v))
	objects := []granio.CanvasObject{background, header, subHeader}
	if c.Image != nil {
		objects = append(objects, c.Image)
	}
	if c.Content != nil {
		objects = append(objects, c.Content)
	}
	r := &cardRenderer{
		widget.NewBaseRenderer(objects),
		header, subHeader, c, background,
	}
	r.applyTheme()
	return r
}

// MinSize returns the size that this widget should not shrink below
func (c *Card) MinSize() granio.Size {
	c.ExtendBaseWidget(c)
	return c.BaseWidget.MinSize()
}

// SetContent changes the body of this card to have the specified content.
func (c *Card) SetContent(obj granio.CanvasObject) {
	c.Content = obj

	c.Refresh()
}

// SetImage changes the image displayed above the title for this card.
func (c *Card) SetImage(img *canvas.Image) {
	c.Image = img

	c.Refresh()
}

// SetSubTitle updates the secondary title for this card.
func (c *Card) SetSubTitle(text string) {
	c.Subtitle = text

	c.Refresh()
}

// SetTitle updates the main title for this card.
func (c *Card) SetTitle(text string) {
	c.Title = text

	c.Refresh()
}

type cardRenderer struct {
	widget.BaseRenderer

	header, subHeader *canvas.Text

	card *Card

	background *canvas.Rectangle
}

const (
	cardMediaHeight = 128
)

// Layout the components of the card container.
func (c *cardRenderer) Layout(size granio.Size) {
	padding := c.card.Theme().Size(theme.SizeNamePadding)
	pos := granio.NewSquareOffsetPos(padding / 2)
	size = size.Subtract(granio.NewSquareSize(padding))
	c.background.Resize(size)
	c.background.Move(pos)

	if c.card.Image != nil {
		c.card.Image.Move(pos)
		c.card.Image.Resize(granio.NewSize(size.Width, cardMediaHeight))
		pos.Y += cardMediaHeight
	}

	if c.card.Title != "" || c.card.Subtitle != "" {
		titlePad := padding * 2
		size.Width -= titlePad * 2
		pos.X += titlePad
		pos.Y += titlePad

		if c.card.Title != "" {
			height := c.header.MinSize().Height
			c.header.Move(pos)
			c.header.Resize(granio.NewSize(size.Width, height))
			pos.Y += height + padding
		}

		if c.card.Subtitle != "" {
			height := c.subHeader.MinSize().Height
			c.subHeader.Move(pos)
			c.subHeader.Resize(granio.NewSize(size.Width, height))
			pos.Y += height + padding
		}

		size.Width = size.Width + titlePad*2
		pos.X = pos.X - titlePad
		pos.Y += titlePad
	}

	size.Width -= padding * 2
	pos.X += padding
	if c.card.Content != nil {
		height := size.Height - padding*2 - (pos.Y - padding/2) // adjust for content and initial offset
		if c.card.Title != "" || c.card.Subtitle != "" {
			height += padding
			pos.Y -= padding
		}
		c.card.Content.Move(pos.Add(granio.NewPos(0, padding)))
		c.card.Content.Resize(granio.NewSize(size.Width, height))
	}
}

// MinSize calculates the minimum size of a card.
// This is based on the contained text, image and content.
func (c *cardRenderer) MinSize() granio.Size {
	hasHeader := c.card.Title != ""
	hasSubHeader := c.card.Subtitle != ""
	hasImage := c.card.Image != nil
	hasContent := c.card.Content != nil

	padding := c.card.Theme().Size(theme.SizeNamePadding)
	if !hasHeader && !hasSubHeader && !hasContent { // just image, or nothing
		if c.card.Image == nil {
			return granio.NewSize(padding, padding) // empty, just space for border
		}
		return granio.NewSize(c.card.Image.MinSize().Width+padding, cardMediaHeight+padding)
	}

	minSize := granio.NewSize(padding, padding)
	if hasImage {
		minSize = granio.NewSize(minSize.Width, minSize.Height+cardMediaHeight)
	}

	if hasHeader || hasSubHeader {
		titlePad := padding * 2
		minSize = minSize.Add(granio.NewSize(0, titlePad*2))
		if hasHeader {
			headerMin := c.header.MinSize()
			minSize = granio.NewSize(max(minSize.Width, headerMin.Width+titlePad*2+padding),
				minSize.Height+headerMin.Height)
			if hasSubHeader {
				minSize.Height += padding
			}
		}
		if hasSubHeader {
			subHeaderMin := c.subHeader.MinSize()
			minSize = granio.NewSize(max(minSize.Width, subHeaderMin.Width+titlePad*2+padding),
				minSize.Height+subHeaderMin.Height)
		}
	}

	if hasContent {
		contentMin := c.card.Content.MinSize()
		minSize = granio.NewSize(max(minSize.Width, contentMin.Width+padding*3),
			minSize.Height+contentMin.Height+padding*2)
	}

	return minSize
}

func (c *cardRenderer) Refresh() {
	c.header.Text = c.card.Title
	c.header.Refresh()
	c.subHeader.Text = c.card.Subtitle
	c.subHeader.Refresh()

	objects := []granio.CanvasObject{c.background, c.header, c.subHeader}
	if c.card.Image != nil {
		objects = append(objects, c.card.Image)
	}
	if c.card.Content != nil {
		objects = append(objects, c.card.Content)
	}
	c.SetObjects(objects)

	c.applyTheme()
	c.Layout(c.card.Size())
	c.background.Refresh()
	canvas.Refresh(c.card.super())
}

// applyTheme updates this button to match the current theme
func (c *cardRenderer) applyTheme() {
	th := c.card.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	if c.header != nil {
		c.header.TextSize = th.Size(theme.SizeNameHeadingText)
		c.header.Color = th.Color(theme.ColorNameForeground, v)
	}
	if c.subHeader != nil {
		c.subHeader.TextSize = th.Size(theme.SizeNameText)
		c.subHeader.Color = th.Color(theme.ColorNameForeground, v)
	}
	if c.card.Content != nil {
		c.card.Content.Refresh()
	}
	c.background.FillColor = th.Color(theme.ColorNameBackground, v)
	c.background.Shadow.Color = th.Color(theme.ColorNameShadow, v)
	c.background.CornerRadius = th.Size(theme.SizeNameCardRadius)
}
