package tutorials

import (
	"image/color"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func makeAnimationScreen(_ granio.Window) granio.CanvasObject {
	curves := makeAnimationCurves()
	curves.Move(granio.NewPos(0, 140+theme.Padding()))
	return container.NewWithoutLayout(makeAnimationCanvas(), curves)
}

func makeAnimationCanvas() granio.CanvasObject {
	rect := canvas.NewRectangle(color.Black)
	rect.Resize(granio.NewSize(410, 140))

	a := canvas.NewColorRGBAAnimation(
		color.NRGBA{R: 0x29, G: 0x6f, B: 0xf6, A: 0xaa},
		color.NRGBA{R: 0x8b, G: 0xc3, B: 0x4a, A: 0xaa},
		time.Second*3, func(c color.Color) {
			rect.FillColor = c
			canvas.Refresh(rect)
		},
	)
	a.RepeatCount = granio.AnimationRepeatForever
	a.AutoReverse = true
	a.Start()

	var a2 *granio.Animation
	i := widget.NewIcon(theme.CheckButtonCheckedIcon())
	a2 = canvas.NewPositionAnimation(granio.NewPos(0, 0), granio.NewPos(350, 80), time.Second*3, func(p granio.Position) {
		i.Move(p)

		width := 10 + (p.X / 7)
		i.Resize(granio.NewSize(width, width))
	})
	a2.RepeatCount = granio.AnimationRepeatForever
	a2.AutoReverse = true
	a2.Curve = granio.AnimationLinear
	a2.Start()

	OnChangeFuncs = append(OnChangeFuncs, a.Stop, a2.Stop)

	running := true
	var toggle *widget.Button
	toggle = widget.NewButton("Stop", func() {
		if running {
			a.Stop()
			a2.Stop()
			toggle.SetText("Start")
		} else {
			a.Start()
			a2.Start()
			toggle.SetText("Stop")
		}
		running = !running
	})
	toggle.Resize(toggle.MinSize())
	toggle.Move(granio.NewPos(152, 54))
	return container.NewWithoutLayout(rect, i, toggle)
}

func makeAnimationCurves() granio.CanvasObject {
	label1, box1, a1 := makeAnimationCurveItem("EaseInOut", granio.AnimationEaseInOut, 0)
	label2, box2, a2 := makeAnimationCurveItem("EaseIn", granio.AnimationEaseIn, 30+theme.Padding())
	label3, box3, a3 := makeAnimationCurveItem("EaseOut", granio.AnimationEaseOut, 60+theme.Padding()*2)
	label4, box4, a4 := makeAnimationCurveItem("Linear", granio.AnimationLinear, 90+theme.Padding()*3)

	OnChangeFuncs = append(OnChangeFuncs, a1.Stop, a2.Stop, a3.Stop, a4.Stop)

	start := widget.NewButton("Compare", func() {
		a1.Start()
		a2.Start()
		a3.Start()
		a4.Start()
	})
	start.Resize(start.MinSize())
	start.Move(granio.NewPos(0, 120+theme.Padding()*4))
	return container.NewWithoutLayout(label1, label2, label3, label4, box1, box2, box3, box4, start)
}

func makeAnimationCurveItem(label string, curve granio.AnimationCurve, yOff float32) (
	text *widget.Label, box granio.CanvasObject, anim *granio.Animation,
) {
	text = widget.NewLabel(label)
	text.Alignment = granio.TextAlignCenter
	text.Resize(granio.NewSize(380, 30))
	text.Move(granio.NewPos(0, yOff))
	box = newThemedBox()
	box.Resize(granio.NewSize(30, 30))
	box.Move(granio.NewPos(0, yOff))

	anim = canvas.NewPositionAnimation(
		granio.NewPos(0, yOff), granio.NewPos(380, yOff), time.Second, func(p granio.Position) {
			box.Move(p)
			box.Refresh()
		})
	anim.Curve = curve
	anim.AutoReverse = true
	anim.RepeatCount = 1
	return text, box, anim
}

// themedBox is a simple box that change its background color according
// to the selected theme
type themedBox struct {
	widget.BaseWidget
}

func newThemedBox() *themedBox {
	b := &themedBox{}
	b.ExtendBaseWidget(b)
	return b
}

func (b *themedBox) CreateRenderer() granio.WidgetRenderer {
	b.ExtendBaseWidget(b)
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameForeground))
	return &themedBoxRenderer{bg: bg, objects: []granio.CanvasObject{bg}}
}

type themedBoxRenderer struct {
	bg      *canvas.Rectangle
	objects []granio.CanvasObject
}

func (r *themedBoxRenderer) Destroy() {
}

func (r *themedBoxRenderer) Layout(size granio.Size) {
	r.bg.Resize(size)
}

func (r *themedBoxRenderer) MinSize() granio.Size {
	return r.bg.MinSize()
}

func (r *themedBoxRenderer) Objects() []granio.CanvasObject {
	return r.objects
}

func (r *themedBoxRenderer) Refresh() {
	r.bg.FillColor = theme.Color(theme.ColorNameForeground)
	r.bg.Refresh()
}
