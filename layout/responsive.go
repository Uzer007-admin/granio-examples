package layout

import (
	"fmt"
	"log"
	"math"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// responsive layout provides a granio.Layout that is responsive to the window size.
// All granio.CanvasObject are resized and positionned following the rules you decide.
//
// It is strongly inspired by Bootstrap's grid system. But instead of using 12 columns,
// we use width ratio.
// By default, a standard granio.CanvasObject will always be width to 1 * containerSize and place in vertical.
// If you want to change the behavior, you can use Responsive() function that registers the layout configuration.
//
// Example:
//    layout := NewResponsiveLayout(
//         Responsive(label, 1, .5, .25, .5), // small, medium, large, xlarge ratio
//    }
// Note that a responsive layout can handle others layouts, responsive or not.

// responsiveBreakpoint is a integer representing a breakpoint size as defined in Bootstrap.
//
// See: https://getbootstrap.com/docs/4.0/layout/overview/#responsive-breakpoints
type responsiveBreakpoint uint16

const (
	// SMALL is the smallest breakpoint (mobile vertical).
	SMALL responsiveBreakpoint = 576

	// MEDIUM is the medium breakpoint (mobile horizontal, tablet vertical).
	MEDIUM responsiveBreakpoint = 768

	// LARGE is the largest breakpoint (tablet horizontal, small desktop).
	LARGE responsiveBreakpoint = 992

	// XLARGE is the largest breakpoint (large desktop).
	XLARGE responsiveBreakpoint = 1200

	// SM is an alias for SMALL
	SM responsiveBreakpoint = SMALL

	// MD is an alias for MEDIUM
	MD responsiveBreakpoint = MEDIUM

	// LG is an alias for LARGE
	LG responsiveBreakpoint = LARGE

	// XL is an alias for XLARGE
	XL responsiveBreakpoint = XLARGE
)

// ResponsiveConfiguration is the configuration for a responsive object. It's
// a simple map from the breakpoint to the size ratio from it's container.
// Breakpoint is a uint16 that should be set from const SMALL, MEDIUM, LARGE and XLARGE.
type responsiveConfig map[responsiveBreakpoint]float32

// newResponsiveConf return a new responsive configuration.
// The optional ratios must
// be 0 < ratio <= 1 and  passed in this order:
//
//	Responsive(object, smallRatio, mediumRatio, largeRatio, xlargeRatio)
//
// They are set to previous value if a value is not passed, or 1.0 if there is no previous value.
func newResponsiveConf(ratios ...float32) responsiveConfig {
	responsive := responsiveConfig{}

	if len(ratios) > 4 {
		log.Println("Responsive: you declared more than 4 ratios, only the first 4 will be used")
	}

	// basic check
	for _, i := range ratios {
		if i <= 0 || i > 1 {
			message := "Responsive: size must be > 0 and <= 1, got: %f"
			panic(fmt.Errorf(message, i))
		}
	}

	// Set default values
	for index, bp := range []responsiveBreakpoint{SMALL, MEDIUM, LARGE, XLARGE} {
		if len(ratios) <= index {
			if index == 0 {
				ratios = append(ratios, 1)
			} else {
				ratios = append(ratios, ratios[index-1])
			}
		}
		responsive[bp] = ratios[index]
	}
	return responsive
}

// ResponsiveLayout is the layout that will adapt objects with the responsive rules. See NewResponsiveLayout
// for details.
type ResponsiveLayout struct{}

// Layout will place the size and place the objects following the configured reponsive rules.
//
// Implements: granio.Layout
func (resp *ResponsiveLayout) Layout(objects []granio.CanvasObject, containerSize granio.Size) {
	// yes, it may happen
	if len(objects) == 0 || objects[0] == nil {
		return
	}

	// Responsive is based on the window size, so we need to get it
	window := granio.CurrentApp().Driver().CanvasForObject(objects[0])
	if window == nil {
		return
	}

	// this will be updatad for each element to know where to place
	// the next object.
	pos := granio.NewPos(0, 0)

	// to calculate the next pos.Y when a new line is needed
	maxHeight := float32(0)

	// objects in a line
	line := []granio.CanvasObject{}

	// cast windowSize.Width to responsiveBreakpoint (uint16)
	ww := responsiveBreakpoint(window.Size().Width)

	// For each object, place it at the right position (pos) and resize it.
	for _, o := range objects {
		if o == nil || !o.Visible() {
			continue
		}

		// get tht configuration
		ro, ok := o.(*responsiveWidget)
		if !ok {
			log.Fatal("A non responsive object has been packed inside a ResponsibleLayout. This is impossible.")
		}
		conf := ro.responsiveConfig

		line = append(line, o) // add the container to the line
		size := o.MinSize()    // get some informations

		// adapt object witdh from the configuration
		if ww <= SMALL {
			size.Width = conf[SMALL] * containerSize.Width
		} else if ww <= MEDIUM {
			size.Width = conf[MEDIUM] * containerSize.Width
		} else if ww <= LARGE {
			size.Width = conf[LARGE] * containerSize.Width
		} else {
			size.Width = conf[XLARGE] * containerSize.Width
		}

		// place and resize the element
		o.Resize(size)
		o.Move(pos)

		// next element X position
		pos = pos.Add(granio.NewPos(size.Width+theme.Padding(), 0))

		maxHeight = resp.maxFloat32(maxHeight, size.Height)

		// Manage end of line, the next position overflows, so go to next line.
		if pos.X >= containerSize.Width-theme.Padding() {
			// we now know the number of object in a line, fix padding
			resp.fixPaddingOnLine(line)
			line = []granio.CanvasObject{}
			pos.X = 0          // back to left
			pos.Y += maxHeight // move to the next line
			maxHeight = 0
		}
	}
	resp.fixPaddingOnLine(line) // fix padding for the last line
}

// MinSize return the minimum size ot the layout.
//
// Implements: granio.Layout
func (resp *ResponsiveLayout) MinSize(objects []granio.CanvasObject) granio.Size {
	if len(objects) == 0 {
		return granio.NewSize(0, 0)
	}

	var h, w, maxHeight float32
	currentY := objects[0].Position().Y

	for _, o := range objects {
		if o == nil || !o.Visible() {
			continue
		}
		w = resp.maxFloat32(o.MinSize().Width, w) + theme.Padding()
		if o.Position().Y != currentY {
			currentY = o.Position().Y
			// new line, so we can add the maxHeight to h
			h += maxHeight

			// drop the line
			maxHeight = 0
		}
		maxHeight = resp.maxFloat32(maxHeight, o.MinSize().Height)
	}
	h += maxHeight + theme.Padding()
	return granio.NewSize(w, h)
}

// fixPaddingOnLine fix the space between the objects in a line.
func (resp *ResponsiveLayout) fixPaddingOnLine(line []granio.CanvasObject) {
	if len(line) <= 1 {
		return
	}
	for i, o := range line {
		s := o.Size()
		s.Width -= theme.Padding() / float32(len(line)-1)
		o.Resize(s)
		if i > 0 {
			p := o.Position()
			p.X -= theme.Padding() * float32(i)
			o.Move(p)
		}
	}
}

// math.Max only works with float64, so let's make our own
func (resp *ResponsiveLayout) maxFloat32(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

// NewResponsiveLayout return a responsive layout that will adapt objects with the responsive rules. To
// configure the rule, each object could be encapsulated by a "Responsive" object.
//
// Example:
//
//	container := NewResponsiveLayout(
//	    Responsive(label, 1, .5, .25),  // 100% for small, 50% for medium, 25% for large
//	    Responsive(button, 1, .5, .25), // ...
//	    label2,                         // this will be placed and resized with default behaviors
//	                                    // => 1, 1, 1
//	)
func NewResponsiveLayout(o ...granio.CanvasObject) *granio.Container {
	r := &ResponsiveLayout{}

	objects := []granio.CanvasObject{}
	for _, unknowObject := range o {
		if _, ok := unknowObject.(*responsiveWidget); !ok {
			unknowObject = Responsive(unknowObject)
		}
		objects = append(objects, unknowObject)
	}

	return container.New(r, objects...)
}

type responsiveWidget struct {
	widget.BaseWidget

	render           granio.CanvasObject
	responsiveConfig responsiveConfig
}

var _ granio.Widget = (*responsiveWidget)(nil)

// Responsive register the object with a responsive configuration.
// The optional ratios must
// be 0 < ratio <= 1 and  passed in this order:
//
//	Responsive(object, smallRatio, mediumRatio, largeRatio, xlargeRatio)
//
// They are set to previous value if a value is not passed, or 1.0 if there is no previous value.
// The returned object is not modified.
func Responsive(object granio.CanvasObject, breakpointRatio ...float32) granio.CanvasObject {
	ro := &responsiveWidget{render: object, responsiveConfig: newResponsiveConf(breakpointRatio...)}
	ro.ExtendBaseWidget(ro)
	return ro
}

func (ro *responsiveWidget) CreateRenderer() granio.WidgetRenderer {
	if ro.render == nil {
		return nil
	}
	return widget.NewSimpleRenderer(ro.render)
}
