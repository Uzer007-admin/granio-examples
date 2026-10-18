package software

import (
	"image"
	"image/draw"
	"sync"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/app"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/internal/painter/software"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
	"github.com/Uzer007-admin/granio/v2/theme"
)

const canvasDefaultSize = 100

// WindowlessCanvas provides functionality for a canvas to operate without a window
//
// Since: 2.9
type WindowlessCanvas interface {
	granio.Canvas

	Padded() bool
	Resize(granio.Size)
	SetPadded(bool)
	SetScale(float32)
}

// NewCanvas creates a new canvas in memory that can render without hardware support.
func NewCanvas() WindowlessCanvas {
	return createCanvas(software.NewPainter(), false)
}

// NewCanvasWithPainter creates a new canvas in memory that can render without hardware support
// which uses the given driver.Painter for #Capture().
//
// Since: 2.8
func NewCanvasWithPainter(painter driver.Painter) WindowlessCanvas {
	return createCanvas(painter, false)
}

// NewTransparentCanvas creates a new canvas in memory that can render without hardware support without a background color.
//
// Since: 2.2
func NewTransparentCanvas() WindowlessCanvas {
	return createCanvas(software.NewPainter(), true)
}

// NewTransparentCanvasWithPainter creates a new canvas in memory that can render without hardware support
// which uses the given driver.Painter for #Capture() without a background color.
//
// Since: 2.8
func NewTransparentCanvasWithPainter(painter driver.Painter) WindowlessCanvas {
	return createCanvas(painter, true)
}

func createCanvas(painter driver.Painter, transparent bool) WindowlessCanvas {
	c := &canvas{
		focusMgr:    app.NewFocusManager(nil),
		padded:      true,
		painter:     painter,
		scale:       1.0,
		size:        granio.NewSize(canvasDefaultSize, canvasDefaultSize),
		transparent: transparent,
	}
	c.overlays.Canvas = c
	return c
}

type canvas struct {
	size    granio.Size
	resized bool
	scale   float32

	content     granio.CanvasObject
	overlays    internal.OverlayStack
	focusMgr    *app.FocusManager
	padded      bool
	transparent bool

	onTypedRune func(rune)
	onTypedKey  func(*granio.KeyEvent)

	granio.ShortcutHandler
	painter      driver.Painter
	propertyLock sync.RWMutex
}

func (c *canvas) Capture() image.Image {
	cache.Clean(true)
	size := c.Size()
	bounds := image.Rect(0, 0, scale.ToScreenCoordinate(c, size.Width), scale.ToScreenCoordinate(c, size.Height))
	img := image.NewNRGBA(bounds)
	if !c.transparent {
		draw.Draw(img, bounds, image.NewUniform(theme.Color(theme.ColorNameBackground)), image.Point{}, draw.Src)
	}

	if c.painter != nil {
		draw.Draw(img, bounds, c.painter.Paint(c), image.Point{}, draw.Over)
	}

	return img
}

func (c *canvas) Content() granio.CanvasObject {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.content
}

func (c *canvas) Focus(obj granio.Focusable) {
	c.focusManager().Focus(obj)
}

func (c *canvas) FocusNext() {
	c.focusManager().FocusNext()
}

func (c *canvas) FocusPrevious() {
	c.focusManager().FocusPrevious()
}

func (c *canvas) Focused() granio.Focusable {
	return c.focusManager().Focused()
}

func (c *canvas) InteractiveArea() (granio.Position, granio.Size) {
	return granio.NewPos(0, 0), c.Size()
}

func (c *canvas) OnTypedKey() func(*granio.KeyEvent) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.onTypedKey
}

func (c *canvas) OnTypedRune() func(rune) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.onTypedRune
}

func (c *canvas) Overlays() granio.OverlayStack {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	return &c.overlays
}

func (c *canvas) Padded() bool {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.padded
}

func (c *canvas) PixelCoordinateForPosition(pos granio.Position) (x, y int) {
	return int(pos.X * c.scale), int(pos.Y * c.scale)
}

func (*canvas) Refresh(granio.CanvasObject) {
}

func (c *canvas) Resize(size granio.Size) {
	c.propertyLock.Lock()
	c.resized = true
	c.propertyLock.Unlock()

	c.doResize(size)
}

func (c *canvas) Scale() float32 {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.scale
}

func (c *canvas) SetContent(content granio.CanvasObject) {
	c.propertyLock.Lock()
	c.content = content
	c.focusMgr = app.NewFocusManager(c.content)
	resized := c.resized
	c.propertyLock.Unlock()

	if content == nil {
		return
	}

	minSize := content.MinSize()
	if c.padded {
		minSize = minSize.Add(granio.NewSquareSize(theme.Padding() * 2))
	}

	if resized {
		c.doResize(internal.MaxSizes(c.Size(), minSize))
	} else {
		c.doResize(minSize)
	}
}

func (c *canvas) SetOnTypedKey(handler func(*granio.KeyEvent)) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.onTypedKey = handler
}

func (c *canvas) SetOnTypedRune(handler func(rune)) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.onTypedRune = handler
}

func (c *canvas) SetPadded(padded bool) {
	c.propertyLock.Lock()
	c.padded = padded
	c.propertyLock.Unlock()

	c.doResize(c.Size())
}

func (c *canvas) SetScale(s float32) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.scale = s
}

func (c *canvas) Size() granio.Size {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.size
}

func (c *canvas) Unfocus() {
	c.focusManager().Focus(nil)
}

func (c *canvas) doResize(size granio.Size) {
	c.propertyLock.Lock()
	content := c.content
	overlays := c.overlays
	padded := c.padded
	c.size = size
	c.propertyLock.Unlock()

	if content == nil {
		return
	}

	// Ensure testcanvas mimics real canvas.Resize behavior
	fullPos, fullSize := c.InteractiveArea()
	for _, overlay := range overlays.List() {
		overlay.Move(fullPos)
		overlay.Resize(fullSize)
	}

	if padded {
		padding := theme.Padding()
		content.Resize(size.Subtract(granio.NewSquareSize(padding * 2)))
		content.Move(granio.NewSquareOffsetPos(padding))
	} else {
		content.Resize(size)
		content.Move(granio.NewPos(0, 0))
	}
}

func (c *canvas) focusManager() *app.FocusManager {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if focusMgr := c.overlays.TopFocusManager(); focusMgr != nil {
		return focusMgr
	}
	return c.focusMgr
}
