package binding

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"

	"github.com/Uzer007-admin/granio/v2"
)

var errAsyncClosed = errors.New("asynchronous data source is closed")

type asyncRequest struct {
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	err      error
	finished bool
}

type asyncSource[T any] struct {
	mu      sync.Mutex
	applyMu sync.Mutex
	parent  context.Context
	loader  func(context.Context) (T, error)
	apply   func(T) error
	current *asyncRequest
	closed  bool
}

// Reload starts a background load and cancels the preceding request. Providers
// must honor the supplied context. A stale result never overwrites a newer one.
func (s *asyncSource[T]) Reload() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return errAsyncClosed
	}
	if err := s.parent.Err(); err != nil {
		s.mu.Unlock()
		return err
	}
	if s.current != nil {
		s.current.cancel()
	}
	ctx, cancel := context.WithCancel(s.parent)
	request := &asyncRequest{ctx: ctx, cancel: cancel, done: make(chan struct{})}
	s.current = request
	s.mu.Unlock()
	go func() {
		value, err := s.loader(ctx)
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		// Serialize commits, including providers that return after cancellation.
		// Reload and Close do not acquire this lock, so data listeners can safely
		// request another load from inside a change notification.
		granio.Do(func() {
			defer cancel()
			if ctx.Err() != nil {
				err = ctx.Err()
			}
			s.applyMu.Lock()
			s.mu.Lock()
			valid := !s.closed && s.current == request && ctx.Err() == nil
			s.mu.Unlock()
			if valid && err == nil {
				err = s.apply(value)
			}
			s.mu.Lock()
			request.err, request.finished = err, true
			close(request.done)
			s.mu.Unlock()
			s.applyMu.Unlock()
		})
	}()
	return nil
}

// Loading reports whether the current request is still running.
func (s *asyncSource[T]) Loading() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.closed && s.current != nil && !s.current.finished && s.current.ctx.Err() == nil
}

// LoadError returns the current request's error after completion. Failed loads
// leave the previously committed data available to widgets.
func (s *asyncSource[T]) LoadError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current == nil || !s.current.finished {
		if s.current != nil {
			return s.current.ctx.Err()
		}
		return s.parent.Err()
	}
	return s.current.err
}

// Wait waits for the request active at the time of this call. Cancelling this
// wait does not cancel the source; use Close or the source's parent context.
// Call Wait from a background goroutine: committing a result uses the UI thread.
func (s *asyncSource[T]) Wait(ctx context.Context) error {
	s.mu.Lock()
	request := s.current
	s.mu.Unlock()
	if request == nil {
		return s.parent.Err()
	}
	select {
	case <-request.done:
		return request.err
	default:
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-request.ctx.Done():
		select {
		case <-request.done:
			return request.err
		default:
			return request.ctx.Err()
		}
	case <-request.done:
		return request.err
	}
}

// Close cancels the current request and prevents further reloads. It does not
// wait for a provider that ignores cancellation, and retains committed data.
func (s *asyncSource[T]) Close() {
	s.mu.Lock()
	s.closed = true
	if s.current != nil {
		s.current.cancel()
	}
	s.mu.Unlock()
}

// AsyncList is a normal bindable List whose snapshots load in the background.
// Bind it with widget.NewListWithData, or use rows of slices for Table data.
type AsyncList[T any] struct {
	List[T]
	*asyncSource[[]T]
}

// Get returns a copy of the committed slice. Values of T are copied shallowly.
func (l *AsyncList[T]) Get() ([]T, error) {
	values, err := l.List.Get()
	return slices.Clone(values), err
}

// NewAsyncList creates a list and starts its first load. Snapshots are copied
// before committing; the provider must not mutate a returned slice concurrently.
func NewAsyncList[T any](ctx context.Context, loader func(context.Context) ([]T, error), comparator func(T, T) bool) *AsyncList[T] {
	list := NewList(comparator)
	source := &asyncSource[[]T]{parent: ctx, loader: loader,
		apply: func(values []T) error {
			sameLength := list.Length() == len(values)
			err := list.Set(slices.Clone(values))
			if err == nil && sameLength {
				list.(*boundList[T]).trigger()
			}
			return err
		}}
	result := &AsyncList[T]{List: list, asyncSource: source}
	_ = result.Reload()
	return result
}

// TreeSnapshot holds a complete hierarchy and its node values. Root has ID "".
type TreeSnapshot[T any] struct {
	Children map[string][]string
	Values   map[string]T
}

// AsyncTree is a bindable Tree loaded as atomic background snapshots. It works
// with widget.NewTreeWithData, including retaining selection by stable node IDs.
type AsyncTree[T any] struct {
	Tree[T]
	*asyncSource[TreeSnapshot[T]]
}

// Get returns copies of the hierarchy and value maps. Values of T are copied
// shallowly, while the child ID slices are independent of the source.
func (t *AsyncTree[T]) Get() (map[string][]string, map[string]T, error) {
	children, values, err := t.Tree.Get()
	copy := make(map[string][]string, len(children))
	for parent, ids := range children {
		copy[parent] = slices.Clone(ids)
	}
	return copy, maps.Clone(values), err
}

// NewAsyncTree creates a tree and starts its first load. Returned maps and child
// slices are copied before committing. Providers must honor context cancellation.
func NewAsyncTree[T any](ctx context.Context, loader func(context.Context) (TreeSnapshot[T], error), comparator func(T, T) bool) *AsyncTree[T] {
	tree := NewTree(comparator)
	source := &asyncSource[TreeSnapshot[T]]{parent: ctx, loader: loader,
		apply: func(snapshot TreeSnapshot[T]) error {
			children := make(map[string][]string, len(snapshot.Children))
			for parent, ids := range snapshot.Children {
				children[parent] = slices.Clone(ids)
			}
			previous, _, _ := tree.Get()
			sameHierarchy := treeChildrenEqual(previous, children)
			err := tree.Set(children, maps.Clone(snapshot.Values))
			if err == nil && sameHierarchy {
				tree.(*boundTree[T]).trigger()
			}
			return err
		}}
	result := &AsyncTree[T]{Tree: tree, asyncSource: source}
	_ = result.Reload()
	return result
}
