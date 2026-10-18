package mobile

import (
	"context"
	"image"
	"math"
	"slices"
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/mobile"
	"github.com/Uzer007-admin/granio/v2/internal/app"
	intdriver "github.com/Uzer007-admin/granio/v2/internal/driver"
	"github.com/Uzer007-admin/granio/v2/internal/driver/common"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var _ granio.Canvas = (*canvas)(nil)

type canvas struct {
	common.Canvas
	content        granio.CanvasObject
	device         *device
	initialized    bool
	lastTapDown    map[int]time.Time
	lastTapDownPos map[int]granio.Position
	lastTapDelta   map[int]granio.Delta
	menu           granio.CanvasObject
	padded         bool
	scale          float32
	size           granio.Size
	touched        map[int]mobile.Touchable
	windowHead     *granio.Container

	dragOffset     granio.Position
	dragStart      granio.Position
	dragging       granio.Draggable
	draggingOuter  granio.Draggable
	otherDirection container.ScrollDirection

	onTypedKey  func(event *granio.KeyEvent)
	onTypedRune func(rune)

	touchCancelFunc context.CancelFunc
	touchCancelLock sync.Mutex
	touchLastTapped granio.CanvasObject
	touchTapCount   int
}

func newCanvas(dev granio.Device) granio.Canvas {
	d, _ := dev.(*device)
	ret := &canvas{
		Canvas: common.Canvas{
			OnFocus:   d.handleKeyboard,
			OnUnfocus: d.HideVirtualKeyboard,
		},
		device:         d,
		lastTapDown:    make(map[int]time.Time),
		lastTapDownPos: make(map[int]granio.Position),
		lastTapDelta:   make(map[int]granio.Delta),
		padded:         true,
		scale:          dev.SystemScaleForWindow(nil), // we don't need a window parameter on mobile,
		touched:        make(map[int]mobile.Touchable),
	}
	ret.Initialize(ret, ret.overlayChanged)
	return ret
}

func (c *canvas) Capture() image.Image {
	return c.Painter().Capture(c)
}

func (c *canvas) Content() granio.CanvasObject {
	return c.content
}

func (c *canvas) InteractiveArea() (granio.Position, granio.Size) {
	var pos granio.Position
	var size granio.Size
	if c.device == nil {
		// running in test mode
		size = c.Size()
	} else {
		safeLeft := float32(c.device.safeLeft) / c.scale
		safeTop := float32(c.device.safeTop) / c.scale
		safeRight := float32(c.device.safeRight) / c.scale
		safeBottom := float32(c.device.safeBottom) / c.scale
		pos = granio.NewPos(safeLeft, safeTop)
		size = c.size.SubtractWidthHeight(safeLeft+safeRight, safeTop+safeBottom)
	}
	if c.windowHeadIsDisplacing() {
		offset := c.windowHead.MinSize().Height
		pos = pos.AddXY(0, offset)
		size = size.SubtractWidthHeight(0, offset)
	}
	return pos, size
}

func (c *canvas) MinSize() granio.Size {
	return c.size // TODO check
}

func (c *canvas) OnTypedKey() func(*granio.KeyEvent) {
	return c.onTypedKey
}

func (c *canvas) OnTypedRune() func(rune) {
	return c.onTypedRune
}

func (c *canvas) PixelCoordinateForPosition(pos granio.Position) (x, y int) {
	return int(pos.X * c.scale), int(pos.Y * c.scale)
}

func (c *canvas) Resize(size granio.Size) {
	if size == c.size {
		return
	}

	c.sizeContent(size)
}

func (c *canvas) Scale() float32 {
	return c.scale
}

func (c *canvas) SetContent(content granio.CanvasObject) {
	c.applyContent(content)
	c.sizeContent(c.Size()) // fixed window size for mobile, cannot stretch to new content
	c.SetDirty()
}

func (c *canvas) SetOnTypedKey(typed func(*granio.KeyEvent)) {
	c.onTypedKey = typed
}

func (c *canvas) SetOnTypedRune(typed func(rune)) {
	c.onTypedRune = typed
}

func (c *canvas) Size() granio.Size {
	return c.size
}

func (c *canvas) applyContent(content granio.CanvasObject) {
	c.content = content
	c.SetContentTreeAndFocusMgr(content)
}

func (c *canvas) applyThemeOutOfTreeObjects() {
	if c.menu != nil {
		app.ApplyThemeTo(c.menu, c) // Ensure our menu gets the theme change message as it's out-of-tree
	}
	if c.windowHead != nil {
		app.ApplyThemeTo(c.windowHead, c) // Ensure our child windows get the theme change message as it's out-of-tree
	}
}

func (c *canvas) findObjectAtPositionMatching(pos granio.Position, test func(object granio.CanvasObject) bool) (granio.CanvasObject, granio.Position, int) {
	if c.menu != nil {
		return intdriver.FindObjectAtPositionMatching(pos, test, c.Overlays().Top(), c.menu)
	}

	var wh granio.CanvasObject
	if c.windowHead != nil {
		wh = c.windowHead
	}
	return intdriver.FindObjectAtPositionMatching(pos, test, c.Overlays().Top(), wh, c.content)
}

func (c *canvas) overlayChanged() {
	c.device.handleKeyboard(c.Focused())
	c.SetDirty()
}

func (c *canvas) setMenu(menu granio.CanvasObject) {
	c.menu = menu
	c.SetMenuTreeAndFocusMgr(menu)
}

func (c *canvas) setWindowHead(head *granio.Container) {
	if c.padded {
		head = container.NewPadded(head)
	}
	c.windowHead = head
	c.SetMobileWindowHeadTree(head)
}

func (c *canvas) sizeContent(size granio.Size) {
	if c.content == nil { // window may not be configured yet
		return
	}

	c.size = size
	areaPos, areaSize := c.InteractiveArea()

	if c.windowHead != nil {
		var headSize granio.Size
		headPos := areaPos
		if c.windowHeadIsDisplacing() {
			headSize = granio.NewSize(areaSize.Width, c.windowHead.MinSize().Height)
			headPos = headPos.SubtractXY(0, headSize.Height)
		} else {
			headSize = c.windowHead.MinSize()
		}
		c.windowHead.Resize(headSize)
		c.windowHead.Move(headPos)
	}

	for _, overlay := range c.Overlays().List() {
		overlay.Resize(areaSize)
		overlay.Move(areaPos)
	}

	if c.padded {
		c.content.Resize(areaSize.Subtract(granio.NewSize(theme.Padding()*2, theme.Padding()*2)))
		c.content.Move(areaPos.Add(granio.NewPos(theme.Padding(), theme.Padding())))
	} else {
		c.content.Resize(areaSize)
		c.content.Move(areaPos)
	}
}

func (c *canvas) tapDown(pos granio.Position, tapID int) {
	c.lastTapDown[tapID] = time.Now()
	c.lastTapDownPos[tapID] = pos
	c.dragging = nil
	c.draggingOuter = nil

	co, objPos, layer := c.findObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
		switch object.(type) {
		case mobile.Touchable, granio.Focusable:
			return true
		}

		return false
	})

	if wid, ok := co.(mobile.Touchable); ok {
		touchEv := &mobile.TouchEvent{}
		touchEv.ID = tapID
		touchEv.Position = objPos
		touchEv.AbsolutePosition = pos
		wid.TouchDown(touchEv)
		c.touched[tapID] = wid
	}

	if layer != 1 { // 0 - overlay, 1 - window head / menu, 2 - content
		if wid, ok := co.(granio.Focusable); !ok || wid != c.Focused() {
			c.Unfocus()
		}
	}
}

