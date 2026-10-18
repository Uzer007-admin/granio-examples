package container

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	intTheme "github.com/Uzer007-admin/granio/v2/internal/theme"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// TabItem represents a single view in a tab view.
// The Text and Icon are used for the tab button and the Content is shown when the corresponding tab is active.
//
// Since: 1.4
type TabItem struct {
	Text    string
	Icon    granio.Resource
	Content granio.CanvasObject

	button *tabButton

	disabled bool
}

// Disabled returns whether or not the TabItem is disabled.
//
// Since: 2.3
func (ti *TabItem) Disabled() bool {
	return ti.disabled
}

func (ti *TabItem) disable() {
	ti.disabled = true
	if ti.button != nil {
		ti.button.Disable()
	}
}

func (ti *TabItem) enable() {
	ti.disabled = false
	if ti.button != nil {
		ti.button.Enable()
	}
}

// TabLocation is the location where the tabs of a tab container should be rendered
//
// Since: 1.4
type TabLocation int

// TabLocation values
const (
	TabLocationTop TabLocation = iota
	TabLocationLeading
	TabLocationBottom
	TabLocationTrailing
)

// NewTabItem creates a new item for a tabbed widget - each item specifies the content and a label for its tab.
//
// Since: 1.4
func NewTabItem(text string, content granio.CanvasObject) *TabItem {
	return &TabItem{Text: text, Content: content}
}

// NewTabItemWithIcon creates a new item for a tabbed widget - each item specifies the content and a label with an icon for its tab.
//
// Since: 1.4
func NewTabItemWithIcon(text string, icon granio.Resource, content granio.CanvasObject) *TabItem {
	return &TabItem{Text: text, Icon: icon, Content: content}
}

type baseTabs interface {
	granio.Widget

	onUnselected() func(*TabItem)
	onSelected() func(*TabItem)

	items() []*TabItem
	applyItems([]*TabItem)

	getCurrent() int
	setCurrent(int)

	tabLocation() TabLocation

	transitioning() bool
	setTransitioning(bool)
}

func isMobile(b baseTabs) bool {
	d := granio.CurrentDevice()
	mobile := intTheme.FeatureForWidget(intTheme.FeatureNameDeviceIsMobile, b)
	if is, ok := mobile.(bool); ok {
		return is
	}

	return d.IsMobile()
}

func tabsAdjustedLocation(l TabLocation, b baseTabs) TabLocation {
	// Mobile has limited screen space, so don't put app tab bar on long edges
	if isMobile(b) {
		if o := granio.CurrentDevice().Orientation(); granio.IsVertical(o) {
			if l == TabLocationLeading {
				return TabLocationTop
			}
			if l == TabLocationTrailing {
				return TabLocationBottom
			}
		} else {
			if l == TabLocationTop {
				return TabLocationLeading
			}
			if l == TabLocationBottom {
				return TabLocationTrailing
			}
		}
	}

	return l
}

func buildPopUpMenu(t baseTabs, button *widget.Button, items []*granio.MenuItem) *widget.PopUpMenu {
	d := granio.CurrentApp().Driver()
	c := d.CanvasForObject(button)
	if c == nil {
		// Overflow button detached from its canvas; nothing to host the
		// pop-up on (see fyne-io/fyne#5965).
		return nil
	}
	popUpMenu := widget.NewPopUpMenu(granio.NewMenu("", items...), c)
	buttonPos := d.AbsolutePositionForObject(button)
	buttonSize := button.Size()
	popUpMin := popUpMenu.MinSize()
	var popUpPos granio.Position
	switch t.tabLocation() {
	case TabLocationLeading:
		popUpPos.X = buttonPos.X + buttonSize.Width
		popUpPos.Y = buttonPos.Y + buttonSize.Height - popUpMin.Height
	case TabLocationTrailing:
		popUpPos.X = buttonPos.X - popUpMin.Width
		popUpPos.Y = buttonPos.Y + buttonSize.Height - popUpMin.Height
	case TabLocationTop:
		popUpPos.X = buttonPos.X + buttonSize.Width - popUpMin.Width
		popUpPos.Y = buttonPos.Y + buttonSize.Height
	case TabLocationBottom:
		popUpPos.X = buttonPos.X + buttonSize.Width - popUpMin.Width
		popUpPos.Y = buttonPos.Y - popUpMin.Height
	}
	if popUpPos.X < 0 {
		popUpPos.X = 0
	}
	if popUpPos.Y < 0 {
		popUpPos.Y = 0
	}
	popUpMenu.ShowAtPosition(popUpPos)
	return popUpMenu
}

