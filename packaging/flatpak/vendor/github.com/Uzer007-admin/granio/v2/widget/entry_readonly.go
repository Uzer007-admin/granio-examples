package widget

// SetReadOnly changes editing permission without discarding text, selection,
// bindings or undo history. Entering readonly cancels an uncommitted preedit.
func (e *Entry) SetReadOnly(readonly bool) {
	if e.ReadOnly == readonly {
		return
	}
	e.CancelComposition()
	e.ReadOnly = readonly
	e.Refresh()
}
