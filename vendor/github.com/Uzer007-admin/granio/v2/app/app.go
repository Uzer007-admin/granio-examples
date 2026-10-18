// Package app provides app implementations for working with Fyne graphical interfaces.
// The fastest way to get started is to call app.New() which will normally load a new desktop application.
// If the "ci" tag is passed to go (go run -tags ci myapp.go) it will run an in-memory application.
package app // import "github.com/Uzer007-admin/granio/v2/app"

import (
	"strconv"
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/internal/app"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	intRepo "github.com/Uzer007-admin/granio/v2/internal/repository"
	"github.com/Uzer007-admin/granio/v2/internal/scheduler"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/storage/repository"
)

// Declare conformity with App interface
var _ granio.App = (*fyneApp)(nil)

type fyneApp struct {
	driver    granio.Driver
	clipboard granio.Clipboard
	icon      granio.Resource
	uniqueID  string
	missingID bool // true if the developer did not supply their own ID

	cache     granio.Cache
	cloud     granio.CloudProvider
	lifecycle app.Lifecycle
	settings  *settings
	storage   granio.Storage
	prefs     granio.Preferences
	scheduler *scheduler.Scheduler

	secretPrefs     *secretPreferences
	secretPrefsLock sync.Mutex
}

func (a *fyneApp) CloudProvider() granio.CloudProvider {
	return a.cloud
}

func (a *fyneApp) Icon() granio.Resource {
	if a.icon != nil {
		return a.icon
	}

	if a.Metadata().Icon == nil || len(a.Metadata().Icon.Content()) == 0 {
		return nil
	}
	return a.Metadata().Icon
}

func (a *fyneApp) SetIcon(icon granio.Resource) {
	a.icon = icon
}

func (a *fyneApp) UniqueID() string {
	if a.uniqueID != "" {
		return a.uniqueID
	}
	if a.Metadata().ID != "" {
		return a.Metadata().ID
	}

	a.uniqueID = "missing-id-" + strconv.FormatInt(time.Now().Unix(), 10) // This is a fake unique - it just has to not be reused...
	a.missingID = true
	return a.uniqueID
}

func (a *fyneApp) NewWindow(title string) granio.Window {
	return a.driver.CreateWindow(title)
}

func (a *fyneApp) Run() {
	go a.lifecycle.RunEventQueue(a.driver.DoFromGoroutine)
	<-a.lifecycle.EventQueueReady()

	if !a.driver.Device().IsMobile() {
		a.settings.watchSettings()
		defer a.settings.stopWatching()
	}

	if !build.MigratedToFyneDo() && build.HasHints {
		async.PrintFyneDoWarning()
	}
	a.driver.Run()
}

func (a *fyneApp) Quit() {
	for _, window := range a.driver.AllWindows() {
		window.Close()
	}

	a.settings.stopWatching()
	a.driver.Quit()
}

func (a *fyneApp) Driver() granio.Driver {
	return a.driver
}

// Settings returns the application settings currently configured.
func (a *fyneApp) Settings() granio.Settings {
	return a.settings
}

func (a *fyneApp) Storage() granio.Storage {
	return a.storage
}

func (a *fyneApp) Preferences() granio.Preferences {
	if a.missingID {
		granio.LogError("Preferences API requires a unique ID, use app.NewWithID() or the FyneApp.toml ID field", nil)
	}
	return a.prefs
}

func (a *fyneApp) SecretPreferences(password []byte) (granio.Preferences, error) {
	if a.missingID {
		granio.LogError("SecretPreferences API requires a unique ID, use app.NewWithID() or the FyneApp.toml ID field", nil)
	}
	return a.secretPreferencesFrom(a.newSecretStore, password)
}

func (a *fyneApp) secretPreferencesFrom(newStore func([]byte) secretStore, password []byte) (granio.Preferences, error) {
	a.secretPrefsLock.Lock()
	defer a.secretPrefsLock.Unlock()
	if a.secretPrefs == nil { // created on first use as the fallback store needs the password
		p, err := newSecretPreferences(a, newStore, password)
		if err != nil { // not kept, so the stored data is never overwritten and the app can try again
			return nil, err
		}
		a.secretPrefs = p
	}
	return a.secretPrefs, nil
}

func (a *fyneApp) Lifecycle() granio.Lifecycle {
	return &a.lifecycle
}

func (a *fyneApp) newDefaultPreferences() *preferences {
	p := newPreferences(a)
	if a.uniqueID != "" {
		p.load()
	}
	return p
}

func (a *fyneApp) Cache() granio.Cache {
	return a.cache
}

func (a *fyneApp) Clipboard() granio.Clipboard {
	return a.clipboard
}

// New returns a new application instance with the default driver and no unique ID (unless specified in FyneApp.toml)
func New() granio.App {
	if meta.ID == "" {
		checkLocalMetadata() // if no ID passed, check if it was in toml
		if meta.ID == "" {
			internal.LogHint("Applications should be created with a unique ID using app.NewWithID()")
		}
	}
	return NewWithID(meta.ID)
}

func makeStoreDocs(id string, s *store) *internal.Docs {
	storageRoot := s.a.storageRoot()
	if id == "" || storageRoot == "" {
		return &internal.Docs{} // an empty impl to avoid crashes
	}

	uri, err := storage.ParseURI(storageRoot)
	if err != nil {
		uri = storage.NewFileURI(storageRoot)
	}

	exists, err := storage.Exists(uri)
	if !exists || err != nil {
		err = storage.CreateListable(uri)
		if err != nil {
			granio.LogError("Failed to create app storage space", err)
		}
	}

	docRoot, _ := s.docRootURI()
	return &internal.Docs{RootDocURI: docRoot}
}

func newAppWithDriver(d granio.Driver, clipboard granio.Clipboard, id string) granio.App {
	newApp := &fyneApp{uniqueID: id, clipboard: clipboard, driver: d}
	granio.SetCurrentApp(newApp)

	newApp.prefs = newApp.newDefaultPreferences()
	newApp.lifecycle.InitEventQueue()
	newApp.lifecycle.SetOnStoppedHookExecuted(func() {
		if prefs, ok := newApp.prefs.(*preferences); ok {
			prefs.forceImmediateSave()
		}
		newApp.secretPrefsLock.Lock()
		secretPrefs := newApp.secretPrefs
		newApp.secretPrefsLock.Unlock()
		if secretPrefs != nil {
			secretPrefs.forceImmediateSave()
		}
	})

	newApp.registerRepositories() // for web this may provide docs / settings
	newApp.cache = makeCache(newApp)
	newApp.scheduler = scheduler.New(newApp.cache, newApp.SendNotification)
	newApp.scheduler.Start()
	newApp.settings = loadSettings()
	store := &store{a: newApp}
	store.Docs = makeStoreDocs(id, store)
	newApp.storage = store

	httpHandler := intRepo.NewHTTPRepository()
	repository.Register("http", httpHandler)
	repository.Register("https", httpHandler)
	return newApp
}

// marker interface to pass system tray to supporting drivers
type systrayDriver interface {
	SetSystemTrayMenu(*granio.Menu)
	SetSystemTrayIcon(granio.Resource)
	SetSystemTrayWindow(granio.Window)
}
