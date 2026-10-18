package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
	"unicode/utf8"
)

func (e *Entry) AccessibilityScrollTextRange(start, end int, alignTop bool) bool {
	if e.Password || start < 0 || end < start || end > utf8.RuneCountInString(e.Text) {
		return false
	}
	if e.scroller == nil {
		return true
	}
	provider, th := e.textProvider(), e.Theme()
	pad := th.Size(theme.SizeNameInnerPadding)
	row, col := e.rowColFromTextPos(start)
	if !alignTop {
		row, col = e.rowColFromTextPos(end)
	}
	y, height := provider.rowGeometry(row)
	if height <= 0 {
		height = provider.charMinSize(false, e.TextStyle, th.Size(theme.SizeNameText)).Height
	}
	x := provider.lineSizeToColumn(col, row, th.Size(theme.SizeNameText), pad).Width
	offset := e.scroller.Offset
	size := e.scroller.Size()
	if x < offset.X {
		offset.X = x
	} else if x >= offset.X+size.Width {
		offset.X = x - size.Width + 1
	}
	offset.Y = y
	if !alignTop {
		offset.Y = y + height + 2*pad - size.Height
	}
	e.scroller.ScrollToOffset(granio.NewPos(max(0, offset.X), max(0, offset.Y)))
	return true
}

func (e *Entry) AccessibilityTextRectangles(start, end int) []granio.AccessibleTextRectangle {
	if e.Password {
		return nil
	}
	count := utf8.RuneCountInString(e.Text)
	if start < 0 || end < start || end > count {
		return nil
	}
	provider, th := e.textProvider(), e.Theme()
	pad, border, textSize := th.Size(theme.SizeNameInnerPadding), th.Size(theme.SizeNameInputBorder), th.Size(theme.SizeNameText)
	offset, origin, viewport := granio.Position{}, granio.Position{}, e.Size()
	if e.scroller != nil {
		offset, origin, viewport = e.scroller.Offset, e.scroller.Position(), e.scroller.Size()
	}
	first, firstCol := e.rowColFromTextPos(start)
	last, lastCol := e.rowColFromTextPos(end)
	result := make([]granio.AccessibleTextRectangle, 0, min(32, last-first+1))
	for row := first; row <= last && len(result) < 4096; row++ {
		a, b := 0, provider.rowLength(row)
		if row == first {
			a = firstCol
		}
		if row == last {
			b = lastCol
		}
		left := provider.lineSizeToColumn(a, row, textSize, pad).Width - border/2
		right := provider.lineSizeToColumn(b, row, textSize, pad).Width - border/2
		y, height := provider.rowGeometry(row)
		if height <= 0 {
			height = provider.charMinSize(e.Password, e.TextStyle, textSize).Height
		}
		left, right = left-offset.X, right-offset.X
		top, bottom := y+pad-border-offset.Y, y+pad-border-offset.Y+height
		if right < 0 || left > viewport.Width || bottom <= 0 || top >= viewport.Height {
			continue
		}
		left, right = max(0, left), min(viewport.Width, max(right, left+1))
		top, bottom = max(0, top), min(viewport.Height, bottom)
		if right > left && bottom > top {
			result = append(result, granio.AccessibleTextRectangle{Position: origin.Add(granio.NewPos(left, top)), Size: granio.NewSize(right-left, bottom-top)})
		}
	}
	return result
}

func (e *Entry) AccessibilityTextOffsetAtPosition(position granio.Position) int {
	if e.Password {
		return 0
	}
	if e.scroller != nil {
		position = position.Subtract(e.scroller.Position()).Add(e.scroller.Offset)
	}
	provider, th := e.textProvider(), e.Theme()
	pad, size := th.Size(theme.SizeNameInnerPadding), th.Size(theme.SizeNameText)
	row := 0
	for row+1 < provider.rows() {
		y, height := provider.rowGeometry(row)
		if position.Y < y+pad+height {
			break
		}
		row++
	}
	low, high := 0, provider.rowLength(row)
	for low < high {
		mid := (low + high) / 2
		if provider.lineSizeToColumn(mid, row, size, pad).Width < position.X {
			low = mid + 1
		} else {
			high = mid
		}
	}
	if low > 0 {
		a, b := provider.lineSizeToColumn(low-1, row, size, pad).Width, provider.lineSizeToColumn(low, row, size, pad).Width
		if position.X < (a+b)/2 {
			low--
		}
	}
	return textPosFromRowCol(row, low, provider)
}