func removeIndex(t baseTabs, index int) {
	items := t.items()
	if index < 0 || index >= len(items) {
		return
	}
	setItems(t, append(items[:index], items[index+1:]...))
	if s := t.getCurrent(); index < s {
		t.setCurrent(s - 1)
	}
}

func removeItem(t baseTabs, item *TabItem) {
	for index, existingItem := range t.items() {
		if existingItem == item {
			removeIndex(t, index)
			break
		}
	}
}

func selected(t baseTabs) *TabItem {
	selected := t.getCurrent()
	items := t.items()
	if selected < 0 || selected >= len(items) {
		return nil
	}
	return items[selected]
}

func selectIndex(t baseTabs, index int) {
	selected := t.getCurrent()

	if selected == index {
		// No change, so do nothing
		return
	}

	items := t.items()

	if f := t.onUnselected(); f != nil && selected >= 0 && selected < len(items) {
		// Notification of unselected
		f(items[selected])
	}

	if index < 0 || index >= len(items) {
		// Out of bounds, so do nothing
		return
	}

	t.setTransitioning(true)
	t.setCurrent(index)
	t.Refresh()

	if f := t.onSelected(); f != nil {
		// Notification of selected
		f(items[index])
	}
}

func selectItem(t baseTabs, item *TabItem) {
	for i, child := range t.items() {
		if child == item {
			selectIndex(t, i)
			return
		}
	}
}

func setItems(t baseTabs, items []*TabItem) {
	if build.HasHints && mismatchedTabItems(items) {
		internal.LogHint("Tab items should all have the same type of content (text, icons or both)")
	}
	t.applyItems(items)
	selected := t.getCurrent()
	count := len(items)
	switch {
	case count == 0:
		// No items available to be selected
		selectIndex(t, -1) // Unsure OnUnselected gets called if applicable
		t.setCurrent(-1)
	case selected < 0:
		// Current is first tab item
		selectIndex(t, 0)
	case selected >= count:
		// Current doesn't exist, select last tab
		selectIndex(t, count-1)
	}
}

func disableIndex(t baseTabs, index int) {
	items := t.items()
	if index < 0 || index >= len(items) {
		return
	}

	item := items[index]
	item.disable()

	if selected(t) == item {
		// the disabled tab is currently selected, so select the first enabled tab
		for i, it := range items {
			if !it.Disabled() {
				selectIndex(t, i)
				break
			}
		}
	}

	if selected(t) == item {
		selectIndex(t, -1) // no other tab is able to be selected
	}
}

func disableItem(t baseTabs, item *TabItem) {
	for i, it := range t.items() {
		if it == item {
			disableIndex(t, i)
			return
		}
	}
}

func enableIndex(t baseTabs, index int) {
	items := t.items()
	if index < 0 || index >= len(items) {
		return
	}

	item := items[index]
	item.enable()
}

func enableItem(t baseTabs, item *TabItem) {
	for i, it := range t.items() {
		if it == item {
			enableIndex(t, i)
			return
		}
	}
}

type baseTabsRenderer struct {
	positionAnimation, sizeAnimation *granio.Animation

	lastIndicatorPos    granio.Position
	lastIndicatorSize   granio.Size
	lastIndicatorHidden bool

	action             *widget.Button
	bar                *granio.Container
	divider, indicator *canvas.Rectangle

	tabs baseTabs
}

func (*baseTabsRenderer) Destroy() {
}

func (r *baseTabsRenderer) applyTheme(t baseTabs) {
	if r.action != nil {
		r.action.SetIcon(moreIcon(t))
	}
	th := theme.CurrentForWidget(t)
	v := granio.CurrentApp().Settings().ThemeVariant()

	r.divider.FillColor = th.Color(theme.ColorNameShadow, v)
	r.indicator.FillColor = th.Color(theme.ColorNamePrimary, v)
	r.indicator.CornerRadius = th.Size(theme.SizeNameSelectionRadius)

	for _, tab := range r.tabs.items() {
		tab.Content.Refresh()
	}
}

