//go:build accessibility && linux && !wasm && !test_web_driver

package glfw

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/scale"
)

const atspiRoot = dbus.ObjectPath("/org/a11y/atspi/accessible/root")
const atspiAccessible = "org.a11y.atspi.Accessible"
const atspiComponent = "org.a11y.atspi.Component"
const atspiAction = "org.a11y.atspi.Action"
const atspiText = "org.a11y.atspi.Text"
const atspiEditable = "org.a11y.atspi.EditableText"
const atspiValue = "org.a11y.atspi.Value"
const atspiSelection = "org.a11y.atspi.Selection"
const atspiTable = "org.a11y.atspi.Table"

type atspiReference struct {
	Bus  string
	Path dbus.ObjectPath
}
type atspiRectangle struct{ X, Y, Width, Height int32 }
type atspiRelation struct {
	Type    uint32
	Targets []atspiReference
}
type atspiActionInfo struct{ Name, Description, KeyBinding string }
type atspiElement struct {
	node      accessibleNode
	path      dbus.ObjectPath
	parent    dbus.ObjectPath
	children  []dbus.ObjectPath
	rectangle atspiRectangle
	window    *window
}
type atspiManager struct {
	mu              sync.RWMutex
	conn            *dbus.Conn
	elements        map[dbus.ObjectPath]atspiElement
	windows         map[*window][]dbus.ObjectPath
	parent          atspiReference
	busName, locale string
	appID           int32
	next            uint64
}

var linuxAccessibility = &atspiManager{elements: make(map[dbus.ObjectPath]atspiElement), windows: make(map[*window][]dbus.ObjectPath), parent: atspiReference{Path: "/org/a11y/atspi/null"}}
var atspiConnectOnce sync.Once

func (*window) initAccessibilityForWindow() {
	atspiConnectOnce.Do(func() { go linuxAccessibility.connect() })
}

func (m *atspiManager) connect() {
	session, err := dbus.ConnectSessionBus()
	if err != nil {
		return
	}
	defer session.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var address string
	err = session.Object("org.a11y.Bus", "/org/a11y/bus").CallWithContext(ctx, "org.a11y.Bus.GetAddress", 0).Store(&address)
	if err != nil {
		return
	}
	conn, err := dbus.Connect(address)
	if err != nil {
		return
	}
	m.mu.Lock()
	m.conn = conn
	m.busName = conn.Names()[0]
	m.locale = string(granio.CurrentDevice().Locale())
	m.elements[atspiRoot] = atspiElement{path: atspiRoot, node: accessibleNode{name: granio.CurrentApp().UniqueID(), role: granio.AccessibleRoleContainer}}
	m.mu.Unlock()
	m.export(atspiRoot, true)
	_ = conn.Export(&atspiCacheObject{m}, atspiCachePath, atspiCache)
	_ = conn.Export(&atspiCacheObject{m}, atspiCachePath, "org.freedesktop.DBus.Properties")
	var parent atspiReference
	if conn.Object("org.a11y.atspi.Registry", atspiRoot).CallWithContext(ctx, "org.a11y.atspi.Socket.Embed", 0, m.reference(atspiRoot)).Store(&parent) == nil {
		m.mu.Lock()
		m.parent = parent
		m.mu.Unlock()
	}
	queueAccessibility(func() {
		m.mu.RLock()
		windows := make([]*window, 0, len(m.windows))
		for window := range m.windows {
			windows = append(windows, window)
		}
		m.mu.RUnlock()
		for _, window := range windows {
			window.updateAccessibility()
		}
	})
}

func (m *atspiManager) reference(path dbus.ObjectPath) atspiReference {
	if path == "" {
		return atspiReference{Path: "/org/a11y/atspi/null"}
	}
	m.mu.RLock()
	name := m.busName
	m.mu.RUnlock()
	return atspiReference{Bus: name, Path: path}
}

