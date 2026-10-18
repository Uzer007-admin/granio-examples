package widget

import (
	"image/color"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	fynecolor "github.com/Uzer007-admin/granio/v2/internal/color"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var _ granio.Widget = (*Activity)(nil)

// Activity is used to indicate that something is happening that should be waited for,
// or is in the background (depending on usage).
//
// Since: 2.5
type Activity struct {
	BaseWidget

	started bool
}

// NewActivity returns a widget for indicating activity
//
// Since: 2.5
func NewActivity() *Activity {
	a := &Activity{}
	a.ExtendBaseWidget(a)
	return a
}

// MinSize implements the [granio.CanvasObject] interface.
func (a *Activity) MinSize() granio.Size {
	a.ExtendBaseWidget(a)
	return a.BaseWidget.MinSize()
}

// Start the activity indicator animation
func (a *Activity) Start() {
	if a.started {
		return // already started
	}

	a.started = true

	a.Refresh()
}

// Stop the activity indicator animation
func (a *Activity) Stop() {
	if !a.started {
		return // already stopped
	}

	a.started = false

	a.Refresh()
}

// CreateRenderer implements the [granio.Widget] interface.
func (a *Activity) CreateRenderer() granio.WidgetRenderer {
	dots := make([]granio.CanvasObject, 3)
	v := granio.CurrentApp().Settings().ThemeVariant()
	for i := range dots {
		dots[i] = canvas.NewCircle(a.Theme().Color(theme.ColorNameForeground, v))
	}
	r := &activityRenderer{dots: dots, parent: a}
	r.anim = &granio.Animation{
		Duration:    time.Second * 2,
		RepeatCount: granio.AnimationRepeatForever,
		Tick:        r.animate,
	}
	r.updateColor()

	if a.started {
		r.start()
	}

	return r
}

var _ granio.WidgetRenderer = (*activityRenderer)(nil)

type activityRenderer struct {
	anim   *granio.Animation
	dots   []granio.CanvasObject
	parent *Activity

	bound      granio.Size
	maxCol     color.NRGBA
	maxRad     float32
	wasStarted bool
}

func (a *activityRenderer) Destroy() {
	a.parent.started = false
	a.stop()
}

func (a *activityRenderer) Layout(size granio.Size) {
	a.maxRad = min(size.Width, size.Height) / 2
	a.bound = size

	if a.parent.started && !granio.CurrentApp().Settings().ShowAnimations() {
		a.drawStaticEllipsis()
	}
}

func (a *activityRenderer) MinSize() granio.Size {
	return granio.NewSquareSize(a.parent.Theme().Size(theme.SizeNameInlineIcon))
}

func (a *activityRenderer) Objects() []granio.CanvasObject {
	return a.dots
}

func (a *activityRenderer) Refresh() {
	if a.parent.started {
		if !a.wasStarted {
			a.start()
		}
	} else if a.wasStarted {
		a.stop()
	}

	a.updateColor()
}

func (a *activityRenderer) animate(done float32) {
	off := done * 2
	if off > 1 {
		off = 2 - off
	}

	off1 := (done + 0.25) * 2
	if done >= 0.75 {
		off1 = (done - 0.75) * 2
	}
	if off1 > 1 {
		off1 = 2 - off1
	}

	off2 := (done + 0.75) * 2
	if done >= 0.25 {
		off2 = (done - 0.25) * 2
	}
	if off2 > 1 {
		off2 = 2 - off2
	}

	a.scaleDot(a.dots[0].(*canvas.Circle), off)
	a.scaleDot(a.dots[1].(*canvas.Circle), off1)
	a.scaleDot(a.dots[2].(*canvas.Circle), off2)
}

func (a *activityRenderer) scaleDot(dot *canvas.Circle, off float32) {
	rad := a.maxRad - a.maxRad*off/1.2 //revive:disable-line:add-constant
	mid := granio.NewPos(a.bound.Width/2, a.bound.Height/2)

	dot.Move(mid.Subtract(granio.NewSquareOffsetPos(rad)))
	dot.Resize(granio.NewSquareSize(rad * 2))

	alpha := uint8(float32(a.maxCol.A) * off)
	dot.FillColor = color.NRGBA{R: a.maxCol.R, G: a.maxCol.G, B: a.maxCol.B, A: alpha}
	dot.Refresh()
}

func (a *activityRenderer) start() {
	a.wasStarted = true
	if !granio.CurrentApp().Settings().ShowAnimations() {
		a.drawStaticEllipsis()
		return
	}
	a.anim.Start()
}

func (a *activityRenderer) stop() {
	a.wasStarted = false
	a.anim.Stop()
	if !granio.CurrentApp().Settings().ShowAnimations() {
		a.hideDots()
	}
}

// drawStaticEllipsis places the three dots in a horizontal row, like an
// ellipsis. Used when animations are disabled so the widget still indicates
// something is happening without any motion.
func (a *activityRenderer) drawStaticEllipsis() {
	th := a.parent.Theme()
	innerPad := th.Size(theme.SizeNameInnerPadding)
	d := min(a.bound.Width/4, a.bound.Height)
	if d > th.Size(theme.SizeNameInlineIcon)/2 {
		d -= innerPad
	}
	if d <= 0 {
		a.hideDots()
		return
	}
	radius := d / 2
	totalW := 4 * d
	startX := (a.bound.Width - totalW) / 2
	cy := a.bound.Height / 2
	fill := color.NRGBA{R: a.maxCol.R, G: a.maxCol.G, B: a.maxCol.B, A: a.maxCol.A}
	for i, obj := range a.dots {
		dot, _ := obj.(*canvas.Circle)
		cx := startX + radius + float32(i)*3*radius
		dot.Move(granio.NewPos(cx-radius, cy-radius))
		dot.Resize(granio.NewSquareSize(d))
		dot.FillColor = fill
		dot.Refresh()
	}
}

func (a *activityRenderer) hideDots() {
	for _, obj := range a.dots {
		dot, _ := obj.(*canvas.Circle)
		dot.Resize(granio.NewSquareSize(0))
		dot.Refresh()
	}
}

func (a *activityRenderer) updateColor() {
	v := granio.CurrentApp().Settings().ThemeVariant()
	rr, gg, bb, aa := fynecolor.ToNRGBA(a.parent.Theme().Color(theme.ColorNameForeground, v))
	a.maxCol = color.NRGBA{R: rr, G: gg, B: bb, A: aa}
}