func (r *baseTabsRenderer) layout(t baseTabs, size granio.Size) {
	var (
		barPos, dividerPos, contentPos    granio.Position
		barSize, dividerSize, contentSize granio.Size
	)

	barMin := r.bar.MinSize()

	th := theme.CurrentForWidget(t)
	padding := th.Size(theme.SizeNamePadding)
	dividerWidth := th.Size(theme.SizeNameSeparatorThickness)
	switch t.tabLocation() {
	case TabLocationTop:
		barHeight := barMin.Height
		barPos = granio.NewPos(0, 0)
		barSize = granio.NewSize(size.Width, barHeight)
		dividerPos = granio.NewPos(0, barHeight)
		dividerSize = granio.NewSize(size.Width, dividerWidth)
		contentPos = granio.NewPos(0, barHeight+padding)
		contentSize = granio.NewSize(size.Width, size.Height-barHeight-padding)
	case TabLocationLeading:
		barWidth := barMin.Width
		barPos = granio.NewPos(0, 0)
		barSize = granio.NewSize(barWidth, size.Height)
		dividerPos = granio.NewPos(barWidth, 0)
		dividerSize = granio.NewSize(dividerWidth, size.Height)
		contentPos = granio.NewPos(barWidth+padding, 0)
		contentSize = granio.NewSize(size.Width-barWidth-padding, size.Height)
	case TabLocationBottom:
		barHeight := barMin.Height
		barPos = granio.NewPos(0, size.Height-barHeight)
		barSize = granio.NewSize(size.Width, barHeight)
		dividerPos = granio.NewPos(0, size.Height-barHeight-padding)
		dividerSize = granio.NewSize(size.Width, dividerWidth)
		contentPos = granio.NewPos(0, 0)
		contentSize = granio.NewSize(size.Width, size.Height-barHeight-padding)
	case TabLocationTrailing:
		barWidth := barMin.Width
		barPos = granio.NewPos(size.Width-barWidth, 0)
		barSize = granio.NewSize(barWidth, size.Height)
		dividerPos = granio.NewPos(size.Width-barWidth-padding, 0)
		dividerSize = granio.NewSize(dividerWidth, size.Height)
		contentPos = granio.NewPos(0, 0)
		contentSize = granio.NewSize(size.Width-barWidth-padding, size.Height)
	}

	r.bar.Move(barPos)
	r.bar.Resize(barSize)
	r.divider.Move(dividerPos)
	r.divider.Resize(dividerSize)
	selected := t.getCurrent()
	for i, ti := range t.items() {
		if i == selected {
			ti.Content.Move(contentPos)
			ti.Content.Resize(contentSize)
			ti.Content.Show()
		} else {
			ti.Content.Hide()
		}
	}
}

func (r *baseTabsRenderer) minSize(t baseTabs) granio.Size {
	th := theme.CurrentForWidget(t)
	pad := th.Size(theme.SizeNamePadding)
	buttonPad := pad
	barMin := r.bar.MinSize()
	tabsMin := r.bar.Objects[0].MinSize()
	accessory := r.bar.Objects[1]
	accessoryMin := accessory.MinSize()
	if scroll, ok := r.bar.Objects[0].(*Scroll); ok && len(scroll.Content.(*granio.Container).Objects) == 0 {
		tabsMin = granio.Size{} // scroller forces 32 where we don't need any space
		buttonPad = 0
	} else if group, ok := r.bar.Objects[0].(*granio.Container); ok && len(group.Objects) > 0 {
		tabsMin = group.Objects[0].MinSize()
		buttonPad = 0
	}
	if !accessory.Visible() || accessoryMin.Width == 0 {
		buttonPad = 0
		accessoryMin = granio.Size{}
	}

	contentMin := granio.NewSize(0, 0)
	for _, content := range t.items() {
		contentMin = internal.MaxSizes(contentMin, content.Content.MinSize())
	}

	switch t.tabLocation() {
	case TabLocationLeading, TabLocationTrailing:
		return granio.NewSize(barMin.Width+contentMin.Width+pad,
			max(contentMin.Height, accessoryMin.Height+buttonPad+tabsMin.Height))
	default:
		return granio.NewSize(max(contentMin.Width, accessoryMin.Width+buttonPad+tabsMin.Width),
			barMin.Height+contentMin.Height+pad)
	}
}

