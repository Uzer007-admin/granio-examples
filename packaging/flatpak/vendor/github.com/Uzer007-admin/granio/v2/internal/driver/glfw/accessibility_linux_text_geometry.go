//go:build accessibility && linux && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
	"unicode/utf8"
)

func (o *atspiTextObject) GetTextBeforeOffset(offset int32, boundary uint32) (string, int32, int32, *dbus.Error) {
	_, a, _, err := o.GetTextAtOffset(offset, boundary)
	if err != nil || a <= 0 {
		return "", a, a, err
	}
	return o.GetTextAtOffset(a-1, boundary)
}
func (o *atspiTextObject) GetTextAfterOffset(offset int32, boundary uint32) (string, int32, int32, *dbus.Error) {
	_, _, b, err := o.GetTextAtOffset(offset, boundary)
	if err != nil {
		return "", b, b, err
	}
	return o.GetTextAtOffset(b, boundary)
}
func (o *atspiTextObject) GetAttributeRun(offset int32, includeDefaults bool) (map[string]string, int32, int32, *dbus.Error) {
	e, err := o.element()
	return map[string]string{}, 0, int32(utf8.RuneCountInString(e.node.value)), err
}
func (o *atspiTextObject) GetAttributes(offset int32) (map[string]string, int32, int32, *dbus.Error) {
	return o.GetAttributeRun(offset, false)
}
func (o *atspiTextObject) GetAttributeValue(offset int32, name string) (string, *dbus.Error) {
	return "", nil
}
func (o *atspiTextObject) GetDefaultAttributes() (map[string]string, *dbus.Error) {
	return map[string]string{}, nil
}
func (o *atspiTextObject) GetDefaultAttributeSet() (map[string]string, *dbus.Error) {
	return map[string]string{}, nil
}
func (o *atspiTextObject) GetCharacterExtents(offset int32, coordinates uint32) (atspiRectangle, *dbus.Error) {
	return o.GetRangeExtents(offset, offset+1, coordinates)
}
func (o *atspiTextObject) GetRangeExtents(start, end int32, coordinates uint32) (atspiRectangle, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return atspiRectangle{}, err
	}
	if coordinates > 2 || start < 0 || end < start || int(end) > utf8.RuneCountInString(e.node.value) {
		return atspiRectangle{}, dbus.NewError("org.a11y.atspi.Error.InvalidArguments", nil)
	}
	var result atspiRectangle
	accessibilityUI(func() {
		geometry, ok := e.node.object.(granio.AccessibleTextGeometry)
		if !ok {
			return
		}
		rects := geometry.AccessibilityTextRectangles(int(start), int(end))
		for i, r := range rects {
			x, y := e.rectangle.X+int32(scale.ToScreenCoordinate(e.window.canvas, r.Position.X)), e.rectangle.Y+int32(scale.ToScreenCoordinate(e.window.canvas, r.Position.Y))
			width, height := int32(scale.ToScreenCoordinate(e.window.canvas, r.Size.Width)), int32(scale.ToScreenCoordinate(e.window.canvas, r.Size.Height))
			if i == 0 {
				result = atspiRectangle{x, y, width, height}
			} else {
				right, bottom := max(result.X+result.Width, x+width), max(result.Y+result.Height, y+height)
				result.X, result.Y = min(result.X, x), min(result.Y, y)
				result.Width, result.Height = right-result.X, bottom-result.Y
			}
		}
	})
	if coordinates != 0 {
		o.manager.mu.RLock()
		origin := o.manager.elements[e.parent].rectangle
		if coordinates == 1 {
			paths := o.manager.windows[e.window]
			if len(paths) > 0 {
				origin = o.manager.elements[paths[0]].rectangle
			}
		}
		o.manager.mu.RUnlock()
		result.X -= origin.X
		result.Y -= origin.Y
	}
	return result, nil
}
func (o *atspiTextObject) GetOffsetAtPoint(x, y int32, coordinates uint32) (int32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return -1, err
	}
	if coordinates > 2 {
		return -1, dbus.NewError("org.a11y.atspi.Error.InvalidArguments", nil)
	}
	if coordinates != 0 {
		o.manager.mu.RLock()
		origin := o.manager.elements[e.parent].rectangle
		if coordinates == 1 {
			paths := o.manager.windows[e.window]
			if len(paths) > 0 {
				origin = o.manager.elements[paths[0]].rectangle
			}
		}
		o.manager.mu.RUnlock()
		x += origin.X
		y += origin.Y
	}
	offset := int32(-1)
	accessibilityUI(func() {
		if geometry, ok := e.node.object.(granio.AccessibleTextGeometry); ok {
			offset = int32(geometry.AccessibilityTextOffsetAtPosition(granio.NewPos(scale.ToFyneCoordinate(e.window.canvas, int(x-e.rectangle.X)), scale.ToFyneCoordinate(e.window.canvas, int(y-e.rectangle.Y)))))
		}
	})
	return offset, nil
}
