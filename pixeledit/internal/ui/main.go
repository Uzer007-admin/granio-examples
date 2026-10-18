package ui

import (
	"path/filepath"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func (e *editor) fileOpen() {
	open := dialog.NewFileOpen(func(read granio.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.win)
			return
		}
		if read == nil {
			return
		}

		e.LoadFile(read)
	}, e.win)

	open.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
	open.Show()
}

func (e *editor) fileReset() {
	win := granio.CurrentApp().Driver().AllWindows()[0]
	dialog.ShowConfirm("Reset content?", "Are you sure you want to re-load the image?",
		func(ok bool) {
			if !ok {
				return
			}

			e.Reload()
		}, win)
}

func (e *editor) fileSave() {
	e.Save()
}

func (e *editor) fileSaveAs() {
	open := dialog.NewFileSave(func(write granio.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.win)
			return
		}
		if write == nil {
			return
		}

		e.SaveAs(write)
	}, e.win)

	open.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
	open.Show()
}

func buildToolbar(e *editor) granio.CanvasObject {
	return widget.NewToolbar(
		&widget.ToolbarAction{Icon: theme.FolderOpenIcon(), OnActivated: e.fileOpen},
		&widget.ToolbarAction{Icon: theme.CancelIcon(), OnActivated: e.fileReset},
		&widget.ToolbarAction{Icon: theme.DocumentSaveIcon(), OnActivated: e.fileSave},
	)
}

func (e *editor) buildMainMenu() *granio.MainMenu {
	recents := granio.NewMenuItem("Open Recent", nil)
	recents.ChildMenu = e.loadRecentMenu()

	file := granio.NewMenu("File",
		granio.NewMenuItem("Open ...", e.fileOpen),
		recents,
		granio.NewMenuItemSeparator(),
		granio.NewMenuItem("Reset ...", e.fileReset),
		granio.NewMenuItem("Save", e.fileSave),
		granio.NewMenuItem("Save As ...", e.fileSaveAs),
	)

	return granio.NewMainMenu(file)
}

func (e *editor) loadRecentMenu() *granio.Menu {
	var items []*granio.MenuItem
	for _, item := range e.loadRecent() {
		u := item
		label := filepath.Base(item.String())

		items = append(items, granio.NewMenuItem(label, func() {
			read, err := storage.OpenFileFromURI(u)
			if err != nil {
				granio.LogError("Unable to open file \""+u.String()+"\"", err)
				return
			}
			e.LoadFile(read)
		}))
	}

	if e.recentMenu == nil {
		e.recentMenu = granio.NewMenu("Recent Items", items...)
	} else {
		e.recentMenu.Items = items
	}

	return e.recentMenu
}

// BuildUI creates the main window of our pixel edit application
func (e *editor) BuildUI(w granio.Window) {
	palette := newPalette(e)
	toolbar := buildToolbar(e)
	e.win = w
	w.SetMainMenu(e.buildMainMenu())

	content := granio.NewContainerWithLayout(layout.NewBorderLayout(toolbar, e.status, palette, nil),
		toolbar, e.status, palette, e.buildUI())
	w.SetContent(content)
}
