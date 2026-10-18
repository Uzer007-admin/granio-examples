//go:build accessibility && linux && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2"
)

func (m *atspiManager) pathForKey(w *window, key accessibleKey) dbus.ObjectPath {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, path := range m.windows[w] {
		if nodeKey(m.elements[path].node) == key {
			return path
		}
	}
	return ""
}
func (o *atspiSelectionObject) selectedIDs() ([]string, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return nil, err
	}
	var ids []string
	accessibilityUI(func() {
		if source, ok := e.node.object.(granio.AccessibleVirtualCollection); ok {
			ids = source.AccessibilityVirtualSelection()
		}
	})
	return ids, nil
}
