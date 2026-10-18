package canvas

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
)

// Declare conformity with CanvasObject interface
var _ granio.CanvasObject = (*ArbitraryPolygon)(nil)

// ArbitraryPolygon describes a colored arbitrary polygon primitive in a Fyne canvas.
// The polygon is defined by a list of vertex positions in clockwise order,
// relative to the object (top-left is (0,0), bottom-right is (width,height)).
// Each corner can have an individually specified rounding radius.
// The implementation is limited to 16 vertices.
//
// Since: 2.8
type ArbitraryPolygon struct {
	baseObject

	Points           []granio.Position // Vertices in coordinates relative to the object. If NormalizedPoints is true, these are (0.0 to 1.0), otherwise absolute
	NormalizedPoints bool              // True if Points are specified in normalized coordinates (0.0 to 1.0) relative to the object's size
	CornerRadii      []float32         // Per-corner rounding radius, must match len(Points), missing entries default to 0
	FillColor        color.Color       // The polygon fill color
	StrokeColor      color.Color       // The polygon stroke color
	StrokeWidth      float32           // The stroke width of the polygon
}

// Hide will set this arbitrary polygon to not be visible
func (p *ArbitraryPolygon) Hide() {
	p.baseObject.Hide()

	repaint(p)
}

// Move the arbitrary polygon to a new position, relative to its parent / canvas
func (p *ArbitraryPolygon) Move(pos granio.Position) {
	if p.Position() == pos {
		return
	}

	p.baseObject.Move(pos)

	repaint(p)
}

// Refresh causes this arbitrary polygon to be redrawn with its configured state.
func (p *ArbitraryPolygon) Refresh() {
	Refresh(p)
}

// Resize on an arbitrary polygon updates the new size of this object.
func (p *ArbitraryPolygon) Resize(s granio.Size) {
	if s == p.Size() {
		return
	}

	p.baseObject.Resize(s)

	Refresh(p)
}

// NewArbitraryPolygon returns a new ArbitraryPolygon instance
func NewArbitraryPolygon(points []granio.Position, fill color.Color) *ArbitraryPolygon {
	return &ArbitraryPolygon{
		Points:    points,
		FillColor: fill,
	}
}
