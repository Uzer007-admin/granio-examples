package widget

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/test"
)

func TestNewAnimatedGif(t *testing.T) {
	gif, err := NewAnimatedGif(storage.NewFileURI("./testdata/gif/earth.gif"))
	assert.Nil(t, err)

	w := test.NewWindow(gif)
	defer w.Close()
	w.SetPadded(false)
	w.Resize(granio.NewSize(128, 128))

	test.AssertImageMatches(t, "gif/initial.png", w.Canvas().Capture())

	gif.Start()
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, -1, gif.remaining)
}

func TestAnimatedGif_MinSize(t *testing.T) {
	f, err := os.Open("./testdata/gif/earth.gif")
	assert.Nil(t, err)

	r, err := io.ReadAll(f)
	assert.Nil(t, err)

	res := granio.NewStaticResource("earth.gif", r)
	gif, _ := NewAnimatedGifFromResource(res)
	assert.True(t, gif.min.IsZero())

	gif.SetMinSize(granio.NewSize(10.0, 10.0))
	assert.Equal(t, float32(10), gif.MinSize().Width)
	assert.Equal(t, float32(10), gif.MinSize().Height)
}
