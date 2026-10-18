package canvas

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/svg"
	"github.com/Uzer007-admin/granio/v2/storage"
)

// ImageLoadOptions bounds an image load. Zero values use defaults of 64 MiB of
// encoded data and 64 million bitmap pixels. Larger limits may be supplied for
// trusted large images. SVG raster dimensions are determined by the canvas.
type ImageLoadOptions struct {
	MaxEncodedBytes int64
	MaxPixels       uint64
}

var imageHTTPClient = &http.Client{Timeout: 30 * time.Second}

// LoadImage loads a URI and decodes bitmaps without touching the UI. Call it from
// a background goroutine, then attach or update the image using granio.Do. HTTP
// requests honor context cancellation and require a successful status code.
func LoadImage(ctx context.Context, uri granio.URI) (*Image, error) {
	return LoadImageWithOptions(ctx, uri, ImageLoadOptions{})
}

// LoadImageWithOptions loads an image with explicit resource limits.
func LoadImageWithOptions(ctx context.Context, uri granio.URI, options ImageLoadOptions) (*Image, error) {
	if ctx == nil || uri == nil {
		return nil, errors.New("image load requires a context and URI")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if options.MaxEncodedBytes < 0 {
		return nil, errors.New("image byte limit must be non-negative")
	}
	if options.MaxEncodedBytes == 0 {
		options.MaxEncodedBytes = 64 << 20
	}
	if options.MaxPixels == 0 {
		options.MaxPixels = 64_000_000
	}
	var reader io.ReadCloser
	if uri.Scheme() == "http" || uri.Scheme() == "https" {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, uri.String(), nil)
		if err != nil {
			return nil, err
		}
		response, err := imageHTTPClient.Do(request)
		if err != nil {
			return nil, err
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			response.Body.Close()
			return nil, fmt.Errorf("image request returned HTTP %d", response.StatusCode)
		}
		reader = response.Body
	} else {
		var err error
		reader, err = storage.Reader(uri)
		if err != nil {
			return nil, err
		}
	}
	defer reader.Close()
	limit := options.MaxEncodedBytes
	if limit < int64(^uint64(0)>>1) {
		limit++
	}
	data, err := io.ReadAll(io.LimitReader(contextImageReader{ctx: ctx, reader: reader}, limit))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > options.MaxEncodedBytes {
		return nil, errors.New("encoded image exceeds byte limit")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// A content hash separates equally named images and reuses identical SVGs.
	sum := sha256.Sum256(data)
	resource := granio.NewStaticResource(fmt.Sprintf("%x-%s", sum[:12], path.Base(uri.Path())), data)
	if svg.IsResourceSVG(resource) {
		decoder, err := svg.NewDecoder(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return &Image{Resource: resource, icon: decoder, isSVG: true, aspect: decoder.Config().Aspect}, nil
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if config.Width <= 0 || config.Height <= 0 || uint64(config.Width) > options.MaxPixels/uint64(config.Height) {
		return nil, errors.New("decoded image exceeds pixel limit")
	}
	decoded, _, err := image.Decode(contextImageReader{ctx: ctx, reader: bytes.NewReader(data)})
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return NewImageFromImage(decoded), nil
}

type contextImageReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextImageReader) Read(data []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(data)
}

// NewImageFromURIAsync returns an empty image immediately, loading and decoding
// its data in the background. Completion and onLoaded run on the UI thread.
// Cancelling ctx prevents a late result from changing the image or calling the
// callback. FillMode, ScaleMode and other display settings remain configurable.
func NewImageFromURIAsync(ctx context.Context, uri granio.URI, onLoaded func(error)) *Image {
	result := &Image{}
	go func() {
		loaded, err := LoadImage(ctx, uri)
		if ctx == nil || ctx.Err() != nil {
			return
		}
		granio.Do(func() {
			if ctx.Err() != nil {
				return
			}
			if err == nil {
				result.Image, result.Resource = loaded.Image, loaded.Resource
				result.Refresh()
			}
			if onLoaded != nil {
				onLoaded(err)
			}
		})
	}()
	return result
}
