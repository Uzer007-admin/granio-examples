package widget

import "github.com/Uzer007-admin/granio/v2"

// One spacing decision is shared by layout, minimum height, caret and selection.
// Explicit semantic runs opt into interline spacing; legacy runs keep their
// historic paragraph-only spacing, including entries' wrapped lines.
func (t *RichText) rowSpacing(row int, legacy float32) float32 {
	if row < 0 || row+1 >= len(t.rowBounds) {
		return 0
	}
	bound := &t.rowBounds[row]
	spacing, semantic := float32(0), false
	for _, segment := range bound.segments {
		if text, ok := segment.(*TextSegment); ok && text.Style.TextRole != granio.TextRoleDefault {
			spacing = max(spacing, text.typography().LineSpacing)
			semantic = true
		}
	}
	if semantic {
		return spacing
	}
	last := rowEndSegment(bound)
	if !last.Inline() && t.rowBounds[row+1].segments[0] != last {
		return legacy
	}
	return 0
}
