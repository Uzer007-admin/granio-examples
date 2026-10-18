package theme

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
)

// DefaultEmojiFont returns the font resource for the built-in emoji font.
// This may return nil if the application was packaged without an emoji font.
//
// Since: 2.4
func DefaultEmojiFont() granio.Resource {
	return emoji
}

// DefaultTextBoldFont returns the font resource for the built-in bold font style.
func DefaultTextBoldFont() granio.Resource {
	return bold
}

// DefaultTextBoldItalicFont returns the font resource for the built-in bold and italic font style.
func DefaultTextBoldItalicFont() granio.Resource {
	return bolditalic
}

// DefaultTextFont returns the font resource for the built-in regular font style.
func DefaultTextFont() granio.Resource {
	return regular
}

// DefaultTextItalicFont returns the font resource for the built-in italic font style.
func DefaultTextItalicFont() granio.Resource {
	return italic
}

// DefaultTextMonospaceFont returns the font resource for the built-in monospace font face.
func DefaultTextMonospaceFont() granio.Resource {
	return monospace
}

// DefaultSymbolFont returns the font resource for the built-in symbol font.
//
// Since: 2.2
func DefaultSymbolFont() granio.Resource {
	return symbol
}

// Font looks up the font for current theme and text style.
//
// Since: 2.5
func Font(style granio.TextStyle) granio.Resource {
	return safeFontLookup(style)
}

// TextBoldFont returns the font resource for the bold font style.
func TextBoldFont() granio.Resource {
	return safeFontLookup(granio.TextStyle{Bold: true})
}

// TextBoldItalicFont returns the font resource for the bold and italic font style.
func TextBoldItalicFont() granio.Resource {
	return safeFontLookup(granio.TextStyle{Bold: true, Italic: true})
}

// TextColor returns the theme's standard text color - this is actually the foreground color since 1.4.
//
// Deprecated: Use theme.ForegroundColor() colour instead.
func TextColor() color.Color {
	return safeColorLookup(ColorNameForeground, currentVariant())
}

// TextFont returns the font resource for the regular font style.
func TextFont() granio.Resource {
	return safeFontLookup(granio.TextStyle{})
}

// TextItalicFont returns the font resource for the italic font style.
func TextItalicFont() granio.Resource {
	return safeFontLookup(granio.TextStyle{Italic: true})
}

// TextMonospaceFont returns the font resource for the monospace font face.
func TextMonospaceFont() granio.Resource {
	return safeFontLookup(granio.TextStyle{Monospace: true})
}

// SymbolFont returns the font resource for the symbol font style.
//
// Since: 2.4
func SymbolFont() granio.Resource {
	return safeFontLookup(granio.TextStyle{Symbol: true})
}

func safeFontLookup(s granio.TextStyle) granio.Resource {
	font := Current().Font(s)
	if font != nil {
		return font
	}
	granio.LogError("Loaded theme returned nil font", nil)

	if s.Monospace {
		return DefaultTextMonospaceFont()
	}
	if s.Bold {
		if s.Italic {
			return DefaultTextBoldItalicFont()
		}
		return DefaultTextBoldFont()
	}
	if s.Italic {
		return DefaultTextItalicFont()
	}
	if s.Symbol {
		return DefaultSymbolFont()
	}

	return DefaultTextFont()
}
