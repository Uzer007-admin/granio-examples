package glfw

import (
	"image"
	"math"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/app"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	"github.com/Uzer007-admin/granio/v2/internal/driver"
	"github.com/Uzer007-admin/granio/v2/internal/driver/common"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// Declare conformity with Canvas interface
var _ granio.Canvas = (*glCanvas)(nil)

type glCanvas struct {
	common.Canvas

	content granio.CanvasObject
	menu    granio.CanvasObject
	padded  bool
	size    granio.Size

	onTypedRune func(rune)
	onTypedKey  func(*granio.KeyEvent)
	onKeyDown   func(*granio.KeyEvent)
	onKeyUp     func(*granio.KeyEvent)
	// shortcut    granio.ShortcutHandler

	scale, detectedScale, texScale float32

	context         driver.WithContext
	webExtraWindows *container.MultipleWindows
}

func (c *glCanvas) Capture() image.Image {
	var img image.Image
	c.context.(*window).RunWithContext(func() {
		if painter, ok := c.Painter().(interface {
			CaptureBack(granio.Canvas) image.Image
		}); ok {
			updateGLContext(c.context.(*window))
			c.FreeDirtyTextures()
			c.paint(c.Size())
			img = painter.CaptureBack(c)
		} else {
			img = c.Painter().Capture(c)
		}
	})
	return img
}

func (c *glCanvas) Content() granio.CanvasObject {
	return c.content
}

func (c *glCanvas) Focus(object granio.Focusable) {
	if c.Focused() == object {
		return
	}
	if window, ok := c.context.(*window); ok {
		window.resetInputMethod()
	}
	c.Canvas.Focus(object)
	if window, ok := c.context.(*window); ok {
		window.updateInputMethod()
	}
}

func (c *glCanvas) FocusNext() {
	if window, ok := c.context.(*window); ok {
		window.resetInputMethod()
	}
	c.Canvas.FocusNext()
	if window, ok := c.context.(*window); ok {
		window.updateInputMethod()
	}
}

func (c *glCanvas) FocusPrevious() {
	if window, ok := c.context.(*window); ok {
		window.resetInputMethod()
	}
	c.Canvas.FocusPrevious()
	if window, ok := c.context.(*window); ok {
		window.updateInputMethod()
	}
}

func (c *glCanvas) FocusLost() {
	if window, ok := c.context.(*window); ok {
		window.resetInputMethod()
	}
	c.Canvas.FocusLost()
	if window, ok := c.context.(*window); ok {
		window.updateInputMethod()
	}
}

func (c *glCanvas) DismissMenu() bool {
	if c.menu != nil && c.menu.(*MenuBar).IsActive() {
		c.menu.(*MenuBar).Toggle()
		return true
	}
	return false
}

func (c *glCanvas) InteractiveArea() (granio.Position, granio.Size) {
	return granio.Position{}, c.Size()
}

func (c *glCanvas) MinSize() granio.Size {
	return c.canvasSize(c.content.MinSize())
}

func (c *glCanvas) OnKeyDown() func(*granio.KeyEvent) {
	return c.onKeyDown
}

func (c *glCanvas) OnKeyUp() func(*granio.KeyEvent) {
	return c.onKeyUp
}

func (c *glCanvas) OnTypedKey() func(*granio.KeyEvent) {
	return c.onTypedKey
}

func (c *glCanvas) OnTypedRune() func(rune) {
	return c.onTypedRune
}

func (c *glCanvas) Padded() bool {
	return c.padded
}

func (c *glCanvas) PixelCoordinateForPosition(pos granio.Position) (x, y int) {
	multiple := c.scale * c.texScale
	scaleInt := func(x float32) int {
		return int(math.Round(float64(x * multiple)))
	}

	return scaleInt(pos.X), scaleInt(pos.Y)
}

func (c *glCanvas) Resize(size granio.Size) {
	// This might not be the ideal solution, but it effectively avoid the first frame to be blurry due to the
	// rounding of the size to the loower integer when scale == 1. It does not affect the other cases as far as we tested.
	// This can easily be seen with fyne/cmd/hello and a scale == 1 as the text will happear blurry without the following line.
	var nearestSize granio.Size
	if c.scale == 1 {
		nearestSize = granio.NewSize(float32(math.Ceil(float64(size.Width))), float32(math.Ceil(float64(size.Height))))
	} else {
		nearestSize = size
	}

	c.size = nearestSize

	if c.webExtraWindows != nil {
		c.webExtraWindows.Resize(size)
	}
	for _, overlay := range c.Overlays().List() {
		overlay.Resize(nearestSize)
	}

	content := c.content
	contentSize := c.contentSize(nearestSize)
	contentPos := c.contentPos()
	menu := c.menu
	menuHeight := c.menuHeight()

	content.Resize(contentSize)
	content.Move(contentPos)

	if menu != nil {
		menu.Refresh()
		menu.Resize(granio.NewSize(nearestSize.Width, menuHeight))
	}
}

func (c *glCanvas) Scale() float32 {
	return c.scale
}

// SetDirty wakes the native loop when a background resource finishes loading.
func (c *glCanvas) SetDirty() {
	c.Canvas.SetDirty()
	wakeMainLoop()
}

func (c *glCanvas) SetContent(content granio.CanvasObject) {
	newSize := internal.MaxSizes(c.size, c.canvasSize(content.MinSize()))

	c.applyContent(content)

	c.Resize(newSize)
	c.SetDirty()
}

func (c *glCanvas) SetOnKeyDown(typed func(*granio.KeyEvent)) {
	c.onKeyDown = typed
}

func (c *glCanvas) SetOnKeyUp(typed func(*granio.KeyEvent)) {
	c.onKeyUp = typed
}

func (c *glCanvas) SetOnTypedKey(typed func(*granio.KeyEvent)) {
	c.onTypedKey = typed
}

func (c *glCanvas) SetOnTypedRune(typed func(rune)) {
	c.onTypedRune = typed
}

func (c *glCanvas) SetPadded(padded bool) {
	c.padded = padded

	c.content.Move(c.contentPos())
}

func (c *glCanvas) reloadScale() {
	w, _ := c.context.(*window)
	windowVisible := w.visible
	if !windowVisible {
		return
	}

	c.scale = w.calculatedScale()
	c.SetDirty()

	c.context.RescaleContext()
}

func (c *glCanvas) Size() granio.Size {
	return c.size
}

func (c *glCanvas) ToggleMenu() {
	if c.menu != nil {
		c.menu.(*MenuBar).Toggle()
	}
}

func (c *glCanvas) applyContent(content granio.CanvasObject) {
	c.content = content
	c.SetContentTreeAndFocusMgr(content)
}

func (c *glCanvas) buildMenu(w *window, m *granio.MainMenu) {
	c.setMenuOverlay(nil)
	if m == nil {
		return
	}
	if build.HasNativeMenu {
		setupNativeMenu(w, m)
	} else {
		c.setMenuOverlay(buildMenuOverlay(m, w))
	}
}

// canvasSize computes the needed canvas size for the given content size
func (c *glCanvas) canvasSize(contentSize granio.Size) granio.Size {
	canvasSize := contentSize.Add(granio.NewSize(0, c.menuHeight()))
	if c.Padded() {
		return canvasSize.Add(granio.NewSquareSize(theme.Padding() * 2))
	}
	return canvasSize
}

func (c *glCanvas) contentPos() granio.Position {
	contentPos := granio.NewPos(0, c.menuHeight())
	if c.Padded() {
		return contentPos.Add(granio.NewSquareOffsetPos(theme.Padding()))
	}
	return contentPos
}

func (c *glCanvas) contentSize(canvasSize granio.Size) granio.Size {
	contentSize := granio.NewSize(canvasSize.Width, canvasSize.Height-c.menuHeight())
	if c.Padded() {
		return contentSize.Subtract(granio.NewSquareSize(theme.Padding() * 2))
	}
	return contentSize
}

func (c *glCanvas) menuHeight() float32 {
	if c.menu == nil {
		return 0 // no menu or native menu -> does not consume space on the canvas
	}

	return c.menu.MinSize().Height
}

func (c *glCanvas) overlayChanged() {
	c.SetDirty()
}

func (c *glCanvas) paint(size granio.Size) {
	clips := &internal.ClipStack{}
	if c.Content() == nil {
		return
	}
	c.Painter().Clear()

	paint := func(node *common.RenderCacheNode, pos granio.Position) {
		obj := node.Obj()
		if driver.IsClip(obj) {
			inner := clips.Push(pos, obj.Size())
			c.Painter().StartClipping(inner.Rect())
		}
		if size.Width <= 0 || size.Height <= 0 { // iconifying on Windows can do bad things
			return
		}
		c.Painter().Paint(obj, pos, size, clips.Top())
	}
	afterPaint := func(node *common.RenderCacheNode, pos granio.Position) {
		if driver.IsClip(node.Obj()) {
			clips.Pop()
			if top := clips.Top(); top != nil {
				c.Painter().StartClipping(top.Rect())
			} else {
				c.Painter().StopClipping()
			}
		}

		if build.Mode == granio.BuildDebug {
			c.DrawDebugOverlay(node.Obj(), pos, size, clips.Top())
		}
	}
	c.WalkTrees(paint, afterPaint)
	// The painter batches text and plain rectangles, so the tail of the tree can
	// still be sitting in a queue when the walk ends.
	if p, ok := c.Painter().(interface{ FlushGlyphs() }); ok {
		p.FlushGlyphs()
	}
}

func (c *glCanvas) setMenuOverlay(b granio.CanvasObject) {
	c.menu = b
	c.SetMenuTreeAndFocusMgr(b)

	if c.menu != nil && !c.size.IsZero() {
		c.content.Resize(c.contentSize(c.size))
		c.content.Move(c.contentPos())

		c.menu.Refresh()
		c.menu.Resize(granio.NewSize(c.size.Width, c.menu.MinSize().Height))
	}
}

func (c *glCanvas) applyThemeOutOfTreeObjects() {
	if c.menu != nil {
		app.ApplyThemeTo(c.menu, c) // Ensure our menu gets the theme change message as it's out-of-tree
	}

	c.SetPadded(c.padded) // refresh the padding for potential theme differences
}

func newCanvas() *glCanvas {
	c := &glCanvas{scale: 1.0, texScale: 1.0, padded: true}
	connectKeyboard(c)
	c.Initialize(c, c.overlayChanged)
	c.applyContent(&canvas.Rectangle{FillColor: theme.Color(theme.ColorNameBackground)})
	return c
}
