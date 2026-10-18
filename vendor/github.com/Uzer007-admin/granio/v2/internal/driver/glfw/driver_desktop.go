//go:build !wasm && !test_web_driver

package glfw

import (
	"bytes"
	"image/png"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio-systray"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/software"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/goos"
	"github.com/Uzer007-admin/granio/v2/internal/painter"
	"github.com/Uzer007-admin/granio/v2/internal/svg"
	"github.com/Uzer007-admin/granio/v2/lang"
	"github.com/Uzer007-admin/granio/v2/theme"
)

const systrayIconSize = 64

var (
	systrayIcon    granio.Resource
	systrayRunning bool
)

func (d *gLDriver) CreateWindow(title string) (win granio.Window) {
	async.EnsureMain(func() {
		win = d.newWindow(title, true)
	})
	return win
}

func (*gLDriver) HasSecondaryDisplay() bool {
	monitors := glfw.GetMonitors()
	if len(monitors) == 1 {
		return false
	}

	primaryTop, primaryLeft := monitors[0].GetPos()
	for _, m := range monitors[1:] {
		top, left := m.GetPos()
		if top != primaryTop || left != primaryLeft {
			return true
		}
	}

	return false // all the monitors had same origin, thus mirroring
}

func (d *gLDriver) SetSystemTrayMenu(m *granio.Menu) {
	if !systrayRunning {
		systrayRunning = true
		d.runSystray(m)
	}

	d.refreshSystray(m)
}

func (d *gLDriver) runSystray(m *granio.Menu) {
	d.trayStart, d.trayStop = systray.RunWithExternalLoop(func() {
		if systrayIcon != nil {
			d.SetSystemTrayIcon(systrayIcon)
		} else if granio.CurrentApp().Icon() != nil {
			d.SetSystemTrayIcon(granio.CurrentApp().Icon())
		} else {
			d.SetSystemTrayIcon(theme.BrokenImageIcon())
		}

		// Some XDG systray crash without a title (See #3678)
		if runtime.GOOS == goos.Linux || goos.IsBSD(runtime.GOOS) {
			app := granio.CurrentApp()
			title := app.Metadata().Name
			if title == "" {
				title = app.UniqueID()
			}

			systray.SetTitle(title)
		}

		if m != nil {
			// it must be refreshed after init, so an earlier call would have been ineffective
			runOnMain(func() {
				d.refreshSystray(m)
			})
		}
	}, func() {
		// anything required for tear-down
	})

	// the only way we know the app was asked to quit is if this window is asked to close...
	w := d.CreateWindow("SystrayMonitor")
	w.(*window).create()
	w.SetCloseIntercept(d.Quit)
}

// systrayShortcutKeys maps the few key names that Fyne spells differently to the
// platform neutral names that the systray package understands.
var systrayShortcutKeys = map[granio.KeyName]string{
	granio.KeyEnter:    "Enter",
	granio.KeyPageDown: "PageDown",
	granio.KeyPageUp:   "PageUp",
}

func systrayShortcutKey(key granio.KeyName) string {
	if name, ok := systrayShortcutKeys[key]; ok {
		return name
	}
	return string(key)
}

func systrayModifiers(mod granio.KeyModifier) (mods systray.KeyModifier) {
	if mod&granio.KeyModifierShift != 0 {
		mods |= systray.KeyModifierShift
	}
	if mod&granio.KeyModifierControl != 0 {
		mods |= systray.KeyModifierControl
	}
	if mod&granio.KeyModifierAlt != 0 {
		mods |= systray.KeyModifierAlt
	}
	if mod&granio.KeyModifierSuper != 0 {
		mods |= systray.KeyModifierSuper
	}
	return mods
}

