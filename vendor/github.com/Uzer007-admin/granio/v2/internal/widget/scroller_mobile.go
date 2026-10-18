//go:build ci || no_glfw || android || ios || mobile

package widget

import "github.com/Uzer007-admin/granio/v2"

// DragEnd will stop scrolling on mobile has stopped
func (s *Scroll) DragEnd() {
}

// Dragged will scroll on any drag - bar or otherwise - for mobile
func (s *Scroll) Dragged(e *granio.DragEvent) {
	if s.updateOffset(e.Dragged.DX, e.Dragged.DY) {
		s.refreshWithoutOffsetUpdate()
	}
}