func (c *canvas) tapMove(pos granio.Position, tapID int,
	dragCallback func(granio.Draggable, *granio.DragEvent),
) {
	previousPos := c.lastTapDownPos[tapID]
	deltaX := pos.X - previousPos.X
	deltaY := pos.Y - previousPos.Y

	if c.dragging == nil && (math.Abs(float64(deltaX)) < tapMoveThreshold && math.Abs(float64(deltaY)) < tapMoveThreshold) {
		return
	}
	c.lastTapDownPos[tapID] = pos
	offset := granio.Delta{DX: deltaX, DY: deltaY}
	c.lastTapDelta[tapID] = offset

	co, objPos, _ := c.findObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
		if _, ok := object.(granio.Draggable); ok {
			return true
		} else if _, ok := object.(mobile.Touchable); ok {
			return true
		}

		return false
	})
	var scrollOtherDirection granio.CanvasObject
	if scr, ok := co.(*container.Scroll); ok {
		switch scr.Direction {
		case container.ScrollHorizontalOnly:
			c.otherDirection = container.ScrollVerticalOnly
		case container.ScrollVerticalOnly:
			c.otherDirection = container.ScrollHorizontalOnly
		}
		if c.otherDirection != container.ScrollBoth {
			scrollOtherDirection, _, _ = c.findObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
				if scr, ok := object.(*container.Scroll); ok {
					return scr.Direction == c.otherDirection || scr.Direction == container.ScrollBoth
				}

				return false
			})
		}
	}

	if c.touched[tapID] != nil {
		if touch, ok := co.(mobile.Touchable); !ok || c.touched[tapID] != touch {
			touchEv := &mobile.TouchEvent{}
			touchEv.ID = tapID
			touchEv.Position = objPos
			touchEv.AbsolutePosition = pos
			c.touched[tapID].TouchCancel(touchEv)
			c.touched[tapID] = nil
		}
	}

	if c.dragging == nil {
		drag, ok := co.(granio.Draggable)
		if !ok {
			return
		}

		c.dragging = drag
		c.dragOffset = previousPos.Subtract(objPos)
		c.dragStart = co.Position()
		if scrollOtherDirection != nil {
			c.draggingOuter, _ = scrollOtherDirection.(granio.Draggable)
		}
	}

	ev := &granio.DragEvent{}
	draggedObjDelta := c.dragStart.Subtract(c.dragging.(granio.CanvasObject).Position())
	ev.Position = pos.Subtract(c.dragOffset).Add(draggedObjDelta)
	ev.Dragged = offset

	dragCallback(c.dragging, ev)
	if c.draggingOuter != nil {
		if c.otherDirection == container.ScrollVerticalOnly {
			ev.Dragged.DX = 0
		} else {
			ev.Dragged.DY = 0
		}
		dragCallback(c.draggingOuter, ev)
	}
}

