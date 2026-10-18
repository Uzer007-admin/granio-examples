//go:build windows && !android && !ios && !mobile && !tinygo && !noos && !tamago

package theme

import (
	"github.com/Uzer007-admin/granio/v2"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
	"sync"
)

var windowsUIFonts [4]granio.Resource
var windowsUIFontsOnce sync.Once

func loadSystemFonts() [4]granio.Resource {
	windowsUIFontsOnce.Do(func() { windowsUIFonts = readWindowsUIFonts() })
	return windowsUIFonts
}
func readWindowsUIFonts() (fonts [4]granio.Resource) {
	dir, err := windows.GetWindowsDirectory()
	if err != nil {
		return
	}
	for i, name := range [...]string{"segoeui.ttf", "segoeuib.ttf", "segoeuii.ttf", "segoeuiz.ttf"} {
		data, err := os.ReadFile(filepath.Join(dir, "Fonts", name))
		if err != nil {
			continue
		}
		font, err := sfnt.Parse(data)
		if err != nil {
			continue
		}
		var buffer sfnt.Buffer
		valid := true
		for _, r := range "AaЯяЖжЁё09" {
			glyph, err := font.GlyphIndex(&buffer, r)
			if err != nil || glyph == 0 {
				valid = false
				break
			}
		}
		if valid {
			fonts[i] = granio.NewStaticResource(name, data)
		}
	}
	return
}
