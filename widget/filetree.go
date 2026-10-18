package widget

import (
	"sort"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// FileTree extends widget.Tree to display a file system hierarchy.
type FileTree struct {
	widget.Tree
	Filter       storage.FileFilter
	ShowRootPath bool
	Sorter       func(granio.URI, granio.URI) bool

	listCache     map[widget.TreeNodeID][]widget.TreeNodeID
	listableCache map[widget.TreeNodeID]granio.ListableURI
	uriCache      map[widget.TreeNodeID]granio.URI
}

// NewFileTree creates a new FileTree from the given root URI.
func NewFileTree(root granio.URI) *FileTree {
	tree := &FileTree{
		Tree: widget.Tree{
			Root: root.String(),
			CreateNode: func(branch bool) granio.CanvasObject {
				var icon granio.CanvasObject
				if branch {
					icon = widget.NewIcon(nil)
				} else {
					icon = widget.NewFileIcon(nil)
				}
				return container.NewBorder(nil, nil, icon, nil, widget.NewLabel("Template Object"))
			},
		},
		listCache:     make(map[widget.TreeNodeID][]widget.TreeNodeID),
		listableCache: make(map[widget.TreeNodeID]granio.ListableURI),
		uriCache:      make(map[widget.TreeNodeID]granio.URI),
	}
	tree.IsBranch = func(id widget.TreeNodeID) bool {
		_, err := tree.toListable(id)
		return err == nil
	}
	tree.ChildUIDs = func(id widget.TreeNodeID) (c []string) {
		listable, err := tree.toListable(id)
		if err != nil {
			granio.LogError("Unable to get lister for "+id, err)
			return
		}

		ids, ok := tree.listCache[id]
		if ok {
			return ids
		}

		uris, err := listable.List()
		if err != nil {
			granio.LogError("Unable to list "+listable.String(), err)
			return
		}

		for _, u := range tree.sort(tree.filter(uris)) {
			// Convert to String
			c = append(c, u.String())
		}

		tree.listCache[id] = c
		return
	}
	tree.UpdateNode = func(id widget.TreeNodeID, branch bool, node granio.CanvasObject) {
		uri, err := tree.toURI(id)
		if err != nil {
			granio.LogError("Unable to parse URI", err)
			return
		}

		c := node.(*granio.Container)
		if branch {
			var r granio.Resource
			if tree.IsBranchOpen(id) {
				// Set open folder icon
				r = theme.FolderOpenIcon()
			} else {
				// Set folder icon
				r = theme.FolderIcon()
			}
			c.Objects[1].(*widget.Icon).SetResource(r)
		} else {
			// Set file uri to update icon
			c.Objects[1].(*widget.FileIcon).SetURI(uri)
		}

		var l string
		if tree.Root == id && tree.ShowRootPath {
			l = id
		} else {
			l = uri.Name()
		}
		c.Objects[0].(*widget.Label).SetText(l)
	}

	// reset sorted child ID cache if the branch is closed - in the future we do FS watch
	tree.OnBranchClosed = func(id widget.TreeNodeID) {
		delete(tree.listCache, id)
	}

	tree.ExtendBaseWidget(tree)
	return tree
}

func (t *FileTree) Refresh() {
	t.Tree.Refresh()

	t.listCache = make(map[widget.TreeNodeID][]widget.TreeNodeID)
	t.listableCache = make(map[widget.TreeNodeID]granio.ListableURI)
	t.uriCache = make(map[widget.TreeNodeID]granio.URI)
}

// MapURI allows an app to return a specific URI for the given uid.
// This can be helpful to make more custom trees based on file structure/
func (t *FileTree) MapURI(uid string, target granio.URI) {
	t.uriCache[uid] = target
	t.Refresh()
}

func (t *FileTree) filter(uris []granio.URI) []granio.URI {
	filter := t.Filter
	if filter == nil {
		return uris
	}
	var filtered []granio.URI
	for _, u := range uris {
		if filter.Matches(u) {
			filtered = append(filtered, u)
		}
	}
	return filtered
}

func (t *FileTree) sort(uris []granio.URI) []granio.URI {
	if sorter := t.Sorter; sorter != nil {
		sort.Slice(uris, func(i, j int) bool {
			return sorter(uris[i], uris[j])
		})
	}
	return uris
}

func (t *FileTree) toListable(id widget.TreeNodeID) (granio.ListableURI, error) {
	listable, ok := t.listableCache[id]
	if ok {
		return listable, nil
	}
	uri, err := t.toURI(id)
	if err != nil {
		return nil, err
	}

	listable, err = storage.ListerForURI(uri)
	if err != nil {
		return nil, err
	}
	t.listableCache[id] = listable
	return listable, nil
}

func (t *FileTree) toURI(id widget.TreeNodeID) (granio.URI, error) {
	uri, ok := t.uriCache[id]
	if ok {
		return uri, nil
	}

	uri, err := storage.ParseURI(id)
	if err != nil {
		return nil, err
	}
	t.uriCache[id] = uri
	return uri, nil
}
