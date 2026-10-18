package test

import (
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	fynedriver "github.com/Uzer007-admin/granio/v2/driver"
	intdriver "github.com/Uzer007-admin/granio/v2/internal/driver"
	"github.com/Uzer007-admin/granio/v2/internal/painter"
	"github.com/Uzer007-admin/granio/v2/internal/painter/software"
	intRepo "github.com/Uzer007-admin/granio/v2/internal/repository"
	"github.com/Uzer007-admin/granio/v2/storage/repository"
)

// SoftwarePainter describes a simple type that can render canvases
//
// Deprecated: Use driver.Painter instead.
type SoftwarePainter = fynedriver.Painter

type driver struct {
	device       device
	painter      fynedriver.Painter
	windows      []granio.Window
	windowsMutex sync.RWMutex
}

// Declare conformity with Driver
var _ granio.Driver = (*driver)(nil)

// NewDriver sets up and registers a new dummy driver for test purpose
func NewDriver() granio.Driver {
	drv := &driver{windowsMutex: sync.RWMutex{}}
	repository.Register("file", intRepo.NewFileRepository())

	httpHandler := intRepo.NewHTTPRepository()
	repository.Register("http", httpHandler)
	repository.Register("https", httpHandler)

	// make a single dummy window for rendering tests
	drv.CreateWindow("")

	return drv
}

// NewDriverWithPainter creates a new dummy driver that will pass the given
// painter to all canvases created
func NewDriverWithPainter(p fynedriver.Painter) granio.Driver {
	return &driver{painter: p}
}

// DoFromGoroutine on a test driver ignores the wait flag as our threading is simple
func (*driver) DoFromGoroutine(f func(), _ bool) {
	// Tests all run on a single (but potentially different per-test) thread
	f()
}

func (d *driver) AbsolutePositionForObject(co granio.CanvasObject) granio.Position {
	c := d.CanvasForObject(co)
	if c == nil {
		return granio.NewPos(0, 0)
	}

	overlays := c.Overlays().List()
	trees := make([]granio.CanvasObject, 0, len(overlays)+1)
	if content := c.Content(); content != nil {
		trees = append(trees, content)
	}
	trees = append(trees, overlays...)
	pos := intdriver.AbsolutePositionForObject(co, trees)
	inset, _ := c.InteractiveArea()
	return pos.Subtract(inset)
}

func (d *driver) AllWindows() []granio.Window {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	return d.windows
}

func (d *driver) CanvasForObject(granio.CanvasObject) granio.Canvas {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	// cheating: probably the last created window is meant
	return d.windows[len(d.windows)-1].Canvas()
}

func (d *driver) CreateWindow(title string) granio.Window {
	p := d.painter
	if p == nil {
		p = software.NewPainter()
	}
	c := NewCanvasWithPainter(p)
	w := &window{canvas: c, driver: d, title: title}

	d.windowsMutex.Lock()
	d.windows = append(d.windows, w)
	d.windowsMutex.Unlock()
	return w
}

func (d *driver) Device() granio.Device {
	return &d.device
}

// RenderedTextSize looks up how bit a string would be if drawn on screen
func (*driver) RenderedTextSize(text string, size float32, style granio.TextStyle, source granio.Resource) (granio.Size, float32) {
	return painter.RenderedTextSize(text, size, style, source)
}

func (*driver) Run() {
	// no-op
}

func (*driver) StartAnimation(a *granio.Animation) {
	// currently no animations in test app, we just initialize it and leave
	if a.AutoReverse {
		a.Tick(0.0)
	} else {
		a.Tick(1.0)
	}
	if a.RepeatCount != granio.AnimationRepeatForever && a.OnFinished != nil {
		a.OnFinished()
	}
}

func (*driver) StopAnimation(*granio.Animation) {
	// currently no animations in test app, do nothing
}

func (*driver) Quit() {
	// no-op
}

func (*driver) Clipboard() granio.Clipboard {
	return nil
}

func (d *driver) removeWindow(w *window) {
	d.windowsMutex.Lock()
	i := 0
	for _, win := range d.windows {
		if win == w {
			break
		}
		i++
	}

	copy(d.windows[i:], d.windows[i+1:])
	d.windows[len(d.windows)-1] = nil // Allow the garbage collector to reclaim the memory.
	d.windows = d.windows[:len(d.windows)-1]

	d.windowsMutex.Unlock()
}

func (*driver) DoubleTapDelay() time.Duration {
	return 300 * time.Millisecond
}

func (*driver) SetDisableScreenBlanking(_ bool) {
	// no-op for test
}
