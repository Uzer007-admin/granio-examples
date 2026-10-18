package widget

import (
	"image/color"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	fynecolor "github.com/Uzer007-admin/granio/v2/internal/color"
	"github.com/Uzer007-admin/granio/v2/internal/svg"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var (
	_ granio.Widget     = (*menuItem)(nil)
	_ desktop.Hoverable = (*menuItem)(nil)
	_ granio.Tappable   = (*menuItem)(nil)
)

// menuItem is a widget for displaying a fyne.menuItem.
type menuItem struct {
	widget.Base
	Item *granio.MenuItem

	alignment     granio.TextAlign
	child, parent *Menu
}

// newMenuItem creates a new menuItem.
func newMenuItem(item *granio.MenuItem, parent *Menu) *menuItem {
	i := &menuItem{Item: item, parent: parent}
	i.alignment = parent.alignment
	i.ExtendBaseWidget(i)
	return i
}

func (i *menuItem) Child() *Menu {
	if i.Item.ChildMenu != nil && i.child == nil {
		child := NewMenu(i.Item.ChildMenu)
		child.Hide()
		child.OnDismiss = i.parent.Dismiss
		child.isSubmenu = true
		i.child = child
	}
	return i.child
}

// CreateRenderer returns a new renderer for the menu item.
func (i *menuItem) CreateRenderer() granio.WidgetRenderer {
	th := i.parent.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	background := canvas.NewRectangle(th.Color(theme.ColorNameHover, v))
	background.CornerRadius = th.Size(theme.SizeNameMenuRadius)
	background.Hide()
	text := canvas.NewText(i.Item.Label, th.Color(theme.ColorNameForeground, v))
	text.Alignment = i.alignment
	objects := []granio.CanvasObject{background, text}
	var expandIcon *canvas.Image
	if i.Item.ChildMenu != nil {
		expandIcon = canvas.NewImageFromResource(th.Icon(theme.IconNameMenuExpand))
		objects = append(objects, expandIcon)
	}
	checkIcon := canvas.NewImageFromResource(th.Icon(theme.IconNameConfirm))
	if !i.Item.Checked {
		checkIcon.Hide()
	}
	var icon *canvas.Image
	if i.Item.Icon != nil {
		icon = canvas.NewImageFromResource(i.Item.Icon)
		objects = append(objects, icon)
	}
	var shortcutTexts []*canvas.Text
	if s, ok := i.Item.Shortcut.(granio.KeyboardShortcut); ok {
		shortcutTexts = textsForShortcut(s, th)
		for _, t := range shortcutTexts {
			objects = append(objects, t)
		}
	}

	objects = append(objects, checkIcon)
	r := &menuItemRenderer{
		BaseRenderer:  widget.NewBaseRenderer(objects),
		i:             i,
		expandIcon:    expandIcon,
		checkIcon:     checkIcon,
		icon:          icon,
		shortcutTexts: shortcutTexts,
		text:          text,
		background:    background,
	}
	r.updateVisuals()
	return r
}

// MouseIn activates the item which shows the submenu if the item has one.
// The submenu of any sibling of the item will be hidden.
func (i *menuItem) MouseIn(*desktop.MouseEvent) {
	i.activate()
}

// MouseMoved does nothing.
func (*menuItem) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut deactivates the item unless it has an open submenu.
func (i *menuItem) MouseOut() {
	if !i.isSubmenuOpen() {
		i.deactivate()
	}
}

// Tapped performs the action of the item and dismisses the menu.
// It does nothing if the item doesn’t have an action.
func (i *menuItem) Tapped(*granio.PointEvent) {
	if i.Item.Disabled {
		return
	}

	if i.Item.Action == nil || i.Item.ChildMenu != nil {
		if granio.CurrentDevice().IsMobile() {
			i.activate()
		}

		return
	}

	i.trigger()
}

func (i *menuItem) activate() {
	if i.Item.Disabled {
		return
	}
	if i.Child() != nil {
		i.Child().Show()
	}
	i.parent.activateItem(i)
}

func (i *menuItem) activateLastSubmenu() bool {
	if i.Child() == nil {
		return false
	}
	if i.isSubmenuOpen() {
		return i.Child().ActivateLastSubmenu()
	}
	i.Child().Show()
	i.Child().ActivateNext()
	return true
}

func (i *menuItem) deactivate() {
	if i.Child() != nil {
		i.Child().Hide()
	}
	i.parent.DeactivateChild()
}

func (i *menuItem) deactivateLastSubmenu() bool {
	if !i.isSubmenuOpen() {
		return false
	}
	if !i.Child().DeactivateLastSubmenu() {
		i.Child().DeactivateChild()
		i.Child().Hide()
	}
	return true
}

func (i *menuItem) isActive() bool {
	return i.parent.activeItem == i
}

func (i *menuItem) isSubmenuOpen() bool {
	return i.Child() != nil && i.Child().Visible()
}

func (i *menuItem) trigger() {
	i.parent.Dismiss()
	if i.Item.Action != nil {
		i.Item.Action()
	}
}

func (i *menuItem) triggerLast() {
	if i.isSubmenuOpen() {
		i.Child().TriggerLast()
		return
	}
	i.trigger()
}

type menuItemRenderer struct {
	widget.BaseRenderer
	i                *menuItem
	background       *canvas.Rectangle
	checkIcon        *canvas.Image
	expandIcon       *canvas.Image
	icon             *canvas.Image
	lastThemePadding float32
	minSize          granio.Size
	shortcutTexts    []*canvas.Text
	text             *canvas.Text
}

func (r *menuItemRenderer) Layout(size granio.Size) {
	th := r.i.parent.Theme()
	innerPad := th.Size(theme.SizeNameInnerPadding)
	pad := th.Size(theme.SizeNamePadding)
	inlineIcon := th.Size(theme.SizeNameInlineIcon)

	leftOffset := innerPad + r.checkSpace()
	rightOffset := size.Width
	iconSize := granio.NewSquareSize(inlineIcon)
	iconTopOffset := (size.Height - inlineIcon) / 2

	if r.expandIcon != nil {
		rightOffset -= inlineIcon
		r.expandIcon.Resize(iconSize)
		r.expandIcon.Move(granio.NewPos(rightOffset, iconTopOffset))
	}

	rightOffset -= innerPad
	textHeight := r.text.MinSize().Height
	for i := len(r.shortcutTexts) - 1; i >= 0; i-- {
		text := r.shortcutTexts[i]
		text.Resize(text.MinSize())
		rightOffset -= text.MinSize().Width
		text.Move(granio.NewPos(rightOffset, innerPad+(textHeight-text.Size().Height)))

		if i == 0 {
			rightOffset -= innerPad
		}
	}

	r.checkIcon.Resize(iconSize)
	r.checkIcon.Move(granio.NewPos(innerPad, iconTopOffset))

	if r.icon != nil {
		r.icon.Resize(iconSize)
		r.icon.Move(granio.NewPos(leftOffset, iconTopOffset))
		leftOffset += inlineIcon
		leftOffset += innerPad
	}

	r.text.Resize(granio.NewSize(rightOffset-leftOffset, textHeight))
	r.text.Move(granio.NewPos(leftOffset, innerPad))

	r.background.Resize(size.Subtract(granio.NewSquareSize(pad)))
	r.background.Move(granio.NewPos(pad/2, pad/2))
}

func (r *menuItemRenderer) MinSize() granio.Size {
	if r.minSizeUnchanged() {
		return r.minSize
	}

	th := r.i.parent.Theme()
	innerPad := th.Size(theme.SizeNameInnerPadding)
	inlineIcon := th.Size(theme.SizeNameInlineIcon)
	innerPad2 := innerPad * 2

	minSize := r.text.MinSize().AddWidthHeight(innerPad2+r.checkSpace(), innerPad2)
	if r.expandIcon != nil {
		minSize = minSize.AddWidthHeight(inlineIcon, 0)
	}
	if r.icon != nil {
		minSize = minSize.AddWidthHeight(inlineIcon+innerPad, 0)
	}
	if r.shortcutTexts != nil {
		var textWidth float32
		for _, text := range r.shortcutTexts {
			textWidth += text.MinSize().Width
		}
		minSize = minSize.AddWidthHeight(textWidth+innerPad, 0)
	}
	r.minSize = minSize
	return r.minSize
}

func (r *menuItemRenderer) updateVisuals() {
	th := r.i.parent.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()
	r.background.CornerRadius = th.Size(theme.SizeNameMenuRadius)
	r.background.StrokeWidth = 0
	if granio.CurrentDevice().IsMobile() || !r.i.isActive() {
		r.background.Hide()
	} else {
		r.background.FillColor = th.Color(theme.ColorNameFocus, v)
		if _, native := activeSystemColors(th); native {
			applySelectionStyle(r.background, th, v)
		}
		r.background.Show()
	}
	r.background.Refresh()
	r.text.Text = r.i.Item.Label
	r.text.Alignment = r.i.alignment
	r.refreshText(r.text, false)
	for _, text := range r.shortcutTexts {
		r.refreshText(text, true)
	}

	if r.i.Item.Checked {
		r.checkIcon.Show()
	} else {
		r.checkIcon.Hide()
	}
	r.updateIcon(r.checkIcon, th.Icon(theme.IconNameConfirm))
	r.updateIcon(r.expandIcon, th.Icon(theme.IconNameMenuExpand))
	r.updateIcon(r.icon, r.i.Item.Icon)
}

func (r *menuItemRenderer) Refresh() {
	r.updateVisuals()
	canvas.Refresh(r.i)
}

func (r *menuItemRenderer) checkSpace() float32 {
	if r.i.parent.containsCheck {
		return theme.IconInlineSize() + theme.InnerPadding()
	}
	return 0
}

func (r *menuItemRenderer) minSizeUnchanged() bool {
	th := r.i.parent.Theme()

	return !r.minSize.IsZero() &&
		r.text.TextSize == th.Size(theme.SizeNameText) &&
		(r.expandIcon == nil || r.expandIcon.Size().Width == th.Size(theme.SizeNameInlineIcon)) &&
		r.lastThemePadding == th.Size(theme.SizeNameInnerPadding)
}

func (r *menuItemRenderer) updateIcon(img *canvas.Image, rsc granio.Resource) {
	if img == nil {
		return
	}
	if r.i.Item.Disabled && svg.IsResourceSVG(rsc) {
		img.Resource = theme.NewDisabledResource(rsc)
	} else {
		img.Resource = rsc
	}
}

func (r *menuItemRenderer) refreshText(text *canvas.Text, shortcut bool) {
	th := r.i.parent.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	text.TextSize = th.Size(theme.SizeNameText)
	if r.i.Item.Disabled {
		text.Color = th.Color(theme.ColorNameDisabled, v)
	} else {
		if shortcut {
			text.Color = shortcutColor(th)
		} else {
			text.Color = th.Color(theme.ColorNameForeground, v)
		}
	}
	text.Refresh()
}

func shortcutColor(th granio.Theme) color.Color {
	if colors, ok := activeSystemColors(th); ok {
		return colors.GrayText
	}
	v := granio.CurrentApp().Settings().ThemeVariant()
	r, g, b, a := fynecolor.ToNRGBA(th.Color(theme.ColorNameForeground, v))
	a = uint8(float32(a) * 0.198)
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

func textsForShortcut(sc granio.KeyboardShortcut, th granio.Theme) (texts []*canvas.Text) {
	// add modifier
	b := strings.Builder{}
	mods := sc.Mod()
	if mods&granio.KeyModifierControl != 0 {
		b.WriteString(textModifierControl)
	}
	if mods&granio.KeyModifierAlt != 0 {
		b.WriteString(textModifierAlt)
	}
	if mods&granio.KeyModifierShift != 0 {
		b.WriteString(textModifierShift)
	}
	if mods&granio.KeyModifierSuper != 0 {
		b.WriteString(textModifierSuper)
	}
	shortColor := shortcutColor(th)
	if b.Len() > 0 {
		t := canvas.NewText(b.String(), shortColor)
		t.TextStyle = styleModifiers
		texts = append(texts, t)
	}
	// add key
	style := defaultStyleKeys
	s, ok := keyTexts[sc.Key()]
	if !ok {
		s = string(sc.Key())
	} else if len(s) == 1 {
		style = granio.TextStyle{Symbol: true}
	}
	t := canvas.NewText(s, shortColor)
	t.TextStyle = style
	texts = append(texts, t)
	return texts
}
