package widget

import (
	"strings"
	"unicode/utf8"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
)

var _ desktop.TextInput = (*Entry)(nil)
var _ desktop.TextInputRangeEditor = (*Entry)(nil)

// StartCompositionRange selects the committed range requested for reconversion.
// A cancelled composition restores the selection that preceded this request.
func (e *Entry) StartCompositionRange(start, end int) bool {
	if e.Disabled() || e.ReadOnly || e.Password || start < 0 || end < start || end > utf8.RuneCountInString(e.Text) {
		return false
	}
	e.CancelComposition()
	e.syncSelectable()
	row, col, selectRow, selectCol, selecting := e.CursorRow, e.CursorColumn, e.sel.selectRow, e.sel.selectColumn, e.sel.selecting
	e.AccessibilitySetSelection(start, end)
	e.UpdateComposition(desktop.CompositionEvent{})
	e.composition.cursorRow, e.composition.cursorColumn = row, col
	e.composition.selectRow, e.composition.selectColumn, e.composition.selecting = selectRow, selectCol, selecting
	return true
}

// ReplaceTextRange applies a native edit as one Undo operation.
func (e *Entry) ReplaceTextRange(text string, start, end int) bool {
	if e.ReadOnly || e.Disabled() {
		return false
	}
	e.CancelComposition()
	if !e.AccessibilitySetSelection(start, end) {
		return false
	}
	if text == "" {
		e.TypedKey(&granio.KeyEvent{Name: granio.KeyBackspace})
	} else {
		e.TypedText(text)
	}
	return true
}

type entryComposition struct {
	text                                             string
	start, end                                       int
	cursorRow, cursorColumn, selectRow, selectColumn int
	selecting                                        bool
	segments                                         []RichTextSegment
	preedit                                          string
	cursor                                           int
	selectionStart, selectionEnd                     int
}

// CompositionActive reports whether uncommitted input is displayed.
func (e *Entry) CompositionActive() bool { return e.composition != nil }

// UpdateComposition displays preedit without modifying Text, bindings, undo
// history or validation. Changing candidates replaces the same captured range.
func (e *Entry) UpdateComposition(event desktop.CompositionEvent) {
	if e.Disabled() || e.ReadOnly {
		e.CancelComposition()
		return
	}
	e.syncSelectable()
	if e.composition == nil {
		start, end := e.CursorTextOffset(), e.CursorTextOffset()
		if e.sel.selecting {
			start, end = e.sel.selection()
		}
		e.composition = &entryComposition{text: e.Text, start: start, end: end,
			cursorRow: e.CursorRow, cursorColumn: e.CursorColumn, selectRow: e.sel.selectRow, selectColumn: e.sel.selectColumn,
			selecting: e.sel.selecting, segments: e.textWidget.Segments}
	}
	state := e.composition
	state.preedit, state.cursor = event.Text, min(max(0, event.Cursor), utf8.RuneCountInString(event.Text))
	state.selectionStart = min(max(0, event.SelectionStart), utf8.RuneCountInString(event.Text))
	state.selectionEnd = min(max(state.selectionStart, event.SelectionEnd), utf8.RuneCountInString(event.Text))
	e.sel.selecting = false
	e.syncSegments()
	e.textWidget.updateRowBounds()
	e.CursorRow, e.CursorColumn = e.rowColFromTextPos(state.start + state.cursor)
	e.syncSelectable()
	e.Refresh()
}

func (e *Entry) compositionSegments() []RichTextSegment {
	state := e.composition
	runes := []rune(state.text)
	start, end := min(state.start, len(runes)), min(state.end, len(runes))
	style := RichTextStyleInline
	style.ColorName, style.TextStyle, style.concealed = theme.ColorNameForeground, e.TextStyle, e.Password
	preeditStyle := style
	preeditStyle.TextStyle.Underline = true
	if state.selectionStart != state.selectionEnd {
		preedit := []rune(state.preedit)
		selectedStyle := preeditStyle
		selectedStyle.TextStyle.Bold = true
		selectedStyle.ColorName = theme.ColorNamePrimary
		return []RichTextSegment{&TextSegment{Text: string(runes[:start]), Style: style}, &TextSegment{Text: string(preedit[:state.selectionStart]), Style: preeditStyle}, &TextSegment{Text: string(preedit[state.selectionStart:state.selectionEnd]), Style: selectedStyle}, &TextSegment{Text: string(preedit[state.selectionEnd:]), Style: preeditStyle}, &TextSegment{Text: string(runes[end:]), Style: style}}
	}
	return []RichTextSegment{
		&TextSegment{Text: string(runes[:start]), Style: style},
		&TextSegment{Text: state.preedit, Style: preeditStyle},
		&TextSegment{Text: string(runes[end:]), Style: style},
	}
}

