package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var (
	_ granio.Requireable = (*Entry)(nil)
	_ granio.Validatable = (*Entry)(nil)
)

// HasValue is used for required validation and returns true if the text is not empty.
//
// Since: 2.8
func (e *Entry) HasValue() bool {
	return e.Text != ""
}

// SetOnRequiredChanged is intended for parent widgets or containers to hook into the required state.
// The function might be overwritten by a parent that cares about child state (e.g. widget.Form).
//
// Since: 2.8
func (e *Entry) SetOnRequiredChanged(callback func(bool)) {
	e.onRequiredChanged = callback
}

// Validate validates the current text in the widget.
func (e *Entry) Validate() (err error) {
	if e.Validator != nil {
		err = e.Validator(e.Text)
	}

	e.SetValidationError(err)
	return err
}

func (e *Entry) validateWithoutRefresh() {
	var err error

	if e.Validator != nil {
		err = e.Validator(e.Text)
	}
	e.setValidationErrorWithoutRefresh(err)
}

// SetOnValidationChanged is intended for parent widgets or containers to hook into the validation.
// The function might be overwritten by a parent that cares about child validation (e.g. widget.Form).
func (e *Entry) SetOnValidationChanged(callback func(error)) {
	e.onValidationChanged = callback

	if e.Validator != nil && callback != nil && e.Text != "" {
		if err := e.Validator(e.Text); err != nil {
			callback(err)
		}
	}
}

// SetValidationError manually updates the validation status until the next input change.
func (e *Entry) SetValidationError(err error) {
	if e.setValidationErrorWithoutRefresh(err) {
		e.Refresh()
	}
}

// setValidationErrorWithoutRefresh sets the validation error and returns a bool to indicate if it changes.
// It assumes that the widget has a validator.
func (e *Entry) setValidationErrorWithoutRefresh(err error) bool {
	if e.AlwaysShowValidationError {
		e.validationError = err
		if e.onValidationChanged != nil {
			e.onValidationChanged(err)
		}
		return true
	}

	gone := err == nil
	if !gone && (e.focused || (!e.hasFocused && e.Text == "")) {
		return false
	}

	e.validationError = err
	if e.onValidationChanged != nil {
		e.onValidationChanged(err)
	}

	return true
}

var _ granio.Widget = (*validationStatus)(nil)

type validationStatus struct {
	BaseWidget
	entry *Entry
}

func newValidationStatus(e *Entry) *validationStatus {
	rs := &validationStatus{
		entry: e,
	}

	rs.ExtendBaseWidget(rs)
	return rs
}

func (r *validationStatus) CreateRenderer() granio.WidgetRenderer {
	icon := &canvas.Image{}
	icon.Hide()
	return &validationStatusRenderer{
		WidgetRenderer: NewSimpleRenderer(icon),
		icon:           icon,
		entry:          r.entry,
	}
}

var _ granio.WidgetRenderer = (*validationStatusRenderer)(nil)

type validationStatusRenderer struct {
	granio.WidgetRenderer
	entry *Entry
	icon  *canvas.Image
}

func (r *validationStatusRenderer) Layout(size granio.Size) {
	iconSize := r.entry.Theme().Size(theme.SizeNameInlineIcon)
	r.icon.Resize(granio.NewSquareSize(iconSize))
	r.icon.Move(granio.NewPos((size.Width-iconSize)/2, (size.Height-iconSize)/2))
}

func (r *validationStatusRenderer) MinSize() granio.Size {
	iconSize := r.entry.Theme().Size(theme.SizeNameInlineIcon)
	return granio.NewSquareSize(iconSize)
}

func (r *validationStatusRenderer) Refresh() {
	th := r.entry.Theme()
	if r.entry.Disabled() {
		r.icon.Hide()
		return
	}

	if r.entry.validationError == nil && r.entry.Text != "" && r.entry.Validator != nil {
		r.icon.Resource = th.Icon(theme.IconNameConfirm)
		r.icon.Show()
	} else if r.entry.validationError != nil && !r.entry.focused && (r.entry.dirty || r.entry.AlwaysShowValidationError) {
		r.icon.Resource = theme.NewErrorThemedResource(th.Icon(theme.IconNameError))
		r.icon.Show()
	} else {
		r.icon.Hide()
	}

	r.icon.Refresh()
}
