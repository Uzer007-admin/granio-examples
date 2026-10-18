package diagramwidget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// LinkPoint identifies the point at which a link end is connected to another diagram element's connection pad
type LinkPoint struct {
	widget.BaseWidget
	link DiagramLink
}

// NewLinkPoint creates an instance of a LinkPoint for a specific DiagramLink
func NewLinkPoint(link DiagramLink) *LinkPoint {
	lp := &LinkPoint{}
	lp.BaseWidget.ExtendBaseWidget(lp)
	lp.link = link
	return lp
}

// CreateRenderer creates the renderere for a LinkPoint
func (lp *LinkPoint) CreateRenderer() granio.WidgetRenderer {
	lpr := &linkPointRenderer{}
	return lpr
}

// GetLink returns the Link to which the LinkPoint belongs
func (lp *LinkPoint) GetLink() DiagramLink {
	return lp.link
}

// IsConnectionAllowed returns true if a connection is permitted with the indicated pad. The
// question is passed to the owning link
func (lp *LinkPoint) IsConnectionAllowed(connectionPad ConnectionPad) bool {
	return lp.link.isConnectionAllowed(lp, connectionPad)
}

// linkPointRenderer
type linkPointRenderer struct {
}

func (lpr *linkPointRenderer) Destroy() {

}

func (lpr *linkPointRenderer) Layout(size granio.Size) {

}

func (lpr *linkPointRenderer) MinSize() granio.Size {
	return granio.NewSize(1, 1)
}

func (lpr *linkPointRenderer) Objects() []granio.CanvasObject {
	obj := []granio.CanvasObject{}
	return obj
}

func (lpr *linkPointRenderer) Refresh() {

}
