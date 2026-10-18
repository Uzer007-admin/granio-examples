// Package test provides utility drivers for running UI tests without rendering to a screen.
package test // import "github.com/Uzer007-admin/granio/v2/test"

import (
	"net/url"
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal"
	intapp "github.com/Uzer007-admin/granio/v2/internal/app"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/internal/painter"
	"github.com/Uzer007-admin/granio/v2/internal/scheduler"
	"github.com/Uzer007-admin/granio/v2/internal/test"
	"github.com/Uzer007-admin/granio/v2/lang"
	"github.com/Uzer007-admin/granio/v2/theme"
)

// ensure we have a dummy app loaded and ready to test
func init() {
	NewApp()
}

type app struct {
	driver       *driver
	settings     *testSettings
	prefs        granio.Preferences
	secretPrefs  granio.Preferences
	propertyLock sync.RWMutex
	storage      granio.Storage
	lifecycle    intapp.Lifecycle
	cache        granio.Cache
	clip         granio.Clipboard
	cloud        granio.CloudProvider

	// user action variables
	appliedTheme              granio.Theme
	lastNotification          *granio.Notification
	scheduledNotifications    map[string]*granio.ScheduledNotification
	lastScheduledNotification *granio.ScheduledNotification
	lastCancelledScheduleID   string
}

func (a *app) CloudProvider() granio.CloudProvider {
	return a.cloud
}

func (*app) Icon() granio.Resource {
	return nil
}

func (*app) SetIcon(granio.Resource) {
	// no-op
}

func (a *app) NewWindow(title string) granio.Window {
	return a.driver.CreateWindow(title)
}

func (*app) OpenURL(_ *url.URL) error {
	// no-op
	return nil
}

func (*app) Run() {
	// no-op
}

func (*app) Quit() {
	// no-op
}

func (a *app) Cache() granio.Cache {
	return a.cache
}

func (a *app) Clipboard() granio.Clipboard {
	return a.clip
}

func (*app) UniqueID() string {
	return "testApp" // TODO should this be randomised?
}

func (a *app) Driver() granio.Driver {
	return a.driver
}

func (a *app) SendNotification(notify *granio.Notification) {
	a.propertyLock.Lock()
	defer a.propertyLock.Unlock()

	a.lastNotification = notify
}

func (a *app) ScheduleNotification(n *granio.Notification, when time.Time) (*granio.ScheduledNotification, error) {
	id, _ := scheduler.NewID()
	scheduled := granio.NewScheduledNotification(id, n, when)

	a.propertyLock.Lock()
	defer a.propertyLock.Unlock()

	if a.scheduledNotifications == nil {
		a.scheduledNotifications = map[string]*granio.ScheduledNotification{}
	}
	a.scheduledNotifications[id] = scheduled
	a.lastScheduledNotification = scheduled
	return scheduled, nil
}

func (a *app) CancelScheduledNotification(id string) error {
	a.propertyLock.Lock()
	defer a.propertyLock.Unlock()

	delete(a.scheduledNotifications, id)
	a.lastCancelledScheduleID = id

	return nil
}

func (a *app) SetCloudProvider(p granio.CloudProvider) {
	if p == nil {
		a.cloud = nil
		return
	}

	a.transitionCloud(p)
}

func (a *app) Settings() granio.Settings {
	return a.settings
}

func (a *app) Preferences() granio.Preferences {
	return a.prefs
}

func (a *app) SecretPreferences([]byte) (granio.Preferences, error) {
	return a.secretPrefs, nil
}

func (a *app) Storage() granio.Storage {
	return a.storage
}

func (a *app) Lifecycle() granio.Lifecycle {
	return &a.lifecycle
}

func (*app) Metadata() granio.AppMetadata {
	return granio.AppMetadata{} // just dummy data
}

func (a *app) lastAppliedTheme() granio.Theme {
	a.propertyLock.Lock()
	defer a.propertyLock.Unlock()

	return a.appliedTheme
}

