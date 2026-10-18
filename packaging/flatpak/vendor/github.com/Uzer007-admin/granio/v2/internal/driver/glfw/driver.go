// Package glfw provides a full Fyne desktop driver that uses the system OpenGL libraries.
// This supports Windows, Mac OS X and Linux using the gl and glfw packages from go-gl.
package glfw

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/jpeg" // allow to use JPEGs as icon source
	"image/png"
	"os"
	"runtime"
	"sync/atomic"

	"github.com/Uzer007-admin/granio-image/ico"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/animation"
	intapp "github.com/Uzer007-admin/granio/v2/internal/app"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/driver"
	"github.com/Uzer007-admin/granio/v2/internal/driver/common"
	"github.com/Uzer007-admin/granio/v2/internal/goos"
	"github.com/Uzer007-admin/granio/v2/internal/painter"
	intRepo "github.com/Uzer007-admin/granio/v2/internal/repository"
	"github.com/Uzer007-admin/granio/v2/storage/repository"
)

var curWindow *window

// NewGLDriver sets up a new Driver instance implemented using the GLFW Go library and OpenGL bindings.
func NewGLDriver() granio.Driver {
	repository.Register(granio.URISchemeFile, intRepo.NewFileRepository())

	return &gLDriver{
		done: make(chan struct{}),
	}
}

type gLDriver struct {
	displayPlatform    atomic.Int32
	windows            []granio.Window
	initialized        bool
	done               chan struct{}
	requestedFrameRate atomic.Int32
	detectedFrameRate  atomic.Int32

	animation       animation.Runner
	animationOwners map[granio.CanvasObject]bool

	currentKeyModifiers granio.KeyModifier // desktop driver only

	trayStart, trayStop func()       // shut down the system tray, if used
	systrayMenu         *granio.Menu // cache the menu set so we know when to refresh
}

// Declare conformity with Driver
var _ granio.Driver = (*gLDriver)(nil)

func (d *gLDriver) AbsolutePositionForObject(co granio.CanvasObject) granio.Position {
	c := d.CanvasForObject(co)
	if c == nil {
		return granio.NewPos(0, 0)
	}

	glc, _ := c.(*glCanvas)
	return driver.AbsolutePositionForObject(co, glc.ObjectTrees())
}

func (d *gLDriver) AllWindows() []granio.Window {
	return d.windows
}

func (*gLDriver) CanvasForObject(obj granio.CanvasObject) granio.Canvas {
	return common.CanvasForObject(obj)
}

func (d *gLDriver) CreateSplashWindow() granio.Window {
	win := d.newWindow("", false)
	win.SetPadded(false)
	win.CenterOnScreen()
	return win
}

func (*gLDriver) Device() granio.Device {
	return &glDevice{}
}

func (*gLDriver) DoFromGoroutine(f func(), wait bool) {
	if wait {
		async.EnsureNotMain(func() {
			runOnMainWithWait(f, true)
		})
	} else {
		runOnMainWithWait(f, false)
	}
}

func (d *gLDriver) Quit() {
	if curWindow != nil {
		if f := granio.CurrentApp().Lifecycle().(*intapp.Lifecycle).OnExitedForeground(); f != nil {
			f()
		}
		curWindow = nil
		if d.trayStop != nil {
			d.trayStop()
		}
	}

	// Only call close once to avoid panic.
	if running.CompareAndSwap(true, false) {
		close(d.done)
		wakeMainLoop()
	}
}

func (*gLDriver) RenderedTextSize(text string, textSize float32, style granio.TextStyle, source granio.Resource) (size granio.Size, baseline float32) {
	return painter.RenderedTextSize(text, textSize, style, source)
}

func (d *gLDriver) Run() {
	if !async.IsMainGoroutine() {
		panic("Run() or ShowAndRun() must be called from main goroutine")
	}

	go d.catchTerm()
	d.runGL()

	// Ensure lifecycle events run to completion before the app exits
	l, _ := granio.CurrentApp().Lifecycle().(*intapp.Lifecycle)
	l.WaitForEvents()
	l.DestroyEventQueue()
}

func (*gLDriver) SetDisableScreenBlanking(disable bool) {
	setDisableScreenBlank(disable)
}

func (d *gLDriver) newWindow(title string, decorate bool) granio.Window {
	var ret *window
	if title == "" {
		title = defaultTitle
	}

	d.init()

	// A window starts with no mouse move outstanding: the zero value would read
	// as one pending at (0,0), which the first click would then apply.
	ret = &window{title: title, decorate: decorate, driver: d, mousePosUpdateProcessed: true}
	ret.frame = newPresentGate(ret)
	ret.canvas = newCanvas()
	ret.canvas.context = ret
	ret.SetIcon(ret.icon)
	d.windows = append(d.windows, ret)
	return ret
}

// a trivial implementation of "focus previous" - return to the most recently opened, or master if set.
// This may not do the right thing if your app has 3 or more windows open, but it was agreed this was not much
// of an issue, and the added complexity to track focus was not needed at this time.
func (d *gLDriver) focusPreviousWindow() {
	var chosen *window
	for _, w := range d.windows {
		win, _ := w.(*window)
		if !win.visible {
			continue
		}
		chosen = win
		if win.master {
			break
		}
	}

	if chosen == nil || chosen.view() == nil {
		return
	}
	chosen.RequestFocus()
}

func (d *gLDriver) init() {
	if !d.initialized {
		d.initialized = true
		d.initGLFW()
	}
}

func (d *gLDriver) initFailed(msg string, err error) {
	granio.LogError(msg, err)

	if running.Load() {
		os.Exit(1) //revive:disable-line:deep-exit
	}

	d.Quit()
}

func convertToPNG(img image.Image) ([]byte, error) {
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, img, bounds.Min, draw.Src)

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, nrgba); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func toOSIcon(icon []byte) ([]byte, error) {
	return toOSIconForRuntime(icon, runtime.GOOS)
}

// toOSIconForRuntime takes the input image bytes and converts it to an image type
// that is suitable for the specified GOOS runtime. Which makes platform-specific icon handling testable.
func toOSIconForRuntime(icon []byte, osName string) ([]byte, error) {
	if osName != goos.Windows && !usesUnixSystrayIcon(osName) {
		return icon, nil
	}

	img, format, err := image.Decode(bytes.NewReader(icon))
	if err != nil {
		return nil, err
	}

	// keep windows behavior: convert to ico
	if osName == goos.Windows {
		buf := &bytes.Buffer{}
		if err = ico.Encode(buf, img); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	if format == "png" {
		return icon, nil
	}

	// Unix systray expects a PNG bitmap.
	return convertToPNG(img)
}

func usesUnixSystrayIcon(osName string) bool {
	return osName == goos.Linux || goos.IsBSD(osName)
}