func (e *Entry) restoreComposition() {
	state := e.composition
	if state == nil {
		return
	}
	e.composition = nil
	e.textWidget.Segments = state.segments
	e.CursorRow, e.CursorColumn = state.cursorRow, state.cursorColumn
	e.sel.selectRow, e.sel.selectColumn, e.sel.selecting = state.selectRow, state.selectColumn, state.selecting
	e.syncSegments()
	e.textWidget.updateRowBounds()
	e.syncSelectable()
}

// CancelComposition removes preedit and restores the original caret/selection.
func (e *Entry) CancelComposition() {
	if e.composition == nil {
		return
	}
	e.restoreComposition()
	e.Refresh()
}

// CommitComposition inserts the final result as one change and one undo action.
func (e *Entry) CommitComposition(text string) {
	if e.ReadOnly || e.Disabled() {
		e.CancelComposition()
		return
	}
	e.restoreComposition()
	if text == "" {
		e.Refresh()
		return
	}
	e.TypedText(text)
}

// CommitTextEdit applies a native surrounding-text edit as one undoable change.
func (e *Entry) CommitTextEdit(text string, before, after int) {
	if e.ReadOnly || e.Disabled() {
		e.CancelComposition()
		return
	}
	e.restoreComposition()
	if before > 0 || after > 0 {
		cursor := e.CursorTextOffset()
		e.AccessibilitySetSelection(max(0, cursor-before), min(utf8.RuneCountInString(e.Text), cursor+after))
	}
	if text == "" {
		if e.sel.selecting {
			e.TypedKey(&granio.KeyEvent{Name: granio.KeyBackspace})
		}
		return
	}
	e.TypedText(text)
}

// TypedText inserts committed text as a single undo action. It also replaces an
// active selection atomically, and sends OnChanged once for the final value.
func (e *Entry) TypedText(text string) {
	if e.Disabled() || e.ReadOnly || text == "" {
		return
	}
	e.restoreComposition()
	e.syncSelectable()
	if !e.MultiLine {
		text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", " "), "\n", " ")
	}
	start, end := e.CursorTextOffset(), e.CursorTextOffset()
	if e.sel.selecting {
		start, end = e.sel.selection()
	}
	provider := e.textProvider()
	removed := provider.deleteFromTo(start, end)
	inserted := []rune(text)
	provider.insertAt(start, inserted)
	e.undoStack.Add(&entryReplaceAction{position: start, removed: removed, inserted: inserted})
	e.updateText(provider.String(), false)
	e.CursorRow, e.CursorColumn = e.rowColFromTextPos(start + len(inserted))
	e.sel.selecting = false
	e.syncSelectable()
	e.validateWithoutRefresh()
	if e.OnChanged != nil {
		e.OnChanged(e.Text)
	}
	e.Refresh()
}

// TextInputState returns the surrounding text and caret location for native IME.
func (e *Entry) TextInputState() desktop.TextInputState {
	e.syncSelectable()
	state := desktop.TextInputState{Text: e.Text, Cursor: e.CursorTextOffset(),
		CaretPosition: e.CursorPosition(), CaretSize: granio.NewSize(1, theme.SizeForWidget(theme.SizeNameText, e)), Sensitive: e.Password}
	if e.scroller != nil {
		state.CaretPosition = state.CaretPosition.Add(e.scroller.Position()).Subtract(e.scroller.Offset)
	}
	state.SelectionStart, state.SelectionEnd = state.Cursor, state.Cursor
	if e.composition != nil {
		state.Cursor = e.composition.start
		state.SelectionStart, state.SelectionEnd = e.composition.start, e.composition.end
	} else if e.sel.selecting {
		state.SelectionStart, state.SelectionEnd = e.sel.selection()
	}
	if e.Password {
		state.Text, state.Cursor, state.SelectionStart, state.SelectionEnd = "", 0, 0, 0
	}
	return state
}

type entryReplaceAction struct {
	position          int
	removed, inserted []rune
}

func (a *entryReplaceAction) Undo(text string) string { return a.replace(text, a.inserted, a.removed) }
func (a *entryReplaceAction) Redo(text string) string { return a.replace(text, a.removed, a.inserted) }
func (a *entryReplaceAction) replace(text string, remove, insert []rune) string {
	runes := []rune(text)
	return string(runes[:a.position]) + string(insert) + string(runes[a.position+len(remove):])
}
