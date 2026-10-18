package canvas

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// Declare conformity with CanvasObject interface
var _ granio.CanvasObject = (*Text)(nil)

// Text describes a text primitive in a Fyne canvas.
// A text object can have a style set which will apply to the whole string.
// No formatting or text parsing will be performed
type Text struct {
	baseObject
	Alignment granio.TextAlign // The alignment of the text content

	Color     color.Color      // The main text draw color
	Text      string           // The string content of this Text
	TextSize  float32          // Size of the text - if the Canvas scale is 1.0 this will be equivalent to point size
	TextStyle granio.TextStyle // The style of the text content

	// FontSource defines a resource that can be used instead of the theme for looking up the font.
	// When a font source is set the `TextStyle` may not be effective, as it will be limited to the styles
	// present in the data provided.
	//
	// Since: 2.5
	FontSource granio.Resource
}

// Hide will set this text to not be visible
func (t *Text) Hide() {
	t.baseObject.Hide()

	repaint(t)
}

// MinSize returns the minimum size of this text object based on its font size and content.
// This is normally determined by the render implementation.
func (t *Text) MinSize() granio.Size {
	s, _ := granio.CurrentApp().Driver().RenderedTextSize(t.Text, t.TextSize, t.TextStyle, t.FontSource)
	return s
}

// Move the text to a new position, relative to its parent / canvas
func (t *Text) Move(pos granio.Position) {
	if t.Position() == pos {
		return
	}

	t.baseObject.Move(pos)

	repaint(t)
}

// Resize on a text updates the new size of this object, which may not result in a visual change, depending on alignment.
func (t *Text) Resize(s granio.Size) {
	if s == t.Size() {
		return
	}

	t.baseObject.Resize(s)
	Refresh(t)
}

// SetMinSize has no effect as the smallest size this canvas object can be is based on its font size and content.
func (*Text) SetMinSize(granio.Size) {
	// no-op
}

// Refresh causes this text to be redrawn with its configured state.
func (t *Text) Refresh() {
	Refresh(t)
}

// NewText returns a new Text implementation
func NewText(text string, c color.Color) *Text {
	return &Text{
		Color:    c,
		Text:     text,
		TextSize: theme.TextSize(),
	}
}
