package mobile

import "github.com/Uzer007-admin/granio/v2"

func (d *driver) StartAnimation(a *granio.Animation) {
	d.animation.Start(a)
}

func (d *driver) StopAnimation(a *granio.Animation) {
	d.animation.Stop(a)
}
