package theme

import (
	"math"

	"github.com/Uzer007-admin/granio/v2"
)

const (
	// SizeNameCaptionText is the name of theme lookup for helper text size, normally smaller than regular text size.
	//
	// Since: 2.0
	SizeNameCaptionText granio.ThemeSizeName = "helperText"

	// SizeNameInlineIcon is the name of theme lookup for inline icons size.
	//
	// Since: 2.0
	SizeNameInlineIcon granio.ThemeSizeName = "iconInline"

	// SizeNameInnerPadding is the name of theme lookup for internal widget padding size.
	//
	// Since: 2.3
	SizeNameInnerPadding granio.ThemeSizeName = "innerPadding"

	// SizeNameInnerWindowRadius is the name of theme lookup for inner window corner radius.
	//
	// Since: 2.8
	SizeNameInnerWindowRadius granio.ThemeSizeName = "innerWindowRadius"

	// SizeNameLineSpacing is the name of theme lookup for between text line spacing.
	//
	// Since: 2.3
	SizeNameLineSpacing granio.ThemeSizeName = "lineSpacing"

	// SizeNamePadding is the name of theme lookup for padding size.
	//
	// Since: 2.0
	SizeNamePadding granio.ThemeSizeName = "padding"

	// SizeNameScrollBar is the name of theme lookup for the scrollbar size.
	//
	// Since: 2.0
	SizeNameScrollBar granio.ThemeSizeName = "scrollBar"

	// SizeNameScrollBarSmall is the name of theme lookup for the shrunk scrollbar size.
	//
	// Since: 2.0
	SizeNameScrollBarSmall granio.ThemeSizeName = "scrollBarSmall"

	// SizeNameSeparatorThickness is the name of theme lookup for the thickness of a separator.
	//
	// Since: 2.0
	SizeNameSeparatorThickness granio.ThemeSizeName = "separator"

	// SizeNameSplitThickness is the name of theme lookup for the thickness of a split divider.
	//
	// Since: 2.8
	SizeNameSplitThickness granio.ThemeSizeName = "split"

	// SizeNameText is the name of theme lookup for text size.
	//
	// Since: 2.0
	SizeNameText granio.ThemeSizeName = "text"

	// SizeNameHeadingText is the name of theme lookup for text size of a heading.
	//
	// Since: 2.1
	SizeNameHeadingText granio.ThemeSizeName = "headingText"

	// SizeNameSubHeadingText is the name of theme lookup for text size of a sub-heading.
	//
	// Since: 2.1
	SizeNameSubHeadingText granio.ThemeSizeName = "subHeadingText"

	// SizeNameInputBorder is the name of theme lookup for input border size.
	//
	// Since: 2.0
	SizeNameInputBorder granio.ThemeSizeName = "inputBorder"

	// SizeNameInputRadius is the name of theme lookup for input corner radius.
	//
	// Since: 2.4
	SizeNameInputRadius granio.ThemeSizeName = "inputRadius"

	// SizeNameModalBlurRadius is the name of theme lookup for blurring modal popup backgrounds.
	//
	// Since: 2.8
	SizeNameModalBlurRadius granio.ThemeSizeName = "modalBlurRadius"

	// SizeNameSelectionRadius is the name of theme lookup for selection corner radius.
	//
	// Since: 2.4
	SizeNameSelectionRadius granio.ThemeSizeName = "selectionRadius"

	// SizeNameScrollBarRadius is the name of theme lookup for the scroll bar corner radius.
	//
	// Since: 2.5
	SizeNameScrollBarRadius granio.ThemeSizeName = "scrollBarRadius"

	// SizeNameWindowButtonHeight is the name of the height for an inner window titleBar button.
	//
	// Since: 2.6
	SizeNameWindowButtonHeight granio.ThemeSizeName = "windowButtonHeight"

	// SizeNameWindowButtonRadius is the name of the radius for an inner window titleBar button.
	//
	// Since: 2.6
	SizeNameWindowButtonRadius granio.ThemeSizeName = "windowButtonRadius"

	// SizeNameWindowButtonIcon is the name of the width of an inner window titleBar button.
	//
	// Since: 2.6
	SizeNameWindowButtonIcon granio.ThemeSizeName = "windowButtonIcon"

	// SizeNameWindowShadowActiveRadius controls the overall size of inner window shadows when the
	// window is active (has focus).
	//
	// Since: 2.9
	SizeNameWindowShadowActiveRadius granio.ThemeSizeName = "windowShadowActiveRadius"

	// SizeNameWindowShadowRadius controls the overall size of inner window shadows.
	//
	// Since: 2.9
	SizeNameWindowShadowRadius granio.ThemeSizeName = "windowShadowRadius"

	// SizeNameWindowTitleBarHeight is the height for inner window titleBars.
	//
	// Since: 2.6
	SizeNameWindowTitleBarHeight granio.ThemeSizeName = "windowTitleBarHeight"

	// SizeNameButtonRadius is the name of theme lookup for button corner radius.
	//
	// Since: 2.8
	SizeNameButtonRadius granio.ThemeSizeName = "buttonRadius"

	// SizeNameCardRadius is the name of theme lookup for card widget corner radius.
	//
	// Since: 2.8
	SizeNameCardRadius granio.ThemeSizeName = "cardRadius"

	// SizeNameDialogRadius is the name of theme lookup for dialog and modal popup corner radius.
	//
	// Since: 2.8
	SizeNameDialogRadius granio.ThemeSizeName = "dialogRadius"

	// SizeNamePopupRadius is the name of theme lookup for popup corner radius.
	//
	// Since: 2.8
	SizeNamePopupRadius granio.ThemeSizeName = "popupRadius"

	// SizeNameMenuRadius is the name of theme lookup for menu corner radius.
	//
	// Since: 2.8
	SizeNameMenuRadius granio.ThemeSizeName = "menuRadius"
)

