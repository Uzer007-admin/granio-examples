package glfw

import "github.com/Uzer007-admin/granio/v2"

func (d *gLDriver) StartAnimation(a *granio.Animation) {
	d.animation.Start(a)
	wakeMainLoop()
}

func (d *gLDriver) StopAnimation(a *granio.Animation) {
	d.animation.Stop(a)
}
