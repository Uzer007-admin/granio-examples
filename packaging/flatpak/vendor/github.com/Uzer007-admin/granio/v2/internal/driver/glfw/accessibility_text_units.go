//go:build accessibility && !wasm && !test_web_driver

package glfw

import (
	"github.com/go-text/typesetting/segmenter"
	"slices"
	"sort"
)

// Shared Unicode boundaries for UIA and AT-SPI. Native UTF-16 conversion happens
// only at the Windows boundary; grapheme movement never splits emoji or accents.
func accessibleTextBoundaries(text string, unit int) []int {
	runes := []rune(text)
	boundaries := []int{0}
	if unit == 0 || unit == 1 {
		var s segmenter.Segmenter
		s.Init(runes)
		if unit == 0 {
			it := s.GraphemeIterator()
			for it.Next() {
				g := it.Grapheme()
				boundaries = append(boundaries, g.Offset+len(g.Text))
			}
		} else {
			it := s.WordIterator()
			for it.Next() {
				w := it.Word()
				boundaries = append(boundaries, w.Offset, w.Offset+len(w.Text))
			}
		}
	} else if unit == 2 || unit == 3 {
		for i, r := range runes {
			if r == '\n' && (unit == 2 || i == 0 || runes[i-1] == '\n') {
				boundaries = append(boundaries, i+1)
			}
		}
	}
	if unit == 5 {
		for i, r := range runes {
			if r == '.' || r == '!' || r == '?' || r == '。' || r == '！' || r == '？' || r == '\n' {
				boundaries = append(boundaries, i+1)
			}
		}
	}
	boundaries = append(boundaries, len(runes))
	slices.Sort(boundaries)
	return slices.Compact(boundaries)
}

func accessibleTextSpan(text string, offset, unit int) (int, int) {
	b := accessibleTextBoundaries(text, unit)
	offset = min(max(offset, 0), b[len(b)-1])
	i := sort.Search(len(b), func(i int) bool { return b[i] > offset })
	if i == len(b) {
		return b[len(b)-1], b[len(b)-1]
	}
	return b[max(0, i-1)], b[i]
}