func itemForMenuItem(i *granio.MenuItem, parent *systray.MenuItem) *systray.MenuItem {
	if i.IsSeparator {
		if parent != nil {
			parent.AddSeparator()
		} else {
			systray.AddSeparator()
		}
		return nil
	}

	var item *systray.MenuItem
	if i.Checked {
		if parent != nil {
			item = parent.AddSubMenuItemCheckbox(i.Label, "", true)
		} else {
			item = systray.AddMenuItemCheckbox(i.Label, "", true)
		}
	} else {
		if parent != nil {
			item = parent.AddSubMenuItem(i.Label, "")
		} else {
			item = systray.AddMenuItem(i.Label, "")
		}
	}
	if i.Disabled {
		item.Disable()
	}
	if s, ok := i.Shortcut.(granio.KeyboardShortcut); ok {
		item.SetShortcut(systrayModifiers(s.Mod()), systrayShortcutKey(s.Key()))
	}
	if i.Icon != nil {
		data := i.Icon.Content()
		if svg.IsResourceSVG(i.Icon) {
			b := &bytes.Buffer{}
			res := i.Icon
			if runtime.GOOS == goos.Windows && isDark() { // windows menus don't match dark mode so invert icons
				res = theme.NewInvertedThemedResource(i.Icon)
			}
			img := painter.PaintImage(canvas.NewImageFromResource(res), nil, systrayIconSize, systrayIconSize)
			err := png.Encode(b, img)
			if err != nil {
				granio.LogError("Failed to encode SVG icon for menu", err)
			} else {
				data = b.Bytes()
			}
		}

		img, err := toOSIcon(data)
		if err != nil {
			granio.LogError("Failed to convert systray icon", err)
		} else {
			if _, ok := i.Icon.(*theme.ThemedResource); ok {
				item.SetTemplateIcon(img, img)
			} else {
				item.SetIcon(img)
			}
		}
	}
	return item
}

func (d *gLDriver) refreshSystray(m *granio.Menu) {
	d.systrayMenu = m

	systray.ResetMenu()
	d.refreshSystrayMenu(m, nil)

	addMissingQuitForMenu(m, d)
}

func (d *gLDriver) refreshSystrayMenu(m *granio.Menu, parent *systray.MenuItem) {
	if m == nil {
		return
	}

	for _, i := range m.Items {
		item := itemForMenuItem(i, parent)
		if item == nil {
			continue // separator
		}
		if i.ChildMenu != nil {
			d.refreshSystrayMenu(i.ChildMenu, item)
		}

		fn := i.Action
		go func() {
			for range item.ClickedCh {
				if fn != nil {
					runOnMain(fn)
				}
			}
		}()
	}
}

func (*gLDriver) SetSystemTrayIcon(resource granio.Resource) {
	systrayIcon = resource // in case we need it later

	// only macOS supports SVG system tray
	if runtime.GOOS != goos.Darwin && svg.IsResourceSVG(resource) {
		img := canvas.NewImageFromResource(resource)
		c := software.NewTransparentCanvas()
		c.SetContent(img)
		c.SetPadded(false)
		c.Resize(granio.NewSquareSize(systrayIconSize))

		buf := &bytes.Buffer{}
		err := png.Encode(buf, c.Capture())
		if err != nil {
			granio.LogError("Failed to encode SVG icon for system tray icon", err)
			return
		}
		resource = granio.NewStaticResource(resource.Name()+".png", buf.Bytes())
	}

	img, err := toOSIcon(resource.Content())
	if err != nil {
		granio.LogError("Failed to convert systray icon", err)
		return
	}

	if _, ok := resource.(*theme.ThemedResource); ok {
		systray.SetTemplateIcon(img, img)
	} else {
		systray.SetIcon(img)
	}
}

func (d *gLDriver) SetSystemTrayWindow(w granio.Window) {
	if !systrayRunning {
		systrayRunning = true
		d.runSystray(nil)
	}

	w.SetCloseIntercept(w.Hide)
	glw, _ := w.(*window)
	if glw.decorate {
		systray.SetOnTapped(func() { granio.Do(glw.Show) })
	} else {
		systray.SetOnTapped(func() { granio.Do(glw.toggleVisible) })
	}
}

func (d *gLDriver) SystemTrayMenu() *granio.Menu {
	return d.systrayMenu
}

func (d *gLDriver) CurrentKeyModifiers() granio.KeyModifier {
	return d.currentKeyModifiers
}

// this function should be invoked from a goroutine
func (d *gLDriver) catchTerm() {
	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(terminateSignal)
	select {
	case <-terminateSignal:
		granio.Do(d.Quit)
	case <-d.done:
	}
}

func addMissingQuitForMenu(menu *granio.Menu, d *gLDriver) {
	localQuit := lang.L("Quit")
	var lastItem *granio.MenuItem
	if len(menu.Items) > 0 {
		lastItem = menu.Items[len(menu.Items)-1]
		if lastItem.Label == localQuit {
			lastItem.IsQuit = true
		}
	}
	if lastItem == nil || !lastItem.IsQuit { // make sure the menu always has a quit option
		quitItem := granio.NewMenuItem(localQuit, nil)
		quitItem.IsQuit = true
		menu.Items = append(menu.Items, granio.NewMenuItemSeparator(), quitItem)
	}
	for _, item := range menu.Items {
		if item.IsQuit && item.Action == nil {
			item.Action = d.Quit
		}
	}
}
