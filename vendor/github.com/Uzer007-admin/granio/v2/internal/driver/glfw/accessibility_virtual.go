//go:build accessibility && !wasm && !test_web_driver

package glfw

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Uzer007-admin/granio/v2"
)

type virtualAccessibleObject struct {
	owner granio.CanvasObject
	item  granio.AccessibleVirtualItem
}

func (*virtualAccessibleObject) MinSize() granio.Size                       { return granio.Size{} }
func (*virtualAccessibleObject) Size() granio.Size                          { return granio.Size{} }
func (*virtualAccessibleObject) Position() granio.Position                  { return granio.Position{} }
func (*virtualAccessibleObject) Move(granio.Position)                       {}
func (*virtualAccessibleObject) Resize(granio.Size)                         {}
func (*virtualAccessibleObject) Hide()                                      {}
func (*virtualAccessibleObject) Show()                                      {}
func (*virtualAccessibleObject) Visible() bool                              { return true }
func (*virtualAccessibleObject) Refresh()                                   {}
func (v *virtualAccessibleObject) AccessibilityLabel() string               { return v.item.Label }
func (v *virtualAccessibleObject) AccessibilityRole() granio.AccessibleRole { return v.item.Role }
func (v *virtualAccessibleObject) AccessibilityActions() []string {
	if v.item.ChildCount > 0 {
		return []string{"select", "add-selection", "remove-selection", "focus", "realize", "toggle"}
	}
	return []string{"select", "add-selection", "remove-selection", "focus", "realize"}
}
func (v *virtualAccessibleObject) AccessibilityPerformAction(action string) bool {
	return v.owner.(granio.AccessibleVirtualCollection).AccessibilityVirtualAction(v.item.ID, action)
}

