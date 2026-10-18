package test

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/software"
)

type window struct {
	title              string
	fullScreen         bool
	fixedSize          bool
	focused            bool
	onClosed           func()
	onCloseIntercepted func()

	canvas software.WindowlessCanvas
	driver *driver
	menu   *granio.MainMenu
}

// NewWindow creates and registers a new window for test purposes
func NewWindow(content granio.CanvasObject) granio.Window {
	window := granio.CurrentApp().NewWindow("")
	window.SetContent(content)
	return window
}

func (w *window) Canvas() granio.Canvas {
	return w.canvas
}

func (*window) CenterOnScreen() {
	// no-op
}

func (*window) Clipboard() granio.Clipboard {
	return NewClipboard()
}

func (w *window) Close() {
	if w.onClosed != nil {
		w.onClosed()
	}
	w.focused = false
	w.driver.removeWindow(w)
}

func (w *window) Content() granio.CanvasObject {
	return w.Canvas().Content()
}

func (w *window) FixedSize() bool {
	return w.fixedSize
}

func (w *window) FullScreen() bool {
	return w.fullScreen
}

func (w *window) Hide() {
	w.focused = false
}

func (*window) Icon() granio.Resource {
	return granio.CurrentApp().Icon()
}

func (w *window) MainMenu() *granio.MainMenu {
	return w.menu
}

func (w *window) Padded() bool {
	return w.canvas.Padded()
}

func (w *window) RequestFocus() {
	for _, win := range w.driver.AllWindows() {
		win.(*window).focused = false
	}

	w.focused = true
}

func (w *window) Resize(size granio.Size) {
	w.canvas.Resize(size)
}

func (w *window) SetContent(obj granio.CanvasObject) {
	w.Canvas().SetContent(obj)
}

func (w *window) SetFixedSize(fixed bool) {
	w.fixedSize = fixed
}

func (*window) SetIcon(_ granio.Resource) {
	// no-op
}

func (w *window) SetFullScreen(fullScreen bool) {
	w.fullScreen = fullScreen
}

func (w *window) SetMainMenu(menu *granio.MainMenu) {
	w.menu = menu
}

func (*window) SetMaster() {
	// no-op
}

func (w *window) SetOnClosed(closed func()) {
	w.onClosed = closed
}

func (w *window) SetCloseIntercept(callback func()) {
	w.onCloseIntercepted = callback
}

func (*window) SetOnDropped(func(granio.Position, []granio.URI)) {
}

func (w *window) SetPadded(padded bool) {
	w.canvas.SetPadded(padded)
}

func (w *window) SetTitle(title string) {
	w.title = title
}

func (w *window) Show() {
	w.RequestFocus()
}

func (w *window) ShowAndRun() {
	w.Show()
}

func (w *window) Title() string {
	return w.title
}
