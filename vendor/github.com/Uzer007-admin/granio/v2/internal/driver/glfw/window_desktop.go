//go:build !wasm && !test_web_driver

package glfw

import (
	"bytes"
	"context"
	"image"
	_ "image/png" // for the icon
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/internal/goos"
	"github.com/Uzer007-admin/granio/v2/internal/painter"
	"github.com/Uzer007-admin/granio/v2/internal/painter/gl"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
	"github.com/Uzer007-admin/granio/v2/internal/svg"
	"github.com/Uzer007-admin/granio/v2/storage"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

const (
	defaultTitle              = "Fyne Application"
	disableDPIDetectionEnvKey = "FYNE_DISABLE_DPI_DETECTION"
)

// Input modes.
const (
	CursorMode             glfw.InputMode = glfw.CursorMode
	StickyKeysMode         glfw.InputMode = glfw.StickyKeysMode
	StickyMouseButtonsMode glfw.InputMode = glfw.StickyMouseButtonsMode
	LockKeyMods            glfw.InputMode = glfw.LockKeyMods
	RawMouseMotion         glfw.InputMode = glfw.RawMouseMotion
)

// Cursor mode values.
const (
	CursorNormal   int = glfw.CursorNormal
	CursorHidden   int = glfw.CursorHidden
	CursorDisabled int = glfw.CursorDisabled
)

var cursors [desktop.HiddenCursor + 1]*glfw.Cursor

func initCursors() {
	cursors = [desktop.HiddenCursor + 1]*glfw.Cursor{
		desktop.DefaultCursor:    glfw.CreateStandardCursor(glfw.ArrowCursor),
		desktop.TextCursor:       glfw.CreateStandardCursor(glfw.IBeamCursor),
		desktop.CrosshairCursor:  glfw.CreateStandardCursor(glfw.CrosshairCursor),
		desktop.PointerCursor:    glfw.CreateStandardCursor(glfw.HandCursor),
		desktop.HResizeCursor:    glfw.CreateStandardCursor(glfw.HResizeCursor),
		desktop.VResizeCursor:    glfw.CreateStandardCursor(glfw.VResizeCursor),
		desktop.NESWResizeCursor: glfw.CreateStandardCursor(glfw.ResizeNESWCursor),
		desktop.NWSEResizeCursor: glfw.CreateStandardCursor(glfw.ResizeNWSECursor),
		desktop.HiddenCursor:     nil,
	}
}

// Declare conformity to Window interface
var _ granio.Window = (*window)(nil)

type window struct {
	visuals     windowVisualState
	viewport    *glfw.Window
	frame       presentGate
	created     bool
	decorate    bool
	closing     bool
	fixedSize   bool
	scalePolicy desktop.ScalePolicy

	cursor       desktop.Cursor
	customCursor *glfw.Cursor
	canvas       *glCanvas
	driver       *gLDriver
	title        string
	icon         granio.Resource
	mainmenu     *granio.MainMenu

	master                          bool
	fullScreen, fullScreenSecondary bool
	centered, visible, onTop        bool

	mousePosUpdateProcessed    bool
	newMousePosX, newMousePosY float64
	mousePos                   granio.Position
	mouseDragged               granio.Draggable
	mouseDraggedObjStart       granio.Position
	mouseDraggedOffset         granio.Position
	mouseDragPos               granio.Position
	mouseDragStarted           bool
	mouseButton                desktop.MouseButton
	mouseOver                  desktop.Hoverable
	mouseLastClick             granio.CanvasObject
	mousePressed               granio.CanvasObject
	mouseClickCount            int
	mouseCancelFunc            context.CancelFunc

	onClosed           func()
	onCloseIntercepted func()
	onDataDropped      func(granio.Position, map[string][]byte)
	onFilesDropped     func(granio.Position, []granio.URI)
	dropFormats        []string

	menuTogglePending       granio.KeyName
	menuDeactivationPending granio.KeyName

	xpos, ypos                      int
	width, height                   int
	requestedWidth, requestedHeight int
	shouldWidth, shouldHeight       int
	shouldExpand                    bool

	pending []func()

	lastWalkedTime   time.Time
	sizeLimitWarning sync.Once
}

func (w *window) SetFullScreen(full bool) {
	w.fullScreen = full
	w.fullScreenSecondary = false

	if w.view() != nil {
		async.EnsureMain(func() {
			w.doSetFullScreen(full)
		})
	}
}

func (w *window) RequestAlwaysOnTop() {
	w.onTop = true

	if w.view() != nil {
		async.EnsureMain(func() {
			w.view().SetAttrib(glfw.Floating, glfw.True)
		})
	}
}

func (w *window) RequestFullScreenSecondary() {
	w.fullScreenSecondary = true
	w.fullScreen = true

	if w.view() != nil {
		async.EnsureMain(func() {
			w.doSetFullScreen2(true)
		})
	}
}

func (w *window) RequestPosition(x, y int) {
	w.xpos = x
	w.ypos = y

	if w.view() != nil {
		w.view().SetPos(x, y)
	}
}

func (w *window) CenterOnScreen() {
	if build.IsWayland {
		return
	}

	w.centered = true

	w.runOnMainWhenCreated(w.doCenterOnScreen)
}

func (w *window) SetOnDropped(dropped func(pos granio.Position, items []granio.URI)) {
	w.runOnMainWhenCreated(func() {
		w.onFilesDropped = dropped
		w.viewport.SetDropCallback(func(_ *glfw.Window, names []string) {
			if dropped == nil {
				return
			}

			uris := make([]granio.URI, len(names))
			for i, name := range names {
				uris[i] = storage.NewFileURI(name)
			}

			w.ensurePositionProcessed()
			dropped(w.mousePos, uris)
		})
		w.updateDataDrop()
	})
}

func (w *window) doCenterOnScreen() {
	viewWidth, viewHeight := w.screenSize(w.canvas.size)
	if w.width > viewWidth { // in case our window has not called back to canvas size yet
		viewWidth = w.width
	}
	if w.height > viewHeight {
		viewHeight = w.height
	}

	// get window dimensions in pixels
	monitor := w.getMonitorForWindow()
	monMode := monitor.GetVideoMode()
	if monMode == nil { // monitor was disconnected
		return
	}

	// these come into play when dealing with multiple monitors
	monX, monY := monitor.GetPos()

	// math them to the middle
	newX := (monMode.Width-viewWidth)/2 + monX
	newY := (monMode.Height-viewHeight)/2 + monY

	// set new window coordinates
	w.viewport.SetPos(newX, newY)
}

func (w *window) RequestFocus() {
	if build.IsWayland || w.view() == nil {
		return
	}

	w.runOnMainWhenCreated(w.viewport.Focus)
}

func (w *window) SetIcon(icon granio.Resource) {
	w.icon = icon
	if build.IsWayland {
		return
	}

	if icon == nil {
		appIcon := granio.CurrentApp().Icon()
		if appIcon != nil {
			w.SetIcon(appIcon)
		}
		return
	}

	w.runOnMainWhenCreated(func() {
		if w.icon == nil {
			w.viewport.SetIcon(nil)
			return
		}

		var img image.Image
		if svg.IsResourceSVG(w.icon) {
			img = painter.PaintImage(&canvas.Image{Resource: w.icon}, nil, windowIconSize, windowIconSize)
		} else {
			pix, _, err := image.Decode(bytes.NewReader(w.icon.Content()))
			if err != nil {
				granio.LogError("Failed to decode image for window icon", err)
				return
			}
			img = pix
		}

		w.viewport.SetIcon([]image.Image{img})
	})
}

func (w *window) SetMaster() {
	w.master = true
}

func (w *window) fitContent() {
	if w.canvas.Content() == nil || (w.fullScreen && w.visible) {
		return
	}

	if w.isClosing() {
		return
	}

	minWidth, minHeight := w.minSizeOnScreen()
	view := w.viewport
	w.shouldWidth, w.shouldHeight = w.width, w.height
	if w.width < minWidth || w.height < minHeight {
		if w.width < minWidth {
			w.shouldWidth = minWidth
		}
		if w.height < minHeight {
			w.shouldHeight = minHeight
		}
		w.shouldExpand = true // queue the resize to happen on main
	}
	if w.fixedSize {
		if w.shouldWidth > w.requestedWidth {
			w.requestedWidth = w.shouldWidth
		}
		if w.shouldHeight > w.requestedHeight {
			w.requestedHeight = w.shouldHeight
		}
		view.SetSizeLimits(w.requestedWidth, w.requestedHeight, w.requestedWidth, w.requestedHeight)
	} else {
		view.SetSizeLimits(minWidth, minHeight, glfw.DontCare, glfw.DontCare)
	}
}

// getMonitorScale returns the scale factor for a given monitor, handling platform-specific cases
func getMonitorScale(monitor *glfw.Monitor) float32 {
	const steamDeckIncorrectlyReportedDisplaySize = 60
	widthMm, heightMm := monitor.GetPhysicalSize()
	if runtime.GOOS == goos.Linux && widthMm == steamDeckIncorrectlyReportedDisplaySize && heightMm == steamDeckIncorrectlyReportedDisplaySize { // Steam Deck incorrectly reports 6cm square!
		return 1.0
	}

	videoMode := monitor.GetVideoMode()
	if videoMode == nil { // monitor was disconnected
		return 1.0
	}
	return calculateDetectedScale(widthMm, videoMode.Width)
}

// getScaledMonitorSize returns the monitor dimensions adjusted for scaling
func getScaledMonitorSize(monitor *glfw.Monitor) granio.Size {
	videoMode := monitor.GetVideoMode()
	if videoMode == nil { // monitor was disconnected
		return granio.NewSize(0, 0)
	}
	s := getMonitorScale(monitor)

	scaledWidth := float32(videoMode.Width) / s
	scaledHeight := float32(videoMode.Height) / s
	return granio.NewSize(scaledWidth, scaledHeight)
}

func (w *window) getMonitorForWindow() *glfw.Monitor {
	if !build.IsWayland {
		x, y := w.xpos, w.ypos
		if w.fullScreen {
			x, y = w.viewport.GetPos()
		}
		xOff := x + (w.width / 2)
		yOff := y + (w.height / 2)

		for _, monitor := range glfw.GetMonitors() {
			x, y := monitor.GetPos()

			if x > xOff || y > yOff {
				continue
			}

			scaledSize := getScaledMonitorSize(monitor)
			if x+int(scaledSize.Width) <= xOff || y+int(scaledSize.Height) <= yOff {
				continue
			}

			return monitor
		}
	}

	// try built-in function to detect monitor if above logic didn't succeed
	// if it doesn't work then return primary monitor as default
	monitor := w.viewport.GetMonitor()
	if monitor == nil {
		monitor = glfw.GetPrimaryMonitor()
	}
	return monitor
}

// findSiblingMonitor returns the monitor of an already-visible window in this app, or nil.
func (w *window) findSiblingMonitor() *glfw.Monitor {
	for _, other := range w.driver.AllWindows() {
		ow, ok := other.(*window)
		if !ok || ow == w || !ow.visible || ow.viewport == nil {
			continue
		}
		return ow.getMonitorForWindow()
	}
	return nil
}

func (w *window) detectScale() float32 {
	if build.IsWayland { // Wayland controls scale through content scaling
		return 1
	}

	// check if DPI detection is disabled
	env := os.Getenv(disableDPIDetectionEnvKey)
	if strings.EqualFold(env, "true") || strings.EqualFold(env, "t") || env == "1" {
		return 1
	}

	monitor := w.getMonitorForWindow()
	if monitor == nil {
		return 1
	}

	return getMonitorScale(monitor)
}

func (w *window) moved(_ *glfw.Window, x, y int) {
	w.processMoved(x, y)
}

func (w *window) resized(_ *glfw.Window, width, height int) {
	w.processResized(width, height)
}

func (w *window) scaled(_ *glfw.Window, x, _ float32) {
	if w.closing || !validScale(x) {
		return
	}
	if build.IsWayland {
		w.canvas.texScale = x
		w.canvas.Refresh(w.canvas.content)
	} else {
		// GLFW owns WM_DPICHANGED and its suggested native rectangle. The
		// content-scale callback also fires when a stationary window changes DPI.
		w.canvas.detectedScale = w.detectScale()
		w.canvas.reloadScale()
	}
}

func (w *window) frameSized(_ *glfw.Window, width, height int) {
	w.processFrameSized(width, height)
}

func (w *window) refresh(_ *glfw.Window) {
	w.processRefresh()
}

func (w *window) closed(viewport *glfw.Window) {
	if viewport != nil {
		viewport.SetShouldClose(false) // reset the closed flag until we check the veto in processClosed
	}

	w.processClosed()
}

func fyneToNativeCursor(cursor desktop.Cursor) (*glfw.Cursor, bool) {
	cursorType, standard := cursor.(desktop.StandardCursor)
	if !standard {
		img, x, y := cursor.Image()
		if img == nil {
			return nil, true
		}
		return glfw.CreateCursor(img, x, y), true
	}

	if cursorType < 0 || cursorType >= desktop.StandardCursor(len(cursors)) {
		return cursors[desktop.DefaultCursor], false
	}

	return cursors[cursorType], false
}

func (w *window) SetCursor(cursor *glfw.Cursor) {
	async.EnsureMain(func() {
		w.viewport.SetCursor(cursor)
	})
}

func (w *window) setCustomCursor(rawCursor *glfw.Cursor, isCustomCursor bool) {
	if w.customCursor != nil {
		w.customCursor.Destroy()
		w.customCursor = nil
	}
	if isCustomCursor {
		w.customCursor = rawCursor
	}
}

func (w *window) mouseMoved(_ *glfw.Window, xpos, ypos float64) {
	w.newMousePosX = xpos
	w.newMousePosY = ypos
	w.mousePosUpdateProcessed = false
}

func (w *window) mouseClicked(_ *glfw.Window, btn glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	button, modifiers := convertMouseButton(btn, mods)
	mouseAction := convertAction(action)

	w.processMouseClicked(button, mouseAction, modifiers)
}

func (w *window) mouseScrolled(viewport *glfw.Window, xoff float64, yoff float64) {
	if runtime.GOOS != goos.Darwin && xoff == 0 &&
		(viewport.GetKey(glfw.KeyLeftShift) == glfw.Press ||
			viewport.GetKey(glfw.KeyRightShift) == glfw.Press) {
		xoff, yoff = yoff, xoff
	}

	w.processMouseScrolled(xoff, yoff)
}

func convertMouseButton(btn glfw.MouseButton, mods glfw.ModifierKey) (desktop.MouseButton, granio.KeyModifier) {
	modifier := desktopModifier(mods)
	rightClick := false
	if runtime.GOOS == goos.Darwin {
		if modifier&granio.KeyModifierControl != 0 {
			rightClick = true
			modifier &^= granio.KeyModifierControl
		}
		if modifier&granio.KeyModifierSuper != 0 {
			modifier |= granio.KeyModifierControl
			modifier &^= granio.KeyModifierSuper
		}
	}

	switch btn {
	case glfw.MouseButton1:
		if rightClick {
			return desktop.MouseButtonSecondary, modifier
		}
		return desktop.MouseButtonPrimary, modifier
	case glfw.MouseButton2:
		return desktop.MouseButtonSecondary, modifier
	case glfw.MouseButton3:
		return desktop.MouseButtonTertiary, modifier
	default:
		return 0, modifier
	}
}

//gocyclo:ignore
func glfwKeyToKeyName(key glfw.Key) granio.KeyName {
	switch key {
	// numbers - lookup by code to avoid AZERTY using the symbol name instead of number
	case glfw.Key0, glfw.KeyKP0:
		return granio.Key0
	case glfw.Key1, glfw.KeyKP1:
		return granio.Key1
	case glfw.Key2, glfw.KeyKP2:
		return granio.Key2
	case glfw.Key3, glfw.KeyKP3:
		return granio.Key3
	case glfw.Key4, glfw.KeyKP4:
		return granio.Key4
	case glfw.Key5, glfw.KeyKP5:
		return granio.Key5
	case glfw.Key6, glfw.KeyKP6:
		return granio.Key6
	case glfw.Key7, glfw.KeyKP7:
		return granio.Key7
	case glfw.Key8, glfw.KeyKP8:
		return granio.Key8
	case glfw.Key9, glfw.KeyKP9:
		return granio.Key9

	// non-printable
	case glfw.KeyEscape:
		return granio.KeyEscape
	case glfw.KeyEnter:
		return granio.KeyReturn
	case glfw.KeyTab:
		return granio.KeyTab
	case glfw.KeyBackspace:
		return granio.KeyBackspace
	case glfw.KeyInsert:
		return granio.KeyInsert
	case glfw.KeyDelete:
		return granio.KeyDelete
	case glfw.KeyRight:
		return granio.KeyRight
	case glfw.KeyLeft:
		return granio.KeyLeft
	case glfw.KeyDown:
		return granio.KeyDown
	case glfw.KeyUp:
		return granio.KeyUp
	case glfw.KeyPageUp:
		return granio.KeyPageUp
	case glfw.KeyPageDown:
		return granio.KeyPageDown
	case glfw.KeyHome:
		return granio.KeyHome
	case glfw.KeyEnd:
		return granio.KeyEnd

	case glfw.KeySpace:
		return granio.KeySpace
	case glfw.KeyKPEnter:
		return granio.KeyEnter

	// desktop
	case glfw.KeyLeftShift:
		return desktop.KeyShiftLeft
	case glfw.KeyRightShift:
		return desktop.KeyShiftRight
	case glfw.KeyLeftControl:
		return desktop.KeyControlLeft
	case glfw.KeyRightControl:
		return desktop.KeyControlRight
	case glfw.KeyLeftAlt:
		return desktop.KeyAltLeft
	case glfw.KeyRightAlt:
		return desktop.KeyAltRight
	case glfw.KeyLeftSuper:
		return desktop.KeySuperLeft
	case glfw.KeyRightSuper:
		return desktop.KeySuperRight
	case glfw.KeyMenu:
		return desktop.KeyMenu
	case glfw.KeyPrintScreen:
		return desktop.KeyPrintScreen
	case glfw.KeyCapsLock:
		return desktop.KeyCapsLock

	// functions
	case glfw.KeyF1:
		return granio.KeyF1
	case glfw.KeyF2:
		return granio.KeyF2
	case glfw.KeyF3:
		return granio.KeyF3
	case glfw.KeyF4:
		return granio.KeyF4
	case glfw.KeyF5:
		return granio.KeyF5
	case glfw.KeyF6:
		return granio.KeyF6
	case glfw.KeyF7:
		return granio.KeyF7
	case glfw.KeyF8:
		return granio.KeyF8
	case glfw.KeyF9:
		return granio.KeyF9
	case glfw.KeyF10:
		return granio.KeyF10
	case glfw.KeyF11:
		return granio.KeyF11
	case glfw.KeyF12:
		return granio.KeyF12
	}

	return granio.KeyUnknown
}

func keyCodeToKeyName(code string) granio.KeyName {
	if len(code) != 1 {
		return granio.KeyUnknown
	}

	char := code[0]
	if char >= 'a' && char <= 'z' {
		return granio.KeyName(char ^ ('a' - 'A')) // Corresponding KeyName is uppercase. Convert with simple bit flip.
	}

	switch char {
	case '[':
		return granio.KeyLeftBracket
	case '\\':
		return granio.KeyBackslash
	case ']':
		return granio.KeyRightBracket
	case '\'':
		return granio.KeyApostrophe
	case ',':
		return granio.KeyComma
	case '-':
		return granio.KeyMinus
	case '.':
		return granio.KeyPeriod
	case '/':
		return granio.KeySlash
	case '*':
		return granio.KeyAsterisk
	case '`':
		return granio.KeyBackTick
	case ';':
		return granio.KeySemicolon
	case '+':
		return granio.KeyPlus
	case '=':
		return granio.KeyEqual
	}

	return granio.KeyUnknown
}

func keyToName(code glfw.Key, scancode int) granio.KeyName {
	ret := glfwKeyToKeyName(code)
	if ret != granio.KeyUnknown {
		return ret
	}

	keyName := safeGetKeyName(code, scancode)
	return keyCodeToKeyName(keyName)
}

func safeGetKeyName(key glfw.Key, scancode int) string {
	defer func() {
		if r := recover(); r != nil {
			err, _ := r.(error)
			granio.LogError("Failed to get GLFW key name", err)
		}
	}()
	return glfw.GetKeyName(key, scancode)
}

func convertAction(action glfw.Action) action {
	switch action {
	case glfw.Press:
		return press
	case glfw.Release:
		return release
	case glfw.Repeat:
		return repeat
	}
	panic("Could not convert glfw.Action.")
}

func convertASCII(key glfw.Key) granio.KeyName {
	if key < glfw.KeyA || key > glfw.KeyZ {
		return granio.KeyUnknown
	}

	return granio.KeyName(rune(key))
}

func (w *window) keyPressed(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	keyName := keyToName(key, scancode)
	keyDesktopModifier := desktopModifier(mods)
	w.driver.currentKeyModifiers = desktopModifierCorrected(mods, key, action)
	keyAction := convertAction(action)
	keyASCII := convertASCII(key)

	w.processKeyPressed(keyName, keyASCII, scancode, keyAction, keyDesktopModifier)
}

func desktopModifier(mods glfw.ModifierKey) granio.KeyModifier {
	var m granio.KeyModifier
	if (mods & glfw.ModShift) != 0 {
		m |= granio.KeyModifierShift
	}
	if (mods & glfw.ModControl) != 0 {
		m |= granio.KeyModifierControl
	}
	if (mods & glfw.ModAlt) != 0 {
		m |= granio.KeyModifierAlt
	}
	if (mods & glfw.ModSuper) != 0 {
		m |= granio.KeyModifierSuper
	}
	return m
}

func desktopModifierCorrected(mods glfw.ModifierKey, key glfw.Key, action glfw.Action) granio.KeyModifier {
	// On X11, pressing/releasing modifier keys does not include newly pressed/released keys in 'mod' mask.
	// https://github.com/glfw/glfw/issues/1630
	if action == glfw.Press || action == glfw.Repeat {
		mods |= glfwKeyToModifier(key)
	} else if action == glfw.Release {
		mods &= ^glfwKeyToModifier(key)
	}
	return desktopModifier(mods)
}

func glfwKeyToModifier(key glfw.Key) glfw.ModifierKey {
	switch key {
	case glfw.KeyLeftControl, glfw.KeyRightControl:
		return glfw.ModControl
	case glfw.KeyLeftAlt, glfw.KeyRightAlt:
		return glfw.ModAlt
	case glfw.KeyLeftShift, glfw.KeyRightShift:
		return glfw.ModShift
	case glfw.KeyLeftSuper, glfw.KeyRightSuper:
		return glfw.ModSuper
	default:
		return 0
	}
}

// charInput defines the character with modifiers callback which is called when a
// Unicode character is input.
//
// Characters do not map 1:1 to physical keys, as a key may produce zero, one or more characters.
func (w *window) charInput(_ *glfw.Window, char rune) {
	w.processCharInput(char)
}

func (w *window) focused(_ *glfw.Window, focused bool) {
	w.processFocused(focused)
}

func (*window) DetachCurrentContext() {
	glfw.DetachCurrentContext()
}

func (w *window) RescaleContext() {
	if w.isClosing() {
		return
	}
	w.fitContent()

	if w.fullScreen {
		w.width, w.height = w.viewport.GetSize()
		scaledFull := granio.NewSize(
			scale.ToFyneCoordinate(w.canvas, w.width),
			scale.ToFyneCoordinate(w.canvas, w.height),
		)
		w.canvas.Resize(scaledFull)
		return
	}

	size := internal.MaxSizes(w.canvas.size, w.canvas.MinSize())
	newWidth, newHeight := w.screenSize(size)
	w.viewport.SetSize(newWidth, newHeight)

	// Ensure textures re-rasterize at the new scale
	w.RunWithContext(func() {
		cache.DeleteTextTexturesFor(w.canvas)
	})
	w.canvas.content.Refresh()
}

func (w *window) create() {
	const fallbackScreenSize = 10

	// make the window hidden, we will set it up and then show it later
	glfw.WindowHint(glfw.Visible, glfw.False)
	if w.decorate {
		glfw.WindowHint(glfw.Decorated, glfw.True)
	} else {
		glfw.WindowHint(glfw.Decorated, glfw.False)
	}
	if w.fixedSize {
		glfw.WindowHint(glfw.Resizable, glfw.False)
	} else {
		glfw.WindowHint(glfw.Resizable, glfw.True)
	}
	if w.onTop {
		glfw.WindowHint(glfw.Floating, glfw.True)
	} else {
		glfw.WindowHint(glfw.Floating, glfw.False)
	}
	glfw.WindowHint(glfw.AutoIconify, glfw.False)
	initWindowHints()
	alphaSurface := w.beginVisualCreation()
	if alphaSurface {
		glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	} else {
		glfw.WindowHint(glfw.TransparentFramebuffer, glfw.False)
	}
	if build.IsWayland {
		glfw.WindowHintString(glfw.WaylandAppID, granio.CurrentApp().UniqueID())
	}

	pixWidth, pixHeight := w.screenSize(w.canvas.size)
	pixWidth = int(max(float32(pixWidth), float32(w.width)))
	if pixWidth == 0 {
		pixWidth = fallbackScreenSize
	}
	pixHeight = int(max(float32(pixHeight), float32(w.height)))
	if pixHeight == 0 {
		pixHeight = fallbackScreenSize
	}

	win, err := glfw.CreateWindow(pixWidth, pixHeight, w.title, nil, nil)
	// TransparentFramebuffer is a global GLFW hint. Never leak our reservation
	// into another application window, including after creation failure.
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.False)
	if err != nil {
		w.failVisualCreation()
		w.driver.initFailed("window creation error", err)
		return
	}

	w.viewport = win
	if w.view() == nil { // something went wrong above, it will have been logged
		return
	}

	// macOS 26 places new windows on a different screen than existing app windows;
	// default new windows onto the same monitor as a visible sibling when no position was set.
	if runtime.GOOS == goos.Darwin && !build.IsWayland && w.xpos == 0 && w.ypos == 0 {
		if monitor := w.findSiblingMonitor(); monitor != nil {
			monMode := monitor.GetVideoMode()
			if monMode != nil { // monitor was disconnected
				monX, monY := monitor.GetPos()
				w.xpos = monX + (monMode.Width-pixWidth)/2
				w.ypos = monY + (monMode.Height-pixHeight)/2
			}
		}
	}

	if (w.xpos != 0 || w.ypos != 0) && !build.IsWayland {
		win.SetPos(w.xpos, w.ypos)
	}

	// run the GL init on the draw thread
	w.RunWithContext(func() {
		w.canvas.SetPainter(gl.NewPainter(w.canvas, w))
		w.canvas.Painter().Init()

		if build.IsWayland {
			glfw.SwapInterval(0)
		}
	})

	w.setDarkMode()

	win.SetCloseCallback(w.closed)
	win.SetPosCallback(w.moved)
	win.SetSizeCallback(w.resized)
	win.SetFramebufferSizeCallback(w.frameSized)
	win.SetRefreshCallback(w.refresh)
	win.SetContentScaleCallback(w.scaled)
	win.SetCursorPosCallback(w.mouseMoved)
	win.SetMouseButtonCallback(w.mouseClicked)
	win.SetScrollCallback(w.mouseScrolled)
	win.SetKeyCallback(w.keyPressed)
	win.SetCharCallback(w.charInput)
	win.SetFocusCallback(w.focused)
	w.initInputMethod()

	w.canvas.detectedScale = w.detectScale()
	w.canvas.scale = w.calculatedScale()
	w.canvas.texScale = w.detectTextureScale()
	// update window size now we have scaled detected
	w.fitContent()

	w.drainPendingEvents()

	if w.FixedSize() && (w.requestedWidth == 0 || w.requestedHeight == 0) {
		bigEnough := w.canvas.canvasSize(w.canvas.Content().MinSize())
		w.width, w.height = w.screenSize(bigEnough)
		w.shouldWidth, w.shouldHeight = w.width, w.height
	}

	w.requestedWidth, w.requestedHeight = w.width, w.height
	// order of operation matters so we do these last items in order
	w.viewport.SetSize(w.shouldWidth, w.shouldHeight) // ensure we requested latest size

	// Initialize accessibility support
	w.initAccessibilityForWindow()
	w.applyWindowVisuals()
}

func (w *window) view() *glfw.Window {
	if w.closing {
		return nil
	}
	return w.viewport
}
