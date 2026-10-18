package cache

import (
	"errors"
	"sync/atomic"
)

// Limits bounds retained CPU cache data. Zero disables the corresponding cache.
// Bytes include retained keys, values and pixel buffers, but exclude Go map and
// allocator overhead; entry limits also bound that overhead.
type Limits struct {
	TextEntries int
	TextBytes   int64
	SVGEntries  int
	SVGBytes    int64
}

// Usage describes retained data, using the same accounting as Limits.
type Usage struct {
	TextEntries int
	TextBytes   int64
	SVGEntries  int
	SVGBytes    int64
}

var cacheLimits = func() *atomic.Pointer[Limits] {
	p := new(atomic.Pointer[Limits])
	p.Store(&Limits{TextEntries: 8192, TextBytes: 8 << 20, SVGEntries: 256, SVGBytes: 64 << 20})
	return p
}()

func GetLimits() Limits { return *cacheLimits.Load() }

// SetLimits applies new limits immediately, evicting least recently used data.
func SetLimits(limits Limits) error {
	if limits.TextEntries < 0 || limits.TextBytes < 0 || limits.SVGEntries < 0 || limits.SVGBytes < 0 {
		return errors.New("cache limits must be non-negative")
	}
	cacheLimits.Store(&limits)
	fontSizeLock.Lock()
	trimFontMetrics(limits)
	fontSizeLock.Unlock()
	svgs.mu.Lock()
	svgs.trim(limits)
	svgs.mu.Unlock()
	return nil
}

func GetUsage() Usage {
	fontSizeLock.Lock()
	usage := Usage{TextEntries: len(fontSizeCache), TextBytes: fontSizeBytes}
	fontSizeLock.Unlock()
	svgs.mu.Lock()
	usage.SVGEntries, usage.SVGBytes = len(svgs.entries), svgs.bytes
	svgs.mu.Unlock()
	return usage
}
