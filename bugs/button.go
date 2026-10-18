package bugs

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type bugRenderer struct {
	icon  *canvas.Image
	label *canvas.Text

	objects []granio.CanvasObject
	button  *bugButton
}

const bugSize = 18

// MinSize calculates the minimum size of a bug button. A fixed amount.
func (b *bugRenderer) MinSize() granio.Size {
	return granio.NewSize(bugSize+theme.Padding()*2, bugSize+theme.Padding()*2)
}

// Layout the components of the widget
func (b *bugRenderer) Layout(size granio.Size) {
	inner := size.Subtract(granio.NewSize(theme.Padding()*2, theme.Padding()*2))
	b.icon.Resize(inner)
	b.icon.Move(granio.NewPos(theme.Padding(), theme.Padding()))

	textSize := size.Height * .67
	textMin := granio.MeasureText(b.label.Text, textSize, granio.TextStyle{Bold: true})

	b.label.TextSize = textSize
	b.label.Resize(granio.NewSize(size.Width, textMin.Height))
	b.label.Move(granio.NewPos(0, (size.Height-textMin.Height)/2))
}

// ApplyTheme is called when the bugButton may need to update it's look
func (b *bugRenderer) ApplyTheme() {
	b.label.Color = theme.Color(theme.ColorNameForeground)
	b.Refresh()
}

func (b *bugRenderer) BackgroundColor() color.Color {
	return theme.Color(theme.ColorNameButton)
}

func (b *bugRenderer) Refresh() {
	b.label.Text = b.button.text

	b.icon.Hidden = b.button.icon == nil
	if b.button.icon != nil {
		b.icon.Resource = b.button.icon
		b.icon.Refresh()
	}

	b.Layout(b.button.Size())
	canvas.Refresh(b.button)
}

func (b *bugRenderer) Objects() []granio.CanvasObject {
	return b.objects
}

func (b *bugRenderer) Destroy() {
}

// bugButton widget is a scalable button that has a text label and icon and triggers an event func when clicked
type bugButton struct {
	widget.BaseWidget
	text string
	icon granio.Resource

	tap func(bool)
}

// Tapped is called when a regular tap is reported
func (b *bugButton) Tapped(ev *granio.PointEvent) {
	b.tap(true)
}

// TappedSecondary is called when an alternative tap is reported
func (b *bugButton) TappedSecondary(ev *granio.PointEvent) {
	b.tap(false)
}

func (b *bugButton) CreateRenderer() granio.WidgetRenderer {
	text := canvas.NewText(b.text, theme.Color(theme.ColorNameForeground))
	text.Alignment = granio.TextAlignCenter
	text.TextStyle.Bold = true

	icon := canvas.NewImageFromResource(b.icon)
	icon.FillMode = canvas.ImageFillContain

	objects := []granio.CanvasObject{
		text,
		icon,
	}

	return &bugRenderer{icon, text, objects, b}
}

// SetText allows the button label to be changed
func (b *bugButton) SetText(text string) {
	b.text = text

	b.Refresh()
}

// SetIcon updates the icon on a label - pass nil to hide an icon
func (b *bugButton) SetIcon(icon granio.Resource) {
	b.icon = icon

	b.Refresh()
}

// newButton creates a new button widget with the specified label, themed icon and tap handler
func newButton(label string, icon granio.Resource, tap func(bool)) *bugButton {
	button := &bugButton{text: label, icon: icon, tap: tap}
	button.ExtendBaseWidget(button)
	return button
}