// Only demanded logical providers are retained. Merely exposing a million-row
// child count neither invokes its data callback nor allocates a million objects.
// The map belongs to the UI thread; native readers always dispatch before use.
var virtualAccessibilityWindows sync.Map // *window -> map[accessibleKey]*virtualAccessibleObject
func virtualKey(owner granio.CanvasObject, id string) accessibleKey {
	return accessibleKey{identity: fmt.Sprintf("virtual:%p:%s", owner, id)}
}
func (w *window) requestVirtual(owner granio.CanvasObject, id string) (valid bool) {
	defer func() {
		if recover() != nil {
			valid = false
		}
	}()
	source, ok := owner.(granio.AccessibleVirtualCollection)
	if !ok {
		return false
	}
	stored, _ := virtualAccessibilityWindows.LoadOrStore(w, make(map[accessibleKey]*virtualAccessibleObject))
	items := stored.(map[accessibleKey]*virtualAccessibleObject)
	seen := make(map[string]bool)
	var request func(string) bool
	request = func(id string) bool {
		if seen[id] || len(seen) >= 256 {
			return false
		}
		seen[id] = true
		item, valid := source.AccessibilityVirtualItem(id)
		if !valid {
			delete(items, virtualKey(owner, id))
			return false
		}
		if item.Parent != "" && items[virtualKey(owner, item.Parent)] == nil && !request(item.Parent) {
			return false
		}
		key := virtualKey(owner, id)
		object := items[key]
		if object == nil {
			object = &virtualAccessibleObject{owner: owner}
			items[key] = object
		}
		object.item = item
		return true
	}
	return request(id)
}
func (w *window) mergeVirtualAccessibility(nodes []accessibleNode) []accessibleNode {
	owners := make(map[granio.CanvasObject]int)
	byKey := make(map[accessibleKey]int, len(nodes))
	for i := range nodes {
		n := &nodes[i]
		if collection, ok := n.object.(granio.AccessibleVirtualCollection); ok {
			n.virtualOwner = n.object
			n.virtualCount = collection.AccessibilityVirtualCount("")
			owners[n.object] = i
		}
		if _, ok := n.selectionOwner.(granio.AccessibleVirtualCollection); ok {
			n.virtualOwner = n.selectionOwner
			n.virtualID = strconv.Itoa(n.selectionIndex)
			n.virtualIndex = n.selectionIndex
			if n.tableOwner != nil {
				n.virtualID = fmt.Sprintf("%d:%d", n.row, n.column)
			}
			if n.role == granio.AccessibleRoleTreeItem {
				n.virtualID = strings.TrimPrefix(n.identity, fmt.Sprintf("tree:%p:", n.selectionOwner))
			}
			n.identity = virtualKey(n.virtualOwner, n.virtualID).identity
		}
		byKey[nodeKey(*n)] = i
	}
	stored, ok := virtualAccessibilityWindows.Load(w)
	if !ok {
		return nodes
	}
	items := stored.(map[accessibleKey]*virtualAccessibleObject)
	selected := make(map[granio.CanvasObject]map[string]bool)
	for owner := range owners {
		ids := owner.(granio.AccessibleVirtualCollection).AccessibilityVirtualSelection()
		set := make(map[string]bool, len(ids))
		for _, id := range ids {
			set[id] = true
		}
		selected[owner] = set
	}
	for key, object := range items {
		parent, exists := owners[object.owner]
		if !exists {
			delete(items, key)
			continue
		}
		if i, visible := byKey[key]; visible {
			nodes[i].virtualParent = object.item.Parent
			nodes[i].virtualIndex = object.item.Index
			nodes[i].virtualCount = object.item.ChildCount
			continue
		}
		item := object.item
		if item.Role == granio.AccessibleRoleCell {
			_, cols := object.owner.(granio.AccessibleTable).AccessibilityTableSize()
			item.Index = item.Row*cols + item.Column
			object.item.Index = item.Index
		}
		if id, valid := object.owner.(granio.AccessibleVirtualCollection).AccessibilityVirtualChild(item.Parent, item.Index); !valid || id != item.ID {
			delete(items, key)
			continue
		}
		item.State.Selected = selected[object.owner][item.ID]
		node := accessibleNode{object: object, parent: parent, name: item.Label, value: item.Value, role: item.Role, state: item.State, identity: key.identity, actions: object.AccessibilityActions(), virtualOwner: object.owner, virtualID: item.ID, virtualParent: item.Parent, virtualIndex: item.Index, virtualCount: item.ChildCount, offscreen: true, selectionOwner: object.owner, selectionIndex: item.Index, row: item.Row, column: item.Column}
		if item.Role == granio.AccessibleRoleCell {
			node.tableOwner = object.owner
			_, cols := object.owner.(granio.AccessibleTable).AccessibilityTableSize()
			node.selectionIndex = item.Row*cols + item.Column
		}
		byKey[key] = len(nodes)
		nodes = append(nodes, node)
	}
	// Build logical tree ancestry for demanded nodes, including closed branches.
	for i := range nodes {
		n := &nodes[i]
		if n.virtualOwner == nil || n.virtualID == "" {
			continue
		}
		if n.virtualParent != "" {
			if parent, ok := byKey[virtualKey(n.virtualOwner, n.virtualParent)]; ok {
				n.parent = parent
			}
		}
	}
	for i := range nodes {
		nodes[i].children = nil
	}
	for i := 1; i < len(nodes); i++ {
		p := nodes[i].parent
		nodes[p].children = append(nodes[p].children, i)
	}
	return nodes
}
func (w *window) virtualChild(node accessibleNode, index int) (accessibleKey, bool) {
	source, ok := node.virtualOwner.(granio.AccessibleVirtualCollection)
	if !ok {
		return accessibleKey{}, false
	}
	id, ok := source.AccessibilityVirtualChild(node.virtualID, index)
	if !ok || !w.requestVirtual(node.virtualOwner, id) {
		return accessibleKey{}, false
	}
	return virtualKey(node.virtualOwner, id), true
}
