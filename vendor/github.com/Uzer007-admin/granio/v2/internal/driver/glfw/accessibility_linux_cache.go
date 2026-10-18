//go:build accessibility && linux && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/dbus/v5"
	"slices"
)

const atspiCache = "org.a11y.atspi.Cache"
const atspiCachePath = dbus.ObjectPath("/org/a11y/atspi/cache")

// The post-2015 protocol carries indices/counts, not an eagerly expanded child
// array. A virtual collection remains lazy even during the initial bulk query.
type atspiCacheItem struct {
	Object, Application, Parent atspiReference
	Index, ChildCount           int32
	Interfaces                  []string
	Name                        string
	Role                        uint32
	Description                 string
	States                      []uint32
}
type atspiCacheObject struct{ manager *atspiManager }

func (o *atspiCacheObject) item(path dbus.ObjectPath) (atspiCacheItem, *dbus.Error) {
	object := &atspiObject{o.manager, path}
	e, err := object.element()
	if err != nil {
		return atspiCacheItem{}, err
	}
	index, err := object.GetIndexInParent()
	if err != nil {
		return atspiCacheItem{}, err
	}
	interfaces, err := object.GetInterfaces()
	if err != nil {
		return atspiCacheItem{}, err
	}
	role, err := object.GetRole()
	if err != nil {
		return atspiCacheItem{}, err
	}
	states, err := object.GetState()
	if err != nil {
		return atspiCacheItem{}, err
	}
	count := len(e.children)
	if e.node.virtualOwner != nil {
		count = e.node.virtualCount
	}
	parent := o.manager.reference(e.parent)
	if path == atspiRoot {
		o.manager.mu.RLock()
		parent = o.manager.parent
		o.manager.mu.RUnlock()
	}
	return atspiCacheItem{o.manager.reference(path), o.manager.reference(atspiRoot), parent, index, int32(count), interfaces, e.node.name, role, "", states}, nil
}
func (o *atspiCacheObject) GetItems() ([]atspiCacheItem, *dbus.Error) {
	o.manager.mu.RLock()
	paths := make([]dbus.ObjectPath, 0, len(o.manager.elements))
	for path := range o.manager.elements {
		paths = append(paths, path)
	}
	o.manager.mu.RUnlock()
	slices.Sort(paths)
	items := make([]atspiCacheItem, 0, len(paths))
	for _, path := range paths {
		item, err := o.item(path)
		if err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}
func (o *atspiCacheObject) Get(iface, name string) (dbus.Variant, *dbus.Error) {
	if iface == atspiCache && name == "version" {
		return dbus.MakeVariant(uint32(1)), nil
	}
	return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", nil)
}
func (o *atspiCacheObject) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	return map[string]dbus.Variant{"version": dbus.MakeVariant(uint32(1))}, nil
}
