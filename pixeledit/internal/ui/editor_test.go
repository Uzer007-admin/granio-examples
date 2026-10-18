package ui

import (
	"image"
	"image/color"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Uzer007-admin/granio-examples/pixeledit/internal/api"
)

func uriForTestFile(name string) granio.URI {
	path := filepath.Join(".", "testdata", name+".png")

	return storage.NewURI("file://" + path)
}

func testFile(name string) granio.URIReadCloser {
	read, err := storage.OpenFileFromURI(uriForTestFile(name))
	if err != nil {
		granio.LogError("Unable to open file \""+name+"\"", err)
		return nil
	}

	return read
}

func testFileWrite(name string) granio.URIWriteCloser {
	write, err := storage.SaveFileToURI(uriForTestFile(name))
	if err != nil {
		granio.LogError("Unable to save file \""+name+"\"", err)
		return nil
	}

	return write
}

func TestEditor_LoadFile(t *testing.T) {
	file := testFile("8x8")
	e := testEditorWithFile(file)

	assert.Equal(t, color.RGBA{A: 255}, e.PixelColor(0, 0))
	assert.Equal(t, color.RGBA{R: 255, G: 255, B: 255, A: 255}, e.PixelColor(1, 0))
}

func TestEditor_Reset(t *testing.T) {
	file := testFile("8x8")
	e := testEditorWithFile(file)

	assert.Equal(t, color.RGBA{A: 255}, e.PixelColor(0, 0))

	red := color.RGBA{255, 0, 0, 255}
	e.SetPixelColor(0, 0, red)
	assert.Equal(t, red, e.PixelColor(0, 0))

	e.Reload()
	assert.Equal(t, color.RGBA{A: 255}, e.PixelColor(0, 0))
}

func TestEditor_Save(t *testing.T) {
	origFile := testFile("8x8")
	defer origFile.Close()
	outURI := storage.NewFileURI(filepath.Join(t.TempDir(), "8x8.png"))
	outFile, err := storage.Writer(outURI)
	require.NoError(t, err)
	content, err := ioutil.ReadAll(origFile)
	require.NoError(t, err)
	_, err = outFile.Write(content)
	require.NoError(t, err)
	require.NoError(t, outFile.Close())

	file, err := storage.Reader(outURI)
	require.NoError(t, err)
	e := testEditorWithFile(file)

	assert.Equal(t, color.RGBA{A: 255}, e.PixelColor(0, 0))

	red := color.RGBA{255, 0, 0, 255}
	e.SetPixelColor(0, 0, red)
	assert.Equal(t, red, e.PixelColor(0, 0))

	e.Save()

	e.Reload()
	assert.Equal(t, red, e.PixelColor(0, 0))
}

func TestEditorFGColor(t *testing.T) {
	e := NewEditor()

	assert.Equal(t, color.Black, e.FGColor())
}

func TestEditor_SetFGColor(t *testing.T) {
	e := NewEditor()

	fg := color.White
	e.SetFGColor(fg)
	assert.Equal(t, fg, e.FGColor())
}

func TestEditor_PixelColor(t *testing.T) {
	file := testFile("8x8")
	e := testEditorWithFile(file)

	assert.Equal(t, color.RGBA{A: 255}, e.PixelColor(0, 0))
	assert.Equal(t, color.RGBA{R: 0, G: 0, B: 0, A: 0}, e.PixelColor(9, 9))
}

func TestEditor_SetPixelColor(t *testing.T) {
	file := testFile("8x8")
	e := testEditorWithFile(file)

	assert.Equal(t, color.RGBA{A: 255}, e.PixelColor(0, 0))
	col := color.RGBA{R: 255, G: 255, B: 0, A: 128}
	e.SetPixelColor(1, 1, col)
	assert.Equal(t, col, e.PixelColor(1, 1))
}

func TestEditor_fixEncoding(t *testing.T) {
	size := 4
	nonRGBA := image.NewCMYK(image.Rect(0, 0, size, size))

	fixed := fixEncoding(nonRGBA)
	assert.Equal(t, image.Rect(0, 0, size, size), fixed.Bounds())
}

func TestEditor_isSupported(t *testing.T) {
	e := NewEditor().(*editor)

	assert.True(t, e.isSupported("test.png"))
	assert.False(t, e.isPNG("test.jpg"))
}

func TestEditor_isPNG(t *testing.T) {
	e := NewEditor().(*editor)

	assert.True(t, e.isPNG("test.png"))
	assert.True(t, e.isPNG("BIG.PNG"))
	assert.False(t, e.isPNG("wrong.ping"))
}

func testEditorWithFile(path granio.URIReadCloser) api.Editor {
	e := NewEditor()
	e.(*editor).win = test.NewWindow(nil)
	e.LoadFile(path)

	return e
}