func (a *app) transitionCloud(p granio.CloudProvider) {
	if a.cloud != nil {
		a.cloud.Cleanup(a)
	}

	err := p.Setup(a)
	if err != nil {
		granio.LogError("Failed to set up cloud provider "+p.ProviderName(), err)
		return
	}
	a.cloud = p

	listeners := a.prefs.ChangeListeners()
	if pp, ok := p.(granio.CloudProviderPreferences); ok {
		a.prefs = pp.CloudPreferences(a)
	} else {
		a.prefs = internal.NewInMemoryPreferences()
	}
	if store, ok := p.(granio.CloudProviderStorage); ok {
		a.storage = store.CloudStorage(a)
	} else {
		a.storage = &testStorage{}
	}

	for _, l := range listeners {
		a.prefs.AddChangeListener(l)
		l() // assume that preferences have changed because we replaced the provider
	}

	// after transition ensure settings listener is fired
	a.settings.apply()
}

// NewApp returns a new dummy app used for testing.
// It loads a test driver which creates a virtual window in memory for testing.
func NewApp() granio.App {
	settings := &testSettings{scale: 1.0, theme: Theme()}
	prefs := internal.NewInMemoryPreferences()
	secretPrefs := internal.NewInMemoryPreferences()
	store := &testStorage{}
	testApp := &app{
		settings: settings, prefs: prefs, secretPrefs: secretPrefs, storage: store, driver: NewDriver().(*driver),
		clip: NewClipboard(), cache: makeCache(),
	}
	settings.app = testApp
	root, _ := store.docRootURI()
	store.Docs = &internal.Docs{RootDocURI: root}
	painter.ClearFontCache()
	cache.ResetThemeCaches()
	granio.SetCurrentApp(testApp)
	lang.SetLocale(testApp.Driver().Device().Locale())

	return testApp
}

type testSettings struct {
	primaryColor string
	scale        float32
	theme        granio.Theme

	listeners       []func(granio.Settings)
	changeListeners []chan granio.Settings
	propertyLock    sync.RWMutex
	app             *app
}

func (s *testSettings) AddChangeListener(listener chan granio.Settings) {
	s.propertyLock.Lock()
	defer s.propertyLock.Unlock()
	s.changeListeners = append(s.changeListeners, listener)
}

func (s *testSettings) AddListener(listener func(granio.Settings)) {
	s.propertyLock.Lock()
	defer s.propertyLock.Unlock()
	s.listeners = append(s.listeners, listener)
}

func (*testSettings) BuildType() granio.BuildType {
	return granio.BuildStandard
}

func (s *testSettings) PrimaryColor() string {
	if s.primaryColor != "" {
		return s.primaryColor
	}

	return theme.ColorBlue
}

func (s *testSettings) SetTheme(t granio.Theme) {
	s.propertyLock.Lock()
	s.theme = t
	s.propertyLock.Unlock()

	s.apply()
}

func (*testSettings) ShowAnimations() bool {
	return true
}

func (s *testSettings) Theme() granio.Theme {
	s.propertyLock.RLock()
	defer s.propertyLock.RUnlock()

	if s.theme == nil {
		return test.DarkTheme(theme.DefaultTheme())
	}

	return s.theme
}

func (*testSettings) ThemeVariant() granio.ThemeVariant {
	return 2 // not a preference
}

func (s *testSettings) Scale() float32 {
	s.propertyLock.RLock()
	defer s.propertyLock.RUnlock()
	return s.scale
}

func (s *testSettings) apply() {
	s.propertyLock.RLock()
	listeners := s.changeListeners
	listenersFns := s.listeners
	s.propertyLock.RUnlock()

	for _, listener := range listeners {
		listener <- s
	}

	s.app.propertyLock.Lock()
	painter.ClearFontCache()
	cache.ResetThemeCaches()
	intapp.ApplySettings(s, s.app)
	s.app.propertyLock.Unlock()

	for _, l := range listenersFns {
		l(s)
	}

	s.app.propertyLock.Lock()
	s.app.appliedTheme = s.Theme()
	s.app.propertyLock.Unlock()
}
