package widget

import (
	"github.com/Uzer007-admin/granio/v2/internal/cache"
)

// BeginThemeChange protects the viewport before Fyne refreshes child renderers.
// Those children must not reveal the caret while a theme/DPI/text-scale reflow
// is in progress. Rune offsets survive new wrapping; valid scroll offsets are
// restored, and a smaller document clamps them through the normal Scroll API.
func (e *Entry) BeginThemeChange() func() {
	if e.scroller == nil || e.sel == nil {
		return nil
	}
	offset, previousPreserve := e.scroller.Offset, e.preserveScroll
	text := e.Text
	cursor := e.CursorTextOffset()
	selectRow, selectCol, selecting := e.sel.selectRow, e.sel.selectColumn, e.sel.selecting
	selection := textPosFromRowCol(selectRow, selectCol, e.textProvider())
	e.preserveScroll = true
	return func() {
		if e.Text == text {
			e.CursorRow, e.CursorColumn = e.rowColFromTextPos(cursor)
			e.sel.selectRow, e.sel.selectColumn = e.rowColFromTextPos(selection)
			e.sel.selecting = selecting
			e.syncSelectable()
			if renderer, ok := cache.CachedRenderer(e.content); ok {
				renderer.Refresh()
			}
			e.scroller.ScrollToOffset(offset)
		}
		e.preserveScroll = previousPreserve
	}
}
