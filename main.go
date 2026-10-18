// Package main provides various examples of Fyne API capabilities.
package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/cmd/granio_settings/settings"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/Uzer007-admin/granio-demo/data"
	"github.com/Uzer007-admin/granio-demo/tutorials"
)

const preferenceCurrentTutorial = "currentTutorial"

var topWindow granio.Window

func main() {
	a := app.NewWithID("io.github.uzer007_admin.granio.demo")
	a.SetIcon(data.FyneLogo)
	w := a.NewWindow("Granio Demo")
	makeTray(a, w)
	logLifecycle(a)
	topWindow = w

	w.SetMainMenu(makeMenu(a, w))
	w.SetMaster()

	content := container.NewStack()
	title := widget.NewLabel("Component name")
	intro := widget.NewLabel("An introduction would probably go\nhere, as well as a")
	intro.Wrapping = granio.TextWrapWord

	top := container.NewVBox(title, widget.NewSeparator(), intro)
	setTutorial := func(t tutorials.Tutorial) {
		if granio.CurrentDevice().IsMobile() {
			child := a.NewWindow(t.Title)
			topWindow = child
			child.SetContent(t.View(topWindow))
			child.Show()
			child.SetOnClosed(func() {
				topWindow = w
			})
			return
		}

		title.SetText(t.Title)
		isMarkdown := len(t.Intro) == 0
		if !isMarkdown {
			intro.SetText(t.Intro)
		}

		if t.Title == "Welcome" || isMarkdown {
			top.Hide()
		} else {
			top.Show()
		}

		content.Objects = []granio.CanvasObject{t.View(w)}
		content.Refresh()
	}

	tutorial := container.NewBorder(
		top, nil, nil, nil, content)
	if granio.CurrentDevice().IsMobile() {
		w.SetContent(makeNav(setTutorial, false))
	} else {
		split := container.NewHSplit(makeNav(setTutorial, true), tutorial)
		split.Offset = 0.2
		w.SetContent(split)
	}
	w.Resize(granio.NewSize(640, 460))
	w.ShowAndRun()
}

