package cache

import (
	"image/color"
	"sync"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
)

// fontSizeCache is read by every canvas.Text on every frame the canvas is
// walked, so it is a typed map: async.Map keys by interface, which boxes the
// struct and hashes it field by field through reflection on each lookup.
var (
	fontSizeLock           sync.Mutex
	fontSizeCache          = map[fontSizeEntry]*fontMetric{}
	fontSizeBytes          int64
	fontOldest, fontNewest *fontMetric
)

type fontMetric struct {
	expiringCache
	size         granio.Size
	baseLine     float32
	entry        fontSizeEntry
	bytes        int64
	older, newer *fontMetric
}

type fontSizeEntry struct {
	Text   string
	Size   float32
	Style  granio.TextStyle
	Source string
}

type FontCacheEntry struct {
	fontSizeEntry

	Canvas granio.Canvas
	Color  color.Color
}

// GetFontMetrics looks up a calculated size and baseline required for the specified text parameters.
func GetFontMetrics(text string, fontSize float32, style granio.TextStyle, source granio.Resource) (size granio.Size, base float32) {
	name := ""
	if source != nil {
		name = source.Name()
	}
	fontSizeLock.Lock()
	defer fontSizeLock.Unlock()
	ret, ok := fontSizeCache[fontSizeEntry{text, fontSize, style, name}]
	if !ok {
		return granio.Size{Width: 0, Height: 0}, 0
	}
	ret.setAlive()
	touchFontMetric(ret)
	return ret.size, ret.baseLine
}

// SetFontMetrics stores a calculated font size and baseline for parameters that were missing from the cache.
func SetFontMetrics(text string, fontSize float32, style granio.TextStyle, source granio.Resource, size granio.Size, base float32) {
	name := ""
	if source != nil {
		name = source.Name()
	}
	ent := fontSizeEntry{text, fontSize, style, name}
	metric := &fontMetric{size: size, baseLine: base, entry: ent,
		bytes: int64(len(text)+len(name)) + int64(unsafe.Sizeof(fontMetric{})) + int64(unsafe.Sizeof(fontSizeEntry{}))}
	metric.setAlive()
	fontSizeLock.Lock()
	limits := GetLimits()
	if old := fontSizeCache[ent]; old != nil {
		removeFontMetric(old)
	}
	if limits.TextEntries == 0 || metric.bytes > limits.TextBytes {
		fontSizeLock.Unlock()
		return
	}
	fontSizeCache[ent] = metric
	fontSizeBytes += metric.bytes
	touchFontMetric(metric)
	trimFontMetrics(limits)
	fontSizeLock.Unlock()
}

// ClearFontMetrics clears the fontSizeCache (for testing/benchmarks)
func ClearFontMetrics() {
	fontSizeLock.Lock()
	fontSizeCache = map[fontSizeEntry]*fontMetric{}
	fontSizeBytes, fontOldest, fontNewest = 0, nil, nil
	fontSizeLock.Unlock()
}

// destroyExpiredFontMetrics destroys expired fontSizeCache entries
func destroyExpiredFontMetrics(now time.Time, recentEpoch ...uint64) {
	fontSizeLock.Lock()
	for _, v := range fontSizeCache {
		if v.isExpired(now, recentEpoch...) {
			removeFontMetric(v)
		}
	}
	fontSizeLock.Unlock()
}

// All LRU operations run under fontSizeLock and reuse the metric allocation.
func touchFontMetric(metric *fontMetric) {
	if metric == fontNewest {
		return
	}
	if metric.older != nil {
		metric.older.newer = metric.newer
	} else if metric == fontOldest {
		fontOldest = metric.newer
	}
	if metric.newer != nil {
		metric.newer.older = metric.older
	}
	metric.older, metric.newer = fontNewest, nil
	if fontNewest != nil {
		fontNewest.newer = metric
	} else {
		fontOldest = metric
	}
	fontNewest = metric
}

func removeFontMetric(metric *fontMetric) {
	if metric.older != nil {
		metric.older.newer = metric.newer
	} else {
		fontOldest = metric.newer
	}
	if metric.newer != nil {
		metric.newer.older = metric.older
	} else {
		fontNewest = metric.older
	}
	delete(fontSizeCache, metric.entry)
	fontSizeBytes -= metric.bytes
	metric.older, metric.newer = nil, nil
}

func trimFontMetrics(limits Limits) {
	for fontOldest != nil && (len(fontSizeCache) > limits.TextEntries || fontSizeBytes > limits.TextBytes) {
		removeFontMetric(fontOldest)
	}
}
