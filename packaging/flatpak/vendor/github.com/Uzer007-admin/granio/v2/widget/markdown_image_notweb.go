//go:build !wasm && !test_web_driver

package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/yuin/goldmark/ast"
)

func parseMarkdownImage(t *ast.Image) []RichTextSegment {
	dest := string(t.Destination)
	u, err := storage.ParseURI(dest)
	if err != nil {
		u = storage.NewFileURI(dest)
	}
	return []RichTextSegment{&ImageSegment{Source: u, Title: string(t.Title), Alignment: granio.TextAlignCenter}}
}
