//go:build accessibility && !wasm && !test_web_driver

package glfw

import (
	"fmt"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	internalWidget "github.com/Uzer007-admin/granio/v2/internal/widget"
)

type accessibleNode struct {
	object                               granio.CanvasObject
	parent                               int
	children                             []int
	name, value                          string
	role                                 granio.AccessibleRole
	state                                granio.AccessibleState
	position                             granio.Position
	size                                 granio.Size
	actions                              []string
	cursor, selectionStart, selectionEnd int
	identity                             string
	rows, columns, row, column           int
	tableOwner                           granio.CanvasObject
	selectionOwner                       granio.CanvasObject
	selectionIndex                       int
	numeric                              bool
	minimum, maximum, step, number       float64
	multiple                             bool
	virtualOwner                         granio.CanvasObject
	virtualID, virtualParent             string
	virtualIndex, virtualCount           int
	offscreen                            bool
}

type accessibleKey struct {
	object   granio.CanvasObject
	identity string
}

func nodeKey(node accessibleNode) accessibleKey {
	if node.identity != "" {
		return accessibleKey{identity: node.identity}
	}
	return accessibleKey{object: node.object}
}
func accessibilityUI(fn func()) {
	if async.IsMainGoroutine() {
		fn()
	} else {
		runOnMain(fn)
	}
}

// Native assistive-technology callbacks belong to this driver's UI queue. Using
// the process-wide CurrentApp driver would dispatch through unrelated test/apps.
func queueAccessibility(fn func()) { runOnMainWithWait(fn, false) }

// Snapshot only visible controls and renderer children. Virtual collections
// therefore expose viewport nodes without walking their complete data sources.
func (w *window) accessibilitySnapshot() []accessibleNode {
	nodes := []accessibleNode{{parent: -1, name: w.title, role: granio.AccessibleRoleWindow, size: w.canvas.Size(), state: granio.AccessibleState{Active: curWindow == w}, offscreen: !w.visible}}
	seen := make(map[granio.CanvasObject]bool)
	tableCells := make(map[granio.CanvasObject]struct {
		cell  granio.AccessibleTableCell
		owner granio.CanvasObject
	})
	var walk func(granio.CanvasObject, granio.Position, int, granio.Position, granio.Size)
	walk = func(object granio.CanvasObject, position granio.Position, parent int, clipPosition granio.Position, clipSize granio.Size) {
		if object == nil || !object.Visible() || seen[object] {
			return
		}
		seen[object] = true
		position = position.Add(object.Position())
		size := object.Size()
		offscreen := position.X+size.Width <= clipPosition.X || position.Y+size.Height <= clipPosition.Y || position.X >= clipPosition.X+clipSize.Width || position.Y >= clipPosition.Y+clipSize.Height
		role := granio.AccessibleRoleContainer
		accessible, isAccessible := object.(granio.Accessible)
		cell, isCell := tableCells[object]
		if isAccessible || isCell {
			node := accessibleNode{object: object, parent: parent, position: position, size: size, offscreen: offscreen}
			valid := false
			func() {
				// A broken custom accessibility method must not crash the event loop.
				defer func() { _ = recover() }()
				if isAccessible {
					role = accessible.AccessibilityRole()
					node.name = accessible.AccessibilityLabel()
				}
				node.role = role
				if value, ok := object.(granio.AccessibleValue); ok {
					node.value = value.AccessibilityValue()
				}
				if state, ok := object.(granio.AccessibleStateProvider); ok {
					node.state = state.AccessibilityState()
				}
				if focusable, ok := object.(granio.Focusable); ok {
					node.state.Focused = focusable == w.canvas.Focused()
				}
				if actions, ok := object.(granio.AccessibleAction); ok {
					node.actions = actions.AccessibilityActions()
				}
				if text, ok := object.(granio.AccessibleEditableText); ok {
					node.cursor, node.selectionStart, node.selectionEnd = text.AccessibilityTextSelection()
				}
				if identity, ok := object.(granio.AccessibleIdentity); ok {
					node.identity = identity.AccessibilityIdentity()
				}
				if item, ok := object.(granio.AccessibleCollectionItem); ok {
					node.selectionOwner, node.selectionIndex = item.AccessibilityCollectionIndex()
				}
				if value, ok := object.(granio.AccessibleNumeric); ok {
					node.number, node.minimum, node.maximum, node.step = value.AccessibilityNumericValue()
					node.numeric = true
				}
				if selection, ok := object.(granio.AccessibleSelection); ok {
					node.multiple = selection.AccessibilityCanSelectMultiple()
				}
				if table, ok := object.(granio.AccessibleTable); ok {
					node.rows, node.columns = table.AccessibilityTableSize()
					for _, item := range table.AccessibilityVisibleCells() {
						tableCells[item.Object] = struct {
							cell  granio.AccessibleTableCell
							owner granio.CanvasObject
						}{item, object}
					}
				}
				if isCell {
					role = granio.AccessibleRoleCell
					node.role = role
					node.row, node.column = cell.cell.Row, cell.cell.Column
					node.tableOwner = cell.owner
					node.selectionOwner = cell.owner
					_, columns := cell.owner.(granio.AccessibleTable).AccessibilityTableSize()
					node.selectionIndex = node.row*columns + node.column
					node.state.Selected = cell.cell.Selected
					node.actions = []string{"select", "add-selection", "remove-selection"}
					node.identity = fmt.Sprintf("cell:%p:%d:%d", cell.owner, node.row, node.column)
				}
				valid = true
			}()
			if !valid {
				return
			}
			if node.state.Sensitive {
				node.value = ""
			}
			// Layout containers inside a collection have no independent semantic
			// role. Keep row/cell providers as direct children of their selection
			// container, rather than hiding them behind renderer implementation.
			collectionParent := nodes[parent].role == granio.AccessibleRoleList || nodes[parent].role == granio.AccessibleRoleTree || nodes[parent].role == granio.AccessibleRoleTable
			if role != granio.AccessibleRoleContainer || !collectionParent {
				id := len(nodes)
				nodes = append(nodes, node)
				nodes[parent].children = append(nodes[parent].children, id)
				parent = id
				// Leaf control renderers often contain text duplicating their value.
				if role != granio.AccessibleRoleContainer && role != granio.AccessibleRoleList && role != granio.AccessibleRoleTable && role != granio.AccessibleRoleTree {
					return
				}
			}
		}
		if container, ok := object.(*granio.Container); ok {
			for _, child := range container.Objects {
				walk(child, position, parent, clipPosition, clipSize)
			}
		} else if widget, ok := object.(granio.Widget); ok {
			if _, ok := object.(*internalWidget.Scroll); ok {
				left, top := max(position.X, clipPosition.X), max(position.Y, clipPosition.Y)
				right, bottom := min(position.X+size.Width, clipPosition.X+clipSize.Width), min(position.Y+size.Height, clipPosition.Y+clipSize.Height)
				clipPosition, clipSize = granio.NewPos(left, top), granio.NewSize(max(0, right-left), max(0, bottom-top))
			}
			if renderer := cache.Renderer(widget); renderer != nil {
				for _, child := range renderer.Objects() {
					walk(child, position, parent, clipPosition, clipSize)
				}
			}
		}
	}
	walk(w.canvas.Content(), granio.Position{}, 0, granio.Position{}, w.canvas.Size())
	walk(w.canvas.menu, granio.Position{}, 0, granio.Position{}, w.canvas.Size())
	for _, overlay := range w.canvas.Overlays().List() {
		walk(overlay, granio.Position{}, 0, granio.Position{}, w.canvas.Size())
	}
	return w.mergeVirtualAccessibility(nodes)
}
