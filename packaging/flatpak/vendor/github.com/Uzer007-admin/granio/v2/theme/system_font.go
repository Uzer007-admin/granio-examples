package theme

type FontPolicy uint8

const (
	FontBundled FontPolicy = iota
	// FontSystem loads installed Windows UI faces in memory, without distributing
	// their files. Other platforms and missing styles keep the bundled fallback.
	FontSystem
)
