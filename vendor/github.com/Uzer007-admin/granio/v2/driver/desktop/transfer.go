package desktop

import (
	"errors"

	"github.com/Uzer007-admin/granio/v2"
)

// ErrDataDropUnsupported means the current backend cannot receive MIME drops.
var ErrDataDropUnsupported = errors.New("MIME drag-and-drop is unavailable on this backend")

// DataDropWindow accepts external drops as MIME representations. The callback
// runs on the UI thread. Formats are ordered by preference; unsupported offers
// are rejected. SetOnDropped continues to handle file drops independently.
// Native MIME targets currently support Windows and Linux. Callers should handle
// ErrDataDropUnsupported on other backends.
type DataDropWindow interface {
	SetOnDataDropped(formats []string, callback func(granio.Position, map[string][]byte)) error
}
