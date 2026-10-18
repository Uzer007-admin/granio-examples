package mobile

import (
	"math"
	"runtime"
	"strconv"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	fynecanvas "github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/driver/mobile"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/animation"
	intapp "github.com/Uzer007-admin/granio/v2/internal/app"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	intdriver "github.com/Uzer007-admin/granio/v2/internal/driver"
	"github.com/Uzer007-admin/granio/v2/internal/driver/common"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/app"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/event/key"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/event/lifecycle"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/event/mouse"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/event/paint"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/event/size"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/event/touch"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/gl"
	"github.com/Uzer007-admin/granio/v2/internal/goos"
	"github.com/Uzer007-admin/granio/v2/internal/painter"
	pgl "github.com/Uzer007-admin/granio/v2/internal/painter/gl"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
	"github.com/Uzer007-admin/granio/v2/theme"
)

const (
	tapMoveDecay        = 0.92                   // how much should the scroll continue decay on each frame?
	tapMoveEndThreshold = 2.0                    // at what offset will we stop decaying?
	tapMoveThreshold    = 4.0                    // how far can we move before it is a drag
	tapSecondaryDelay   = 300 * time.Millisecond // how long before secondary tap
	tapDoubleDelay      = 500 * time.Millisecond // max duration between taps for a DoubleTap event
)

// Configuration is the system information about the current device
type Configuration struct {
	SystemTheme granio.ThemeVariant
}

// ConfiguredDriver is a simple type that allows packages to hook into configuration changes of this driver.
type ConfiguredDriver interface {
	SetOnConfigurationChanged(func(*Configuration))
}

type driver struct {
	app   app.App
	glctx gl.Context

	windows     []granio.Window
	device      device
	animation   animation.Runner
	currentSize size.Event

	theme           granio.ThemeVariant
	onConfigChanged func(*Configuration)
	painting        bool
	running         bool
	queuedFuncs     *async.UnboundedChan[func()]
}

// Declare conformity with Driver
var (
	_ granio.Driver    = (*driver)(nil)
	_ ConfiguredDriver = (*driver)(nil)
)

func init() {
	runtime.LockOSThread()
}

func (d *driver) DoFromGoroutine(fn func(), wait bool) {
	caller := func() {
		if d.queuedFuncs == nil {
			fn() // before the app actually starts
			return
		}
		var done chan struct{}
		if wait {
			done = common.DonePool.Get()
			defer common.DonePool.Put(done)
		}

		d.queuedFuncs.In() <- func() {
			fn()
			if wait {
				done <- struct{}{}
			}
		}

		if wait {
			<-done
		}
	}

	if wait {
		async.EnsureNotMain(caller)
	} else {
		caller()
	}
}

func (d *driver) CreateWindow(title string) granio.Window {
	c, _ := newCanvas(granio.CurrentDevice()).(*canvas)
	ret := &window{title: title, canvas: c, isChild: len(d.windows) > 0}
	c.applyContent(&fynecanvas.Rectangle{FillColor: theme.Color(theme.ColorNameBackground)})
	c.SetPainter(pgl.NewPainter(c, ret))
	d.windows = append(d.windows, ret)
	return ret
}

func (d *driver) AllWindows() []granio.Window {
	return d.windows
}

// currentWindow returns the most recently opened window - we can only show one at a time.
func (d *driver) currentWindow() *window {
	if len(d.windows) == 0 {
		return nil
	}

	var last *window
	for i := len(d.windows) - 1; i >= 0; i-- {
		last, _ = d.windows[i].(*window)
		if last.visible {
			return last
		}
	}

	return last
}

func (*driver) Clipboard() granio.Clipboard {
	return NewClipboard()
}

func (*driver) RenderedTextSize(text string, textSize float32, style granio.TextStyle, source granio.Resource) (granio.Size, float32) {
	return painter.RenderedTextSize(text, textSize, style, source)
}

func (d *driver) CanvasForObject(granio.CanvasObject) granio.Canvas {
	if len(d.windows) == 0 {
		return nil
	}

	// TODO figure out how we handle multiple windows...
	return d.currentWindow().Canvas()
}

