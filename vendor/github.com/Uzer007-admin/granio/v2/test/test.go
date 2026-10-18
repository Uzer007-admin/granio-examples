package test

import (
	"os"
	"path/filepath"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	intdriver "github.com/Uzer007-admin/granio/v2/internal/driver"
)

// DoubleTap simulates a double left mouse click on the specified object.
func DoubleTap(obj granio.DoubleTappable) {
	ev, c := prepareTap(obj, granio.NewPos(1, 1))
	handleFocusOnTap(c, obj)
	obj.DoubleTapped(ev)
}

// Drag drags at an absolute position on the canvas.
// deltaX/Y is the dragging distance: <0 for dragging up/left, >0 for dragging down/right.
func Drag(c granio.Canvas, pos granio.Position, deltaX, deltaY float32) {
	matches := func(object granio.CanvasObject) bool {
		_, ok := object.(granio.Draggable)
		return ok
	}
	o, p, _ := intdriver.FindObjectAtPositionMatching(pos, matches, c.Overlays().Top(), c.Content())
	if o == nil {
		return
	}
	e := &granio.DragEvent{
		PointEvent: granio.PointEvent{Position: p},
		Dragged:    granio.Delta{DX: deltaX, DY: deltaY},
	}
	o.(granio.Draggable).Dragged(e)
	o.(granio.Draggable).DragEnd()
}

// FocusNext focuses the next focusable on the canvas.
//
// Deprecated: Use granio.Canvas#FocusNext() instead.
func FocusNext(c granio.Canvas) {
	c.FocusNext()
}

// FocusPrevious focuses the previous focusable on the canvas.
//
// Deprecated: Use granio.Canvas#FocusPrevious() instead.
func FocusPrevious(c granio.Canvas) {
	c.FocusPrevious()
}

// LaidOutObjects returns all granio.CanvasObject starting at the given granio.CanvasObject which is laid out previously.
func LaidOutObjects(o granio.CanvasObject) (objects []granio.CanvasObject) {
	if o != nil {
		objects = layoutAndCollect(objects, o, internal.MaxSizes(o.MinSize(), o.Size()))
	}
	return objects
}

// MoveMouse simulates a mouse movement to the given position.
func MoveMouse(c granio.Canvas, pos granio.Position) {
	if granio.CurrentDevice().IsMobile() {
		return
	}

	tc, _ := c.(*canvas)
	var oldHovered, hovered desktop.Hoverable
	if tc != nil {
		oldHovered = tc.hovered
	}
	matches := func(object granio.CanvasObject) bool {
		_, ok := object.(desktop.Hoverable)
		return ok
	}
	o, p, _ := intdriver.FindObjectAtPositionMatching(pos, matches, c.Overlays().Top(), c.Content())
	if o != nil {
		hovered, _ = o.(desktop.Hoverable)
		me := &desktop.MouseEvent{
			PointEvent: granio.PointEvent{
				AbsolutePosition: pos,
				Position:         p,
			},
		}
		if hovered == oldHovered {
			hovered.MouseMoved(me)
		} else {
			if oldHovered != nil {
				oldHovered.MouseOut()
			}
			hovered.MouseIn(me)
		}
	} else if oldHovered != nil {
		oldHovered.MouseOut()
	}
	if tc != nil {
		tc.hovered = hovered
	}
}

// RenderObjectToMarkup renders the given [github.com/Uzer007-admin/granio/v2.CanvasObject] to a markup string.
//
// Since: 2.6
func RenderObjectToMarkup(o granio.CanvasObject) string {
	c := NewCanvas()
	c.SetPadded(false)
	size := internal.MaxSizes(o.MinSize(), o.Size())
	c.SetContent(o)
	c.Resize(size) // ensure we are large enough for current size

	return snapshot(c)
}

// RenderToMarkup renders the given [github.com/Uzer007-admin/granio/v2.Canvas] to a markup string.
//
// Since: 2.6
func RenderToMarkup(c granio.Canvas) string {
	return snapshot(c)
}

// Scroll scrolls at an absolute position on the canvas.
// deltaX/Y is the scrolling distance: <0 for scrolling up/left, >0 for scrolling down/right.
func Scroll(c granio.Canvas, pos granio.Position, deltaX, deltaY float32) {
	matches := func(object granio.CanvasObject) bool {
		_, ok := object.(granio.Scrollable)
		return ok
	}
	o, _, _ := intdriver.FindObjectAtPositionMatching(pos, matches, c.Overlays().Top(), c.Content())
	if o == nil {
		return
	}

	e := &granio.ScrollEvent{Scrolled: granio.Delta{DX: deltaX, DY: deltaY}}
	o.(granio.Scrollable).Scrolled(e)
}

