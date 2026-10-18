package widget

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	col "github.com/Uzer007-admin/granio/v2/internal/color"
	"github.com/Uzer007-admin/granio/v2/internal/svg"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// ButtonAlign represents the horizontal alignment of a button.
type ButtonAlign int

// ButtonIconPlacement represents the ordering of icon & text within a button.
type ButtonIconPlacement int

// ButtonImportance represents how prominent the button should appear
//
// Since: 1.4
//
// Deprecated: Use widget.Importance instead
type ButtonImportance = Importance

// ButtonStyle determines the behaviour and rendering of a button.
type ButtonStyle int

const (
	// ButtonAlignCenter aligns the icon and the text centrally.
	ButtonAlignCenter ButtonAlign = iota
	// ButtonAlignLeading aligns the icon and the text with the leading edge.
	ButtonAlignLeading
	// ButtonAlignTrailing aligns the icon and the text with the trailing edge.
	ButtonAlignTrailing
)

const (
	// ButtonIconLeadingText aligns the icon on the leading edge of the text.
	ButtonIconLeadingText ButtonIconPlacement = iota
	// ButtonIconTrailingText aligns the icon on the trailing edge of the text.
	ButtonIconTrailingText
)

var (
	_ granio.Focusable  = (*Button)(nil)
	_ granio.Accessible = (*Button)(nil)
)

// Button widget has a text label and triggers an event func when clicked
type Button struct {
	DisableableWidget
	Text string
	Icon granio.Resource
	// Specify how prominent the button should be, High will highlight the button and Low will remove some decoration.
	//
	// Since: 1.4
	Importance    Importance
	Alignment     ButtonAlign
	IconPlacement ButtonIconPlacement

	OnTapped       func() `json:"-"`
	AccessibleName string

	hovered, focused bool
	tapAnim          *granio.Animation
	isAnimating      bool
	pressed          bool
	stateAnimation   *granio.Animation
	stateBackground  color.Color
	stateTarget      color.Color
}

// NewButton creates a new button widget with the set label and tap handler
func NewButton(label string, tapped func()) *Button {
	button := &Button{
		Text:     label,
		OnTapped: tapped,
	}

	button.ExtendBaseWidget(button)
	return button
}

// NewButtonWithIcon creates a new button widget with the specified label, themed icon and tap handler
func NewButtonWithIcon(label string, icon granio.Resource, tapped func()) *Button {
	button := &Button{
		Text:     label,
		Icon:     icon,
		OnTapped: tapped,
	}

	button.ExtendBaseWidget(button)
	return button
}

// AccessibilityLabel for a button is the text, if there is some, otherwise the name of the icon.
//
// Since: 2.8
func (b *Button) AccessibilityLabel() string {
	if b.AccessibleName != "" {
		return b.AccessibleName
	}
	if b.Text != "" {
		return b.Text
	}

	if b.Icon != nil {
		return b.Icon.Name()
	}
	return ""
}

// AccessibilityRole for a button is granio.AccessibleRoleButton.
//
// Since: 2.8
func (*Button) AccessibilityRole() granio.AccessibleRole {
	return granio.AccessibleRoleButton
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (b *Button) CreateRenderer() granio.WidgetRenderer {
	b.ExtendBaseWidget(b)
	th := b.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	seg := &TextSegment{Text: b.Text, Style: RichTextStyleStrong}
	seg.Style.Alignment = granio.TextAlignCenter
	text := NewRichText(seg)
	text.inset = granio.NewSquareSize(th.Size(theme.SizeNameInnerPadding))

	background := canvas.NewRectangle(th.Color(theme.ColorNameButton, v))
	background.CornerRadius = th.Size(theme.SizeNameButtonRadius)
	tapBG := canvas.NewRectangle(color.Transparent)
	b.tapAnim = newButtonTapAnimation(tapBG, b, th)
	b.tapAnim.Curve = granio.AnimationEaseOut
	objects := []granio.CanvasObject{
		background,
		tapBG,
		text,
	}
	r := &buttonRenderer{
		BaseRenderer: widget.NewBaseRenderer(objects),
		background:   background,
		tapBG:        tapBG,
		button:       b,
		label:        text,
		layout:       layout.NewHBoxLayout(),
	}
	r.updateIconAndText()
	r.applyTheme()
	return r
}

// Cursor returns the cursor type of this widget
func (*Button) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}

// FocusGained is a hook called by the focus handling logic after this object gained the focus.
func (b *Button) FocusGained() {
	b.focused = true
	b.Refresh()
}