func (w *window) updateAccessibility() {
	m := linuxAccessibility
	m.mu.Lock()
	if _, exists := m.windows[w]; !exists {
		m.windows[w] = nil
	}
	conn := m.conn
	m.mu.Unlock()
	if conn == nil || w.viewport == nil {
		return
	}
	snapshot := w.accessibilitySnapshot()
	x, y := w.viewport.GetPos()
	m.mu.Lock()
	previous := m.windows[w]
	oldByObject := make(map[accessibleKey]dbus.ObjectPath, len(previous))
	for _, path := range previous {
		if node := m.elements[path].node; node.object != nil {
			oldByObject[nodeKey(node)] = path
		}
	}
	paths := make([]dbus.ObjectPath, len(snapshot))
	for i := range snapshot {
		if i == 0 && len(previous) > 0 {
			paths[i] = previous[0]
		} else {
			paths[i] = oldByObject[nodeKey(snapshot[i])]
		}
		if paths[i] == "" {
			m.next++
			paths[i] = dbus.ObjectPath(fmt.Sprintf("/org/a11y/atspi/accessible/node%d", m.next))
		}
	}
	added, removed := make([]dbus.ObjectPath, 0), make([]dbus.ObjectPath, 0)
	oldPaths, newPaths := make(map[dbus.ObjectPath]bool, len(previous)), make(map[dbus.ObjectPath]bool, len(paths))
	for _, p := range previous {
		oldPaths[p] = true
	}
	for _, p := range paths {
		newPaths[p] = true
	}
	for _, path := range previous {
		if !newPaths[path] {
			removed = append(removed, path)
		}
	}
	for _, path := range paths {
		if !oldPaths[path] {
			added = append(added, path)
		}
	}
	changed := make([]atspiElement, 0)
	removedIndices := make(map[dbus.ObjectPath]int, len(removed))
	for _, path := range removed {
		old := m.elements[path]
		removedIndices[path] = slices.Index(m.elements[old.parent].children, path)
	}
	for i, node := range snapshot {
		parent := atspiRoot
		if node.parent >= 0 {
			parent = paths[node.parent]
		}
		children := make([]dbus.ObjectPath, len(node.children))
		for n, id := range node.children {
			children[n] = paths[id]
		}
		position := node.position
		old, existed := m.elements[paths[i]]
		if existed && (old.node.name != node.name || old.node.value != node.value || old.node.state != node.state || old.node.cursor != node.cursor || old.node.number != node.number || old.node.selectionStart != node.selectionStart || old.node.selectionEnd != node.selectionEnd) {
			changed = append(changed, old)
		}
		m.elements[paths[i]] = atspiElement{node: node, path: paths[i], parent: parent, children: children, window: w,
			rectangle: atspiRectangle{int32(x + scale.ToScreenCoordinate(w.canvas, position.X)), int32(y + scale.ToScreenCoordinate(w.canvas, position.Y)),
				int32(scale.ToScreenCoordinate(w.canvas, node.size.Width)), int32(scale.ToScreenCoordinate(w.canvas, node.size.Height))}}
	}
	childChanges := make([]struct {
		parent, path dbus.ObjectPath
		kind         string
		index        int
	}, 0, len(added)+len(removed))
	for _, path := range removed {
		old := m.elements[path]
		childChanges = append(childChanges, struct {
			parent, path dbus.ObjectPath
			kind         string
			index        int
		}{old.parent, path, "remove", removedIndices[path]})
	}
	for _, path := range added {
		current := m.elements[path]
		childChanges = append(childChanges, struct {
			parent, path dbus.ObjectPath
			kind         string
			index        int
		}{current.parent, path, "add", slices.Index(m.elements[current.parent].children, path)})
	}
	for _, path := range removed {
		delete(m.elements, path)
	}
	m.windows[w] = paths
	root := m.elements[atspiRoot]
	root.children = nil
	for _, window := range m.windows {
		if len(window) > 0 {
			root.children = append(root.children, window[0])
		}
	}
	m.elements[atspiRoot] = root
	m.mu.Unlock()
	for _, path := range added {
		m.export(path, false)
		if item, err := (&atspiCacheObject{m}).item(path); err == nil {
			_ = conn.Emit(atspiCachePath, atspiCache+".AddAccessible", item)
		}
	}
	for _, path := range removed {
		m.unexport(path)
		_ = conn.Emit(atspiCachePath, atspiCache+".RemoveAccessible", m.reference(path))
	}
	for _, change := range childChanges {
		if change.parent != atspiRoot {
			m.event(change.parent, "ChildrenChanged", change.kind, int32(max(0, change.index)), dbus.MakeVariant(m.reference(change.path)))
		}
	}
	if len(previous) == 0 {
		m.event(atspiRoot, "ChildrenChanged", "add", 0, dbus.MakeVariant(m.reference(paths[0])))
	}
	for _, old := range changed {
		current, _ := (&atspiObject{manager: m, path: old.path}).element()
		if old.node.name != current.node.name {
			m.event(old.path, "PropertyChange", "accessible-name", 0, dbus.MakeVariant(current.node.name))
		}
		if old.node.value != current.node.value {
			before, after := []rune(old.node.value), []rune(current.node.value)
			prefix := 0
			for prefix < len(before) && prefix < len(after) && before[prefix] == after[prefix] {
				prefix++
			}
			a, b := len(before), len(after)
			for a > prefix && b > prefix && before[a-1] == after[b-1] {
				a--
				b--
			}
			if a > prefix {
				m.eventDetails(old.path, "TextChanged", "delete", int32(prefix), int32(a-prefix), dbus.MakeVariant(string(before[prefix:a])))
			}
			if b > prefix {
				m.eventDetails(old.path, "TextChanged", "insert", int32(prefix), int32(b-prefix), dbus.MakeVariant(string(after[prefix:b])))
			}
		}
		if old.node.cursor != current.node.cursor {
			m.event(old.path, "TextCaretMoved", "", int32(current.node.cursor), dbus.MakeVariant(int32(0)))
		}
		if old.node.selectionStart != current.node.selectionStart || old.node.selectionEnd != current.node.selectionEnd {
			m.event(old.path, "TextSelectionChanged", "", 0, dbus.MakeVariant(int32(0)))
		}
		if old.node.number != current.node.number && current.node.numeric {
			m.event(old.path, "PropertyChange", "accessible-value", 0, dbus.MakeVariant(current.node.number))
		}
		if old.node.state.Active != current.node.state.Active {
			kind := "Deactivate"
			if current.node.state.Active {
				kind = "Activate"
			}
			_ = conn.Emit(old.path, "org.a11y.atspi.Event.Window."+kind, "", int32(0), int32(0), dbus.MakeVariant(int32(0)), map[string]dbus.Variant{})
		}
		for _, change := range []struct {
			name          string
			before, after bool
		}{{"focused", old.node.state.Focused, current.node.state.Focused}, {"checked", old.node.state.Checked, current.node.state.Checked}, {"selected", old.node.state.Selected, current.node.state.Selected}, {"expanded", old.node.state.Expanded, current.node.state.Expanded}, {"enabled", !old.node.state.Disabled, !current.node.state.Disabled}, {"read-only", old.node.state.ReadOnly, current.node.state.ReadOnly}} {
			if change.before != change.after {
				value := int32(0)
				if change.after {
					value = 1
				}
				m.event(old.path, "StateChanged", change.name, value, dbus.MakeVariant(int32(0)))
			}
		}
	}
	_ = conn
}

