package embedded

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/mobile"
	"github.com/Uzer007-admin/granio/v2/driver/software"
	"github.com/Uzer007-admin/granio/v2/internal/driver"
)

const (
	tapMoveDecay        = 0.92                   // how much should the scroll continue decay on each frame?
	tapMoveEndThreshold = 2.0                    // at what offset will we stop decaying?
	tapMoveThreshold    = 4.0                    // how far can we move before it is a drag
	tapSecondaryDelay   = 300 * time.Millisecond // how long before secondary tap
	tapDoubleDelay      = 500 * time.Millisecond // max duration between taps for a DoubleTap event
)

type touchCanvas struct {
	software.WindowlessCanvas

	lastTapDown    map[int]time.Time
	lastTapDownPos map[int]granio.Position
	lastTapDelta   map[int]granio.Delta

	dragOffset granio.Position
	dragStart  granio.Position
	dragging   granio.Draggable

	touched         map[int]mobile.Touchable
	touchCancelFunc context.CancelFunc
	touchCancelLock sync.Mutex
	touchLastTapped granio.CanvasObject
	touchTapCount   int
}

func newTouchCanvas() *touchCanvas {
	ret := &touchCanvas{
		WindowlessCanvas: software.NewCanvas(),
		lastTapDown:      make(map[int]time.Time),
		lastTapDownPos:   make(map[int]granio.Position),
		lastTapDelta:     make(map[int]granio.Delta),
		touched:          make(map[int]mobile.Touchable),
	}
	return ret
}

func (c *touchCanvas) tapDown(pos granio.Position, tapID int) {
	c.lastTapDown[tapID] = time.Now()
	c.lastTapDownPos[tapID] = pos
	c.dragging = nil

	co, objPos, layer := driver.FindObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
		switch object.(type) {
		case mobile.Touchable, granio.Focusable:
			return true
		}

		return false
	}, nil, nil, c.Content())

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

func (c *touchCanvas) tapMove(pos granio.Position, tapID int,
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

	co, objPos, _ := driver.FindObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
		if _, ok := object.(granio.Draggable); ok {
			return true
		} else if _, ok := object.(mobile.Touchable); ok {
			return true
		}

		return false
	}, nil, nil, c.Content())

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
	}

	ev := &granio.DragEvent{}
	draggedObjDelta := c.dragStart.Subtract(c.dragging.(granio.CanvasObject).Position())
	ev.Position = pos.Subtract(c.dragOffset).Add(draggedObjDelta)
	ev.Dragged = offset

	dragCallback(c.dragging, ev)
}

func (c *touchCanvas) tapUp(pos granio.Position, tapID int,
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

		c.dragging = nil
		return
	}

	duration := time.Since(c.lastTapDown[tapID])

	co, objPos, _ := driver.FindObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
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
	}, nil, nil, c.Content())

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
			tapAltCallback(wid, ev)
		}
	}
}

func (c *touchCanvas) waitForDoubleTap(co granio.CanvasObject, ev *granio.PointEvent, tapCallback func(granio.Tappable, *granio.PointEvent), doubleTapCallback func(granio.DoubleTappable, *granio.PointEvent)) {
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
