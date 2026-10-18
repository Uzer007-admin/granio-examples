package container

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	intWidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// MultipleWindows is a container that handles multiple `InnerWindow` containers.
// Each inner window can be dragged, resized and the stacking will change when the title bar is tapped.
//
// Since: 2.5
type MultipleWindows struct {
	widget.BaseWidget

	Windows []*InnerWindow

	content *granio.Container
}

// NewMultipleWindows creates a new `MultipleWindows` container to manage many inner windows.
// The initial window list is passed optionally to this constructor function.
// You can add new more windows to this container by calling `Add` or updating the `Windows`
// field and calling `Refresh`.
//
// Since: 2.5
func NewMultipleWindows(wins ...*InnerWindow) *MultipleWindows {
	m := &MultipleWindows{Windows: wins}
	m.ExtendBaseWidget(m)
	return m
}

// Add appends an [InnerWindow] to the managed windows.
func (m *MultipleWindows) Add(w *InnerWindow) {
	m.Windows = append(m.Windows, w)
	m.refreshChildren()
}

// CreateRenderer implements the [granio.Widget] interface.
func (m *MultipleWindows) CreateRenderer() granio.WidgetRenderer {
	m.content = New(&multiWinLayout{})
	m.refreshChildren()
	return widget.NewSimpleRenderer(intWidget.NewScroll(m.content))
}

// Refresh implements the [granio.CanvasObject] interface.
func (m *MultipleWindows) Refresh() {
	m.refreshChildren()
	//	m.BaseWidget.Refresh()
}

// RaiseToTop asks this multiple window container to raise a specific inner window above others.
//
// Since: 2.8
func (m *MultipleWindows) RaiseToTop(w *InnerWindow) {
	id := -1
	for i, ww := range m.Windows {
		if ww == w {
			id = i
			break
		}
	}
	if id == -1 {
		return
	}

	windows := append(m.Windows[:id], m.Windows[id+1:]...)
	m.Windows = append(windows, w)
	m.refreshChildren()
}

// Top returns the topmost window in this multiple stack (the one visible above all others).
// If there are no windows it will return nil.
//
// Since: 2.8
func (m *MultipleWindows) Top() *InnerWindow {
	if len(m.Windows) == 0 {
		return nil
	}

	return m.Windows[len(m.Windows)-1]
}

func (m *MultipleWindows) refreshChildren() {
	if m.content == nil {
		return
	}

	objs := make([]granio.CanvasObject, len(m.Windows))
	for i, w := range m.Windows {
		objs[i] = w

		m.setupChild(w)
		w.inactive = i < len(m.Windows)-1
		w.Refresh()
	}
	m.content.Objects = objs
	m.content.Refresh()
}

func (m *MultipleWindows) setupChild(w *InnerWindow) {
	w.OnDragged = func(ev *granio.DragEvent) {
		abs := granio.CurrentApp().Driver().AbsolutePositionForObject(m)
		rel := ev.AbsolutePosition.Subtract(abs)
		if rel.X < 0 || rel.Y < 0 || rel.X > m.Size().Width || rel.Y > m.Size().Height {
			return
		}

		w.Move(w.Position().Add(ev.Dragged))
	}
	w.OnResized = func(ev *granio.DragEvent) {
		size := w.Size().Add(ev.Dragged)
		w.Resize(internal.MaxSizes(size, w.MinSize()))
	}
	w.OnTappedBar = func() {
		m.RaiseToTop(w)
	}
}

type multiWinLayout struct{}

func (*multiWinLayout) Layout(objects []granio.CanvasObject, _ granio.Size) {
	for _, w := range objects { // update the windows so they have real size
		w.Resize(internal.MaxSizes(w.MinSize(), w.Size()))
	}
}

func (*multiWinLayout) MinSize(_ []granio.CanvasObject) granio.Size {
	return granio.Size{}
}
