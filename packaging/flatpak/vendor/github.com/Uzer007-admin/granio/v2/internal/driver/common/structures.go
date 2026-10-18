package common

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/build"
)

type deduplicatedObjectQueue struct {
	queue *async.CanvasObjectQueue
	dedup async.Map[granio.CanvasObject, struct{}] //revive:disable-line:nested-structs -- The empty struct here is a common pattern to simulate a set via a map.
}

// In adds an object to the queue if it is not already present.
func (q *deduplicatedObjectQueue) In(obj granio.CanvasObject) {
	_, exists := q.dedup.Load(obj)
	if exists {
		return
	}

	q.queue.In(obj)
	q.dedup.Store(obj, struct{}{})
}

// Out removes and returns the next object from the queue.
// It assumes that the whole queue is drained and defers clearing
// the deduplication map until it is empty.
func (q *deduplicatedObjectQueue) Out() granio.CanvasObject {
	if q.queue.Len() == 0 {
		q.dedup.Clear()
		return nil
	}

	out := q.queue.Out()
	if !build.MigratedToFyneDo() {
		q.dedup.Delete(out)
	}
	return out
}

// Len returns the number of elements in the queue.
func (q *deduplicatedObjectQueue) Len() uint64 {
	return q.queue.Len()
}

type renderCacheTree struct {
	root *RenderCacheNode
}

// RenderCacheNode represents a node in a render cache tree.
type RenderCacheNode struct {
	// structural data
	firstChild  *RenderCacheNode
	nextSibling *RenderCacheNode
	obj         granio.CanvasObject
	parent      *RenderCacheNode
	// cache data
	minSize granio.Size
}

// Obj returns the node object.
func (r *RenderCacheNode) Obj() granio.CanvasObject {
	return r.obj
}

type overlayStack struct {
	internal.OverlayStack

	renderCaches []*renderCacheTree
}

func (o *overlayStack) Add(overlay granio.CanvasObject) {
	if overlay == nil {
		return
	}

	o.renderCaches = append(o.renderCaches, &renderCacheTree{root: &RenderCacheNode{obj: overlay}})
	o.OverlayStack.Add(overlay)
}

func (o *overlayStack) Remove(overlay granio.CanvasObject) {
	if overlay == nil || len(o.List()) == 0 {
		return
	}

	o.OverlayStack.Remove(overlay)
	overlayCount := len(o.List())

	clear(o.renderCaches[overlayCount:])
	o.renderCaches = o.renderCaches[:overlayCount]
}
