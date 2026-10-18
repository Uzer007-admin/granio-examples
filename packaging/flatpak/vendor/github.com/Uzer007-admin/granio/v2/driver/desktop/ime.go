package desktop

import "github.com/Uzer007-admin/granio/v2"

// CompositionEvent describes uncommitted input-method text. All offsets count
// Unicode runes, rather than UTF-8 bytes or platform UTF-16 code units.
type CompositionEvent struct {
	Text                         string
	Cursor                       int
	SelectionStart, SelectionEnd int
}

// TextInputState provides the committed surrounding text and caret geometry to
// a native input method. CaretPosition is relative to the widget. Password input
// omits surrounding text and marks Sensitive so backends can disable learning.
type TextInputState struct {
	Text                                 string
	Cursor, SelectionStart, SelectionEnd int
	CaretPosition                        granio.Position
	CaretSize                            granio.Size
	Sensitive                            bool
}

// TextInput receives composition separately from committed text. Implementations
// must keep preedit out of data bindings, validators and undo history.
type TextInput interface {
	UpdateComposition(CompositionEvent)
	CommitComposition(string)
	CancelComposition()
	CompositionActive() bool
	TextInputState() TextInputState
}

// TextInputEditor supports native replacement of surrounding text. Offsets count
// runes before and after the committed caret, independently of UTF-8 encoding.
type TextInputEditor interface {
	CommitTextEdit(text string, before, after int)
}

// TextInputRangeEditor supports reconversion and dictation replacements outside
// the current selection. Ranges refer to committed text and count Unicode runes.
type TextInputRangeEditor interface {
	ReplaceTextRange(text string, start, end int) bool
	StartCompositionRange(start, end int) bool
}