// FocusLost is a hook called by the focus handling logic after this object lost the focus.
func (b *Button) FocusLost() {
	b.focused = false
	b.Refresh()
}

// MinSize returns the size that this widget should not shrink below
func (b *Button) MinSize() granio.Size {
	b.ExtendBaseWidget(b)
	return b.BaseWidget.MinSize()
}

// MouseIn is called when a desktop pointer enters the widget
func (b *Button) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

// MouseMoved is called when a desktop pointer hovers over the widget
func (*Button) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut is called when a desktop pointer exits the widget
func (b *Button) MouseOut() {
	b.hovered = false
	b.pressed = false
	b.Refresh()
}

func (b *Button) MouseDown(event *desktop.MouseEvent) {
	if event.Button == desktop.MouseButtonPrimary && !b.Disabled() {
		b.pressed = true
		b.Refresh()
	}
}
func (b *Button) MouseUp(event *desktop.MouseEvent) {
	if event.Button == desktop.MouseButtonPrimary {
		b.pressed = false
		b.Refresh()
	}
}

// SetIcon updates the icon on a label - pass nil to hide an icon
func (b *Button) SetIcon(icon granio.Resource) {
	b.Icon = icon

	b.Refresh()
}

// SetText allows the button label to be changed
func (b *Button) SetText(text string) {
	b.Text = text

	b.Refresh()
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (b *Button) Tapped(*granio.PointEvent) {
	impl := b.super()
	if c := granio.CurrentApp().Driver().CanvasForObject(impl); c != nil {
		c.Focus(nil) // the focus manager won't get this Tap and we are not focusable
	}

	if b.Disabled() {
		return
	}

	b.tapAnimation()

	if onTapped := b.OnTapped; onTapped != nil {
		onTapped()
	}
}

// TypedRune is a hook called by the input handling logic on text input events if this object is focused.
func (*Button) TypedRune(rune) {
}

// TypedKey is a hook called by the input handling logic on key events if this object is focused.
func (b *Button) TypedKey(ev *granio.KeyEvent) {
	if ev.Name == granio.KeySpace {
		b.Tapped(nil)
	}
}

func (b *Button) tapAnimation() {
	if b.tapAnim == nil {
		return
	}
	b.tapAnim.Stop()
	b.isAnimating = false

	if granio.CurrentApp().Settings().ShowAnimations() {
		b.isAnimating = true
		b.tapAnim.Start()
	}
}

type buttonRenderer struct {
	widget.BaseRenderer

	icon       *canvas.Image
	label      *RichText
	background *canvas.Rectangle
	tapBG      *canvas.Rectangle
	button     *Button
	layout     granio.Layout
}

// Layout the components of the button widget
func (r *buttonRenderer) Layout(size granio.Size) {
	r.background.Resize(size)
	if !r.button.isAnimating {
		// if we are animating let the animation control the tapBG size
		r.tapBG.Resize(size)
	}

	th := r.button.Theme()
	padding := r.padding(th)
	hasIcon := r.icon != nil
	hasLabel := r.label.Segments[0].(*TextSegment).Text != ""
	if !hasIcon && !hasLabel {
		// Nothing to layout
		return
	}
	iconSize := granio.NewSquareSize(th.Size(theme.SizeNameInlineIcon))
	labelSize := r.label.MinSize()

	if hasLabel {
		if hasIcon {
			// Both
			var objects []granio.CanvasObject
			if r.button.IconPlacement == ButtonIconLeadingText {
				objects = append(objects, r.icon, r.label)
			} else {
				objects = append(objects, r.label, r.icon)
			}
			r.icon.SetMinSize(iconSize)
			minSize := r.layout.MinSize(objects)
			r.layout.Layout(objects, minSize)
			pos := alignedPosition(r.button.Alignment, padding, minSize, size)
			labelOff := (minSize.Height - labelSize.Height) / 2
			r.label.Move(r.label.Position().Add(pos).AddXY(0, labelOff))
			r.icon.Move(r.icon.Position().Add(pos))
		} else {
			// Label Only
			r.label.Move(alignedPosition(r.button.Alignment, padding, labelSize, size))
			r.label.Resize(labelSize)
		}
	} else {
		// Icon Only
		r.icon.Move(alignedPosition(r.button.Alignment, padding, iconSize, size))
		r.icon.Resize(iconSize)
	}
}

// MinSize calculates the minimum size of a button.
// This is based on the contained text, any icon that is set and a standard
// amount of padding added.
func (r *buttonRenderer) MinSize() (size granio.Size) {
	th := r.button.Theme()
	hasIcon := r.icon != nil
	hasLabel := r.label.Segments[0].(*TextSegment).Text != ""
	iconSize := granio.NewSquareSize(th.Size(theme.SizeNameInlineIcon))
	labelSize := r.label.MinSize()
	if hasLabel {
		size.Width = labelSize.Width
	}
	if hasIcon {
		if hasLabel {
			size.Width += th.Size(theme.SizeNamePadding)
		}
		size.Width += iconSize.Width
	}
	size.Height = max(labelSize.Height, iconSize.Height)
	size = size.Add(r.padding(th))
	size = minimumComponentHeight(size, th, granio.ComponentButton)
	return size
}

func (r *buttonRenderer) Refresh() {
	th := r.button.Theme()
	r.label.inset = granio.NewSquareSize(th.Size(theme.SizeNameInnerPadding))

	r.label.Segments[0].(*TextSegment).Text = r.button.Text
	r.updateIconAndText()
	r.applyTheme()

	r.background.Refresh()
	r.Layout(r.button.Size())
	canvas.Refresh(r.button.super())
}

// applyTheme updates this button to match the current theme
// must be called with the button propertyLock RLocked
func (r *buttonRenderer) applyTheme() {
	th := r.button.Theme()
	fgColorName, bgColorName, bgBlendName := r.buttonColorNames()
	_, systemColors := activeSystemColors(th)
	if bg := r.background; bg != nil {
		v := granio.CurrentApp().Settings().ThemeVariant()
		bgColor := color.Color(color.Transparent)
		if bgColorName != "" {
			bgColor = th.Color(bgColorName, v)
		}
		if bgBlendName != "" {
			bgColor = blendColor(bgColor, th.Color(bgBlendName, v))
		}
		r.updateStateBackground(bgColor, th)
		if colors, ok := activeSystemColors(th); ok {
			bg.StrokeColor, bg.StrokeWidth = colors.WindowText, 1
			if r.button.focused && !r.button.Disabled() {
				bg.StrokeColor, bg.StrokeWidth = colors.Highlight, 2
			}
		} else {
			bg.StrokeColor, bg.StrokeWidth = nil, 0
		}
		state := granio.VisualState(0)
		if r.button.focused {
			state |= granio.VisualFocused
		}
		if r.button.hovered {
			state |= granio.VisualHovered
		}
		if r.button.pressed {
			state |= granio.VisualPressed
		}
		if r.button.Disabled() {
			state |= granio.VisualDisabled
		}
		if style, modern := componentStyle(th, granio.ComponentButton, state); modern {
			bg.StrokeColor, bg.StrokeWidth = style.Border, style.BorderWidth
			if style.FocusWidth > 0 {
				bg.StrokeColor, bg.StrokeWidth = style.Focus, style.FocusWidth
			}
		}
		bg.CornerRadius = th.Size(theme.SizeNameButtonRadius)
		bg.Refresh()
	}

	r.label.Segments[0].(*TextSegment).Style.ColorName = fgColorName
	if systemColors && fgColorName == theme.ColorNameForeground && bgColorName == theme.ColorNameButton {
		r.label.Segments[0].(*TextSegment).Style.ColorName = theme.ColorNameForegroundOnButton
		fgColorName = theme.ColorNameForegroundOnButton
	}
	r.label.Refresh()
	if r.icon != nil && r.icon.Resource != nil {
		icon := r.icon.Resource
		if r.button.Importance != MediumImportance && r.button.Importance != LowImportance {
			if thRes, ok := icon.(granio.ThemedResource); ok {
				if thRes.ThemeColorName() != fgColorName {
					icon = theme.NewColoredResource(icon, fgColorName)
				}
			}
		}
		r.icon.Resource = icon
		r.icon.Refresh()
	}
}

func (r *buttonRenderer) buttonColorNames() (foreground, background, backgroundBlend granio.ThemeColorName) {
	foreground = theme.ColorNameForeground
	b := r.button
	if b.Disabled() {
		foreground = theme.ColorNameDisabled
		if b.Importance != LowImportance {
			background = theme.ColorNameDisabledButton
		}
	} else if b.focused {
		if _, modern := b.Theme().(granio.ComponentTheme); !modern {
			backgroundBlend = theme.ColorNameFocus
		}
	} else if b.hovered {
		backgroundBlend = theme.ColorNameHover
	}
	if b.pressed && !b.Disabled() {
		backgroundBlend = theme.ColorNamePressed
	}
	if background == "" {
		switch b.Importance {
		case DangerImportance:
			foreground = theme.ColorNameForegroundOnError
			background = theme.ColorNameError
		case HighImportance:
			foreground = theme.ColorNameForegroundOnPrimary
			background = theme.ColorNamePrimary
		case LowImportance:
			if backgroundBlend != "" {
				background = theme.ColorNameButton
			}
		case SuccessImportance:
			foreground = theme.ColorNameForegroundOnSuccess
			background = theme.ColorNameSuccess
		case WarningImportance:
			foreground = theme.ColorNameForegroundOnWarning
			background = theme.ColorNameWarning
		default:
			background = theme.ColorNameButton
		}
	}
	return foreground, background, backgroundBlend
}

func (*buttonRenderer) padding(th granio.Theme) granio.Size {
	return granio.NewSquareSize(th.Size(theme.SizeNameInnerPadding) * 2)
}

// must be called with r.button.propertyLock RLocked
func (r *buttonRenderer) updateIconAndText() {
	if r.button.Icon != nil && !r.button.Hidden {
		icon := r.button.Icon
		if r.icon == nil {
			r.icon = canvas.NewImageFromResource(icon)
			r.icon.FillMode = canvas.ImageFillContain
			r.SetObjects([]granio.CanvasObject{r.background, r.tapBG, r.label, r.icon})
		}
		// TODO support disabling bitmap resource not just SVG
		if r.button.Disabled() && svg.IsResourceSVG(icon) {
			icon = theme.NewDisabledResource(icon)
		}
		r.icon.Resource = icon
		r.icon.Refresh()
		r.icon.Show()
	} else if r.icon != nil {
		r.icon.Hide()
	}
	if r.button.Text == "" {
		r.label.Hide()
	} else {
		r.label.Show()
	}
	r.label.Refresh()
}

func alignedPosition(align ButtonAlign, padding, objectSize, layoutSize granio.Size) (pos granio.Position) {
	pos.Y = (layoutSize.Height - objectSize.Height) / 2
	switch align {
	case ButtonAlignCenter:
		pos.X = (layoutSize.Width - objectSize.Width) / 2
	case ButtonAlignLeading:
		pos.X = padding.Width / 2
	case ButtonAlignTrailing:
		pos.X = layoutSize.Width - objectSize.Width - padding.Width/2
	}
	return pos
}

func blendColor(under, over color.Color) color.Color {
	// This alpha blends with the over operator, and accounts for RGBA() returning alpha-premultiplied values
	dstR, dstG, dstB, dstA := under.RGBA()
	srcR, srcG, srcB, srcA := over.RGBA()
	blend := func(src, dst, alpha uint32) uint16 {
		return uint16((src + dst - (dst * alpha / 0xFFFF)) & 0xFFFF)
	}
	// We create an RGBA64 here because the color components are already alpha-premultiplied 16-bit values (they're just stored in uint32s).
	return color.RGBA64{
		R: blend(srcR, dstR, srcA),
		G: blend(srcG, dstG, srcA),
		B: blend(srcB, dstB, srcA),
		A: blend(srcA, dstA, srcA),
	}
}

func newButtonTapAnimation(bg *canvas.Rectangle, w granio.Widget, th granio.Theme) *granio.Animation {
	v := granio.CurrentApp().Settings().ThemeVariant()
	duration := canvas.DurationStandard
	if _, modern := th.(granio.MotionTheme); modern {
		duration = theme.Motion(th, granio.MotionPressed).Duration
	}
	animation := granio.NewAnimation(duration, func(done float32) {
		mid := w.Size().Width / 2
		size := mid * done
		bg.Resize(granio.NewSize(size*2, w.Size().Height))
		bg.Move(granio.NewPos(mid-size, 0))

		r, g, b, a := col.ToNRGBA(th.Color(theme.ColorNamePressed, v))
		fade := a - uint8(float32(a)*done)
		if fade > 0 {
			bg.FillColor = &color.NRGBA{R: r, G: g, B: b, A: fade}
		} else {
			bg.FillColor = color.Transparent
		}
		canvas.Refresh(bg)
		if done == 1.0 {
			if btn, ok := w.(*Button); ok {
				btn.isAnimating = false
			}
		}
	})
	if _, modern := th.(granio.MotionTheme); modern {
		animation.Object = w
	}
	return animation
}
