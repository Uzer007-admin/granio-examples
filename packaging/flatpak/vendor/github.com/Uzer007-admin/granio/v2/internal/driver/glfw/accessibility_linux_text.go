//go:build accessibility && linux && !wasm && !test_web_driver

package glfw

import (
	"unicode/utf8"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2"
)

type atspiTextObject struct{ *atspiObject }

func (o *atspiTextObject) GetText(start, end int32) (string, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return "", err
	}
	runes := []rune(e.node.value)
	if end == -1 {
		end = int32(len(runes))
	}
	if start < 0 || end < start || int(end) > len(runes) {
		return "", dbus.NewError("org.a11y.atspi.Error.InvalidArguments", nil)
	}
	return string(runes[start:end]), nil
}
func (o *atspiTextObject) GetCharacterAtOffset(offset int32) (int32, *dbus.Error) {
	value, err := o.GetText(offset, offset+1)
	if err != nil {
		return -1, err
	}
	char, _ := utf8.DecodeRuneInString(value)
	return int32(char), nil
}
func (o *atspiTextObject) GetStringAtOffset(offset int32, granularity uint32) (string, int32, int32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return "", 0, 0, err
	}
	runes := []rune(e.node.value)
	if offset < 0 || int(offset) >= len(runes) {
		return "", offset, offset, nil
	}
	kind := int(granularity)
	if granularity == 2 {
		kind = 5
	} else if granularity >= 3 {
		kind = int(granularity) - 1
	}
	start, end := accessibleTextSpan(e.node.value, int(offset), kind)
	return string(runes[start:end]), int32(start), int32(end), nil
}
func (o *atspiTextObject) GetTextAtOffset(offset int32, boundary uint32) (string, int32, int32, *dbus.Error) {
	granularity := uint32(0)
	if boundary == 1 || boundary == 2 {
		granularity = 1
	} else if boundary == 3 || boundary == 4 {
		granularity = 2
	} else if boundary >= 5 {
		granularity = 3
	}
	return o.GetStringAtOffset(offset, granularity)
}
func (o *atspiTextObject) GetNSelections() (int32, *dbus.Error) {
	e, err := o.element()
	if e.node.selectionStart != e.node.selectionEnd {
		return 1, err
	}
	return 0, err
}
func (o *atspiTextObject) GetSelection(index int32) (int32, int32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return 0, 0, err
	}
	if index != 0 {
		return 0, 0, dbus.NewError("org.a11y.atspi.Error.InvalidIndex", nil)
	}
	return int32(e.node.selectionStart), int32(e.node.selectionEnd), nil
}
func (o *atspiTextObject) edit(action func(granio.AccessibleEditableText) bool) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	if e.node.state.Disabled {
		return false, nil
	}
	if _, ok := e.node.object.(granio.AccessibleEditableText); !ok {
		return false, nil
	}
	applied := false
	accessibilityUI(func() {
		if current, err := o.element(); err == nil && current.node.object == e.node.object {
			applied = action(e.node.object.(granio.AccessibleEditableText))
			if applied {
				e.window.updateAccessibility()
			}
		}
	})
	return applied, nil
}
func (o *atspiTextObject) SetTextContents(text string) (bool, *dbus.Error) {
	return o.edit(func(e granio.AccessibleEditableText) bool { return e.AccessibilitySetText(text) })
}
func (o *atspiTextObject) SetCaretOffset(offset int32) (bool, *dbus.Error) {
	return o.selectRange(offset, offset)
}
func (o *atspiTextObject) selectRange(start, end int32) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	if start < 0 || end < start || int(end) > utf8.RuneCountInString(e.node.value) {
		return false, nil
	}
	return o.edit(func(e granio.AccessibleEditableText) bool { return e.AccessibilitySetSelection(int(start), int(end)) })
}
func (o *atspiTextObject) AddSelection(start, end int32) (bool, *dbus.Error) {
	return o.selectRange(start, end)
}
func (o *atspiTextObject) SetSelection(index, start, end int32) (bool, *dbus.Error) {
	if index != 0 {
		return false, nil
	}
	return o.selectRange(start, end)
}
func (o *atspiTextObject) RemoveSelection(index int32) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil || index != 0 {
		return false, err
	}
	return o.selectRange(int32(e.node.cursor), int32(e.node.cursor))
}
func (o *atspiTextObject) InsertText(position int32, text string, length int32) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	runes := []rune(e.node.value)
	inserted := []rune(text)
	if position < 0 || int(position) > len(runes) || length < 0 || int(length) > len(inserted) {
		return false, nil
	}
	value := string(runes[:position]) + string(inserted[:length]) + string(runes[position:])
	return o.SetTextContents(value)
}
func (o *atspiTextObject) DeleteText(start, end int32) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	runes := []rune(e.node.value)
	if start < 0 || end < start || int(end) > len(runes) {
		return false, nil
	}
	return o.SetTextContents(string(runes[:start]) + string(runes[end:]))
}
func (o *atspiTextObject) CopyText(start, end int32) *dbus.Error {
	text, err := o.GetText(start, end)
	if err == nil {
		queueAccessibility(func() { NewClipboard().SetContent(text) })
	}
	return err
}
func (o *atspiTextObject) CutText(start, end int32) (bool, *dbus.Error) {
	if err := o.CopyText(start, end); err != nil {
		return false, err
	}
	return o.DeleteText(start, end)
}
func (o *atspiTextObject) PasteText(position int32) (bool, *dbus.Error) {
	return o.edit(func(e granio.AccessibleEditableText) bool {
		if !e.AccessibilitySetSelection(int(position), int(position)) {
			return false
		}
		text := NewClipboard().Content()
		value := []rune(e.AccessibilityValue())
		if int(position) > len(value) {
			return false
		}
		return e.AccessibilitySetText(string(value[:position]) + text + string(value[position:]))
	})
}