func logLifecycle(a granio.App) {
	a.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	a.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	a.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	a.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

func makeMenu(a granio.App, w granio.Window) *granio.MainMenu {
	newItem := granio.NewMenuItem("New", nil)
	checkedItem := granio.NewMenuItem("Checked", nil)
	checkedItem.Checked = true
	disabledItem := granio.NewMenuItem("Disabled", nil)
	disabledItem.Disabled = true
	otherItem := granio.NewMenuItem("Other", nil)
	mailItem := granio.NewMenuItem("Mail", func() { fmt.Println("Menu New->Other->Mail") })
	mailItem.Icon = theme.MailComposeIcon()
	otherItem.ChildMenu = granio.NewMenu("",
		granio.NewMenuItem("Project", func() { fmt.Println("Menu New->Other->Project") }),
		mailItem,
	)
	fileItem := granio.NewMenuItem("File", func() { fmt.Println("Menu New->File") })
	fileItem.Icon = theme.FileIcon()
	dirItem := granio.NewMenuItem("Directory", func() { fmt.Println("Menu New->Directory") })
	dirItem.Icon = theme.FolderIcon()
	newItem.ChildMenu = granio.NewMenu("",
		fileItem,
		dirItem,
		otherItem,
	)

	openSettings := func() {
		w := a.NewWindow("Настройки Granio")
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(granio.NewSize(440, 520))
		w.Show()
	}
	showAbout := func() {
		w := a.NewWindow("About")
		w.SetContent(widget.NewLabel("Демонстрация компонентов Granio."))
		w.Show()
	}
	aboutItem := granio.NewMenuItem("About", showAbout)
	settingsItem := granio.NewMenuItem("Settings", openSettings)
	settingsShortcut := &desktop.CustomShortcut{KeyName: granio.KeyComma, Modifier: granio.KeyModifierShortcutDefault}
	settingsItem.Shortcut = settingsShortcut
	w.Canvas().AddShortcut(settingsShortcut, func(shortcut granio.Shortcut) {
		openSettings()
	})

	cutShortcut := &granio.ShortcutCut{Clipboard: a.Clipboard()}
	cutItem := granio.NewMenuItem("Cut", func() {
		shortcutFocused(cutShortcut, a.Clipboard(), w.Canvas().Focused())
	})
	cutItem.Shortcut = cutShortcut
	copyShortcut := &granio.ShortcutCopy{Clipboard: a.Clipboard()}
	copyItem := granio.NewMenuItem("Copy", func() {
		shortcutFocused(copyShortcut, a.Clipboard(), w.Canvas().Focused())
	})
	copyItem.Shortcut = copyShortcut
	pasteShortcut := &granio.ShortcutPaste{Clipboard: a.Clipboard()}
	pasteItem := granio.NewMenuItem("Paste", func() {
		shortcutFocused(pasteShortcut, a.Clipboard(), w.Canvas().Focused())
	})
	pasteItem.Shortcut = pasteShortcut
	performFind := func() { fmt.Println("Menu Find") }
	findItem := granio.NewMenuItem("Find", performFind)
	findItem.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyF, Modifier: granio.KeyModifierShortcutDefault | granio.KeyModifierAlt | granio.KeyModifierShift | granio.KeyModifierControl | granio.KeyModifierSuper}
	w.Canvas().AddShortcut(findItem.Shortcut, func(shortcut granio.Shortcut) {
		performFind()
	})

	helpMenu := granio.NewMenu("Help",
		granio.NewMenuItem("Documentation", func() {
			u, _ := url.Parse("https://github.com/Uzer007-admin/granio-docs")
			_ = a.OpenURL(u)
		}),
		granio.NewMenuItem("Support", func() {
			u, _ := url.Parse("https://github.com/Uzer007-admin/granio/issues")
			_ = a.OpenURL(u)
		}),
		granio.NewMenuItemSeparator(),
		granio.NewMenuItem("Исходный проект Fyne", func() {
			u, _ := url.Parse("https://github.com/fyne-io/fyne")
			_ = a.OpenURL(u)
		}))

	// a quit item will be appended to our first (File) menu
	file := granio.NewMenu("File", newItem, checkedItem, disabledItem)
	device := granio.CurrentDevice()
	if !device.IsMobile() && !device.IsBrowser() {
		file.Items = append(file.Items, granio.NewMenuItemSeparator(), settingsItem)
	}
	file.Items = append(file.Items, aboutItem)
	main := granio.NewMainMenu(
		file,
		granio.NewMenu("Edit", cutItem, copyItem, pasteItem, granio.NewMenuItemSeparator(), findItem),
		helpMenu,
	)
	checkedItem.Action = func() {
		checkedItem.Checked = !checkedItem.Checked
		main.Refresh()
	}
	return main
}

func makeTray(a granio.App, w granio.Window) {
	if desk, ok := a.(desktop.App); ok {
		h := granio.NewMenuItem("Hello", func() {})
		h.Icon = theme.HomeIcon()
		menu := granio.NewMenu("Hello World", h)
		h.Action = func() {
			log.Println("System tray menu tapped")
			h.Label = "Welcome"
			menu.Refresh()
		}
		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayWindow(w)
	}
}

func makeNav(setTutorial func(tutorial tutorials.Tutorial), loadPrevious bool) granio.CanvasObject {
	a := granio.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return tutorials.TutorialIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := tutorials.TutorialIndex[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) granio.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj granio.CanvasObject) {
			t, ok := tutorials.Tutorials[uid]
			if !ok {
				granio.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			if t, ok := tutorials.Tutorials[uid]; ok {
				for _, f := range tutorials.OnChangeFuncs {
					f()
				}
				tutorials.OnChangeFuncs = nil // Loading a page registers a new cleanup.

				a.Preferences().SetString(preferenceCurrentTutorial, uid)
				setTutorial(t)
			}
		},
	}

	if loadPrevious {
		currentPref := a.Preferences().StringWithFallback(preferenceCurrentTutorial, "welcome")
		tree.Select(currentPref)
	}

	themes := container.NewGridWithColumns(2,
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantDark})
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(&forcedVariant{Theme: theme.DefaultTheme(), variant: theme.VariantLight})
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, tree)
}

func shortcutFocused(s granio.Shortcut, cb granio.Clipboard, f granio.Focusable) {
	switch sh := s.(type) {
	case *granio.ShortcutCopy:
		sh.Clipboard = cb
	case *granio.ShortcutCut:
		sh.Clipboard = cb
	case *granio.ShortcutPaste:
		sh.Clipboard = cb
	}
	if focused, ok := f.(granio.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}