// CaptionTextSize returns the size for caption text.
func CaptionTextSize() float32 {
	return Current().Size(SizeNameCaptionText)
}

// IconInlineSize is the standard size of icons which appear within buttons, labels etc.
func IconInlineSize() float32 {
	return Current().Size(SizeNameInlineIcon)
}

// InnerPadding is the standard gap between element content and the outside edge of a widget.
//
// Since: 2.3
func InnerPadding() float32 {
	return Current().Size(SizeNameInnerPadding)
}

// InputBorderSize returns the input border size (or underline size for an entry).
//
// Since: 2.0
func InputBorderSize() float32 {
	return Current().Size(SizeNameInputBorder)
}

// InputRadiusSize returns the input radius size.
//
// Since: 2.4
func InputRadiusSize() float32 {
	return Current().Size(SizeNameInputRadius)
}

// LineSpacing is the default gap between multiple lines of text.
//
// Since: 2.3
func LineSpacing() float32 {
	return Current().Size(SizeNameLineSpacing)
}

// Padding is the standard gap between elements and the border around interface elements.
func Padding() float32 {
	return Current().Size(SizeNamePadding)
}

// ScrollBarSize is the width (or height) of the bars on a ScrollContainer.
func ScrollBarSize() float32 {
	return Current().Size(SizeNameScrollBar)
}

// ScrollBarSmallSize is the width (or height) of the minimized bars on a ScrollContainer.
func ScrollBarSmallSize() float32 {
	return Current().Size(SizeNameScrollBarSmall)
}

// SelectionRadiusSize returns the selection highlight radius size.
//
// Since: 2.4
func SelectionRadiusSize() float32 {
	return Current().Size(SizeNameSelectionRadius)
}

// SeparatorThicknessSize is the standard thickness of the separator widget.
//
// Since: 2.0
func SeparatorThicknessSize() float32 {
	return Current().Size(SizeNameSeparatorThickness)
}

// Size looks up the specified size for current theme.
//
// Since: 2.5
func Size(name granio.ThemeSizeName) float32 {
	return Current().Size(name)
}

// SizeForWidget looks up the specified size for the requested widget using the current theme.
// If the widget theme has been overridden that theme will be used.
//
// Since: 2.5
func SizeForWidget(name granio.ThemeSizeName, w granio.Widget) float32 {
	return CurrentForWidget(w).Size(name)
}

// TextHeadingSize returns the text size for header text.
//
// Since: 2.1
func TextHeadingSize() float32 {
	return Current().Size(SizeNameHeadingText)
}

// TextSize returns the standard text size.
func TextSize() float32 {
	return Current().Size(SizeNameText)
}

// TextSubHeadingSize returns the text size for sub-header text.
//
// Since: 2.1
func TextSubHeadingSize() float32 {
	return Current().Size(SizeNameSubHeadingText)
}

func (*builtinTheme) Size(s granio.ThemeSizeName) float32 {
	//revive:disable:add-constant
	//revive:disable:identical-switch-branches
	switch s {
	case SizeNameSeparatorThickness:
		return 1
	case SizeNameInlineIcon:
		return 20
	case SizeNameInnerPadding:
		return 8
	case SizeNameLineSpacing:
		return 4
	case SizeNamePadding:
		return 4
	case SizeNameScrollBar:
		return 12
	case SizeNameScrollBarSmall:
		return 3
	case SizeNameText:
		return 14
	case SizeNameHeadingText:
		return 24
	case SizeNameSubHeadingText:
		return 18
	case SizeNameCaptionText:
		return 11
	case SizeNameInnerWindowRadius:
		return 5
	case SizeNameInputBorder:
		return 1
	case SizeNameInputRadius:
		return 5
	case SizeNameButtonRadius:
		return 5
	case SizeNameModalBlurRadius:
		return 1.5
	case SizeNameSelectionRadius:
		return 3
	case SizeNameScrollBarRadius:
		return 3
	case SizeNameWindowButtonHeight:
		return 16
	case SizeNameWindowButtonRadius:
		return math.MaxFloat32 // MaxRadius without the import loop
	case SizeNameWindowButtonIcon:
		return 14
	case SizeNameWindowShadowActiveRadius:
		return 20
	case SizeNameWindowShadowRadius:
		return 10
	case SizeNameWindowTitleBarHeight:
		return 26
	case SizeNameCardRadius:
		return 5
	case SizeNameDialogRadius:
		return 10
	case SizeNamePopupRadius:
		return 5
	case SizeNameMenuRadius:
		return 3

	default:
		return 0
	}
	//revive:enable:identical-switch-branches
	//revive:enable:add-constant
}
