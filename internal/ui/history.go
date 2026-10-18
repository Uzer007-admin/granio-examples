package ui

import (
	"fmt"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/storage"
)

const (
	recentCountKey  = "recentCount"
	recentFormatKey = "recent-%d"
)

func (e *editor) loadRecent() []granio.URI {
	pref := granio.CurrentApp().Preferences()
	count := pref.Int(recentCountKey)

	var recent []granio.URI
	for i := 0; i < count; i++ {
		key := fmt.Sprintf(recentFormatKey, i)
		u, _ := storage.ParseURI(pref.String(key))
		recent = append(recent, u)
	}

	return recent
}

func (e *editor) addRecent(u granio.URI) {
	pref := granio.CurrentApp().Preferences()
	recent := e.loadRecent()

	recent = append([]granio.URI{u}, recent...)
	if len(recent) > 5 {
		recent = recent[:5]
	}

	pref.SetInt(recentCountKey, len(recent))
	for i, uri := range recent {
		key := fmt.Sprintf(recentFormatKey, i)
		pref.SetString(key, uri.String())
	}

	e.loadRecentMenu()
	e.win.SetMainMenu(e.win.MainMenu())
}
