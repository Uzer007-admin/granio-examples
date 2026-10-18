package cache

import (
	"image"
	"sync"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
)

var svgs = svgCache{entries: make(map[string]*svgInfo)}

type svgCache struct {
	mu             sync.Mutex
	entries        map[string]*svgInfo
	bytes          int64
	oldest, newest *svgInfo
}

func (c *svgCache) Len() int { c.mu.Lock(); defer c.mu.Unlock(); return len(c.entries) }
func (c *svgCache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]*svgInfo)
	c.bytes, c.oldest, c.newest = 0, nil, nil
	c.mu.Unlock()
}

func (c *svgCache) Load(key string) (*svgInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	info, ok := c.entries[key]
	if ok && info != c.newest {
		if info.older != nil {
			info.older.newer = info.newer
		} else {
			c.oldest = info.newer
		}
		if info.newer != nil {
			info.newer.older = info.older
		}
		info.older, info.newer = c.newest, nil
		c.newest.newer = info
		c.newest = info
	}
	return info, ok
}

func (c *svgCache) Store(key string, info *svgInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if old := c.entries[key]; old != nil {
		c.remove(old)
	}
	limits := GetLimits()
	info.key = key
	info.bytes = int64(len(key)) + int64(unsafe.Sizeof(svgInfo{}))
	if info.pix != nil {
		info.bytes += int64(len(info.pix.Pix))
	}
	if limits.SVGEntries == 0 || info.bytes > limits.SVGBytes {
		return
	}
	info.older = c.newest
	if c.newest != nil {
		c.newest.newer = info
	} else {
		c.oldest = info
	}
	c.newest = info
	c.entries[key] = info
	c.bytes += info.bytes
	c.trim(limits)
}

func (c *svgCache) remove(info *svgInfo) {
	if info.older != nil {
		info.older.newer = info.newer
	} else {
		c.oldest = info.newer
	}
	if info.newer != nil {
		info.newer.older = info.older
	} else {
		c.newest = info.older
	}
	delete(c.entries, info.key)
	c.bytes -= info.bytes
	info.older, info.newer = nil, nil
}

func (c *svgCache) trim(limits Limits) {
	for c.oldest != nil && (len(c.entries) > limits.SVGEntries || c.bytes > limits.SVGBytes) {
		c.remove(c.oldest)
	}
}

// GetSvg gets svg image from cache if it exists.
func GetSvg(name string, o granio.CanvasObject, w int, h int) *image.NRGBA {
	svginfo, ok := svgs.Load(overriddenName(name, o))
	if !ok || svginfo == nil {
		return nil
	}

	if svginfo.w != w || svginfo.h != h {
		return nil
	}

	svginfo.setAlive()
	return svginfo.pix
}

// SetSvg sets a svg into the cache map.
func SetSvg(name string, o granio.CanvasObject, pix *image.NRGBA, w int, h int) {
	sinfo := &svgInfo{
		pix: pix,
		w:   w,
		h:   h,
	}
	sinfo.setAlive()
	svgs.Store(overriddenName(name, o), sinfo)
}

type svgInfo struct {
	expiringCache
	pix          *image.NRGBA
	w, h         int
	key          string
	bytes        int64
	older, newer *svgInfo
}

// destroyExpiredSvgs destroys expired svgs cache data.
func destroyExpiredSvgs(now time.Time, recentEpoch ...uint64) {
	svgs.mu.Lock()
	for _, sinfo := range svgs.entries {
		if sinfo.isExpired(now, recentEpoch...) {
			svgs.remove(sinfo)
		}
	}
	svgs.mu.Unlock()
}

func overriddenName(name string, o granio.CanvasObject) string {
	if o != nil { // for overridden themes get the cache key right
		if over, ok := overrides.Load(o); ok {
			return over.cacheID + name
		}
	}

	return name
}
