package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func restorePopupFocus(c granio.Canvas, focus granio.Focusable) {
	if focus == nil {
		return
	}
	if c.Focused() == focus {
		focus.FocusGained()
	} else {
		c.Focus(focus)
	}
}

func (p *PopUp) animateAppearance() {
	th := p.Theme()
	if _, modern := th.(granio.MotionTheme); !modern || !animationsAllowed(th, granio.MotionSurface) || p.Material != nil {
		return
	}
	r, ok := cache.Renderer(p).(*popUpRenderer)
	if !ok {
		return
	}
	target := r.background.FillColor
	from := th.Color(theme.ColorNameBackground, granio.CurrentApp().Settings().ThemeVariant())
	spec := theme.Motion(th, granio.MotionSurface)
	a := granio.NewAnimation(spec.Duration, func(value float32) {
		r.background.FillColor = tweenColor(from, target, value)
		canvas.Refresh(r.background)
	})
	a.Object, a.Curve = p, spec.Curve
	a.OnFinished = func() {
		if p.appearanceAnimation == a {
			p.appearanceAnimation = nil
		}
	}
	p.appearanceAnimation = a
	a.Start()
}
