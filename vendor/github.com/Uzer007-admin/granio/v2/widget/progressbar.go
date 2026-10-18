package widget

import (
	"image/color"
	"strconv"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	col "github.com/Uzer007-admin/granio/v2/internal/color"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/theme"
)

type progressRenderer struct {
	widget.BaseRenderer
	background, bar canvas.Rectangle
	label           canvas.Text
	ratio           float32
	progress        *ProgressBar
}

// MinSize calculates the minimum size of a progress bar.
// This is simply the "100%" label size plus padding.
func (p *progressRenderer) MinSize() granio.Size {
	text := "100%"
	if format := p.progress.TextFormatter; format != nil {
		text = format()
	}

	th := p.progress.Theme()
	padding := th.Size(theme.SizeNameInnerPadding) * 2
	size := granio.MeasureText(text, p.label.TextSize, p.label.TextStyle)
	return size.AddWidthHeight(padding, padding)
}

func (p *progressRenderer) calculateRatio() {
	if p.progress.Value < p.progress.Min {
		p.progress.Value = p.progress.Min
	}
	if p.progress.Value > p.progress.Max {
		p.progress.Value = p.progress.Max
	}

	delta := p.progress.Max - p.progress.Min
	p.ratio = float32((p.progress.Value - p.progress.Min) / delta)
}

func (p *progressRenderer) updateBar() {
	p.Layout(p.progress.Size()) // Make sure that bar length updates.

	// Don't draw rectangles when they can't be seen.
	p.background.Hidden = p.ratio == 1.0
	p.bar.Hidden = p.ratio == 0.0

	if text := p.progress.TextFormatter; text != nil {
		p.label.Text = text()
		return
	}

	p.label.Text = strconv.Itoa(int(p.ratio*100)) + "%"
}

// Layout the components of the progress bar widget
func (p *progressRenderer) Layout(size granio.Size) {
	p.calculateRatio()

	p.bar.Resize(granio.NewSize(size.Width*p.ratio, size.Height))
	p.background.Resize(size)
	p.label.Resize(size)
}

// applyTheme updates the progress bar to match the current theme
func (p *progressRenderer) applyTheme() {
	th := p.progress.Theme()
	v := granio.CurrentApp().Settings().ThemeVariant()

	primaryColor := th.Color(theme.ColorNamePrimary, v)
	inputRadius := th.Size(theme.SizeNameInputRadius)

	p.background.FillColor = progressBlendColor(primaryColor)
	p.background.CornerRadius = inputRadius
	p.bar.FillColor = primaryColor
	p.bar.CornerRadius = inputRadius
	p.label.Color = th.Color(theme.ColorNameForegroundOnPrimary, v)
	p.label.TextSize = th.Size(theme.SizeNameText)
}

func (p *progressRenderer) Refresh() {
	p.applyTheme()
	p.updateBar()
	p.background.Refresh()
	p.bar.Refresh()
	p.label.Refresh()
	canvas.Refresh(p.progress.super())
}

// ProgressBar widget creates a horizontal panel that indicates progress
type ProgressBar struct {
	BaseWidget

	Min, Max, Value float64

	// TextFormatter can be used to have a custom format of progress text.
	// If set, it overrides the percentage readout and runs each time the value updates.
	//
	// Since: 1.4
	TextFormatter func() string `json:"-"`

	binder basicBinder
}

// Bind connects the specified data source to this ProgressBar.
// The current value will be displayed and any changes in the data will cause the widget to update.
//
// Since: 2.0
func (p *ProgressBar) Bind(data binding.Float) {
	p.binder.SetCallback(p.updateFromData)
	p.binder.Bind(data)
}

// SetValue changes the current value of this progress bar (from p.Min to p.Max).
// The widget will be refreshed to indicate the change.
func (p *ProgressBar) SetValue(v float64) {
	p.Value = v
	p.Refresh()
}

// MinSize returns the size that this widget should not shrink below
func (p *ProgressBar) MinSize() granio.Size {
	p.ExtendBaseWidget(p)
	return p.BaseWidget.MinSize()
}

// CreateRenderer is a private method to Fyne which links this widget to its renderer
func (p *ProgressBar) CreateRenderer() granio.WidgetRenderer {
	p.ExtendBaseWidget(p)
	if p.Min == 0 && p.Max == 0 {
		p.Max = 1.0
	}

	renderer := &progressRenderer{progress: p}
	renderer.label.Alignment = granio.TextAlignCenter
	renderer.applyTheme()
	renderer.updateBar()

	renderer.SetObjects([]granio.CanvasObject{&renderer.background, &renderer.bar, &renderer.label})
	return renderer
}

// Unbind disconnects any configured data source from this ProgressBar.
// The current value will remain at the last value of the data source.
//
// Since: 2.0
func (p *ProgressBar) Unbind() {
	p.binder.Unbind()
}

// NewProgressBar creates a new progress bar widget.
// The default Min is 0 and Max is 1, Values set should be between those numbers.
// The display will convert this to a percentage.
func NewProgressBar() *ProgressBar {
	bar := &ProgressBar{Min: 0, Max: 1}
	bar.ExtendBaseWidget(bar)
	return bar
}

// NewProgressBarWithData returns a progress bar connected with the specified data source.
//
// Since: 2.0
func NewProgressBarWithData(data binding.Float) *ProgressBar {
	p := NewProgressBar()
	p.Bind(data)
	return p
}

func progressBlendColor(clr color.Color) color.Color {
	r, g, b, a := col.ToNRGBA(clr)
	faded := a / 2
	return &color.NRGBA{R: r, G: g, B: b, A: faded}
}

func (p *ProgressBar) updateFromData(data binding.DataItem) {
	if data == nil {
		return
	}
	floatSource, ok := data.(binding.Float)
	if !ok {
		return
	}

	val, err := floatSource.Get()
	if err != nil {
		granio.LogError("Error getting current data value", err)
		return
	}
	p.SetValue(val)
}