func (r *baseTabsRenderer) moveIndicator(pos granio.Position, siz granio.Size, th granio.Theme, animate bool) {
	isSameState := r.lastIndicatorPos == pos && r.lastIndicatorSize == siz &&
		r.lastIndicatorHidden == r.indicator.Hidden
	if isSameState {
		return
	}

	if r.positionAnimation != nil {
		r.positionAnimation.Stop()
		r.positionAnimation = nil
	}
	if r.sizeAnimation != nil {
		r.sizeAnimation.Stop()
		r.sizeAnimation = nil
	}

	v := granio.CurrentApp().Settings().ThemeVariant()
	r.indicator.FillColor = th.Color(theme.ColorNamePrimary, v)
	if r.indicator.Position().IsZero() {
		r.indicator.Move(pos)
		r.indicator.Resize(siz)
		r.indicator.Refresh()
		return
	}

	r.lastIndicatorPos = pos
	r.lastIndicatorSize = siz
	r.lastIndicatorHidden = r.indicator.Hidden

	if animate && granio.CurrentApp().Settings().ShowAnimations() {
		r.positionAnimation = canvas.NewPositionAnimation(r.indicator.Position(), pos, canvas.DurationShort, func(p granio.Position) {
			r.indicator.Move(p)
			r.indicator.Refresh()
			if pos == p {
				r.positionAnimation.Stop()
				r.positionAnimation = nil
			}
		})
		r.sizeAnimation = canvas.NewSizeAnimation(r.indicator.Size(), siz, canvas.DurationShort, func(s granio.Size) {
			r.indicator.Resize(s)
			r.indicator.Refresh()
			if siz == s {
				r.sizeAnimation.Stop()
				r.sizeAnimation = nil
			}
		})

		r.positionAnimation.Start()
		r.sizeAnimation.Start()
	} else {
		r.indicator.Move(pos)
		r.indicator.Resize(siz)
		r.indicator.Refresh()
	}
}

func (r *baseTabsRenderer) objects(t baseTabs) []granio.CanvasObject {
	objects := []granio.CanvasObject{r.bar, r.divider, r.indicator}
	if i, is := t.getCurrent(), t.items(); i >= 0 && i < len(is) {
		objects = append(objects, is[i].Content)
	}
	return objects
}

func (r *baseTabsRenderer) refresh(t baseTabs) {
	r.applyTheme(t)

	r.bar.Refresh()
	r.divider.Refresh()
	r.indicator.Refresh()
}

type buttonIconPosition int

const (
	buttonIconInline buttonIconPosition = iota
	buttonIconTop
)

var (
	_ granio.Widget     = (*tabButton)(nil)
	_ granio.Tappable   = (*tabButton)(nil)
	_ desktop.Hoverable = (*tabButton)(nil)
)

type tabButton struct {
	widget.DisableableWidget
	hovered       bool
	icon          granio.Resource
	iconPosition  buttonIconPosition
	importance    widget.Importance
	onTapped      func()
	onClosed      func()
	text          string
	textAlignment granio.TextAlign

	tabs baseTabs
}

func (b *tabButton) CreateRenderer() granio.WidgetRenderer {
	b.ExtendBaseWidget(b)
	th := b.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	background := canvas.NewRectangle(th.Color(theme.ColorNameHover, v))
	background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
	background.Hide()
	icon := canvas.NewImageFromResource(b.icon)
	if b.icon == nil {
		icon.Hide()
	}

	label := canvas.NewText(b.text, th.Color(theme.ColorNameForeground, v))
	label.TextStyle.Bold = true

	buttonClose := &tabCloseButton{
		parent: b,
		onTapped: func() {
			if f := b.onClosed; f != nil {
				f()
			}
		},
	}
	buttonClose.ExtendBaseWidget(buttonClose)
	buttonClose.Hide()

	objects := []granio.CanvasObject{background, label, buttonClose, icon}
	return &tabButtonRenderer{
		button:     b,
		background: background,
		icon:       icon,
		label:      label,
		close:      buttonClose,
		objects:    objects,
	}
}

func (b *tabButton) MinSize() granio.Size {
	b.ExtendBaseWidget(b)
	return b.BaseWidget.MinSize()
}

func (b *tabButton) MouseIn(*desktop.MouseEvent) {
	if isMobile(b.tabs) { // mobile simulation still has desktop events
		return
	}

	b.hovered = true
	b.Refresh()
}

func (*tabButton) MouseMoved(*desktop.MouseEvent) {
}

func (b *tabButton) MouseOut() {
	if isMobile(b.tabs) { // mobile simulation still has desktop events
		return
	}

	b.hovered = false
	b.Refresh()
}

func (b *tabButton) Tapped(*granio.PointEvent) {
	if b.Disabled() {
		return
	}

	b.onTapped()
}

type tabButtonRenderer struct {
	button     *tabButton
	background *canvas.Rectangle
	icon       *canvas.Image
	label      *canvas.Text
	close      *tabCloseButton
	objects    []granio.CanvasObject
}

