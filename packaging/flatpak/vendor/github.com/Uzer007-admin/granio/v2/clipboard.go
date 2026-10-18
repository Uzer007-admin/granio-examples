package granio

// Clipboard represents the system clipboard interface
type Clipboard interface {
	// Content returns the clipboard content
	Content() string
	// SetContent sets the clipboard content
	SetContent(content string)
}

// ClipboardData is the optional desktop clipboard interface for typed transfers.
// WriteData replaces the clipboard with all supplied representations. ReadData
// returns a caller-owned copy. MIME types follow the IANA media-type syntax.
type ClipboardData interface {
	Clipboard
	Formats() ([]string, error)
	ReadData(mimeType string) ([]byte, error)
	WriteData(map[string][]byte) error
}

const (
	MIMETextPlain   = "text/plain"
	MIMETextHTML    = "text/html"
	MIMETextURIList = "text/uri-list"
	MIMEImagePNG    = "image/png"
)