func (m *atspiManager) event(path dbus.ObjectPath, kind, detail string, number int32, value dbus.Variant) {
	m.eventDetails(path, kind, detail, number, 0, value)
}
func (m *atspiManager) eventDetails(path dbus.ObjectPath, kind, detail string, first, second int32, value dbus.Variant) {
	m.mu.RLock()
	conn := m.conn
	m.mu.RUnlock()
	if conn != nil {
		_ = conn.Emit(path, "org.a11y.atspi.Event.Object."+kind, detail, first, second, value, map[string]dbus.Variant{})
	}
}

func (w *window) cleanupAccessibilityForWindow() {
	m := linuxAccessibility
	m.mu.Lock()
	paths := m.windows[w]
	delete(m.windows, w)
	for _, path := range paths {
		delete(m.elements, path)
	}
	root := m.elements[atspiRoot]
	root.children = nil
	for _, window := range m.windows {
		if len(window) > 0 {
			root.children = append(root.children, window[0])
		}
	}
	m.elements[atspiRoot] = root
	m.mu.Unlock()
	for _, path := range paths {
		m.unexport(path)
	}
	virtualAccessibilityWindows.Delete(w)
}

func (m *atspiManager) export(path dbus.ObjectPath, application bool) {
	object := &atspiObject{manager: m, path: path}
	_ = m.conn.Export(object, path, atspiAccessible)
	_ = m.conn.Export(object, path, atspiComponent)
	_ = m.conn.Export(object, path, atspiAction)
	e, _ := object.element()
	if _, ok := e.node.object.(granio.AccessibleEditableText); ok {
		_ = m.conn.Export(&atspiTextObject{object}, path, atspiText)
		_ = m.conn.Export(&atspiTextObject{object}, path, atspiEditable)
	}
	if e.node.numeric {
		_ = m.conn.Export(&atspiValueObject{object}, path, atspiValue)
	}
	if _, ok := e.node.object.(granio.AccessibleSelection); ok {
		_ = m.conn.Export(&atspiSelectionObject{object}, path, atspiSelection)
	}
	if _, ok := e.node.object.(granio.AccessibleTable); ok {
		_ = m.conn.Export(&atspiTableObject{object}, path, atspiTable)
	}
	_ = m.conn.Export(object, path, "org.freedesktop.DBus.Properties")
	if application {
		_ = m.conn.Export(object, path, "org.a11y.atspi.Application")
	}
}
func (m *atspiManager) unexport(path dbus.ObjectPath) {
	if m.conn == nil {
		return
	}
	for _, iface := range []string{atspiAccessible, atspiComponent, atspiAction, atspiText, atspiEditable, atspiValue, atspiSelection, atspiTable, "org.freedesktop.DBus.Properties"} {
		_ = m.conn.Export(nil, path, iface)
	}
}

