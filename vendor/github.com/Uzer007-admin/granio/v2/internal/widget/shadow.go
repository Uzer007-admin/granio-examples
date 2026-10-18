package widget

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// All known values for ElevationLevel.
const (
	BaseLevel             ElevationLevel = 0
	MenuBarLevel          ElevationLevel = 1
	MenuLevel             ElevationLevel = 4
	ButtonLevel           ElevationLevel = 6
	CardLevel             ElevationLevel = 6
	PopUpLevel            ElevationLevel = 8
	SubmergedContentLevel ElevationLevel = 8
	DialogLevel           ElevationLevel = 24
)

// ShadowType constants
const (
	ShadowAround ShadowType = iota
	ShadowLeft
	ShadowRight
	ShadowBottom
	ShadowTop
)

const (
	shadowBlurRadiusLevelCard    = 8
	shadowBlurRadiusLevelDialog  = 18
	shadowBlurRadiusLevelMenu    = 6
	shadowBlurRadiusLevelMenuBar = 2
	shadowBlurRadiusLevelPopUp   = 14

	shadowVerticalOffsetLevelCard    = 2
	shadowVerticalOffsetLevelDialog  = 6
	shadowVerticalOffsetLevelMenu    = 2
	shadowVerticalOffsetLevelMenuBar = 1
	shadowVerticalOffsetLevelPopUp   = 4

	shadowOffsetRatioHorizontal = 0.2
	shadowOffsetRatioVertical   = shadowOffsetRatioHorizontal * 2
	shadowRadialCenterOffset    = 0.5
)

// ElevationLevel is the level of elevation of the shadow casting object.
type ElevationLevel int

// Shadow is a widget that renders a shadow.
type Shadow struct {
	Base
	level ElevationLevel
	typ   ShadowType
}

var _ granio.Widget = (*Shadow)(nil)

// ShadowType specifies the type of the shadow.
type ShadowType int

// ApplyShadowForLevel applies Material Design inspired shadow parameters (only downward Y-axis offset + blur) to the given shadow.
// The Variant is always [canvas.DropShadow].
func ApplyShadowForLevel(s *canvas.Shadow, level ElevationLevel, shadowColor color.Color) {
	var blurRadius float32
	var offset granio.Position

	switch {
	case level <= BaseLevel:
		// no shadow
	case level <= MenuBarLevel:
		blurRadius = shadowBlurRadiusLevelMenuBar
		offset = granio.NewPos(0, shadowVerticalOffsetLevelMenuBar)
	case level <= MenuLevel:
		blurRadius = shadowBlurRadiusLevelMenu
		offset = granio.NewPos(0, shadowVerticalOffsetLevelMenu)
	case level <= CardLevel: // equal to ButtonLevel
		blurRadius = shadowBlurRadiusLevelCard
		offset = granio.NewPos(0, shadowVerticalOffsetLevelCard)
	case level <= PopUpLevel:
		blurRadius = shadowBlurRadiusLevelPopUp
		offset = granio.NewPos(0, shadowVerticalOffsetLevelPopUp)
	default: // DialogLevel or more
		blurRadius = shadowBlurRadiusLevelDialog
		offset = granio.NewPos(0, shadowVerticalOffsetLevelDialog)
	}

	s.Color = shadowColor
	s.BlurRadius = blurRadius
	s.Offset = offset
	s.Spread = 0
	s.Variant = canvas.DropShadow
}

// NewShadow create a new Shadow.
func NewShadow(typ ShadowType, level ElevationLevel) *Shadow {
	s := &Shadow{typ: typ, level: level}
	s.ExtendBaseWidget(s)
	return s
}

// CreateRenderer returns a new renderer for the shadow.
func (s *Shadow) CreateRenderer() granio.WidgetRenderer {
	r := &shadowRenderer{s: s}
	r.createShadows()
	return r
}

type shadowRenderer struct {
	BaseRenderer
	b, l, r, t     *canvas.LinearGradient
	bl, br, tl, tr *canvas.RadialGradient
	minSize        granio.Size
	s              *Shadow
}

