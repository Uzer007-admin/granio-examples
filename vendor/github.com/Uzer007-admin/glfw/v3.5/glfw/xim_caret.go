package glfw

import "unicode"

func ximCaret(text []rune, cursor, position, direction int) int {
	cursor = min(max(0, cursor), len(text))
	start, end := cursor, cursor
	for start > 0 && text[start-1] != '\n' {
		start--
	}
	for end < len(text) && text[end] != '\n' {
		end++
	}
	word := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r) || r == '_' }
	switch direction {
	case 0:
		cursor++
	case 1:
		cursor--
	case 2:
		for cursor < len(text) && word(text[cursor]) {
			cursor++
		}
		for cursor < len(text) && !word(text[cursor]) {
			cursor++
		}
	case 3:
		for cursor > 0 && !word(text[cursor-1]) {
			cursor--
		}
		for cursor > 0 && word(text[cursor-1]) {
			cursor--
		}
	case 4, 7:
		if start > 0 {
			previous := start - 1
			for previous > 0 && text[previous-1] != '\n' {
				previous--
			}
			column := cursor - start
			if direction == 7 {
				column = 0
			}
			cursor = min(previous+column, start-1)
		}
	case 5, 6:
		if end < len(text) {
			next := end + 1
			limit := next
			for limit < len(text) && text[limit] != '\n' {
				limit++
			}
			column := cursor - start
			if direction == 6 {
				column = 0
			}
			cursor = min(next+column, limit)
		}
	case 8:
		cursor = start
	case 9:
		cursor = end
	case 10:
		cursor = position
	}
	return min(max(0, cursor), len(text))
}