type atspiObject struct {
	manager *atspiManager
	path    dbus.ObjectPath
}

func (o *atspiObject) element() (atspiElement, *dbus.Error) {
	o.manager.mu.RLock()
	element, found := o.manager.elements[o.path]
	o.manager.mu.RUnlock()
	if !found {
		return atspiElement{}, dbus.NewError("org.a11y.atspi.Error.Defunct", nil)
	}
	return element, nil
}
func (o *atspiObject) GetChildren() ([]atspiReference, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return nil, err
	}
	if e.node.virtualOwner != nil {
		if e.node.virtualCount > 4096 {
			return nil, dbus.NewError("org.a11y.atspi.Error.TooManyChildren", []interface{}{"use GetChildAtIndex"})
		}
		children := make([]atspiReference, 0, e.node.virtualCount)
		for i := 0; i < e.node.virtualCount; i++ {
			child, err := o.GetChildAtIndex(int32(i))
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		}
		return children, nil
	}
	children := make([]atspiReference, len(e.children))
	for i, path := range e.children {
		children[i] = o.manager.reference(path)
	}
	return children, nil
}
func (o *atspiObject) GetChildAtIndex(index int32) (atspiReference, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return o.manager.reference(""), err
	}
	if e.node.virtualOwner != nil {
		path := dbus.ObjectPath("")
		accessibilityUI(func() {
			key, ok := e.window.virtualChild(e.node, int(index))
			if !ok {
				return
			}
			e.window.updateAccessibility()
			path = o.manager.pathForKey(e.window, key)
		})
		if path == "" {
			return o.manager.reference(""), dbus.NewError("org.a11y.atspi.Error.InvalidIndex", nil)
		}
		return o.manager.reference(path), nil
	}
	children, err := o.GetChildren()
	if err != nil {
		return atspiReference{}, err
	}
	if index < 0 || int(index) >= len(children) {
		return atspiReference{}, dbus.NewError("org.a11y.atspi.Error.InvalidIndex", nil)
	}
	return children[index], nil
}
func (o *atspiObject) GetIndexInParent() (int32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return -1, err
	}
	if e.node.virtualOwner != nil && e.node.virtualID != "" {
		return int32(e.node.virtualIndex), nil
	}
	o.manager.mu.RLock()
	parent := o.manager.elements[e.parent]
	o.manager.mu.RUnlock()
	for index, path := range parent.children {
		if path == o.path {
			return int32(index), nil
		}
	}
	return -1, nil
}
func (o *atspiObject) GetRole() (uint32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return 0, err
	}
	if o.path == atspiRoot {
		return 75, nil
	}
	role := map[granio.AccessibleRole]uint32{granio.AccessibleRoleContainer: 39, granio.AccessibleRoleButton: 43, granio.AccessibleRoleText: 29,
		granio.AccessibleRoleInput: 79, granio.AccessibleRoleCheck: 7, granio.AccessibleRoleList: 31, granio.AccessibleRoleListItem: 32,
		granio.AccessibleRoleTree: 65, granio.AccessibleRoleTreeItem: 32, granio.AccessibleRoleTable: 55, granio.AccessibleRoleCell: 56,
		granio.AccessibleRoleSlider: 51, granio.AccessibleRoleProgress: 42, granio.AccessibleRoleLink: 88, granio.AccessibleRoleWindow: 23, granio.AccessibleRoleSelect: 11, granio.AccessibleRoleImage: 27}[e.node.role]
	if role == 0 {
		role = 67
	}
	if e.node.state.Sensitive {
		role = 40
	}
	return role, nil
}
func (o *atspiObject) GetRoleName() (string, *dbus.Error) {
	e, err := o.element()
	return string(e.node.role), err
}
func (o *atspiObject) GetLocalizedRoleName() (string, *dbus.Error)    { return o.GetRoleName() }
func (o *atspiObject) GetRelationSet() ([]atspiRelation, *dbus.Error) { return []atspiRelation{}, nil }
func (o *atspiObject) GetApplication() (atspiReference, *dbus.Error) {
	return o.manager.reference(atspiRoot), nil
}
func (o *atspiObject) GetInterfaces() ([]string, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return nil, err
	}
	result := []string{atspiAccessible, atspiComponent}
	if len(e.node.actions) > 0 {
		result = append(result, atspiAction)
	}
	if _, ok := e.node.object.(granio.AccessibleEditableText); ok {
		result = append(result, atspiText, atspiEditable)
	}
	if e.node.numeric {
		result = append(result, atspiValue)
	}
	if _, ok := e.node.object.(granio.AccessibleSelection); ok {
		result = append(result, atspiSelection)
	}
	if _, ok := e.node.object.(granio.AccessibleTable); ok {
		result = append(result, atspiTable)
	}
	if o.path == atspiRoot {
		result = append(result, "org.a11y.atspi.Application")
	}
	return result, nil
}
func (o *atspiObject) GetAttributes() (map[string]string, *dbus.Error) {
	e, err := o.element()
	return map[string]string{"value": e.node.value}, err
}
func (o *atspiObject) GetState() ([]uint32, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return nil, err
	}
	bits := []uint32{0, 0}
	set := func(bit uint) { bits[bit/32] |= 1 << (bit % 32) }
	if !e.node.offscreen {
		set(25)
	}
	set(30) // showing, visible
	if !e.node.state.Disabled {
		set(8)
		set(24)
	}
	if e.node.state.Focused {
		set(12)
	}
	if e.node.state.Active {
		set(1)
	}
	if e.node.virtualOwner != nil && e.node.virtualID == "" {
		set(31)
	} // manages descendants
	if len(e.node.actions) > 0 {
		set(11)
	}
	if e.node.state.Checked {
		set(4)
	}
	if e.node.state.Selected {
		set(23)
	}
	if e.node.state.Expanded {
		set(10)
	}
	if e.node.role == granio.AccessibleRoleInput {
		set(7)
		set(11)
	}
	return bits, nil
}
func (o *atspiObject) GetExtents(coordinates uint32) (atspiRectangle, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return atspiRectangle{}, err
	}
	if coordinates > 2 {
		return atspiRectangle{}, dbus.NewError("org.a11y.atspi.Error.InvalidArguments", nil)
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
		e.rectangle.X -= origin.X
		e.rectangle.Y -= origin.Y
	}
	return e.rectangle, nil
}
func (o *atspiObject) GetPosition(coordinates uint32) (int32, int32, *dbus.Error) {
	r, err := o.GetExtents(coordinates)
	return r.X, r.Y, err
}
func (o *atspiObject) GetSize() (int32, int32, *dbus.Error) {
	e, err := o.element()
	return e.rectangle.Width, e.rectangle.Height, err
}
func (*atspiObject) GetLayer() (uint32, *dbus.Error)    { return 3, nil }
func (*atspiObject) GetMDIZOrder() (int16, *dbus.Error) { return 0, nil }
func (*atspiObject) GetAlpha() (float64, *dbus.Error)   { return 1, nil }
func (o *atspiObject) GetActions() ([]atspiActionInfo, *dbus.Error) {
	e, err := o.element()
	result := make([]atspiActionInfo, len(e.node.actions))
	for i, name := range e.node.actions {
		result[i].Name = name
	}
	return result, err
}
func (o *atspiObject) GetName(index int32) (string, *dbus.Error) {
	e, err := o.element()
	if index < 0 || int(index) >= len(e.node.actions) {
		return "", dbus.NewError("org.a11y.atspi.Error.InvalidIndex", nil)
	}
	return e.node.actions[index], err
}
func (o *atspiObject) GetLocalizedName(index int32) (string, *dbus.Error) { return o.GetName(index) }
func (*atspiObject) GetDescription(index int32) (string, *dbus.Error)     { return "", nil }
func (*atspiObject) GetKeyBinding(index int32) (string, *dbus.Error)      { return "", nil }
func (o *atspiObject) DoAction(index int32) (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	if e.node.state.Disabled || index < 0 || int(index) >= len(e.node.actions) {
		return false, nil
	}
	action := e.node.actions[index]
	queueAccessibility(func() {
		if current, err := o.element(); err == nil && current.node.object == e.node.object {
			if target, ok := e.node.object.(granio.AccessibleAction); ok {
				target.AccessibilityPerformAction(action)
			}
		}
	})
	return true, nil
}
func (o *atspiObject) GrabFocus() (bool, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return false, err
	}
	for i, action := range e.node.actions {
		if action == "focus" {
			return o.DoAction(int32(i))
		}
	}
	return false, nil
}
func (*atspiObject) GetLocale(uint32) (string, *dbus.Error) {
	return string(granio.CurrentDevice().Locale()), nil
}
func (*atspiObject) GetApplicationBusAddress() (string, *dbus.Error) { return "", nil }
func (o *atspiObject) Get(iface, name string) (dbus.Variant, *dbus.Error) {
	values, err := o.GetAll(iface)
	if err != nil {
		return dbus.Variant{}, err
	}
	value, found := values[name]
	if !found {
		return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", nil)
	}
	return value, nil
}
func (o *atspiObject) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	e, err := o.element()
	if err != nil {
		return nil, err
	}
	if iface == "org.a11y.atspi.Application" {
		o.manager.mu.RLock()
		id := o.manager.appID
		o.manager.mu.RUnlock()
		return map[string]dbus.Variant{"ToolkitName": dbus.MakeVariant("Fyne"), "Version": dbus.MakeVariant("2.8"), "Id": dbus.MakeVariant(id)}, nil
	}
	if iface == atspiText {
		return map[string]dbus.Variant{"CharacterCount": dbus.MakeVariant(int32(utf8.RuneCountInString(e.node.value))), "CaretOffset": dbus.MakeVariant(int32(e.node.cursor))}, nil
	}
	if iface == atspiValue {
		return map[string]dbus.Variant{"CurrentValue": dbus.MakeVariant(e.node.number), "MinimumValue": dbus.MakeVariant(e.node.minimum), "MaximumValue": dbus.MakeVariant(e.node.maximum), "MinimumIncrement": dbus.MakeVariant(e.node.step), "Text": dbus.MakeVariant(e.node.value)}, nil
	}
	if iface == atspiTable {
		table := &atspiTableObject{o}
		rows, err := table.GetSelectedRows()
		if err != nil {
			return nil, err
		}
		columns, err := table.GetSelectedColumns()
		if err != nil {
			return nil, err
		}
		return map[string]dbus.Variant{"NRows": dbus.MakeVariant(int32(e.node.rows)), "NColumns": dbus.MakeVariant(int32(e.node.columns)), "Caption": dbus.MakeVariant(o.manager.reference("")), "Summary": dbus.MakeVariant(o.manager.reference("")), "NSelectedRows": dbus.MakeVariant(int32(len(rows))), "NSelectedColumns": dbus.MakeVariant(int32(len(columns)))}, nil
	}
	if iface == atspiSelection {
		selected, err := (&atspiSelectionObject{o}).selectedIDs()
		if err != nil {
			return nil, err
		}
		count := len(selected)
		return map[string]dbus.Variant{"NSelectedChildren": dbus.MakeVariant(int32(count))}, nil
	}
	if iface != atspiAccessible {
		return nil, dbus.NewError("org.freedesktop.DBus.Error.UnknownInterface", nil)
	}
	parent := o.manager.reference(e.parent)
	if o.path == atspiRoot {
		o.manager.mu.RLock()
		parent = o.manager.parent
		o.manager.mu.RUnlock()
	}
	count := len(e.children)
	if e.node.virtualOwner != nil {
		count = e.node.virtualCount
	}
	return map[string]dbus.Variant{"Name": dbus.MakeVariant(e.node.name), "Description": dbus.MakeVariant(""), "Parent": dbus.MakeVariant(parent),
		"ChildCount": dbus.MakeVariant(int32(count)), "Locale": dbus.MakeVariant(o.manager.locale), "AccessibleId": dbus.MakeVariant(string(o.path)), "version": dbus.MakeVariant(uint32(1))}, nil
}
func (o *atspiObject) Set(iface, name string, value dbus.Variant) *dbus.Error {
	if iface == atspiValue && name == "CurrentValue" {
		if number, ok := value.Value().(float64); ok {
			applied, err := (&atspiValueObject{o}).SetCurrentValue(number)
			if err != nil {
				return err
			}
			if applied {
				return nil
			}
			return dbus.NewError("org.a11y.atspi.Error.InvalidArguments", nil)
		}
	}
	if iface == "org.a11y.atspi.Application" && name == "Id" {
		if id, ok := value.Value().(int32); ok {
			o.manager.mu.Lock()
			o.manager.appID = id
			o.manager.mu.Unlock()
			return nil
		}
	}
	return dbus.NewError("org.freedesktop.DBus.Error.PropertyReadOnly", nil)
}
