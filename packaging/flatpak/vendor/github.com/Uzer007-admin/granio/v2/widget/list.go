package widget

import (
	"fmt"
	"math"
	"sort"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// ListItemID uniquely identifies an item within a list.
type ListItemID = int

// Declare conformity with interfaces.
var (
	_ granio.Widget    = (*List)(nil)
	_ granio.Focusable = (*List)(nil)
)

type listBind struct {
	listener annotatedListener

	oldLength func() int
	oldUpdate func(id ListItemID, item granio.CanvasObject)
}

// List is a widget that pools list items for performance and
// lays the items out in a vertical direction inside of a scroller.
// By default, List requires that all items are the same size, but specific
// rows can have their heights set with SetItemHeight.
//
// Since: 1.4
type List struct {
	BaseWidget
	SelectionMode SelectionMode

	// Length is a callback for returning the number of items in the list.
	Length func() int `json:"-"`

	// CreateItem is a callback invoked to create a new widget to render
	// a row in the list.
	CreateItem func() granio.CanvasObject `json:"-"`

	// UpdateItem is a callback invoked to update a list row widget
	// to display a new row in the list. The UpdateItem callback should
	// only update the given item, it should not invoke APIs that would
	// change other properties of the list itself.
	UpdateItem func(id ListItemID, item granio.CanvasObject) `json:"-"`

	// OnSelected is a callback to be notified when a given item
	// in the list has been selected.
	OnSelected func(id ListItemID) `json:"-"`

	// OnUnselected is a callback to be notified when a given item
	// in the list has been unselected.
	OnUnselected func(id ListItemID) `json:"-"`

	// HideSeparators hides the separators between list rows
	//
	// Since: 2.5
	HideSeparators bool

	// OnHighlighted is a callback to be notified when a given item
	// in the list has been highlighted by keyboard navigation and mouse hover
	//
	// Since: 2.8
	OnHighlighted func(id ListItemID) `json:"-"`

	currentHighlight  ListItemID
	focused           bool
	scroller          *widget.Scroll
	selected          []ListItemID
	itemMin           granio.Size
	itemHeights       map[ListItemID]float32
	offsetY           float32
	offsetUpdated     func(granio.Position)
	minSizeCache      granio.Size
	rowGeometry       collectionGeometry
	selectionAnchor   ListItemID
	selectionAnchored bool

	lastBind *listBind
}

// NewList creates and returns a list widget for displaying items in
// a vertical layout with scrolling and caching for performance.
//
// Since: 1.4
func NewList(length func() int, createItem func() granio.CanvasObject, updateItem func(ListItemID, granio.CanvasObject)) *List {
	list := &List{Length: length, CreateItem: createItem, UpdateItem: updateItem}
	list.ExtendBaseWidget(list)
	return list
}

// NewListWithData creates a new list widget that will display the contents of the provided data.
//
// Since: 2.0
func NewListWithData(data binding.DataList, createItem func() granio.CanvasObject, updateItem func(binding.DataItem, granio.CanvasObject)) *List {
	l := NewList(
		data.Length,
		createItem,
		func(i ListItemID, o granio.CanvasObject) {
			item, err := data.GetItem(i)
			if err != nil {
				granio.LogError(fmt.Sprintf("Error getting data item %d", i), err)
				return
			}
			updateItem(item, o)
		},
	)

	data.AddListener(binding.NewDataListener(l.Refresh))
	return l
}

// Bind connects the specified data source to this List.
// The current contents of the DataList will be used to determine length and content any changes in the data will cause the widget to update.
// The same types of item will be used but the replacement `update` function will be called when an item should update.
// Upon binding all items will update.
//
// Since: 2.8
func (l *List) Bind(data binding.DataList, update func(di binding.DataItem, o granio.CanvasObject)) {
	l.Unbind()

	oldLength := l.Length
	oldUpdate := l.UpdateItem
	l.Length = data.Length
	l.UpdateItem = func(i ListItemID, o granio.CanvasObject) {
		item, err := data.GetItem(i)
		if err != nil {
			granio.LogError(fmt.Sprintf("Error getting data item %d", i), err)
			return
		}
		update(item, o)
	}

	fn := binding.NewDataListener(l.Refresh)
	data.AddListener(fn)
	l.lastBind = &listBind{listener: annotatedListener{data: data, listener: fn}, oldLength: oldLength, oldUpdate: oldUpdate}
	l.Refresh()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (l *List) CreateRenderer() granio.WidgetRenderer {
	l.ExtendBaseWidget(l)

	if f := l.CreateItem; f != nil && l.itemMin.IsZero() {
		item := createItemAndApplyThemeScope(f, l)

		l.itemMin = item.MinSize()
	}

	layout := &granio.Container{Layout: newListLayout(l)}
	l.scroller = widget.NewVScroll(layout)
	layout.Resize(layout.MinSize())
	objects := []granio.CanvasObject{l.scroller}
	return newListRenderer(objects, l, l.scroller, layout)
}

// FocusGained is called after this List has gained focus.
func (l *List) FocusGained() {
	l.focused = true
	l.RefreshItem(l.currentHighlight)
	if f := l.OnHighlighted; f != nil {
		f(l.currentHighlight)
	}
}

// FocusLost is called after this List has lost focus.
func (l *List) FocusLost() {
	l.focused = false
	l.RefreshItem(l.currentHighlight)
}

// MinSize returns the size that this widget should not shrink below.
func (l *List) MinSize() granio.Size {
	l.ExtendBaseWidget(l)
	return l.BaseWidget.MinSize()
}

// RefreshItem refreshes a single item, specified by the item ID passed in.
//
// Since: 2.4
func (l *List) RefreshItem(id ListItemID) {
	l.rowGeometry.dirty = true
	l.minSizeCache = granio.Size{}
	if l.scroller == nil {
		return
	}
	l.BaseWidget.Refresh()
	lo, _ := l.scroller.Content.(*granio.Container).Layout.(*listLayout)
	item, ok := lo.searchVisible(lo.visible, id)
	if ok {
		lo.setupListItem(item, id, l.focused && l.currentHighlight == id)
	}
}

// SetItemHeight supports changing the height of the specified list item. Items normally take the height of the template
// returned from the CreateItem callback. The height parameter uses the same units as a granio.Size type and refers
// to the internal content height not including the divider size.
//
// Since: 2.3
func (l *List) SetItemHeight(id ListItemID, height float32) {
	if l.itemHeights == nil {
		l.itemHeights = make(map[ListItemID]float32)
	}

	refresh := l.itemHeights[id] != height
	l.itemHeights[id] = height

	if refresh {
		l.RefreshItem(id)
	}
}

// Unbind disconnects any configured data source from this List.
// The contents will return to what was displayed before binding.
// Upon unbinding all items will update.
//
// Since: 2.8
func (l *List) Unbind() {
	if l.lastBind == nil {
		return
	}

	l.lastBind.listener.data.RemoveListener(l.lastBind.listener.listener)
	l.UpdateItem = l.lastBind.oldUpdate
	l.Length = l.lastBind.oldLength
	l.lastBind = nil
	l.Refresh()
}

func (l *List) scrollWithoutItemCheckTo(id ListItemID) {
	if l.scroller == nil {
		return
	}

	separatorThickness := l.Theme().Size(theme.SizeNamePadding)
	y := float32(0)
	lastItemHeight := l.itemMin.Height

	if len(l.itemHeights) == 0 {
		y = (float32(id) * l.itemMin.Height) + (float32(id) * separatorThickness)
	} else {
		l.rowGeometry.prepare(l.itemHeights, l.Length(), l.itemMin.Height, separatorThickness)
		y = l.rowGeometry.position(id)
		lastItemHeight = l.itemMin.Height
		if h, ok := l.itemHeights[id]; ok {
			lastItemHeight = h
		}
	}
	if y < l.scroller.Offset.Y {
		l.scroller.Offset.Y = y
	} else if y+l.itemMin.Height > l.scroller.Offset.Y+l.scroller.Size().Height {
		l.scroller.Offset.Y = y + lastItemHeight - l.scroller.Size().Height
	}
	l.offsetUpdated(l.scroller.Offset)
}

// Resize is called when this list should change size. We refresh to ensure invisible items are drawn.
func (l *List) Resize(s granio.Size) {
	l.BaseWidget.Resize(s)
	if l.scroller == nil {
		return
	}

	l.offsetUpdated(l.scroller.Offset)
	l.scroller.Content.(*granio.Container).Layout.(*listLayout).updateList(true)
}

// Highlight scrolls to the item represented by id and highlights it
//
// Since: 2.8
func (l *List) Highlight(id ListItemID) {
	if l.Length() == 0 {
		return
	}

	newID := id
	if id < 0 {
		newID = 0
	}

	if id >= l.Length() {
		newID = l.Length() - 1
	}

	l.scrollWithoutItemCheckTo(newID)
	l.currentHighlight = newID
	if l.OnHighlighted != nil {
		l.OnHighlighted(newID)
	}
	l.Refresh()
}

// Select add the item identified by the given ID to the selection.
func (l *List) Select(id ListItemID) {
	if l.SelectionMode == SelectionMultiple {
		index := sort.SearchInts(l.selected, id)
		if l.Length == nil || id < 0 || id >= l.Length() || (index < len(l.selected) && l.selected[index] == id) {
			return
		}
		l.setSelection(append(append([]ListItemID(nil), l.selected...), id))
		l.scrollWithoutItemCheckTo(id)
		return
	}
	if len(l.selected) > 0 && id == l.selected[0] {
		return
	}
	length := 0
	if f := l.Length; f != nil {
		length = f()
	}
	if id < 0 || id >= length {
		return
	}
	old := l.selected
	l.selected = []ListItemID{id}
	defer func() {
		if f := l.OnUnselected; f != nil && len(old) > 0 {
			f(old[0])
		}
		if f := l.OnSelected; f != nil {
			f(id)
		}
	}()
	l.scrollWithoutItemCheckTo(id)
	l.Refresh()
}

// ScrollTo scrolls to the item represented by id
//
// Since: 2.1
func (l *List) ScrollTo(id ListItemID) {
	length := 0
	if f := l.Length; f != nil {
		length = f()
	}
	if id < 0 || id >= length {
		return
	}
	l.scrollWithoutItemCheckTo(id)
	l.Refresh()
}

// ScrollToBottom scrolls to the end of the list
//
// Since: 2.1
func (l *List) ScrollToBottom() {
	l.scroller.ScrollToBottom()
	l.offsetUpdated(l.scroller.Offset)
}

// ScrollToTop scrolls to the start of the list
//
// Since: 2.1
func (l *List) ScrollToTop() {
	l.scroller.ScrollToTop()
	l.offsetUpdated(l.scroller.Offset)
}

// ScrollToOffset scrolls the list to the given offset position.
//
// Since: 2.5
func (l *List) ScrollToOffset(offset float32) {
	if l.scroller == nil {
		return
	}
	if offset < 0 {
		offset = 0
	}
	contentHeight := l.contentMinSize().Height
	if l.Size().Height >= contentHeight {
		return // content fully visible - no need to scroll
	}
	if offset > contentHeight {
		offset = contentHeight
	}
	l.scroller.ScrollToOffset(granio.NewPos(0, offset))
	l.offsetUpdated(l.scroller.Offset)
}

// GetScrollOffset returns the current scroll offset position
//
// Since: 2.5
func (l *List) GetScrollOffset() float32 {
	return l.offsetY
}

// TypedKey is called if a key event happens while this List is focused.
func (l *List) TypedKey(event *granio.KeyEvent) {
	oldFocus := l.currentHighlight

	switch event.Name {
	case granio.KeySpace:
		if l.SelectionMode == SelectionMultiple {
			l.pointerSelect(l.currentHighlight, selectionModifiers()|granio.KeyModifierControl)
		} else {
			l.Select(l.currentHighlight)
		}
	case granio.KeyDown:
		if f := l.Length; f != nil && l.currentHighlight >= f()-1 {
			return
		}
		l.RefreshItem(l.currentHighlight)
		l.currentHighlight++
		l.scrollWithoutItemCheckTo(l.currentHighlight)
		l.RefreshItem(l.currentHighlight)
	case granio.KeyUp:
		if l.currentHighlight <= 0 {
			return
		}
		l.RefreshItem(l.currentHighlight)
		l.currentHighlight--
		l.scrollWithoutItemCheckTo(l.currentHighlight)
		l.RefreshItem(l.currentHighlight)
	}

	if oldFocus != l.currentHighlight {
		if l.SelectionMode == SelectionMultiple && selectionModifiers()&granio.KeyModifierShift != 0 {
			l.pointerSelect(l.currentHighlight, selectionModifiers())
		}
		if f := l.OnHighlighted; f != nil {
			f(l.currentHighlight)
		}
	}
}

// TypedRune is called if a text event happens while this List is focused.
func (*List) TypedRune(_ rune) {
	// intentionally left blank
}

// Unselect removes the item identified by the given ID from the selection.
func (l *List) Unselect(id ListItemID) {
	if l.SelectionMode == SelectionMultiple {
		for i, selected := range l.selected {
			if selected == id {
				l.setSelection(append(append([]ListItemID(nil), l.selected[:i]...), l.selected[i+1:]...))
				return
			}
		}
		return
	}
	if len(l.selected) == 0 || l.selected[0] != id {
		return
	}

	l.selected = nil
	l.Refresh()
	if f := l.OnUnselected; f != nil {
		f(id)
	}
}

// UnselectAll removes all items from the selection.
//
// Since: 2.1
func (l *List) UnselectAll() {
	if len(l.selected) == 0 {
		return
	}

	selected := l.selected
	l.selected = nil
	l.Refresh()
	if f := l.OnUnselected; f != nil {
		for _, id := range selected {
			f(id)
		}
	}
}

// Refresh causes this List to be redrawn in its current state
func (l *List) Refresh() {
	l.rowGeometry.dirty = true
	l.minSizeCache = granio.Size{}
	l.BaseWidget.Refresh()
}

func (l *List) contentMinSize() granio.Size {
	if !l.minSizeCache.IsZero() {
		return l.minSizeCache
	}

	separatorThickness := l.Theme().Size(theme.SizeNamePadding)
	if l.Length == nil {
		return granio.NewSize(0, 0)
	}
	items := l.Length()

	if len(l.itemHeights) == 0 {
		return granio.NewSize(l.itemMin.Width,
			(l.itemMin.Height+separatorThickness)*float32(items)-separatorThickness)
	}

	l.rowGeometry.prepare(l.itemHeights, items, l.itemMin.Height, separatorThickness)
	size := granio.NewSize(l.itemMin.Width, l.rowGeometry.position(items)-separatorThickness)
	l.minSizeCache = size
	return size
}

// fills l.visibleRowHeights and also returns offY and minRow
func (l *listLayout) calculateVisibleRowHeights(itemHeight float32, length int, th granio.Theme) (offY float32, minRow int) {
	l.visibleRowHeights = l.visibleRowHeights[:0]

	if l.list.scroller.Size().Height <= 0 {
		return offY, minRow
	}

	padding := th.Size(theme.SizeNamePadding)

	if len(l.list.itemHeights) == 0 {
		paddedItemHeight := itemHeight + padding

		offY = float32(math.Floor(float64(l.list.offsetY/paddedItemHeight))) * paddedItemHeight
		minRow = int(math.Floor(float64(offY / paddedItemHeight)))
		maxRow := int(math.Ceil(float64((offY + l.list.scroller.Size().Height) / paddedItemHeight)))

		if minRow > length-1 {
			minRow = length - 1
		}
		if minRow < 0 {
			minRow = 0
			offY = 0
		}

		if maxRow > length-1 {
			maxRow = length - 1
		}

		for i := 0; i <= maxRow-minRow; i++ {
			l.visibleRowHeights = append(l.visibleRowHeights, itemHeight)
		}
		return offY, minRow
	}

	geometry := &l.list.rowGeometry
	geometry.prepare(l.list.itemHeights, length, itemHeight, padding)
	minRow = geometry.itemAt(l.list.offsetY)
	offY = geometry.position(minRow)
	end := geometry.endAt(l.list.offsetY + l.list.scroller.Size().Height)
	for id := minRow; id < end; id++ {
		height := itemHeight
		if custom, ok := l.list.itemHeights[id]; ok {
			height = custom
		}
		l.visibleRowHeights = append(l.visibleRowHeights, height)
	}
	return offY, minRow
}

// Declare conformity with WidgetRenderer interface.
var _ granio.WidgetRenderer = (*listRenderer)(nil)

type listRenderer struct {
	widget.BaseRenderer

	list     *List
	scroller *widget.Scroll
	layout   *granio.Container
}

func newListRenderer(objects []granio.CanvasObject, l *List, scroller *widget.Scroll, layout *granio.Container) *listRenderer {
	lr := &listRenderer{BaseRenderer: widget.NewBaseRenderer(objects), list: l, scroller: scroller, layout: layout}
	lr.scroller.OnScrolled = l.offsetUpdated
	return lr
}

func (l *listRenderer) Layout(size granio.Size) {
	l.scroller.Resize(size)
}

func (l *listRenderer) MinSize() granio.Size {
	return internal.MaxSizes(l.scroller.MinSize(), l.list.itemMin)
}

func (l *listRenderer) Refresh() {
	l.list.minSizeCache = granio.Size{}
	if f := l.list.CreateItem; f != nil {
		item := createItemAndApplyThemeScope(f, l.list)
		l.list.itemMin = item.MinSize()
	}
	l.Layout(l.list.Size())
	l.scroller.Refresh()
	layout, _ := l.layout.Layout.(*listLayout)
	layout.updateList(false)

	for _, s := range layout.separators {
		s.Refresh()
	}
	canvas.Refresh(l.list.super())
}

// Declare conformity with interfaces.
var (
	_ granio.Widget     = (*listItem)(nil)
	_ granio.Tappable   = (*listItem)(nil)
	_ desktop.Hoverable = (*listItem)(nil)
)

type listItem struct {
	BaseWidget
	collection *List
	index      ListItemID

	onTapped          func()
	onHovered         func()
	background        *canvas.Rectangle
	child             granio.CanvasObject
	hovered, selected bool
	clickModifiers    granio.KeyModifier
}

func (li *listItem) MouseDown(event *desktop.MouseEvent) { li.clickModifiers = event.Modifier }
func (*listItem) MouseUp(*desktop.MouseEvent)            {}

func newListItem(child granio.CanvasObject, tapped func()) *listItem {
	li := &listItem{
		child:    child,
		onTapped: tapped,
	}

	li.ExtendBaseWidget(li)
	return li
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (li *listItem) CreateRenderer() granio.WidgetRenderer {
	li.ExtendBaseWidget(li)
	th := li.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	li.background = canvas.NewRectangle(th.Color(theme.ColorNameHover, v))
	li.background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
	li.background.Hide()

	objects := []granio.CanvasObject{li.background, li.child}

	return &listItemRenderer{widget.NewBaseRenderer(objects), li}
}

// MinSize returns the size that this widget should not shrink below.
func (li *listItem) MinSize() granio.Size {
	li.ExtendBaseWidget(li)
	return li.BaseWidget.MinSize()
}

// MouseIn is called when a desktop pointer enters the widget.
func (li *listItem) MouseIn(*desktop.MouseEvent) {
	if li.onHovered != nil {
		li.onHovered()
	}
	li.hovered = true
	li.Refresh()
}

// MouseMoved is called when a desktop pointer hovers over the widget.
func (*listItem) MouseMoved(*desktop.MouseEvent) {
}

// MouseOut is called when a desktop pointer exits the widget.
func (li *listItem) MouseOut() {
	li.hovered = false
	li.Refresh()
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler.
func (li *listItem) Tapped(*granio.PointEvent) {
	if li.onTapped != nil {
		li.selected = true
		li.Refresh()
		li.onTapped()
	}
}

// Declare conformity with the WidgetRenderer interface.
var _ granio.WidgetRenderer = (*listItemRenderer)(nil)

type listItemRenderer struct {
	widget.BaseRenderer

	item *listItem
}

// MinSize calculates the minimum size of a listItem.
// This is based on the size of the status indicator and the size of the child object.
func (li *listItemRenderer) MinSize() granio.Size {
	return li.item.child.MinSize()
}

// Layout the components of the listItem widget.
func (li *listItemRenderer) Layout(size granio.Size) {
	li.item.background.Resize(size)
	li.item.child.Resize(size)
}

func (li *listItemRenderer) Refresh() {
	th := li.item.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	li.item.background.CornerRadius = th.Size(theme.SizeNameSelectionRadius)
	li.item.background.StrokeWidth = 0
	if li.item.selected {
		applySelectionStyle(li.item.background, th, v)
		li.item.background.Show()
	} else if li.item.hovered {
		li.item.background.FillColor = th.Color(theme.ColorNameHover, v)
		li.item.background.Show()
	} else {
		li.item.background.Hide()
	}
	li.item.background.Refresh()
	canvas.Refresh(li.item.super())
}

// Declare conformity with Layout interface.
var _ granio.Layout = (*listLayout)(nil)

type listItemAndID struct {
	item *listItem
	id   ListItemID
}

type listLayout struct {
	list       *List
	separators []granio.CanvasObject
	children   []granio.CanvasObject

	itemPool          async.Pool[*listItem]
	visible           []listItemAndID
	wasVisible        []listItemAndID
	visibleRowHeights []float32
}

func newListLayout(list *List) granio.Layout {
	l := &listLayout{list: list}
	list.offsetUpdated = l.offsetUpdated
	return l
}

func (l *listLayout) Layout([]granio.CanvasObject, granio.Size) {
	l.updateList(true)
}

func (l *listLayout) MinSize([]granio.CanvasObject) granio.Size {
	return l.list.contentMinSize()
}

func (l *listLayout) getItem() *listItem {
	if item := l.itemPool.Get(); item != nil {
		return item
	}

	if f := l.list.CreateItem; f != nil {
		return newListItem(createItemAndApplyThemeScope(f, l.list), nil)
	}

	return nil
}

func (l *listLayout) offsetUpdated(pos granio.Position) {
	if l.list.offsetY == pos.Y {
		return
	}
	l.list.offsetY = pos.Y
	l.updateList(true)
}

func (l *listLayout) setupListItem(li *listItem, id ListItemID, focus bool) {
	li.collection, li.index = l.list, id
	previousIndicator := li.selected
	li.selected = false
	index := sort.SearchInts(l.list.selected, id)
	li.selected = index < len(l.list.selected) && l.list.selected[index] == id
	if focus {
		li.hovered = true
		li.Refresh()
	} else if previousIndicator != li.selected || li.hovered {
		li.hovered = false
		li.Refresh()
	}
	if f := l.list.UpdateItem; f != nil {
		f(id, li.child)
	}
	li.onHovered = func() {
		if f := l.list.OnHighlighted; f != nil {
			f(id)
		}
	}
	li.onTapped = func() {
		if !granio.CurrentDevice().IsMobile() {
			canvas := granio.CurrentApp().Driver().CanvasForObject(l.list.super())
			if canvas != nil {
				canvas.Focus(l.list.super().(granio.Focusable))
			}

			l.list.currentHighlight = id
		}

		if l.list.SelectionMode == SelectionMultiple {
			l.list.pointerSelect(id, li.clickModifiers)
		} else {
			l.list.Select(id)
		}
	}
}

func (l *listLayout) updateList(newOnly bool) {
	th := l.list.Theme()
	separatorThickness := th.Size(theme.SizeNamePadding)
	width := l.list.Size().Width
	length := 0
	if f := l.list.Length; f != nil {
		length = f()
	}
	if l.list.UpdateItem == nil {
		granio.LogError("Missing UpdateItem callback required for List", nil)
	}

	// l.wasVisible now represents the currently visible items, while
	// l.visible will be updated to represent what is visible *after* the update
	l.wasVisible = append(l.wasVisible, l.visible...)
	l.visible = l.visible[:0]

	offY, minRow := l.calculateVisibleRowHeights(l.list.itemMin.Height, length, th)
	if len(l.visibleRowHeights) == 0 && length > 0 { // we can't show anything until we have some dimensions
		return
	}

	oldChildrenLen := len(l.children)
	l.children = l.children[:0]

	y := offY
	for index, itemHeight := range l.visibleRowHeights {
		row := index + minRow
		size := granio.NewSize(width, itemHeight)

		c, ok := l.searchVisible(l.wasVisible, row)
		if !ok {
			c = l.getItem()
			if c == nil {
				continue
			}
			c.Resize(size)
		}

		c.Move(granio.NewPos(0, y))
		c.Resize(size)

		y += itemHeight + separatorThickness
		l.visible = append(l.visible, listItemAndID{id: row, item: c})
		l.children = append(l.children, c)
	}
	l.nilOldSliceData(l.children, len(l.children), oldChildrenLen)

	for _, wasVis := range l.wasVisible {
		if _, ok := l.searchVisible(l.visible, wasVis.id); !ok {
			l.itemPool.Put(wasVis.item)
		}
	}

	l.updateSeparators()

	c, _ := l.list.scroller.Content.(*granio.Container)
	oldObjLen := len(c.Objects)
	c.Objects = c.Objects[:0]
	c.Objects = append(c.Objects, l.children...)
	c.Objects = append(c.Objects, l.separators...)
	l.nilOldSliceData(c.Objects, len(c.Objects), oldObjLen)

	if newOnly {
		for _, vis := range l.visible {
			if _, ok := l.searchVisible(l.wasVisible, vis.id); !ok {
				l.setupListItem(vis.item, vis.id, l.list.focused && l.list.currentHighlight == vis.id)
			}
		}
	} else {
		for _, vis := range l.visible {
			l.setupListItem(vis.item, vis.id, l.list.focused && l.list.currentHighlight == vis.id)
		}

		// a full refresh may change theme, we should drain the pool of unused items instead of refreshing them.
		for l.itemPool.Get() != nil { //revive:disable-line:empty-block
		}
	}

	// we don't need wasVisible now until next call to update; clear and reset its length
	clear(l.wasVisible)
	l.wasVisible = l.wasVisible[:0]
}

func (l *listLayout) updateSeparators() {
	if l.list.HideSeparators {
		l.separators = nil
		return
	}
	if lenChildren := len(l.children); lenChildren > 1 {
		if lenSep := len(l.separators); lenSep > lenChildren {
			l.separators = l.separators[:lenChildren]
		} else {
			for i := lenSep; i < lenChildren; i++ {
				sep := NewSeparator()
				if cache.OverrideThemeMatchingScope(sep, l.list) {
					sep.Refresh()
				}
				l.separators = append(l.separators, sep)
			}
		}
	} else {
		l.separators = nil
	}

	th := l.list.Theme()
	separatorThickness := th.Size(theme.SizeNameSeparatorThickness)
	dividerOff := (th.Size(theme.SizeNamePadding) + separatorThickness) / 2
	for i, child := range l.children {
		if i == 0 {
			continue
		}
		l.separators[i].Move(granio.NewPos(0, child.Position().Y-dividerOff))
		l.separators[i].Resize(granio.NewSize(l.list.Size().Width, separatorThickness))
		l.separators[i].Show()
	}
}

// invariant: visible is in ascending order of IDs
func (*listLayout) searchVisible(visible []listItemAndID, id ListItemID) (*listItem, bool) {
	ln := len(visible)
	idx := sort.Search(ln, func(i int) bool { return visible[i].id >= id })
	if idx < ln && visible[idx].id == id {
		return visible[idx].item, true
	}
	return nil, false
}

func (*listLayout) nilOldSliceData(objs []granio.CanvasObject, length, oldLength int) {
	if oldLength > length {
		objs = objs[:oldLength] // gain view into old data
		clear(objs[length:])
	}
}

func createItemAndApplyThemeScope(f func() granio.CanvasObject, scope granio.Widget) granio.CanvasObject {
	item := f()
	if !cache.OverrideThemeMatchingScope(item, scope) {
		return item
	}

	item.Refresh()
	return item
}
