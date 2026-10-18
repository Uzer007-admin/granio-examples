package widget

import (
	"fmt"
	"math"
	"sync"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// Declare conformity with Widget interface.
var _ granio.Widget = (*GridWrap)(nil)

// GridWrapItemID is the ID of an individual item in the GridWrap widget.
type GridWrapItemID = int

// GridWrap is a widget with an API very similar to widget.List,
// that lays out items in a scrollable wrapping grid similar to container.NewGridWrap.
// It caches and reuses widgets for performance.
//
// Deprecated: Since Fyne 2.4, GridWrap is available in Fyne core
type GridWrap struct {
	widget.BaseWidget

	Length     func() int                                        `json:"-"`
	CreateItem func() granio.CanvasObject                        `json:"-"`
	UpdateItem func(id GridWrapItemID, item granio.CanvasObject) `json:"-"`

	scroller      *container.Scroll
	itemMin       granio.Size
	offsetY       float32
	offsetUpdated func(granio.Position)
}

// NewGridWrap creates and returns a GridWrap widget for displaying items in
// a wrapping grid layout with scrolling and caching for performance.
//
// Deprecated: Since Fyne 2.4, GridWrap is available in Fyne core
func NewGridWrap(length func() int, createItem func() granio.CanvasObject, updateItem func(GridWrapItemID, granio.CanvasObject)) *GridWrap {
	gwList := &GridWrap{Length: length, CreateItem: createItem, UpdateItem: updateItem}
	gwList.ExtendBaseWidget(gwList)
	return gwList
}

// NewGridWrapWithData creates a new GridWrap widget that will display the contents of the provided data.
//
// Deprecated: Since Fyne 2.4, GridWrap is available in Fyne core
func NewGridWrapWithData(data binding.DataList, createItem func() granio.CanvasObject, updateItem func(binding.DataItem, granio.CanvasObject)) *GridWrap {
	gwList := NewGridWrap(
		data.Length,
		createItem,
		func(i GridWrapItemID, o granio.CanvasObject) {
			item, err := data.GetItem(int(i))
			if err != nil {
				granio.LogError(fmt.Sprintf("Error getting data item %d", i), err)
				return
			}
			updateItem(item, o)
		})

	data.AddListener(binding.NewDataListener(gwList.Refresh))
	return gwList
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer.
func (l *GridWrap) CreateRenderer() granio.WidgetRenderer {
	l.ExtendBaseWidget(l)

	if f := l.CreateItem; f != nil && l.itemMin.IsZero() {
		l.itemMin = f().MinSize()
	}

	layout := &granio.Container{Layout: newGridWrapLayout(l)}
	l.scroller = container.NewVScroll(layout)
	layout.Resize(layout.MinSize())

	return newGridWrapRenderer([]granio.CanvasObject{l.scroller}, l, l.scroller, layout)
}

// MinSize returns the size that this widget should not shrink below.
func (l *GridWrap) MinSize() granio.Size {
	l.ExtendBaseWidget(l)

	return l.BaseWidget.MinSize()
}

func (l *GridWrap) scrollTo(id GridWrapItemID) {
	if l.scroller == nil {
		return
	}
	row := math.Floor(float64(id) / float64(l.getColCount()))
	y := float32(row)*l.itemMin.Height + float32(row)*theme.Padding()
	if y < l.scroller.Offset.Y {
		l.scroller.Offset.Y = y
	} else if y+l.itemMin.Height > l.scroller.Offset.Y+l.scroller.Size().Height {
		l.scroller.Offset.Y = y + l.itemMin.Height - l.scroller.Size().Height
	}
	l.offsetUpdated(l.scroller.Offset)
}

// Resize is called when this GridWrap should change size. We refresh to ensure invisible items are drawn.
func (l *GridWrap) Resize(s granio.Size) {
	l.BaseWidget.Resize(s)
	l.offsetUpdated(l.scroller.Offset)
	l.scroller.Content.(*granio.Container).Layout.(*gridWrapLayout).updateList(true)
}

// ScrollTo scrolls to the item represented by id
func (l *GridWrap) ScrollTo(id GridWrapItemID) {
	length := 0
	if f := l.Length; f != nil {
		length = f()
	}
	if id < 0 || int(id) >= length {
		return
	}
	l.scrollTo(id)
	l.Refresh()
}

// ScrollToBottom scrolls to the end of the list
func (l *GridWrap) ScrollToBottom() {
	length := 0
	if f := l.Length; f != nil {
		length = f()
	}
	if length > 0 {
		length--
	}
	l.scrollTo(GridWrapItemID(length))
	l.Refresh()
}

// ScrollToTop scrolls to the start of the list
func (l *GridWrap) ScrollToTop() {
	l.scrollTo(0)
	l.Refresh()
}

// ScrollToOffset scrolls the list to the given offset position
func (l *GridWrap) ScrollToOffset(offset float32) {
	l.scroller.Offset.Y = offset
	l.offsetUpdated(l.scroller.Offset)
}

// GetScrollOffset returns the current scroll offset position
func (l *GridWrap) GetScrollOffset() float32 {
	return l.offsetY
}

// Declare conformity with WidgetRenderer interface.
var _ granio.WidgetRenderer = (*gridWrapRenderer)(nil)

type gridWrapRenderer struct {
	objects []granio.CanvasObject

	list     *GridWrap
	scroller *container.Scroll
	layout   *granio.Container
}

func newGridWrapRenderer(objects []granio.CanvasObject, l *GridWrap, scroller *container.Scroll, layout *granio.Container) *gridWrapRenderer {
	lr := &gridWrapRenderer{objects: objects, list: l, scroller: scroller, layout: layout}
	lr.scroller.OnScrolled = l.offsetUpdated
	return lr
}

func (l *gridWrapRenderer) Layout(size granio.Size) {
	l.scroller.Resize(size)
}

func (l *gridWrapRenderer) MinSize() granio.Size {
	return l.scroller.MinSize().Max(l.list.itemMin)
}

func (l *gridWrapRenderer) Refresh() {
	if f := l.list.CreateItem; f != nil {
		l.list.itemMin = f().MinSize()
	}
	l.Layout(l.list.Size())
	l.scroller.Refresh()
	l.layout.Layout.(*gridWrapLayout).updateList(true)
	canvas.Refresh(l.list)
}

func (l *gridWrapRenderer) Destroy() {
}

func (l *gridWrapRenderer) Objects() []granio.CanvasObject {
	return l.objects
}

// Declare conformity with Layout interface.
var _ granio.Layout = (*gridWrapLayout)(nil)

type gridWrapLayout struct {
	list     *GridWrap
	children []granio.CanvasObject

	itemPool   *syncPool
	visible    map[GridWrapItemID]granio.CanvasObject
	renderLock sync.Mutex
}

func newGridWrapLayout(list *GridWrap) granio.Layout {
	l := &gridWrapLayout{list: list, itemPool: &syncPool{}, visible: make(map[GridWrapItemID]granio.CanvasObject)}
	list.offsetUpdated = l.offsetUpdated
	return l
}

func (l *gridWrapLayout) Layout(_ []granio.CanvasObject, _ granio.Size) {
	l.updateList(true)
}

func (l *gridWrapLayout) MinSize(_ []granio.CanvasObject) granio.Size {
	if lenF := l.list.Length; lenF != nil {
		cols := l.list.getColCount()
		rows := float32(math.Ceil(float64(lenF()) / float64(cols)))
		return granio.NewSize(l.list.itemMin.Width,
			(l.list.itemMin.Height+theme.Padding())*rows-theme.Padding())
	}
	return granio.NewSize(0, 0)
}

func (l *gridWrapLayout) getItem() granio.CanvasObject {
	item := l.itemPool.Obtain()
	if item == nil {
		if f := l.list.CreateItem; f != nil {
			item = f()
		}
	}
	return item
}

func (l *gridWrapLayout) offsetUpdated(pos granio.Position) {
	if l.list.offsetY == pos.Y {
		return
	}
	l.list.offsetY = pos.Y
	l.updateList(false)
}

func (l *gridWrapLayout) setupListItem(li granio.CanvasObject, id GridWrapItemID) {
	if f := l.list.UpdateItem; f != nil {
		f(id, li)
	}
}

func (l *GridWrap) getColCount() int {
	colCount := 1
	width := l.Size().Width
	if width > l.itemMin.Width {
		colCount = int(math.Floor(float64(width+theme.Padding()) / float64(l.itemMin.Width+theme.Padding())))
	}
	return colCount
}

func (l *gridWrapLayout) updateList(refresh bool) {
	// code here is a mashup of listLayout.updateList and gridWrapLayout.Layout

	l.renderLock.Lock()
	defer l.renderLock.Unlock()
	length := 0
	if f := l.list.Length; f != nil {
		length = f()
	}

	colCount := l.list.getColCount()
	visibleRowsCount := int(math.Ceil(float64(l.list.scroller.Size().Height)/float64(l.list.itemMin.Height+theme.Padding()))) + 1

	offY := l.list.offsetY - float32(math.Mod(float64(l.list.offsetY), float64(l.list.itemMin.Height+theme.Padding())))
	minRow := int(offY / (l.list.itemMin.Height + theme.Padding()))
	minItem := GridWrapItemID(minRow * colCount)
	maxRow := int(math.Min(float64(minRow+visibleRowsCount), math.Ceil(float64(length)/float64(colCount))))
	maxItem := GridWrapItemID(math.Min(float64(maxRow*colCount), float64(length-1)))

	if l.list.UpdateItem == nil {
		granio.LogError("Missing UpdateCell callback required for GridWrap", nil)
	}

	wasVisible := l.visible
	l.visible = make(map[GridWrapItemID]granio.CanvasObject)
	var cells []granio.CanvasObject
	y := offY
	curItem := minItem
	for row := minRow; row <= maxRow && curItem <= maxItem; row++ {
		x := float32(0)
		for col := 0; col < colCount && curItem <= maxItem; col++ {
			c, ok := wasVisible[curItem]
			if !ok {
				c = l.getItem()
				if c == nil {
					continue
				}
				c.Resize(l.list.itemMin)
				l.setupListItem(c, curItem)
			}

			c.Move(granio.NewPos(x, y))
			if refresh {
				c.Resize(l.list.itemMin)
				if ok { // refresh visible
					l.setupListItem(c, curItem)
				}
			}

			x += l.list.itemMin.Width + theme.Padding()
			l.visible[curItem] = c
			cells = append(cells, c)
			curItem++
		}
		y += l.list.itemMin.Height + theme.Padding()
	}

	for id, old := range wasVisible {
		if _, ok := l.visible[id]; !ok {
			l.itemPool.Release(old)
		}
	}
	l.children = cells

	objects := l.children
	l.list.scroller.Content.(*granio.Container).Objects = objects
}

type pool interface {
	Obtain() granio.CanvasObject
	Release(granio.CanvasObject)
}

var _ pool = (*syncPool)(nil)

type syncPool struct {
	sync.Pool
}

// Obtain returns an item from the pool for use
func (p *syncPool) Obtain() (item granio.CanvasObject) {
	o := p.Get()
	if o != nil {
		item = o.(granio.CanvasObject)
	}
	return
}

// Release adds an item into the pool to be used later
func (p *syncPool) Release(item granio.CanvasObject) {
	p.Put(item)
}