func (d *driver) AbsolutePositionForObject(co granio.CanvasObject) granio.Position {
	c := d.CanvasForObject(co)
	if c == nil {
		return granio.NewPos(0, 0)
	}

	mc, _ := c.(*canvas)
	pos := intdriver.AbsolutePositionForObject(co, mc.ObjectTrees())
	inset, _ := c.InteractiveArea()
	return pos.Subtract(inset)
}

func (*driver) GoBack() {
	app.GoBack()
}

func (*driver) Quit() {
	// Android and iOS guidelines say this should not be allowed!
}

func (d *driver) Run() {
	if d.running {
		return // Run was called twice.
	}
	d.running = true

	app.Main(func(a app.App) {
		async.SetMainGoroutine()
		d.app = a
		d.queuedFuncs = async.NewUnboundedChan[func()]()

		granio.CurrentApp().Settings().AddListener(func(s granio.Settings) {
			painter.ClearFontCache()
			cache.ResetThemeCaches()
			intapp.ApplySettingsWithCallback(s, granio.CurrentApp(), func(w granio.Window) {
				c, ok := w.Canvas().(*canvas)
				if !ok {
					return
				}
				c.applyThemeOutOfTreeObjects()
			})
		})

		draw := time.NewTicker(time.Second / 60)
		defer func() {
			l, _ := granio.CurrentApp().Lifecycle().(*intapp.Lifecycle)

			// exhaust the event queue
			go func() {
				l.WaitForEvents()
				d.queuedFuncs.Close()
			}()
			for fn := range d.queuedFuncs.Out() {
				fn()
			}

			l.DestroyEventQueue()
		}()

		for {
			select {
			case <-draw.C:
				d.sendPaintEvent()
			case fn := <-d.queuedFuncs.Out():
				fn()
			case e, ok := <-a.Events():
				if !ok {
					return // events channel closed, app done
				}
				current := d.currentWindow()
				if current == nil {
					continue
				}
				c, _ := current.Canvas().(*canvas)

				switch e := a.Filter(e).(type) {
				case lifecycle.Event:
					d.handleLifecycle(e, current)
				case size.Event:
					if e.WidthPx <= 0 {
						continue
					}
					d.currentSize = e
					currentOrientation = e.Orientation
					currentDPI = e.PixelsPerPt * 72
					d.setTheme(e.DarkMode)

					dev := &d.device
					insetChange := dev.safeTop != e.InsetTopPx || dev.safeBottom != e.InsetBottomPx ||
						dev.safeLeft != e.InsetLeftPx || dev.safeRight != e.InsetRightPx
					dev.safeTop = e.InsetTopPx
					dev.safeLeft = e.InsetLeftPx
					dev.safeBottom = e.InsetBottomPx
					dev.safeRight = e.InsetRightPx
					c.scale = granio.CurrentDevice().SystemScaleForWindow(nil)
					c.Painter().SetFrameBufferScale(1.0)

					if insetChange {
						current.canvas.sizeContent(current.canvas.size) // even if size didn't change we invalidate
					}
					// make sure that we paint on the next frame
					c.Content().Refresh()
				case paint.Event:
					d.handlePaint(e, current)
				case touch.Event:
					switch e.Type {
					case touch.TypeBegin:
						d.tapDownCanvas(current, e.X, e.Y, e.Sequence)
					case touch.TypeMove:
						d.tapMoveCanvas(current, e.X, e.Y, e.Sequence)
					case touch.TypeEnd:
						d.tapUpCanvas(current, e.X, e.Y, e.Sequence)
					}
				case mouse.ScrollEvent:
					d.scrollCanvas(current, e)
				case key.Event:
					if runtime.GOOS == goos.Android && e.Code == key.CodeDeleteBackspace && e.Rune < 0 && d.device.keyboardShown {
						break // we are getting release/press on backspace during soft backspace
					}

					switch e.Direction {
					case key.DirPress:
						d.typeDownCanvas(c, e.Rune, e.Code, e.Modifiers)
					case key.DirRelease:
						d.typeUpCanvas(c, e.Rune, e.Code, e.Modifiers)
					}
				}
			}
		}
	})
}

