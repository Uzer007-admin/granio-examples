package theme

import (
	_ "embed"

	"github.com/Uzer007-admin/granio/v2"
)

//go:embed font/NotoSans-Regular.ttf
var notoSansRegular []byte

var regular = &granio.StaticResource{
	StaticName:    "NotoSans-Regular.ttf",
	StaticContent: notoSansRegular,
}

//go:embed font/NotoSans-Bold.ttf
var notoSansBold []byte

var bold = &granio.StaticResource{
	StaticName:    "NotoSans-Bold.ttf",
	StaticContent: notoSansBold,
}

//go:embed font/NotoSans-Italic.ttf
var notoSansItalic []byte

var italic = &granio.StaticResource{
	StaticName:    "NotoSans-Italic.ttf",
	StaticContent: notoSansItalic,
}

//go:embed font/NotoSans-BoldItalic.ttf
var notoSansBoldItalic []byte

var bolditalic = &granio.StaticResource{
	StaticName:    "NotoSans-BoldItalic.ttf",
	StaticContent: notoSansBoldItalic,
}

//go:embed font/DejaVuSansMono-Powerline.ttf
var dejaVuSansMono []byte

var monospace = &granio.StaticResource{
	StaticName:    "DejaVuSansMono-Powerline.ttf",
	StaticContent: dejaVuSansMono,
}

//go:embed font/InterSymbols-Regular.ttf
var interSymbolsRegular []byte

var symbol = &granio.StaticResource{
	StaticName:    "InterSymbols-Regular.ttf",
	StaticContent: interSymbolsRegular,
}