// Tap simulates a left mouse click on the specified object.
func Tap(obj granio.Tappable) {
	TapAt(obj, granio.NewPos(1, 1))
}

// TapAt simulates a left mouse click on the passed object at a specified place within it.
func TapAt(obj granio.Tappable, pos granio.Position) {
	ev, c := prepareTap(obj, pos)
	performTap(c, obj, ev)
}

// TapCanvas taps at an absolute position on the canvas.
func TapCanvas(c granio.Canvas, pos granio.Position) {
	if o, p := findTappable(c, pos); o != nil {
		performTap(c, o.(granio.Tappable), &granio.PointEvent{AbsolutePosition: pos, Position: p})
	}
}

// TapSecondary simulates a right mouse click on the specified object.
func TapSecondary(obj granio.SecondaryTappable) {
	TapSecondaryAt(obj, granio.NewPos(1, 1))
}

// TapSecondaryAt simulates a right mouse click on the passed object at a specified place within it.
func TapSecondaryAt(obj granio.SecondaryTappable, pos granio.Position) {
	ev, c := prepareTap(obj, pos)
	handleFocusOnTap(c, obj)
	obj.TappedSecondary(ev)
}

// Type performs a series of key events to simulate typing of a value into the specified object.
// The focusable object will be focused before typing begins.
// The chars parameter will be input one rune at a time to the focused object.
func Type(obj granio.Focusable, chars string) {
	obj.FocusGained()

	typeChars([]rune(chars), obj.TypedRune, obj.TypedKey)
}

// TypeOnCanvas is like the Type function but it passes the key events to the canvas object
// rather than a focusable widget.
func TypeOnCanvas(c granio.Canvas, chars string) {
	typeChars([]rune(chars), c.OnTypedRune(), c.OnTypedKey())
}

// WidgetRenderer allows test scripts to gain access to the current renderer for a widget.
// This can be used for verifying correctness of rendered components for a widget in unit tests.
func WidgetRenderer(wid granio.Widget) granio.WidgetRenderer {
	return cache.Renderer(wid)
}

func findTappable(c granio.Canvas, pos granio.Position) (o granio.CanvasObject, p granio.Position) {
	matches := func(object granio.CanvasObject) bool {
		_, ok := object.(granio.Tappable)
		return ok
	}
	o, p, _ = intdriver.FindObjectAtPositionMatching(pos, matches, c.Overlays().Top(), c.Content())
	return o, p
}

func handleFocusOnTap(c granio.Canvas, obj any) {
	if c == nil {
		return
	}

	if focus, ok := obj.(granio.Focusable); ok {
		dis, ok := obj.(granio.Disableable)
		if (!ok || !dis.Disabled()) && focus == c.Focused() {
			return
		}
	}

	c.Unfocus()
}

func layoutAndCollect(objects []granio.CanvasObject, o granio.CanvasObject, size granio.Size) []granio.CanvasObject {
	objects = append(objects, o)
	switch c := o.(type) {
	case granio.Widget:
		r := c.CreateRenderer()
		r.Layout(size)
		for _, child := range r.Objects() {
			objects = layoutAndCollect(objects, child, child.Size())
		}
	case *granio.Container:
		if c.Layout != nil {
			c.Layout.Layout(c.Objects, size)
		}
		for _, child := range c.Objects {
			objects = layoutAndCollect(objects, child, child.Size())
		}
	}
	return objects
}

func prepareTap(obj any, pos granio.Position) (*granio.PointEvent, granio.Canvas) {
	d := granio.CurrentApp().Driver()
	ev := &granio.PointEvent{Position: pos}
	var c granio.Canvas
	if co, ok := obj.(granio.CanvasObject); ok {
		c = d.CanvasForObject(co)
		ev.AbsolutePosition = d.AbsolutePositionForObject(co).Add(pos)
	}
	return ev, c
}

func performTap(c granio.Canvas, obj granio.Tappable, ev *granio.PointEvent) {
	handleFocusOnTap(c, obj)
	obj.Tapped(ev)
}

func typeChars(chars []rune, typedRune func(rune), typedKey func(*granio.KeyEvent)) {
	for _, char := range chars {
		if char == '\n' {
			typedKey(&granio.KeyEvent{Name: granio.KeyReturn})
			continue
		}

		typedRune(char)
	}
}

func writeMarkup(path string, markup string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(markup), 0o644)
}
