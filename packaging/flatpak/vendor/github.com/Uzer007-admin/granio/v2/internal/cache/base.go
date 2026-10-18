package cache

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/Uzer007-admin/granio/v2"
)

const cacheCleanCooldown = 10 * time.Second

var (
	ValidDuration     = 1 * time.Minute
	cleanTaskInterval = ValidDuration / 2

	lastClean                     time.Time
	skippedCleanWithCanvasRefresh = false

	// cachedNow is the clock sample setAlive uses. Reading the real clock there
	// is an expensive operation and setAlive is called per-object per paint.
	// cachedNow is updated on Clean, which should be called once per frame.
	// Atomic because tests (and Resize/Refresh from goroutines in non-migrated apps)
	// reach setAlive off the main thread while the draw loop is in Clean.
	cachedNow atomic.Int64
	// Each cleanup attempt advances the epoch. A touch in the preceding epoch
	// is recent even when the cached wall clock came from before a long idle.
	touchEpoch atomic.Uint64

	// testing purpose only
	timeNow = time.Now
)

func init() {
	touchEpoch.Store(1)
	if t, err := time.ParseDuration(os.Getenv("FYNE_CACHE")); err == nil {
		ValidDuration = t
		cleanTaskInterval = ValidDuration / 2
	}
	refreshNow()
}

// refreshNow samples the clock for setAlive and returns the sample.
func refreshNow() time.Time {
	now := timeNow()
	cachedNow.Store(now.UnixNano())
	return now
}

// Clean run cache clean task, it should be called on paint events.
func Clean(canvasRefreshed bool) {
	recentEpoch := touchEpoch.Add(1) - 1
	now := refreshNow()
	// do not run clean task too fast
	if now.Sub(lastClean) < cacheCleanCooldown {
		if canvasRefreshed {
			skippedCleanWithCanvasRefresh = true
		}
		return
	}
	if skippedCleanWithCanvasRefresh {
		skippedCleanWithCanvasRefresh = false
		canvasRefreshed = true
	}
	if !canvasRefreshed && now.Sub(lastClean) < cleanTaskInterval {
		return
	}
	destroyExpiredSvgs(now, recentEpoch)
	destroyExpiredFontMetrics(now, recentEpoch)
	destroyExpiredBlurKernels(now, recentEpoch)
	if canvasRefreshed {
		// Destroy renderers on canvas refresh to avoid flickering screen.
		destroyExpiredRenderers(now, recentEpoch)
		// canvases cache should be invalidated only on canvas refresh, otherwise there wouldn't
		// be a way to recover them later
		destroyExpiredCanvases(now, recentEpoch)
	}
	lastClean = refreshNow()
}

// CleanCanvas performs a complete remove of all the objects that belong to the specified
// canvas. Usually used to free all objects from a closing windows.
func CleanCanvas(canvas granio.Canvas) {
	canvases.Range(func(obj granio.CanvasObject, cinfo *canvasInfo) bool {
		if cinfo.canvas != canvas {
			return true
		}

		canvases.Delete(obj)

		wid, ok := obj.(granio.Widget)
		if !ok {
			return true
		}
		rinfo, ok := renderers.LoadAndDelete(wid)
		if !ok {
			return true
		}
		rinfo.renderer.Destroy()
		overrides.Delete(wid)
		return true
	})
}

// ResetThemeCaches clears all the svg and text size cache maps
func ResetThemeCaches() {
	svgs.Clear()
	ClearFontMetrics()
}

// destroyExpiredCanvases deletes objects from the canvases cache.
func destroyExpiredCanvases(now time.Time, recentEpoch ...uint64) {
	canvases.Range(func(obj granio.CanvasObject, cinfo *canvasInfo) bool {
		if cinfo.isExpired(now, recentEpoch...) {
			canvases.Delete(obj)
		}
		return true
	})
}

// destroyExpiredRenderers deletes the renderer from the cache and calls
// renderer.Destroy()
func destroyExpiredRenderers(now time.Time, recentEpoch ...uint64) {
	renderers.Range(func(wid granio.Widget, rinfo *rendererInfo) bool {
		if rinfo.isExpired(now, recentEpoch...) {
			rinfo.renderer.Destroy()
			overrides.Delete(wid)
			renderers.Delete(wid)
		}
		return true
	})
}

type expiringCache struct {
	expires atomic.Int64 // unix nanos, so setAlive is a load and an add
	touched atomic.Uint64
}

// isExpired check if the cache data is expired.
func (c *expiringCache) isExpired(now time.Time, recentEpoch ...uint64) bool {
	expires := c.expires.Load()
	if len(recentEpoch) != 0 && ValidDuration > 0 {
		if used := c.touched.Load(); used != 0 && used >= recentEpoch[0] {
			// Rebase a recent touch onto the current frame's clock without
			// adding a time.Now call to every object lookup.
			c.expires.CompareAndSwap(expires, now.UnixNano()+ValidDuration.Nanoseconds())
			return false
		}
	}
	return expires < now.UnixNano()
}

// setAlive updates expiration time.
func (c *expiringCache) setAlive() {
	c.touched.Store(touchEpoch.Load())
	c.expires.Store(cachedNow.Load() + ValidDuration.Nanoseconds())
}
