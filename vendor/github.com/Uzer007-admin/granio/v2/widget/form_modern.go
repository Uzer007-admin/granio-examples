package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func thComponent(th granio.Theme) (granio.ComponentTheme, bool) {
	provider, ok := th.(granio.ComponentTheme)
	return provider, ok
}

func (f *Form) updateWrappedHelper(item *FormItem) {
	if item.helperContainer == nil {
		return
	}
	if _, modern := thComponent(f.Theme()); !modern {
		item.helperContainer.Objects = []granio.CanvasObject{item.helperOutput}
		return
	}
	if item.helperWrapped == nil {
		item.helperWrapped = NewRichTextWithText("")
		item.helperWrapped.Wrapping = granio.TextWrapWord
	}
	segment := item.helperWrapped.Segments[0].(*TextSegment)
	segment.Text = item.helperOutput.Text
	segment.Style.TextRole = granio.TextRoleCaption
	segment.Style.ColorName = theme.ColorNamePlaceHolder
	if item.validationError != nil && segment.Text == item.validationError.Error() {
		segment.Style.ColorName = theme.ColorNameError
	}
	if item.helperOutput.Visible() {
		item.helperWrapped.Show()
	} else {
		item.helperWrapped.Hide()
	}
	item.helperContainer.Objects = []granio.CanvasObject{item.helperWrapped}
	item.helperWrapped.Refresh()
}

// Modern forms constrain their label column and measure wrapped labels/errors
// at their actual width before assigning row heights. No editable widget is
// replaced, so focus, binding and selection survive a profile change.
type modernFormLayout struct{ form *Form }

func (l modernFormLayout) columns(objects []granio.CanvasObject, width float32) (float32, float32) {
	th := l.form.Theme()
	pad := th.Size(theme.SizeNamePadding)
	label := min(width*.36, 180*th.Size(theme.SizeNameText)/14)
	return max(0, label), max(0, width-label-pad)
}

func (l modernFormLayout) measure(objects []granio.CanvasObject, width float32, place bool) granio.Size {
	labelWidth, inputWidth := l.columns(objects, width)
	pad := l.form.Theme().Size(theme.SizeNamePadding)
	y, rows := float32(0), 0
	for i := 0; i+1 < len(objects); i += 2 {
		label, input := objects[i], objects[i+1]
		if !label.Visible() && !input.Visible() {
			continue
		}
		label.Resize(granio.NewSize(labelWidth, label.MinSize().Height))
		input.Resize(granio.NewSize(inputWidth, input.MinSize().Height))
		height := max(label.MinSize().Height, input.MinSize().Height)
		if place {
			label.Move(granio.NewPos(0, y))
			input.Move(granio.NewPos(labelWidth+pad, y))
			label.Resize(granio.NewSize(labelWidth, height))
			input.Resize(granio.NewSize(inputWidth, height))
		}
		y += height + pad
		rows++
	}
	if rows != 0 {
		y -= pad
	}
	return granio.NewSize(width, y)
}

func (l modernFormLayout) Layout(objects []granio.CanvasObject, size granio.Size) {
	l.measure(objects, size.Width, true)
}

func (l modernFormLayout) MinSize(objects []granio.CanvasObject) granio.Size {
	minimum := float32(240)
	for i := 1; i < len(objects); i += 2 {
		minimum = max(minimum, objects[i].MinSize().Width/.64)
	}
	result := l.measure(objects, max(minimum, l.form.itemGrid.Size().Width), false)
	result.Width = minimum
	return result
}