func (*tabButtonRenderer) Destroy() {
}

func (r *tabButtonRenderer) Layout(size granio.Size) {
	th := r.button.Theme()
	pad := th.Size(theme.SizeNamePadding)
	r.background.Resize(size)
	padding := r.padding()
	innerSize := size.Subtract(padding)
	innerOffset := granio.NewPos(padding.Width/2, padding.Height/2)
	labelShift := float32(0)
	if r.icon.Visible() {
		shiftPad := pad
		iconSize := r.iconSize()
		var iconOffset granio.Position
		if r.button.iconPosition == buttonIconTop {
			iconOffset = granio.NewPos((innerSize.Width-iconSize)/2, 0)
			shiftPad = 0
		} else {
			iconOffset = granio.NewPos(0, (innerSize.Height-iconSize)/2)
		}
		r.icon.Resize(granio.NewSquareSize(iconSize))
		r.icon.Move(innerOffset.Add(iconOffset))
		labelShift = iconSize + shiftPad
	}
	if r.label.Text != "" {
		var labelOffset granio.Position
		var labelSize granio.Size
		if r.button.iconPosition == buttonIconTop {
			labelOffset = granio.NewPos(0, labelShift)
			labelSize = granio.NewSize(innerSize.Width, r.label.MinSize().Height)
		} else {
			labelOffset = granio.NewPos(labelShift, 0)
			labelSize = granio.NewSize(innerSize.Width-labelShift, innerSize.Height)
		}
		r.label.Resize(labelSize)
		r.label.Move(innerOffset.Add(labelOffset))
	}
	inlineIconSize := th.Size(theme.SizeNameInlineIcon)
	r.close.Move(granio.NewPos(size.Width-inlineIconSize-pad, (size.Height-inlineIconSize)/2))
	r.close.Resize(granio.NewSquareSize(inlineIconSize))
}

func (r *tabButtonRenderer) MinSize() granio.Size {
	th := r.button.Theme()
	var contentWidth, contentHeight float32
	textSize := r.label.MinSize()
	iconSize := r.iconSize()
	padding := th.Size(theme.SizeNamePadding)
	if r.button.iconPosition == buttonIconTop {
		contentWidth = max(textSize.Width, iconSize)
		if r.icon.Visible() {
			contentHeight += iconSize
		}
		if r.label.Text != "" {
			contentHeight += textSize.Height
		}
	} else {
		contentHeight = max(textSize.Height, iconSize)
		if r.icon.Visible() {
			contentWidth += iconSize
		}
		if r.label.Text != "" {
			if r.icon.Visible() {
				contentWidth += padding
			}
			contentWidth += textSize.Width
		}
	}
	if r.button.onClosed != nil {
		inlineIconSize := th.Size(theme.SizeNameInlineIcon)
		contentWidth += inlineIconSize + padding
		contentHeight = max(contentHeight, inlineIconSize)
	}
	return granio.NewSize(contentWidth, contentHeight).Add(r.padding())
}

func (r *tabButtonRenderer) Objects() []granio.CanvasObject {
	return r.objects
}

func (r *tabButtonRenderer) Refresh() {
	th := r.button.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	if r.button.hovered && !r.button.Disabled() {
		r.background.FillColor = th.Color(theme.ColorNameHover, v)
		r.background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
		r.background.Show()
	} else {
		r.background.Hide()
	}
	r.background.Refresh()

	r.label.Text = r.button.text
	r.label.Alignment = r.button.textAlignment
	if !r.button.Disabled() {
		if r.button.importance == widget.HighImportance {
			r.label.Color = th.Color(theme.ColorNamePrimary, v)
		} else {
			r.label.Color = th.Color(theme.ColorNameForeground, v)
		}
	} else {
		r.label.Color = th.Color(theme.ColorNameDisabled, v)
	}
	r.label.TextSize = th.Size(theme.SizeNameText)
	if isMobile(r.button.tabs) {
		r.label.TextSize = th.Size(theme.SizeNameCaptionText)
	}
	if r.button.text == "" {
		r.label.Hide()
	} else {
		r.label.Show()
	}

	r.icon.Resource = r.button.icon
	if r.icon.Resource != nil {
		r.icon.Show()
		switch res := r.icon.Resource.(type) {
		case *theme.ThemedResource:
			if r.button.importance == widget.HighImportance {
				r.icon.Resource = theme.NewPrimaryThemedResource(res)
			}
		case *theme.PrimaryThemedResource:
			if r.button.importance != widget.HighImportance {
				r.icon.Resource = res.Original()
			}
		}
		r.icon.Refresh()
	} else {
		r.icon.Hide()
	}

	if r.button.onClosed != nil && (isMobile(r.button.tabs) || r.button.hovered || r.close.hovered) {
		r.close.Show()
	} else {
		r.close.Hide()
	}
	r.close.Refresh()

	canvas.Refresh(r.button)
}

