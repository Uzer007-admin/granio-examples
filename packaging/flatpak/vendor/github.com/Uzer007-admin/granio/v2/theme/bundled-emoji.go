//go:build !no_emoji

package theme

import (
	_ "embed"

	"github.com/Uzer007-admin/granio/v2"
)

//go:embed font/EmojiOneColor.otf
var emojiFontData []byte

var emoji = &granio.StaticResource{
	StaticName:    "EmojiOneColor.otf",
	StaticContent: emojiFontData,
}