func (*driver) SetDisableScreenBlanking(disable bool) {
	setDisableScreenBlank(disable)
}

func (d *driver) handleLifecycle(e lifecycle.Event, w *window) {
	c, _ := w.Canvas().(*canvas)
	switch e.Crosses(lifecycle.StageAlive) {
	case lifecycle.CrossOn:
		d.onStart()
	case lifecycle.CrossOff:
		d.onStop()
	}
	switch e.Crosses(lifecycle.StageVisible) {
	case lifecycle.CrossOn:
		d.glctx, _ = e.DrawContext.(gl.Context)
		// this is a fix for some android phone to prevent the app from being drawn as a blank screen after being pushed in the background
		c.Content().Refresh()

		d.sendPaintEvent()
	case lifecycle.CrossOff:
		d.glctx = nil
	}
	switch e.Crosses(lifecycle.StageFocused) {
	case lifecycle.CrossOn: // foregrounding
		if f := granio.CurrentApp().Lifecycle().(*intapp.Lifecycle).OnEnteredForeground(); f != nil {
			f()
		}
	case lifecycle.CrossOff: // will enter background
		if runtime.GOOS == goos.Darwin || runtime.GOOS == goos.IOS {
			if d.glctx == nil {
				return
			}

			s := granio.NewSize(float32(d.currentSize.WidthPx)/c.scale, float32(d.currentSize.HeightPx)/c.scale)
			app.BeginPaint()
			d.paintWindow(w, s)
			d.app.Publish()
		}
		if f := granio.CurrentApp().Lifecycle().(*intapp.Lifecycle).OnExitedForeground(); f != nil {
			f()
		}
	}
}

func (d *driver) handlePaint(e paint.Event, w *window) {
	c, _ := w.Canvas().(*canvas)
	if e.Window != 0 { // not all paint events come from hardware
		w.handle = e.Window
	}
	d.painting = false
	if d.glctx == nil || e.External {
		return
	}

	d.animation.TickAnimations()
	needsInit := !c.initialized
	canvasNeedRefresh := c.FreeDirtyTextures() > 0 || c.CheckDirtyAndClear()
	if !needsInit && !canvasNeedRefresh {
		cache.Clean(false)
		return
	}

	// Everything below queues GL calls that only the UI thread can run, and it blocks on each one.
	app.BeginPaint()

	if needsInit {
		c.initialized = true
		c.Painter().Init() // we cannot init until the context is set above
	}

	newSize := granio.NewSize(float32(d.currentSize.WidthPx)/c.scale, float32(d.currentSize.HeightPx)/c.scale)
	if c.EnsureMinSize() {
		c.sizeContent(newSize) // force resize of content
	} else { // if screen changed
		w.Resize(newSize)
	}

	d.paintWindow(w, newSize)
	d.app.Publish()
	w.updateAccessibility()
	cache.Clean(true)
}

func (*driver) onStart() {
	if f := granio.CurrentApp().Lifecycle().(*intapp.Lifecycle).OnStarted(); f != nil {
		f()
	}
}

func (*driver) onStop() {
	l, _ := granio.CurrentApp().Lifecycle().(*intapp.Lifecycle)
	if f := l.OnStopped(); f != nil {
		l.QueueEvent(f)
	}
}

