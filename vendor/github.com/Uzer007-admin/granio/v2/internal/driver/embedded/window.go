package embedded

import (
	"github.com/Uzer007-admin/granio/v2"
)

type noosWindow struct {
	c *touchCanvas
	d *noosDriver

	title string
}

func (w *noosWindow) Title() string {
	return w.title
}

func (w *noosWindow) SetTitle(s string) {
	w.title = s
}

func (*noosWindow) FullScreen() bool {
	return true
}

func (*noosWindow) SetFullScreen(_ bool) {
}

func (w *noosWindow) Resize(s granio.Size) {
	w.c.Resize(s)
}

func (*noosWindow) RequestFocus() {
	// TODO implement me
	panic("implement me")
}

func (*noosWindow) FixedSize() bool {
	return true
}

func (*noosWindow) SetFixedSize(bool) {}

func (*noosWindow) CenterOnScreen() {}

func (w *noosWindow) Padded() bool {
	return w.c.Padded()
}

func (w *noosWindow) SetPadded(pad bool) {
	w.c.SetPadded(pad)
}

func (*noosWindow) Icon() granio.Resource {
	// TODO implement me
	return nil
}

func (*noosWindow) SetIcon(granio.Resource) {
	// TODO implement me
}

func (*noosWindow) SetMaster() {
	// TODO implement me
}

func (*noosWindow) MainMenu() *granio.MainMenu {
	// TODO implement me
	return nil
}

func (*noosWindow) SetMainMenu(*granio.MainMenu) {
	// TODO implement me
}

func (*noosWindow) SetOnClosed(func()) {
	// TODO implement me
}

func (*noosWindow) SetCloseIntercept(func()) {
	// TODO implement me
}

func (*noosWindow) SetOnDropped(func(granio.Position, []granio.URI)) {}

func (w *noosWindow) Show() {
	w.d.renderWindow(w)
}

func (*noosWindow) Hide() {}

func (w *noosWindow) Close() {
	i := -1
	for _, win := range w.d.wins {
		if win == w {
			break
		}
		i++
	}
	if i == -1 {
		return
	}

	copy(w.d.wins[i:], w.d.wins[i+1:])
	w.d.wins[len(w.d.wins)-1] = nil // Allow the garbage collector to reclaim the memory.
	w.d.wins = w.d.wins[:len(w.d.wins)-1]

	if w.d.current > 0 {
		w.d.current--
	}
}

func (w *noosWindow) ShowAndRun() {
	w.Show()
	w.d.Run()
}

func (w *noosWindow) Content() granio.CanvasObject {
	return w.c.Content()
}

func (w *noosWindow) SetContent(object granio.CanvasObject) {
	w.c.SetContent(object)
}

func (w *noosWindow) Canvas() granio.Canvas {
	return w.c
}

func (*noosWindow) Clipboard() granio.Clipboard {
	// TODO implement me
	return nil
}

func newWindow(d *noosDriver) granio.Window {
	return &noosWindow{c: newTouchCanvas(), d: d}
}