func (r *shadowRenderer) Layout(size granio.Size) {
	depth := float32(r.s.level)
	horizontalOffset, verticalOffset := float32(0.0), float32(0.0)
	if r.s.typ == ShadowAround {
		horizontalOffset = depth * shadowOffsetRatioHorizontal
		verticalOffset = depth * shadowOffsetRatioVertical
	}

	if r.tl != nil {
		r.tl.Resize(granio.NewSize(depth, depth))
		r.tl.Move(granio.NewPos(-depth+horizontalOffset, -depth+verticalOffset))
	}
	if r.t != nil {
		r.t.Resize(granio.NewSize(size.Width-horizontalOffset*2, depth))
		r.t.Move(granio.NewPos(horizontalOffset, -depth+verticalOffset))
	}
	if r.tr != nil {
		r.tr.Resize(granio.NewSize(depth, depth))
		r.tr.Move(granio.NewPos(size.Width-horizontalOffset, -depth+verticalOffset))
	}
	if r.r != nil {
		r.r.Resize(granio.NewSize(depth, size.Height-verticalOffset))
		r.r.Move(granio.NewPos(size.Width-horizontalOffset, verticalOffset))
	}
	if r.br != nil {
		r.br.Resize(granio.NewSize(depth, depth))
		r.br.Move(granio.NewPos(size.Width-horizontalOffset, size.Height))
	}
	if r.b != nil {
		r.b.Resize(granio.NewSize(size.Width-horizontalOffset*2, depth))
		r.b.Move(granio.NewPos(horizontalOffset, size.Height))
	}
	if r.bl != nil {
		r.bl.Resize(granio.NewSize(depth, depth))
		r.bl.Move(granio.NewPos(-depth+horizontalOffset, size.Height))
	}
	if r.l != nil {
		r.l.Resize(granio.NewSize(depth, size.Height-verticalOffset))
		r.l.Move(granio.NewPos(-depth+horizontalOffset, verticalOffset))
	}
}

func (r *shadowRenderer) MinSize() granio.Size {
	return r.minSize
}

func (r *shadowRenderer) Refresh() {
	r.refreshShadows()
	r.Layout(r.s.Size())
	canvas.Refresh(r.s)
}

func (r *shadowRenderer) createShadows() {
	th := theme.CurrentForWidget(r.s)
	v := granio.CurrentApp().Settings().ThemeVariant()
	fg := th.Color(theme.ColorNameShadow, v)

	switch r.s.typ {
	case ShadowLeft:
		r.l = canvas.NewHorizontalGradient(color.Transparent, fg)
		r.SetObjects([]granio.CanvasObject{r.l})
	case ShadowRight:
		r.r = canvas.NewHorizontalGradient(fg, color.Transparent)
		r.SetObjects([]granio.CanvasObject{r.r})
	case ShadowBottom:
		r.b = canvas.NewVerticalGradient(fg, color.Transparent)
		r.SetObjects([]granio.CanvasObject{r.b})
	case ShadowTop:
		r.t = canvas.NewVerticalGradient(color.Transparent, fg)
		r.SetObjects([]granio.CanvasObject{r.t})
	case ShadowAround:
		r.tl = canvas.NewRadialGradient(fg, color.Transparent)
		r.tl.CenterOffsetX = shadowRadialCenterOffset
		r.tl.CenterOffsetY = shadowRadialCenterOffset
		r.t = canvas.NewVerticalGradient(color.Transparent, fg)
		r.tr = canvas.NewRadialGradient(fg, color.Transparent)
		r.tr.CenterOffsetX = -shadowRadialCenterOffset
		r.tr.CenterOffsetY = shadowRadialCenterOffset
		r.r = canvas.NewHorizontalGradient(fg, color.Transparent)
		r.br = canvas.NewRadialGradient(fg, color.Transparent)
		r.br.CenterOffsetX = -shadowRadialCenterOffset
		r.br.CenterOffsetY = -shadowRadialCenterOffset
		r.b = canvas.NewVerticalGradient(fg, color.Transparent)
		r.bl = canvas.NewRadialGradient(fg, color.Transparent)
		r.bl.CenterOffsetX = shadowRadialCenterOffset
		r.bl.CenterOffsetY = -shadowRadialCenterOffset
		r.l = canvas.NewHorizontalGradient(color.Transparent, fg)
		r.SetObjects([]granio.CanvasObject{r.tl, r.t, r.tr, r.r, r.br, r.b, r.bl, r.l})
	}
}

func (r *shadowRenderer) refreshShadows() {
	th := theme.CurrentForWidget(r.s)
	v := granio.CurrentApp().Settings().ThemeVariant()
	fg := th.Color(theme.ColorNameShadow, v)

	updateShadowEnd(r.l, fg)
	updateShadowStart(r.r, fg)
	updateShadowStart(r.b, fg)
	updateShadowEnd(r.t, fg)

	updateShadowRadial(r.tl, fg)
	updateShadowRadial(r.tr, fg)
	updateShadowRadial(r.bl, fg)
	updateShadowRadial(r.br, fg)
}

func updateShadowEnd(g *canvas.LinearGradient, fg color.Color) {
	if g == nil {
		return
	}

	g.EndColor = fg
	g.Refresh()
}

func updateShadowRadial(g *canvas.RadialGradient, fg color.Color) {
	if g == nil {
		return
	}

	g.StartColor = fg
	g.Refresh()
}

func updateShadowStart(g *canvas.LinearGradient, fg color.Color) {
	if g == nil {
		return
	}

	g.StartColor = fg
	g.Refresh()
}
