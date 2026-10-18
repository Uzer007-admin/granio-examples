//go:build !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

const idleEventWait = 250 * time.Millisecond

var (
	loopWaiting       atomic.Bool
	nativeEventMu     sync.RWMutex
	nativeEventsReady bool
)

func wakeMainLoop() {
	if !loopWaiting.Load() {
		return
	}
	nativeEventMu.RLock()
	if nativeEventsReady {
		glfw.PostEmptyEvent()
	}
	nativeEventMu.RUnlock()
}

func (d *gLDriver) sampleMonitorFrameRate() {
	maximum := 60
	for _, monitor := range glfw.GetMonitors() {
		if mode := monitor.GetVideoMode(); mode != nil && mode.RefreshRate > maximum {
			maximum = mode.RefreshRate
		}
	}
	d.detectedFrameRate.Store(int32(maximum))
}

func (d *gLDriver) runLoop() {
	var nextFrame, nextMonitorCheck time.Time
	var previousInterval time.Duration
	for {
		select {
		case <-d.done:
			d.finishLoop()
			return
		default:
		}
		// Keep native input responsive even under a long queue of UI callbacks.
		d.pollEvents()
		for processed := 0; processed < 64; processed++ {
			select {
			case f := <-funcQueue.Out():
				executeMainFunc(f)
			default:
				processed = 64
			}
		}
		d.processWindowEvents()
		now := time.Now()
		if d.requestedFrameRate.Load() == 0 && !now.Before(nextMonitorCheck) {
			d.sampleMonitorFrameRate()
			nextMonitorCheck = now.Add(2 * time.Second)
		}
		interval := time.Second / time.Duration(d.FrameRate())
		if interval != previousInterval {
			nextFrame = now
			previousInterval = interval
		}
		if !now.Before(nextFrame) {
			d.prepareAnimationVisibility()
			d.animation.TickAnimations()
			d.drawSingleFrame()
			// Keep the cadence anchored instead of accumulating every wake's
			// scheduling delay. Skip missed periods without burst catch-up.
			if nextFrame.IsZero() {
				nextFrame = now
			}
			completed := time.Now()
			nextFrame = nextFrame.Add((completed.Sub(nextFrame)/interval + 1) * interval)
		}

		// Arm before checking work: publishing a queue item or setting dirty
		// after this point posts an event, so the native wait cannot lose a wake.
		loopWaiting.Store(true)
		select {
		case <-d.done:
			loopWaiting.Store(false)
			d.finishLoop()
			return
		case f := <-funcQueue.Out():
			loopWaiting.Store(false)
			executeMainFunc(f)
			continue
		default:
		}
		wait := idleEventWait
		if d.hasFrameWork() {
			wait = time.Until(nextFrame)
		}
		if wait > 0 {
			glfw.WaitEventsTimeout(wait.Seconds())
		}
		loopWaiting.Store(false)
	}
}

// platform values returned by forcePlatform to override GLFW's auto-detection.
const (
	platformAuto    = ""
	platformX11     = "x11"
	platformWayland = "wayland"
)

func (d *gLDriver) initGLFW() {
	switch forcePlatform() {
	case platformX11:
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformX11))
	case platformWayland:
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformWayland))
	default:
		glfw.InitHint(glfw.PlatformHint, int(glfw.AnyPlatform))
	}
	err := glfw.Init()
	if err != nil {
		granio.LogError("failed to initialise GLFW", err)
		return
	}

	d.displayPlatform.Store(int32(glfw.GetPlatform()))
	initCursors()
	nativeEventMu.Lock()
	nativeEventsReady = true
	nativeEventMu.Unlock()
	build.IsWayland = glfw.GetPlatform() == glfw.PlatformWayland
}

func (*gLDriver) pollEvents() {
	glfw.PollEvents() // This call blocks while window is being resized, which prevents freeDirtyTextures from being called
}

func (*gLDriver) Terminate() {
	nativeEventMu.Lock()
	nativeEventsReady = false
	glfw.Terminate()
	nativeEventMu.Unlock()
}
