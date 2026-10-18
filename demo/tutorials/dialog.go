package tutorials

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/validation"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func confirmCallback(response bool) {
	fmt.Println("Responded with", response)
}

func colorPicked(c color.Color, w granio.Window) {
	log.Println("Color picked:", c)
	rectangle := canvas.NewRectangle(c)
	size := 2 * theme.IconInlineSize()
	rectangle.SetMinSize(granio.NewSize(size, size))
	dialog.ShowCustom("Color Picked", "OK", rectangle, w)
}

// dialogScreen loads demos of the dialogs we support
func dialogScreen(win granio.Window) granio.CanvasObject {
	openFile := widget.NewButton("File Open With Filter (.jpg or .png)", func() {
		fd := dialog.NewFileOpen(func(reader granio.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if reader == nil {
				log.Println("Cancelled")
				return
			}

			imageOpened(reader)
		}, win)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg"}))
		fd.Show()
	})
	saveFile := widget.NewButton("File Save", func() {
		dialog.ShowFileSave(func(writer granio.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if writer == nil {
				log.Println("Cancelled")
				return
			}

			fileSaved(writer, win)
		}, win)
	})
	openFolder := widget.NewButton("Folder Open", func() {
		dialog.ShowFolderOpen(func(list granio.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if list == nil {
				log.Println("Cancelled")
				return
			}

			children, err := list.List()
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			out := fmt.Sprintf("Folder %s (%d children):\n%s", list.Name(), len(children), list.String())
			dialog.ShowInformation("Folder Open", out, win)
		}, win)
	})

	if granio.CurrentDevice().IsBrowser() {
		openFile.Disable()
		saveFile.Disable()
		openFolder.Disable()
	}

	advancedPicker := dialog.NewColorPicker("Pick a Color", "What is your favorite color?", func(c color.Color) {
		colorPicked(c, win)
	}, win)
	advancedPicker.Advanced = true
	return container.NewVScroll(container.NewVBox(
		widget.NewButton("Info", func() {
			dialog.ShowInformation("Information", "You should know this thing...", win)
		}),
		widget.NewButton("Error", func() {
			err := errors.New("a dummy error message")
			dialog.ShowError(err, win)
		}),
		widget.NewButton("Confirm", func() {
			cnf := dialog.NewConfirm("Confirmation", "Are you enjoying this demo?", confirmCallback, win)
			cnf.SetDismissText("Nah")
			cnf.SetConfirmText("Oh Yes!")
			cnf.Show()
		}),
		openFile,
		saveFile,
		openFolder,
		widget.NewButton("Color Picker", func() {
			picker := dialog.NewColorPicker("Pick a Color", "What is your favorite color?", func(c color.Color) {
				colorPicked(c, win)
			}, win)
			picker.Show()
		}),
		widget.NewButton("Advanced Color Picker", func() {
			advancedPicker.Show()
		}),
		widget.NewButton("Form Dialog (Login Form)", func() {
			username := widget.NewEntry()
			username.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]+$`, "username can only contain letters, numbers, '_', and '-'")
			password := widget.NewPasswordEntry()
			password.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]+$`, "password can only contain letters, numbers, '_', and '-'")
			remember := false
			items := []*widget.FormItem{
				widget.NewFormItem("Username", username),
				widget.NewFormItem("Password", password),
				widget.NewFormItem("Remember me", widget.NewCheck("", func(checked bool) {
					remember = checked
				})),
			}

			dialog.ShowForm("Login...", "Log In", "Cancel", items, func(b bool) {
				if !b {
					return
				}
				var rememberText string
				if remember {
					rememberText = "and remember this login"
				}

				log.Println("Please Authenticate", username.Text, password.Text, rememberText)
			}, win)
		}),
	))
}

func imageOpened(f granio.URIReadCloser) {
	if f == nil {
		log.Println("Cancelled")
		return
	}
	defer f.Close()

	showImage(f)
}

func fileSaved(f granio.URIWriteCloser, w granio.Window) {
	defer f.Close()
	_, err := f.Write([]byte("Written by Fyne demo\n"))
	if err != nil {
		dialog.ShowError(err, w)
	}
	err = f.Close()
	if err != nil {
		dialog.ShowError(err, w)
	}
	log.Println("Saved to...", f.URI())
}

func loadImage(f granio.URIReadCloser) *canvas.Image {
	data, err := io.ReadAll(f)
	if err != nil {
		granio.LogError("Failed to load image data", err)
		return nil
	}
	res := granio.NewStaticResource(f.URI().Name(), data)

	return canvas.NewImageFromResource(res)
}

func showImage(f granio.URIReadCloser) {
	img := loadImage(f)
	if img == nil {
		return
	}
	img.FillMode = canvas.ImageFillOriginal

	w := granio.CurrentApp().NewWindow(f.URI().Name())
	w.SetContent(container.NewScroll(img))
	w.Resize(granio.NewSize(320, 240))
	w.Show()
}