func (c *canvas) tapUp(pos granio.Position, tapID int,
	tapCallback func(granio.Tappable, *granio.PointEvent),
	tapAltCallback func(granio.SecondaryTappable, *granio.PointEvent),
	doubleTapCallback func(granio.DoubleTappable, *granio.PointEvent),
	dragCallback func(granio.Draggable, *granio.DragEvent),
) {
	if c.dragging != nil {
		previousDelta := c.lastTapDelta[tapID]
		ev := &granio.DragEvent{Dragged: previousDelta}
		draggedObjDelta := c.dragStart.Subtract(c.dragging.(granio.CanvasObject).Position())
		ev.Position = pos.Subtract(c.dragOffset).Add(draggedObjDelta)
		ev.AbsolutePosition = pos
		dragCallback(c.dragging, ev)
		if c.draggingOuter != nil {
			if c.otherDirection == container.ScrollVerticalOnly {
				ev.Dragged.DX = 0
			} else {
				ev.Dragged.DY = 0
			}
			dragCallback(c.draggingOuter, ev)
		}

		c.dragging = nil
		c.draggingOuter = nil
		c.otherDirection = container.ScrollBoth
		return
	}

	duration := time.Since(c.lastTapDown[tapID])

	if c.menu != nil && c.Overlays().Top() == nil && pos.X > c.menu.Size().Width {
		c.menu.Hide()
		c.menu.Refresh()
		c.setMenu(nil)
		return
	}

	co, objPos, _ := c.findObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
		if _, ok := object.(granio.Tappable); ok {
			return true
		} else if _, ok := object.(granio.SecondaryTappable); ok {
			return true
		} else if _, ok := object.(mobile.Touchable); ok {
			return true
		} else if _, ok := object.(granio.DoubleTappable); ok {
			return true
		}

		return false
	})

	if wid, ok := co.(mobile.Touchable); ok {
		touchEv := &mobile.TouchEvent{}
		touchEv.ID = tapID
		touchEv.Position = objPos
		touchEv.AbsolutePosition = pos
		wid.TouchUp(touchEv)
		c.touched[tapID] = nil
	}

	ev := &granio.PointEvent{
		Position:         objPos,
		AbsolutePosition: pos,
	}

	if duration < tapSecondaryDelay {
		_, doubleTap := co.(granio.DoubleTappable)
		if doubleTap {
			c.touchCancelLock.Lock()
			c.touchTapCount++
			c.touchLastTapped = co
			cancel := c.touchCancelFunc
			c.touchCancelLock.Unlock()
			if cancel != nil {
				cancel()
				return
			}
			go c.waitForDoubleTap(co, ev, tapCallback, doubleTapCallback)
		} else {
			if wid, ok := co.(granio.Tappable); ok {
				tapCallback(wid, ev)
			}
		}
	} else {
		if wid, ok := co.(granio.SecondaryTappable); ok {
			prevOverlay := c.Overlays().Top()
			tapAltCallback(wid, ev)

			// if the secondary tap dismissed an overlay (rather than opening a new
			// one on top), forward the event to the widget underneath
			if prevOverlay != nil && !slices.Contains(c.Overlays().List(), prevOverlay) {
				co2, objPos2, _ := c.findObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
					_, ok := object.(granio.SecondaryTappable)
					return ok
				})
				if sec2, ok := co2.(granio.SecondaryTappable); ok {
					ev2 := &granio.PointEvent{Position: objPos2, AbsolutePosition: pos}
					tapAltCallback(sec2, ev2)
				}
			}
		}
	}
}

