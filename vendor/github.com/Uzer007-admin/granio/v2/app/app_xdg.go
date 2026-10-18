//go:build !ci && !wasm && !test_web_driver && !android && !ios && !mobile && (linux || openbsd || freebsd || netbsd) && !tinygo && !noos && !tamago

package app

import (
	"net/url"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/dbus/v5/desktopportal/notification"
	"github.com/Uzer007-admin/dbus/v5/desktopportal/openuri"
	"github.com/Uzer007-admin/dbus/v5/desktopportal/settings/appearance"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	"github.com/Uzer007-admin/granio/v2/theme"
)

const systemTheme = granio.ThemeVariant(99)

func (*fyneApp) OpenURL(u *url.URL) error {
	if build.IsFlatpak {
		err := openuri.OpenURI("", u.String(), nil)
		if err != nil {
			granio.LogError("Opening url in portal failed", err)
		}
		return err
	}

	cmd := exec.Command("xdg-open", u.String()) //gosec:disable G204 -- It’s the callers responsibility to validate the input.
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Start()
}

// fetch color variant from dbus portal desktop settings.
func findFreedesktopColorScheme() granio.ThemeVariant {
	colorScheme, err := appearance.GetColorScheme()
	if err != nil {
		return systemTheme
	}

	return colorSchemeToThemeVariant(colorScheme)
}

func colorSchemeToThemeVariant(colorScheme appearance.ColorScheme) granio.ThemeVariant {
	switch colorScheme {
	case appearance.Light:
		return theme.VariantLight
	case appearance.Dark:
		return theme.VariantDark
	}

	// Default to light theme to support Gnome's default see https://github.com/fyne-io/fyne/pull/3561
	return theme.VariantLight
}

func (a *fyneApp) SendNotification(n *granio.Notification) {
	if build.IsFlatpak {
		err := a.sendNotificationThroughPortal(n)
		if err != nil {
			granio.LogError("Sending notification using portal failed", err)
		}
		return
	}

	conn, err := dbus.SessionBus() // shared connection, don't close
	if err != nil {
		granio.LogError("Unable to connect to session D-Bus", err)
		return
	}

	appIcon := a.cachedIconPath()
	timeout := int32(0) // we don't support this yet

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call("org.freedesktop.Notifications.Notify", 0, a.uniqueID, uint32(0),
		appIcon, n.Title, n.Content, []string{}, map[string]dbus.Variant{}, timeout)
	if call.Err != nil {
		granio.LogError("Failed to send message to bus", call.Err)
	}
}

// The freedesktop notification spec has no scheduling primitive, so scheduled
// notifications use the in-process scheduler with cache-backed persistence.
func (a *fyneApp) ScheduleNotification(n *granio.Notification, when time.Time) (*granio.ScheduledNotification, error) {
	return a.scheduleViaScheduler(n, when)
}

func (a *fyneApp) CancelScheduledNotification(id string) error {
	return a.cancelViaScheduler(id)
}

// Sending with same ID replaces the old notification.
var notificationID atomic.Uint64

// See https://flatpak.github.io/xdg-desktop-portal/docs/#gdbus-org.freedesktop.portal.Notification.
func (a *fyneApp) sendNotificationThroughPortal(n *granio.Notification) error {
	return notification.Add(
		uint(notificationID.Add(1)),
		notification.Content{
			Title: n.Title,
			Body:  n.Content,
			Icon:  a.uniqueID,
		},
	)
}

// SetSystemTrayMenu creates a system tray item and attaches the specified menu.
// By default, this will use the application icon.
func (a *fyneApp) SetSystemTrayMenu(menu *granio.Menu) {
	if desk, ok := a.Driver().(systrayDriver); ok { // don't use this on mobile tag
		desk.SetSystemTrayMenu(menu)
	}
}

// SetSystemTrayIcon sets a custom image for the system tray icon.
// You should have previously called `SetSystemTrayMenu` to initialise the menu icon.
func (a *fyneApp) SetSystemTrayIcon(icon granio.Resource) {
	if desk, ok := a.Driver().(systrayDriver); ok { // don't use this on mobile tag
		desk.SetSystemTrayIcon(icon)
	}
}

// SetSystemTrayWindow assigns a window to be shown with the system tray menu is tapped.
// You should have previously called `SetSystemTrayMenu` to initialise the menu icon.
func (a *fyneApp) SetSystemTrayWindow(w granio.Window) {
	a.Driver().(systrayDriver).SetSystemTrayWindow(w)
}

func watchTheme(s *settings) {
	s.startSystemAppearance()
}

func (*fyneApp) registerRepositories() {
	// no-op
}

func (s *settings) applyVariant(variant granio.ThemeVariant) {
	s.variant = variant
	s.apply()
}