func (d *driver) paintWindow(window granio.Window, s granio.Size) {
	clips := &internal.ClipStack{}
	c, _ := window.Canvas().(*canvas)

	c.Painter().SetOutputSize(d.currentSize.WidthPx, d.currentSize.HeightPx)

	r, g, b, a := theme.Color(theme.ColorNameBackground).RGBA()
	max16bit := float32(255 * 255)
	d.glctx.ClearColor(float32(r)/max16bit, float32(g)/max16bit, float32(b)/max16bit, float32(a)/max16bit)
	d.glctx.Clear(gl.ColorBufferBit)

	draw := func(node *common.RenderCacheNode, pos granio.Position) {
		obj := node.Obj()
		if intdriver.IsClip(obj) {
			inner := clips.Push(pos, obj.Size())
			c.Painter().StartClipping(inner.Rect())
		}

		if s.Width <= 0 || s.Height <= 0 { // iconifying on Windows can do bad things
			return
		}
		c.Painter().Paint(obj, pos, s, clips.Top())
	}
	afterDraw := func(node *common.RenderCacheNode, pos granio.Position) {
		if intdriver.IsClip(node.Obj()) {
			c.Painter().StopClipping()
			clips.Pop()
			if top := clips.Top(); top != nil {
				c.Painter().StartClipping(top.Rect())
			}
		}

		if build.Mode == granio.BuildDebug {
			c.DrawDebugOverlay(node.Obj(), pos, s, clips.Top())
		}
	}

	c.WalkTrees(draw, afterDraw)
	// The painter batches text and plain rectangles, so the tail of the tree can
	// still be sitting in a queue when the walk ends.
	if p, ok := c.Painter().(interface{ FlushGlyphs() }); ok {
		p.FlushGlyphs()
	}
}

func (d *driver) sendPaintEvent() {
	if d.painting {
		return
	}
	d.app.Send(paint.Event{})
	d.painting = true
}

func (d *driver) setTheme(dark bool) {
	var mode granio.ThemeVariant
	if dark {
		mode = theme.VariantDark
	} else {
		mode = theme.VariantLight
	}

	if d.theme != mode && d.onConfigChanged != nil {
		d.onConfigChanged(&Configuration{SystemTheme: mode})
	}
	d.theme = mode
}

func (*driver) tapDownCanvas(w *window, x, y float32, tapID touch.Sequence) {
	tapX := scale.ToFyneCoordinate(w.canvas, int(x))
	tapY := scale.ToFyneCoordinate(w.canvas, int(y))
	pos := granio.NewPos(tapX, tapY+tapYOffset)

	w.canvas.tapDown(pos, int(tapID))
}

func (*driver) tapMoveCanvas(w *window, x, y float32, tapID touch.Sequence) {
	tapX := scale.ToFyneCoordinate(w.canvas, int(x))
	tapY := scale.ToFyneCoordinate(w.canvas, int(y))
	pos := granio.NewPos(tapX, tapY+tapYOffset)

	co, posMovable, _ := w.canvas.findObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
		if _, ok := object.(mobile.Movable); ok {
			return true
		}
		return false
	})

	if co != nil {
		co.(mobile.Movable).TouchMoved(&mobile.TouchEvent{
			PointEvent: granio.PointEvent{
				Position: posMovable,
			},
			ID: int(tapID),
		})
	}

	if tapID == 0 {
		w.canvas.tapMove(pos, int(tapID), func(wid granio.Draggable, ev *granio.DragEvent) {
			wid.Dragged(ev)
		})
	}
}

