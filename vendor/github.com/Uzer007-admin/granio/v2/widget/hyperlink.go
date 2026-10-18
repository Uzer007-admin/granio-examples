package widget

import (
	"image/color"
	"net/url"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var (
	_ granio.Accessible = (*Hyperlink)(nil)
	_ granio.Focusable  = (*Hyperlink)(nil)
	_ granio.Widget     = (*Hyperlink)(nil)
)

// Hyperlink widget is a text component with appropriate padding and layout.
// When clicked, the default web browser should open with a URL
type Hyperlink struct {
	BaseWidget
	Text      string
	URL       *url.URL
	Alignment granio.TextAlign // The alignment of the Text
	Wrapping  granio.TextWrap  // The wrapping of the Text
	TextStyle granio.TextStyle // The style of the hyperlink text

	// The truncation mode of the hyperlink
	//
	// Since: 2.5
	Truncation granio.TextTruncation

	// The theme size name for the text size of the hyperlink
	//
	// Since: 2.5
	SizeName granio.ThemeSizeName

	// OnTapped overrides the default `granio.OpenURL` call when the link is tapped
	//
	// Since: 2.2
	OnTapped func() `json:"-"`

	textSize         granio.Size // updated in syncSegments
	focused, hovered bool
	provider         RichText

	siblings []*Hyperlink // other visual instances of the same HyperlinkSegment when wrapped in RichText
}

// NewHyperlink creates a new hyperlink widget with the set text content
func NewHyperlink(text string, u *url.URL) *Hyperlink {
	return NewHyperlinkWithStyle(text, u, granio.TextAlignLeading, granio.TextStyle{})
}

// NewHyperlinkWithStyle creates a new hyperlink widget with the set text content
func NewHyperlinkWithStyle(text string, u *url.URL, alignment granio.TextAlign, style granio.TextStyle) *Hyperlink {
	hl := &Hyperlink{
		Text:      text,
		URL:       u,
		Alignment: alignment,
		TextStyle: style,
	}

	return hl
}

// AccessibilityLabel for a hyperlink is the text for the link.
//
// Since: 2.8
func (hl *Hyperlink) AccessibilityLabel() string {
	return hl.Text
}

// AccessibilityRole for a hyperlink is granio.AccessibleRoleLink.
//
// Since: 2.8
func (*Hyperlink) AccessibilityRole() granio.AccessibleRole {
	return granio.AccessibleRoleLink
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (hl *Hyperlink) CreateRenderer() granio.WidgetRenderer {
	hl.ExtendBaseWidget(hl)
	hl.provider.ExtendBaseWidget(&hl.provider)
	hl.syncSegments()

	th := hl.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()
	focus := canvas.NewRectangle(color.Transparent)
	focus.StrokeColor = th.Color(theme.ColorNameFocus, v)
	focus.StrokeWidth = 2
	focus.Hide()
	under := canvas.NewRectangle(th.Color(theme.ColorNameHyperlink, v))
	under.Hide()
	return &hyperlinkRenderer{hl: hl, objects: []granio.CanvasObject{&hl.provider, focus, under}, focus: focus, under: under}
}

// Cursor returns the cursor type of this widget
func (hl *Hyperlink) Cursor() desktop.Cursor {
	if hl.hovered {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

// FocusGained is a hook called by the focus handling logic after this object gained the focus.
func (hl *Hyperlink) FocusGained() {
	hl.focused = true
	hl.BaseWidget.Refresh()
}

// FocusLost is a hook called by the focus handling logic after this object lost the focus.
func (hl *Hyperlink) FocusLost() {
	hl.focused = false
	hl.BaseWidget.Refresh()
}

// MouseIn is a hook that is called if the mouse pointer enters the element.
func (hl *Hyperlink) MouseIn(e *desktop.MouseEvent) {
	hl.MouseMoved(e)
}

// MouseMoved is a hook that is called if the mouse pointer moved over the element.
func (hl *Hyperlink) MouseMoved(e *desktop.MouseEvent) {
	oldHovered := hl.hovered
	hl.hovered = hl.isPosOverText(e.Position)
	if hl.hovered != oldHovered {
		hl.BaseWidget.Refresh()
		for _, s := range hl.siblings {
			s.setHovered(hl.hovered)
		}
	}
}

// MouseOut is a hook that is called if the mouse pointer leaves the element.
func (hl *Hyperlink) MouseOut() {
	changed := hl.hovered
	hl.hovered = false
	if changed {
		hl.BaseWidget.Refresh()
		for _, s := range hl.siblings {
			s.setHovered(false)
		}
	}
}

// setHovered updates the hovered state without propagating to siblings, to avoid recursion.
func (hl *Hyperlink) setHovered(hovered bool) {
	if hl.hovered == hovered {
		return
	}
	hl.hovered = hovered
	hl.BaseWidget.Refresh()
}

func (hl *Hyperlink) focusWidth() float32 {
	th := hl.Theme()

	innerPad := th.Size(theme.SizeNameInnerPadding)
	return min(hl.Size().Width, hl.textSize.Width+innerPad+th.Size(theme.SizeNamePadding)*2) - innerPad
}

func (hl *Hyperlink) focusXPos() float32 {
	innerPad := hl.Theme().Size(theme.SizeNameInnerPadding)

	switch hl.Alignment {
	case granio.TextAlignLeading:
		return innerPad / 2
	case granio.TextAlignCenter:
		return (hl.Size().Width - hl.focusWidth()) / 2
	case granio.TextAlignTrailing:
		return (hl.Size().Width - hl.focusWidth()) - innerPad/2
	default:
		return 0 // unreached
	}
}

func (hl *Hyperlink) isPosOverText(pos granio.Position) bool {
	th := hl.Theme()
	innerPad := th.Size(theme.SizeNameInnerPadding)
	pad := th.Size(theme.SizeNamePadding)
	lineCount := max(1, float32(len(hl.provider.rowBounds)))

	xpos := hl.focusXPos()
	return pos.X >= xpos && pos.X <= xpos+hl.focusWidth() &&
		pos.Y >= innerPad/2 && pos.Y <= hl.textSize.Height*lineCount+pad*2+innerPad/2
}

// Refresh triggers a redraw of the hyperlink.
func (hl *Hyperlink) Refresh() {
	if len(hl.provider.Segments) == 0 {
		return // Not initialized yet.
	}

	hl.syncSegments()
	hl.provider.Refresh()
	hl.BaseWidget.Refresh()
}

// MinSize returns the smallest size this widget can shrink to
func (hl *Hyperlink) MinSize() granio.Size {
	hl.ExtendBaseWidget(hl)
	return hl.BaseWidget.MinSize()
}

// Resize sets a new size for the hyperlink.
// Note this should not be used if the widget is being managed by a Layout within a Container.
func (hl *Hyperlink) Resize(size granio.Size) {
	hl.BaseWidget.Resize(size)

	if len(hl.provider.Segments) == 0 {
		return // Not initialized yet.
	}
	hl.provider.Resize(size)
}

// SetText sets the text of the hyperlink
func (hl *Hyperlink) SetText(text string) {
	hl.Text = text

	if len(hl.provider.Segments) == 0 {
		return // Not initialized yet.
	}
	hl.syncSegments()
	hl.provider.Refresh()
}

// SetURL sets the URL of the hyperlink, taking in a URL type
func (hl *Hyperlink) SetURL(u *url.URL) {
	hl.URL = u
}

// SetURLFromString sets the URL of the hyperlink, taking in a string type
func (hl *Hyperlink) SetURLFromString(str string) error {
	u, err := url.Parse(str)
	if err != nil {
		return err
	}
	hl.SetURL(u)
	return nil
}

// Tapped is called when a pointer tapped event is captured and triggers any change handler
func (hl *Hyperlink) Tapped(e *granio.PointEvent) {
	if len(hl.provider.Segments) != 0 && !hl.isPosOverText(e.Position) {
		return // tapped outside text area
	}
	hl.invokeAction()
}

func (hl *Hyperlink) invokeAction() {
	onTapped := hl.OnTapped

	if onTapped != nil {
		onTapped()
		return
	}
	hl.openURL()
}

// TypedRune is a hook called by the input handling logic on text input events if this object is focused.
func (*Hyperlink) TypedRune(rune) {
}

// TypedKey is a hook called by the input handling logic on key events if this object is focused.
func (hl *Hyperlink) TypedKey(ev *granio.KeyEvent) {
	if ev.Name == granio.KeySpace {
		hl.invokeAction()
	}
}

func (hl *Hyperlink) openURL() {
	u := hl.URL

	if u != nil {
		err := granio.CurrentApp().OpenURL(u)
		if err != nil {
			granio.LogError("Failed to open url", err)
		}
	}
}

func (hl *Hyperlink) syncSegments() {
	th := hl.Theme()

	hl.provider.Wrapping = hl.Wrapping
	hl.provider.Truncation = hl.Truncation

	if len(hl.provider.Segments) == 0 {
		hl.provider.Scroll = widget.ScrollNone
		hl.provider.Segments = []RichTextSegment{
			&TextSegment{
				Style: RichTextStyle{
					Alignment: hl.Alignment,
					ColorName: theme.ColorNameHyperlink,
					Inline:    true,
					TextStyle: hl.TextStyle,
				},
				Text: hl.Text,
			},
		}
	} else {
		segment, _ := hl.provider.Segments[0].(*TextSegment)
		segment.Style.Alignment = hl.Alignment
		segment.Style.TextStyle = hl.TextStyle
		segment.Text = hl.Text
	}

	sizeName := hl.SizeName
	if sizeName == "" {
		sizeName = theme.SizeNameText
	}
	hl.provider.Segments[0].(*TextSegment).Style.SizeName = sizeName
	hl.textSize = granio.MeasureText(hl.Text, th.Size(sizeName), hl.TextStyle)
}

var _ granio.WidgetRenderer = (*hyperlinkRenderer)(nil)

type hyperlinkRenderer struct {
	hl    *Hyperlink
	focus *canvas.Rectangle
	under *canvas.Rectangle

	objects []granio.CanvasObject
}

func (*hyperlinkRenderer) Destroy() {
}

func (r *hyperlinkRenderer) Layout(s granio.Size) {
	th := r.hl.Theme()
	textSize := r.hl.textSize
	innerPad := th.Size(theme.SizeNameInnerPadding)
	w := r.hl.focusWidth()
	xposFocus := r.hl.focusXPos()

	xposUnderline := xposFocus + innerPad/2
	lineCount := float32(len(r.hl.provider.rowBounds))

	r.hl.provider.Resize(s)
	r.focus.Move(granio.NewPos(xposFocus, innerPad/2))
	r.focus.Resize(granio.NewSize(w, textSize.Height*lineCount+innerPad))
	r.under.Move(granio.NewPos(xposUnderline, textSize.Height*lineCount+th.Size(theme.SizeNamePadding)*2))
	r.under.Resize(granio.NewSize(w-innerPad, 1))
}

func (r *hyperlinkRenderer) MinSize() granio.Size {
	return r.hl.provider.MinSize()
}

func (r *hyperlinkRenderer) Objects() []granio.CanvasObject {
	return r.objects
}

func (r *hyperlinkRenderer) Refresh() {
	r.hl.provider.Refresh()
	th := r.hl.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	r.focus.StrokeColor = th.Color(theme.ColorNameFocus, v)
	r.focus.Hidden = !r.hl.focused
	r.focus.Refresh()
	r.under.FillColor = th.Color(theme.ColorNameHyperlink, v)
	r.under.Hidden = !r.hl.hovered
	r.under.Refresh()
}