func (c *canvas) waitForDoubleTap(co granio.CanvasObject, ev *granio.PointEvent, tapCallback func(granio.Tappable, *granio.PointEvent), doubleTapCallback func(granio.DoubleTappable, *granio.PointEvent)) {
	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(tapDoubleDelay))
	c.touchCancelLock.Lock()
	c.touchCancelFunc = cancel
	c.touchCancelLock.Unlock()
	defer cancel()

	<-ctx.Done()
	granio.CurrentApp().Driver().DoFromGoroutine(func() {
		c.touchCancelLock.Lock()
		touchCount := c.touchTapCount
		touchLast := c.touchLastTapped
		c.touchCancelLock.Unlock()

		if touchCount == 2 && touchLast == co {
			if wid, ok := co.(granio.DoubleTappable); ok {
				doubleTapCallback(wid, ev)
			}
		} else {
			if wid, ok := co.(granio.Tappable); ok {
				tapCallback(wid, ev)
			}
		}

		c.touchCancelLock.Lock()
		c.touchTapCount = 0
		c.touchCancelFunc = nil
		c.touchLastTapped = nil
		c.touchCancelLock.Unlock()
	}, true)
}

func (c *canvas) windowHeadIsDisplacing() bool {
	if c.windowHead == nil {
		return false
	}

	chromeBox := c.windowHead
	if c.padded {
		chromeBox, _ = chromeBox.Objects[0].(*granio.Container) // the padded container
	}
	return len(chromeBox.Objects) > 1
}