func (r *tabButtonRenderer) iconSize() float32 {
	iconSize := r.button.Theme().Size(theme.SizeNameInlineIcon)
	if r.button.iconPosition == buttonIconTop {
		return 1.5 * iconSize //revive:disable-line:add-constant
	}

	return iconSize
}

func (r *tabButtonRenderer) padding() granio.Size {
	padding := r.button.Theme().Size(theme.SizeNameInnerPadding)
	if r.label.Text != "" && r.button.iconPosition == buttonIconInline {
		return granio.NewSquareSize(padding * 2)
	}
	return granio.NewSize(padding, padding)
}

var (
	_ granio.Widget     = (*tabCloseButton)(nil)
	_ granio.Tappable   = (*tabCloseButton)(nil)
	_ desktop.Hoverable = (*tabCloseButton)(nil)
)

type tabCloseButton struct {
	widget.BaseWidget
	parent   *tabButton
	hovered  bool
	onTapped func()
}

func (b *tabCloseButton) CreateRenderer() granio.WidgetRenderer {
	b.ExtendBaseWidget(b)
	th := b.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	background := canvas.NewRectangle(th.Color(theme.ColorNameHover, v))
	background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
	background.Hide()
	icon := canvas.NewImageFromResource(theme.CancelIcon())

	return &tabCloseButtonRenderer{
		button:     b,
		background: background,
		icon:       icon,
		objects:    []granio.CanvasObject{background, icon},
	}
}

func (b *tabCloseButton) MinSize() granio.Size {
	b.ExtendBaseWidget(b)
	return b.BaseWidget.MinSize()
}

func (b *tabCloseButton) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	b.parent.Refresh()
}

func (*tabCloseButton) MouseMoved(*desktop.MouseEvent) {
}

func (b *tabCloseButton) MouseOut() {
	b.hovered = false
	b.parent.Refresh()
}

func (b *tabCloseButton) Tapped(*granio.PointEvent) {
	b.onTapped()
}

type tabCloseButtonRenderer struct {
	button     *tabCloseButton
	background *canvas.Rectangle
	icon       *canvas.Image
	objects    []granio.CanvasObject
}

func (*tabCloseButtonRenderer) Destroy() {
}

func (r *tabCloseButtonRenderer) Layout(size granio.Size) {
	r.background.Resize(size)
	r.icon.Resize(size)
}

func (r *tabCloseButtonRenderer) MinSize() granio.Size {
	return granio.NewSquareSize(r.button.Theme().Size(theme.SizeNameInlineIcon))
}

func (r *tabCloseButtonRenderer) Objects() []granio.CanvasObject {
	return r.objects
}

func (r *tabCloseButtonRenderer) Refresh() {
	th := r.button.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	if r.button.hovered {
		r.background.FillColor = th.Color(theme.ColorNameHover, v)
		r.background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
		r.background.Show()
	} else {
		r.background.Hide()
	}
	r.background.Refresh()
	switch res := r.icon.Resource.(type) {
	case *theme.ThemedResource:
		if r.button.parent.importance == widget.HighImportance {
			r.icon.Resource = theme.NewPrimaryThemedResource(res)
		}
	case *theme.PrimaryThemedResource:
		if r.button.parent.importance != widget.HighImportance {
			r.icon.Resource = res.Original()
		}
	}
	r.icon.Refresh()
}

func mismatchedTabItems(items []*TabItem) bool {
	var hasText, hasIcon bool
	for _, tab := range items {
		hasText = hasText || tab.Text != ""
		hasIcon = hasIcon || tab.Icon != nil
	}

	mismatch := false
	for _, tab := range items {
		if (hasText && tab.Text == "") || (hasIcon && tab.Icon == nil) {
			mismatch = true
			break
		}
	}

	return mismatch
}

func moreIcon(t baseTabs) granio.Resource {
	if l := t.tabLocation(); l == TabLocationLeading || l == TabLocationTrailing {
		return theme.MoreVerticalIcon()
	}
	return theme.MoreHorizontalIcon()
}