func (d *driver) tapUpCanvas(w *window, x, y float32, tapID touch.Sequence) {
	tapX := scale.ToFyneCoordinate(w.canvas, int(x))
	tapY := scale.ToFyneCoordinate(w.canvas, int(y))
	pos := granio.NewPos(tapX, tapY+tapYOffset)

	w.canvas.tapUp(pos, int(tapID), func(wid granio.Tappable, ev *granio.PointEvent) {
		wid.Tapped(ev)
	}, func(wid granio.SecondaryTappable, ev *granio.PointEvent) {
		wid.TappedSecondary(ev)
	}, func(wid granio.DoubleTappable, ev *granio.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(wid granio.Draggable, ev *granio.DragEvent) {
		if math.Abs(float64(ev.Dragged.DX)) <= tapMoveEndThreshold && math.Abs(float64(ev.Dragged.DY)) <= tapMoveEndThreshold {
			wid.DragEnd()
			return
		}

		go func() {
			for math.Abs(float64(ev.Dragged.DX)) > tapMoveEndThreshold || math.Abs(float64(ev.Dragged.DY)) > tapMoveEndThreshold {
				if math.Abs(float64(ev.Dragged.DX)) > 0 {
					ev.Dragged.DX *= tapMoveDecay
				}
				if math.Abs(float64(ev.Dragged.DY)) > 0 {
					ev.Dragged.DY *= tapMoveDecay
				}

				d.DoFromGoroutine(func() {
					wid.Dragged(ev)
				}, false)
				time.Sleep(time.Millisecond * 16)
			}

			d.DoFromGoroutine(wid.DragEnd, false)
		}()
	})
}

// scrollCanvas sends a vertical mouse wheel scroll from the simulator to the
// scrollable object under the mouse, if there is one at that position. It
// uses the shared scroll speed, without the desktop driver's acceleration.
func (*driver) scrollCanvas(w *window, e mouse.ScrollEvent) {
	pos := granio.NewPos(scale.ToFyneCoordinate(w.canvas, int(e.X)),
		scale.ToFyneCoordinate(w.canvas, int(e.Y))+tapYOffset)

	co, objPos, _ := w.canvas.findObjectAtPositionMatching(pos, func(object granio.CanvasObject) bool {
		_, ok := object.(granio.Scrollable)
		return ok
	})

	wid, ok := co.(granio.Scrollable)
	if !ok {
		return
	}

	wid.Scrolled(&granio.ScrollEvent{
		PointEvent: granio.PointEvent{
			Position:         objPos,
			AbsolutePosition: pos,
		},
		Scrolled: granio.NewDelta(0, e.ScrollY*common.ScrollSpeed),
	})
}

var keyCodeMap = map[key.Code]granio.KeyName{
	// non-printable
	key.CodeEscape:          granio.KeyEscape,
	key.CodeReturnEnter:     granio.KeyReturn,
	key.CodeTab:             granio.KeyTab,
	key.CodeDeleteBackspace: granio.KeyBackspace,
	key.CodeDeleteForward:   granio.KeyDelete,
	key.CodeInsert:          granio.KeyInsert,
	key.CodePageUp:          granio.KeyPageUp,
	key.CodePageDown:        granio.KeyPageDown,
	key.CodeHome:            granio.KeyHome,
	key.CodeEnd:             granio.KeyEnd,

	key.CodeLeftArrow:  granio.KeyLeft,
	key.CodeRightArrow: granio.KeyRight,
	key.CodeUpArrow:    granio.KeyUp,
	key.CodeDownArrow:  granio.KeyDown,

	key.CodeF1:  granio.KeyF1,
	key.CodeF2:  granio.KeyF2,
	key.CodeF3:  granio.KeyF3,
	key.CodeF4:  granio.KeyF4,
	key.CodeF5:  granio.KeyF5,
	key.CodeF6:  granio.KeyF6,
	key.CodeF7:  granio.KeyF7,
	key.CodeF8:  granio.KeyF8,
	key.CodeF9:  granio.KeyF9,
	key.CodeF10: granio.KeyF10,
	key.CodeF11: granio.KeyF11,
	key.CodeF12: granio.KeyF12,

	key.CodeKeypadEnter: granio.KeyEnter,

	// printable
	key.CodeA:       granio.KeyA,
	key.CodeB:       granio.KeyB,
	key.CodeC:       granio.KeyC,
	key.CodeD:       granio.KeyD,
	key.CodeE:       granio.KeyE,
	key.CodeF:       granio.KeyF,
	key.CodeG:       granio.KeyG,
	key.CodeH:       granio.KeyH,
	key.CodeI:       granio.KeyI,
	key.CodeJ:       granio.KeyJ,
	key.CodeK:       granio.KeyK,
	key.CodeL:       granio.KeyL,
	key.CodeM:       granio.KeyM,
	key.CodeN:       granio.KeyN,
	key.CodeO:       granio.KeyO,
	key.CodeP:       granio.KeyP,
	key.CodeQ:       granio.KeyQ,
	key.CodeR:       granio.KeyR,
	key.CodeS:       granio.KeyS,
	key.CodeT:       granio.KeyT,
	key.CodeU:       granio.KeyU,
	key.CodeV:       granio.KeyV,
	key.CodeW:       granio.KeyW,
	key.CodeX:       granio.KeyX,
	key.CodeY:       granio.KeyY,
	key.CodeZ:       granio.KeyZ,
	key.Code0:       granio.Key0,
	key.CodeKeypad0: granio.Key0,
	key.Code1:       granio.Key1,
	key.CodeKeypad1: granio.Key1,
	key.Code2:       granio.Key2,
	key.CodeKeypad2: granio.Key2,
	key.Code3:       granio.Key3,
	key.CodeKeypad3: granio.Key3,
	key.Code4:       granio.Key4,
	key.CodeKeypad4: granio.Key4,
	key.Code5:       granio.Key5,
	key.CodeKeypad5: granio.Key5,
	key.Code6:       granio.Key6,
	key.CodeKeypad6: granio.Key6,
	key.Code7:       granio.Key7,
	key.CodeKeypad7: granio.Key7,
	key.Code8:       granio.Key8,
	key.CodeKeypad8: granio.Key8,
	key.Code9:       granio.Key9,
	key.CodeKeypad9: granio.Key9,

	key.CodeSemicolon: granio.KeySemicolon,
	key.CodeEqualSign: granio.KeyEqual,

	key.CodeSpacebar:           granio.KeySpace,
	key.CodeApostrophe:         granio.KeyApostrophe,
	key.CodeComma:              granio.KeyComma,
	key.CodeHyphenMinus:        granio.KeyMinus,
	key.CodeKeypadHyphenMinus:  granio.KeyMinus,
	key.CodeFullStop:           granio.KeyPeriod,
	key.CodeKeypadFullStop:     granio.KeyPeriod,
	key.CodeSlash:              granio.KeySlash,
	key.CodeLeftSquareBracket:  granio.KeyLeftBracket,
	key.CodeBackslash:          granio.KeyBackslash,
	key.CodeRightSquareBracket: granio.KeyRightBracket,
	key.CodeGraveAccent:        granio.KeyBackTick,

	key.CodeBackButton: mobile.KeyBack,

	key.CodeLeftShift:    desktop.KeyShiftLeft,
	key.CodeRightShift:   desktop.KeyShiftRight,
	key.CodeLeftControl:  desktop.KeyControlLeft,
	key.CodeRightControl: desktop.KeyControlRight,
	key.CodeLeftAlt:      desktop.KeyAltLeft,
	key.CodeRightAlt:     desktop.KeyAltRight,
	key.CodeLeftGUI:      desktop.KeySuperLeft,
	key.CodeRightGUI:     desktop.KeySuperRight,
}

func keyToName(code key.Code) granio.KeyName {
	ret, ok := keyCodeMap[code]
	if !ok {
		return ""
	}

	return ret
}

func keyModifiers(mod key.Modifiers) granio.KeyModifier {
	var ret granio.KeyModifier
	if mod&key.ModShift != 0 {
		ret |= granio.KeyModifierShift
	}
	if mod&key.ModControl != 0 {
		ret |= granio.KeyModifierControl
	}
	if mod&key.ModAlt != 0 {
		ret |= granio.KeyModifierAlt
	}
	if mod&key.ModMeta != 0 {
		ret |= granio.KeyModifierSuper
	}

	return ret
}

func isKeyModifier(keyName granio.KeyName) bool {
	return keyName == desktop.KeyShiftLeft || keyName == desktop.KeyShiftRight ||
		keyName == desktop.KeyControlLeft || keyName == desktop.KeyControlRight ||
		keyName == desktop.KeyAltLeft || keyName == desktop.KeyAltRight ||
		keyName == desktop.KeySuperLeft || keyName == desktop.KeySuperRight
}

// shortcutForKey mirrors the built-in shortcut detection of the desktop driver.
// All other combinations of a modifier and a key become a desktop.CustomShortcut,
// which widgets such as Entry use for behaviours like word selection or deletion.
func shortcutForKey(keyName granio.KeyName, modifier granio.KeyModifier) granio.Shortcut {
	var shortcut granio.Shortcut
	if modifier == granio.KeyModifierShortcutDefault {
		switch keyName {
		case granio.KeyZ: // detect undo shortcut
			shortcut = &granio.ShortcutUndo{}
		case granio.KeyY: // detect redo shortcut
			shortcut = &granio.ShortcutRedo{}
		case granio.KeyV: // detect paste shortcut
			shortcut = &granio.ShortcutPaste{
				Clipboard: NewClipboard(),
			}
		case granio.KeyC: // detect copy shortcut
			shortcut = &granio.ShortcutCopy{
				Clipboard: NewClipboard(),
			}
		case granio.KeyInsert: // detect copy shortcut (alternative)
			shortcut = &granio.ShortcutCopy{
				Clipboard: NewClipboard(),
				Secondary: true,
			}
		case granio.KeyX: // detect cut shortcut
			shortcut = &granio.ShortcutCut{
				Clipboard: NewClipboard(),
			}
		case granio.KeyA: // detect selectAll shortcut
			shortcut = &granio.ShortcutSelectAll{}
		}
	}

	if modifier == granio.KeyModifierShift {
		switch keyName {
		case granio.KeyInsert: // detect paste shortcut (alternative)
			shortcut = &granio.ShortcutPaste{
				Clipboard: NewClipboard(),
				Secondary: true,
			}
		case granio.KeyDelete: // detect cut shortcut (alternative)
			shortcut = &granio.ShortcutCut{
				Clipboard: NewClipboard(),
				Secondary: true,
			}
		}
	}

	if shortcut == nil && modifier != 0 && !isKeyModifier(keyName) && modifier != granio.KeyModifierShift {
		shortcut = &desktop.CustomShortcut{
			KeyName:  keyName,
			Modifier: modifier,
		}
	}

	return shortcut
}

func runeToPrintable(r rune) rune {
	if strconv.IsPrint(r) {
		return r
	}

	return 0
}

func (d *driver) typeDownCanvas(canvas *canvas, r rune, code key.Code, mod key.Modifiers) {
	keyName := keyToName(code)
	if keyName == granio.KeyTab {
		capture := false
		if ent, ok := canvas.Focused().(granio.Tabbable); ok {
			capture = ent.AcceptsTab()
		}
		if !capture {
			switch mod {
			case 0:
				canvas.FocusNext()
				return
			case key.ModShift:
				canvas.FocusPrevious()
				return
			}
		}
	}

	modifier := keyModifiers(mod)
	r = runeToPrintable(r)
	keyEvent := &granio.KeyEvent{Name: keyName}

	if canvas.Focused() != nil {
		if keyable, ok := canvas.Focused().(desktop.Keyable); ok {
			keyable.KeyDown(keyEvent)
		}

		if shortcut := shortcutForKey(keyName, modifier); shortcut != nil {
			if focusable, ok := canvas.Focused().(granio.Shortcutable); ok {
				focusable.TypedShortcut(shortcut)
				return
			}
		}

		if keyName != "" {
			canvas.Focused().TypedKey(keyEvent)
		}
		if r > 0 {
			canvas.Focused().TypedRune(r)
		}
	} else {
		if keyName != "" {
			if canvas.onTypedKey != nil {
				canvas.onTypedKey(keyEvent)
			} else if keyName == mobile.KeyBack {
				d.GoBack()
			}
		}
		if r > 0 && canvas.onTypedRune != nil {
			canvas.onTypedRune(r)
		}
	}
}

func (*driver) typeUpCanvas(canvas *canvas, _ rune, code key.Code, _ key.Modifiers) {
	if canvas.Focused() == nil {
		return
	}

	keyEvent := &granio.KeyEvent{Name: keyToName(code)}
	if keyable, ok := canvas.Focused().(desktop.Keyable); ok {
		keyable.KeyUp(keyEvent)
	}
}

func (d *driver) Device() granio.Device {
	return &d.device
}

func (d *driver) SetOnConfigurationChanged(f func(*Configuration)) {
	d.onConfigChanged = f
}

func (*driver) DoubleTapDelay() time.Duration {
	return tapDoubleDelay
}

// NewGoMobileDriver sets up a new Driver instance implemented using the Go
// Mobile extension and OpenGL bindings.
func NewGoMobileDriver() granio.Driver {
	d := &driver{
		theme: granio.ThemeVariant(2), // unspecified
	}

	registerRepository(d)
	return d
}
