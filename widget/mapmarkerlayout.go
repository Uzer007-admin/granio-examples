package widget

import (
	"github.com/Uzer007-admin/granio/v2"
)

type mapMarkerLayout struct {
	getPosFromLatLon func(lat, lon float64) granio.Position
}

func (l *mapMarkerLayout) MinSize(objects []granio.CanvasObject) granio.Size {
	w, h := float32(0), float32(0)
	for _, o := range objects {
		childSize := o.MinSize()
		if childSize.Width > w {
			childSize.Width = w
		}
		if childSize.Height > h {
			childSize.Height = h
		}
	}
	return granio.NewSize(w, h)
}

func (l *mapMarkerLayout) Layout(objects []granio.CanvasObject, containerSize granio.Size) {
	for _, o := range objects {
		marker, ok := o.(*mapMarker)
		if !ok {
			continue
		}

		size := o.MinSize()
		o.Resize(size)

		pos := l.getPosFromLatLon(marker.obj.Lat(), marker.obj.Lon())
		off := marker.pinOffset()
		pos.X -= off.X
		pos.Y -= off.Y
		o.Move(pos)
	}
}
