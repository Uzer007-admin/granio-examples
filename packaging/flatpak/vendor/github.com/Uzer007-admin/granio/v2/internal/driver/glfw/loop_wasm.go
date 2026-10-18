//go:build wasm || test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/granio/v2"

	"github.com/Uzer007-admin/granio-gl-js"
	"github.com/Uzer007-admin/granio-glfw-js"
)

func (d *gLDriver) runLoop() {
	frames, request, stop := browserFrameSource(d)
	defer stop()
	request()
	for {
		select {
		case <-d.done:
			d.finishLoop()
			return
		case f := <-funcQueue.Out():
			executeMainFunc(f)
		case <-browserFrameWake:
			request()
		case <-frames:
			d.pollEvents()
			d.processWindowEvents()
			d.animation.TickAnimations()
			d.drawSingleFrame()
			if d.hasFrameWork() {
				request()
			}
		}
	}
}

func (d *gLDriver) initGLFW() {
	err := glfw.Init(gl.ContextWatcher)
	if err != nil {
		granio.LogError("failed to initialise GLFW", err)
		return
	}
}

func (d *gLDriver) pollEvents() {
	glfw.PollEvents() // This call blocks while window is being resized, which prevents freeDirtyTextures from being called
}

func (d *gLDriver) Terminate() {
	glfw.Terminate()
}
