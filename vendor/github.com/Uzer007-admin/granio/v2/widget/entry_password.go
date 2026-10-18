package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var (
	_ desktop.Cursorable = (*passwordRevealer)(nil)
	_ granio.Tappable    = (*passwordRevealer)(nil)
	_ granio.Widget      = (*passwordRevealer)(nil)
)

type passwordRevealer struct {
	BaseWidget

	icon  *canvas.Image
	entry *Entry
}

func newPasswordRevealer(e *Entry) *passwordRevealer {
	th := e.Theme()
	pr := &passwordRevealer{
		icon:  canvas.NewImageFromResource(th.Icon(theme.IconNameVisibilityOff)),
		entry: e,
	}
	pr.ExtendBaseWidget(pr)
	return pr
}

func (r *passwordRevealer) CreateRenderer() granio.WidgetRenderer {
	return &passwordRevealerRenderer{
		WidgetRenderer: NewSimpleRenderer(r.icon),
		icon:           r.icon,
		entry:          r.entry,
	}
}

func (*passwordRevealer) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}

func (r *passwordRevealer) Tapped(*granio.PointEvent) {
	if r.entry.Disabled() {
		return
	}

	r.entry.setFieldsAndRefresh(func() {
		r.entry.Password = !r.entry.Password
	})
	if c := granio.CurrentApp().Driver().CanvasForObject(r); c != nil {
		c.Focus(r.entry.super().(granio.Focusable))
	}
}

var _ granio.WidgetRenderer = (*passwordRevealerRenderer)(nil)

type passwordRevealerRenderer struct {
	granio.WidgetRenderer
	entry *Entry
	icon  *canvas.Image
}

func (r *passwordRevealerRenderer) Layout(size granio.Size) {
	iconSize := r.entry.Theme().Size(theme.SizeNameInlineIcon)
	r.icon.Resize(granio.NewSquareSize(iconSize))
	r.icon.Move(granio.NewPos((size.Width-iconSize)/2, (size.Height-iconSize)/2))
}

func (r *passwordRevealerRenderer) MinSize() granio.Size {
	iconSize := r.entry.Theme().Size(theme.SizeNameInlineIcon)
	return granio.NewSquareSize(iconSize + r.entry.Theme().Size(theme.SizeNameInnerPadding)*2)
}

func (r *passwordRevealerRenderer) Refresh() {
	th := r.entry.Theme()
	if !r.entry.Password {
		r.icon.Resource = th.Icon(theme.IconNameVisibility)
	} else {
		r.icon.Resource = th.Icon(theme.IconNameVisibilityOff)
	}

	if r.entry.Disabled() {
		r.icon.Resource = theme.NewDisabledResource(r.icon.Resource)
	}
	r.icon.Refresh()
}
