package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"image/color"
)

func colorsEqual(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}
func tweenColor(a, b color.Color, progress float32) color.Color {
	if a == nil {
		a = color.Transparent
	}
	if b == nil {
		b = color.Transparent
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	lerp := func(x, y uint32) uint16 { return uint16(float32(x) + (float32(y)-float32(x))*progress) }
	return color.RGBA64{R: lerp(ar, br), G: lerp(ag, bg), B: lerp(ab, bb), A: lerp(aa, ba)}
}
func (r *buttonRenderer) updateStateBackground(target color.Color, th granio.Theme) {
	b := r.button
	if _, modern := th.(granio.MotionTheme); !modern || !animationsAllowed(th, granio.MotionHover) || b.Disabled() {
		if b.stateAnimation != nil {
			b.stateAnimation.Stop()
			b.stateAnimation = nil
		}
		b.stateBackground, b.stateTarget, r.background.FillColor = target, target, target
		return
	}
	if b.stateBackground == nil {
		b.stateBackground = target
	}
	if colorsEqual(b.stateTarget, target) {
		r.background.FillColor = b.stateBackground
		return
	}
	if b.stateAnimation != nil {
		b.stateAnimation.Stop()
	}
	from := b.stateBackground
	b.stateTarget = target
	role := granio.MotionHover
	if b.pressed {
		role = granio.MotionPressed
	}
	spec := theme.Motion(th, role)
	a := granio.NewAnimation(spec.Duration, func(value float32) {
		b.stateBackground = tweenColor(from, target, value)
		r.background.FillColor = b.stateBackground
		canvas.Refresh(r.background) // color only: no text/layout/tree reconstruction
	})
	a.Object, a.Curve = b, spec.Curve
	a.OnFinished = func() {
		if b.stateAnimation == a {
			b.stateAnimation = nil
		}
	}
	b.stateAnimation = a
	r.background.FillColor = from
	a.Start()
}

func (r *buttonRenderer) Destroy() {
	if r.button.stateAnimation != nil {
		r.button.stateAnimation.Stop()
		r.button.stateAnimation = nil
	}
	r.button.stateBackground = r.button.stateTarget
	if r.button.tapAnim != nil {
		r.button.tapAnim.Stop()
	}
}
