package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"

	xwidget "github.com/Uzer007-admin/granio-x/widget"
)

func main() {
	w := app.New().NewWindow("Map Widget")

	m := xwidget.NewMapWithOptions(
		xwidget.WithOsmTiles(),
		xwidget.WithZoomButtons(true),
		xwidget.WithScrollButtons(true),
		xwidget.AtZoomLevel(9),
		xwidget.AtLatLon(55.95, -3.2),
		xwidget.WithMapMarkers([]xwidget.MapMarker{
			xwidget.NewMapMarker(55.9486, -3.1999, "Edinburgh Castle"),
		}),
	)
	w.SetContent(m)

	w.SetPadded(false)
	w.Resize(granio.NewSize(512, 320))
	w.ShowAndRun()
}
