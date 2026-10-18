package tutorials

import (
	"fmt"
	"image/color"
	"net/url"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/validation"
	"github.com/Uzer007-admin/granio/v2/driver/mobile"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/Uzer007-admin/granio-examples/demo/data"
)

const (
	loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque quis consectetur nisi. Suspendisse id interdum felis.
Sed egestas eget tellus eu pharetra. Praesent pulvinar sed massa id placerat. Etiam sem libero, semper vitae consequat ut, volutpat id mi.
Mauris volutpat pellentesque convallis. Curabitur rutrum venenatis orci nec ornare. Maecenas quis pellentesque neque.
Aliquam consectetur dapibus nulla, id maximus odio ultrices ac. Sed luctus at felis sed faucibus. Cras leo augue, congue in velit ut, mattis rhoncus lectus.

Praesent viverra, mauris ut ullamcorper semper, leo urna auctor lectus, vitae vehicula mi leo quis lorem.
Nullam condimentum, massa at tempor feugiat, metus enim lobortis velit, eget suscipit eros ipsum quis tellus. Aenean fermentum diam vel felis dictum semper.
Duis nisl orci, tincidunt ut leo quis, luctus vehicula diam. Sed velit justo, congue id augue eu, euismod dapibus lacus. Proin sit amet imperdiet sapien.
Mauris erat urna, fermentum et quam rhoncus, fringilla consequat ante. Vivamus consectetur molestie odio, ac rutrum erat finibus a.
Suspendisse id maximus felis. Sed mauris odio, mattis eget mi eu, consequat tempus purus.`
)

func makeCardTab(_ granio.Window) granio.CanvasObject {
	card1 := widget.NewCard("Book a table", "Which time suits?",
		widget.NewRadioGroup([]string{"6:30pm", "7:00pm", "7:45pm"}, func(string) {}))
	card2 := widget.NewCard("With media", "No content, with image", nil)
	card2.Image = canvas.NewImageFromResource(data.FyneLogo)
	card3 := widget.NewCard("Title 3", "Another card", widget.NewLabel("Content"))
	return container.NewGridWithColumns(2, container.NewVBox(card1, card3),
		container.NewVBox(card2))
}

func makeEntryTab(_ granio.Window) granio.CanvasObject {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Entry")
	entryDisabled := widget.NewEntry()
	entryDisabled.SetText("Entry (disabled)")
	entryDisabled.Disable()
	entryValidated := newNumEntry()
	entryValidated.SetPlaceHolder("Must contain a number")
	entryMultiLine := widget.NewMultiLineEntry()
	entryMultiLine.SetPlaceHolder("MultiLine Entry")
	entryMultiLine.SetMinRowsVisible(4)

	return container.NewVBox(
		entry,
		entryDisabled,
		entryValidated,
		entryMultiLine)
}

func makeTextGrid() *widget.TextGrid {
	grid := widget.NewTextGridFromString("TextGrid\n\tContent\nZebra")
	grid.SetStyleRange(0, 4, 0, 7,
		&widget.CustomTextGridStyle{BGColor: &color.NRGBA{R: 64, G: 64, B: 192, A: 128}})
	grid.SetRowStyle(1, &widget.CustomTextGridStyle{BGColor: &color.NRGBA{R: 64, G: 192, B: 64, A: 128}})

	white := &widget.CustomTextGridStyle{FGColor: color.White, BGColor: color.Black}
	black := &widget.CustomTextGridStyle{FGColor: color.Black, BGColor: color.White}
	grid.Rows[2].Cells[0].Style = white
	grid.Rows[2].Cells[1].Style = black
	grid.Rows[2].Cells[2].Style = white
	grid.Rows[2].Cells[3].Style = black
	grid.Rows[2].Cells[4].Style = white

	grid.ShowLineNumbers = true
	grid.ShowWhitespace = true

	return grid
}

func makeTextTab(_ granio.Window) granio.CanvasObject {
	label := widget.NewLabel("Label")

	link, err := url.Parse("https://fyne.io/")
	if err != nil {
		granio.LogError("Could not parse URL", err)
	}
	hyperlink := widget.NewHyperlink("Hyperlink", link)

	entryLoremIpsum := widget.NewMultiLineEntry()
	entryLoremIpsum.SetText(loremIpsum)

	label.Alignment = granio.TextAlignLeading
	hyperlink.Alignment = granio.TextAlignLeading

	label.Wrapping = granio.TextWrapWord
	hyperlink.Wrapping = granio.TextWrapWord
	entryLoremIpsum.Wrapping = granio.TextWrapWord

	rich := widget.NewRichTextFromMarkdown(`
# RichText Heading

## A Sub Heading

![title](data/assets/fyne.png)

---

* Item1 in _three_ segments
* Item2
* Item3

Normal **Bold** *Italic* [Link](https://fyne.io/) and some ` + "`Code`" + `.
This styled row should also wrap as expected, but only *when required*.

> An interesting quote here, most likely sharing some very interesting wisdom.`)
	rich.Scroll = container.ScrollBoth

	radioAlign := widget.NewRadioGroup([]string{"Leading", "Center", "Trailing"}, func(s string) {
		var align granio.TextAlign
		switch s {
		case "Leading":
			align = granio.TextAlignLeading
		case "Center":
			align = granio.TextAlignCenter
		case "Trailing":
			align = granio.TextAlignTrailing
		}

		label.Alignment = align
		hyperlink.Alignment = align
		for _, r := range rich.Segments {
			switch t := r.(type) {
			case *widget.TextSegment:
				t.Style.Alignment = align
			case *widget.HyperlinkSegment:
				t.Alignment = align
			case *widget.ImageSegment:
				t.Alignment = align
			}
		}

		label.Refresh()
		hyperlink.Refresh()
		rich.Refresh()
	})
	radioAlign.Horizontal = true
	radioAlign.SetSelected("Leading")

	radioWrap := widget.NewRadioGroup([]string{"Off", "Scroll", "Break", "Word"}, func(s string) {
		var wrap granio.TextWrap
		scroll := container.ScrollBoth
		switch s {
		case "Off":
			wrap = granio.TextWrapOff
			scroll = container.ScrollNone
		case "Scroll":
			wrap = granio.TextWrapOff
		case "Break":
			wrap = granio.TextWrapBreak
		case "Word":
			wrap = granio.TextWrapWord
		}

		label.Wrapping = wrap
		hyperlink.Wrapping = wrap
		entryLoremIpsum.Wrapping = wrap
		entryLoremIpsum.Scroll = scroll
		rich.Wrapping = wrap

		label.Refresh()
		hyperlink.Refresh()
		entryLoremIpsum.Refresh()
		rich.Refresh()
	})
	radioWrap.Horizontal = true
	radioWrap.SetSelected("Word")

	radioTrunc := widget.NewRadioGroup([]string{"Off", "Clip", "Ellipsis"}, func(s string) {
		var trunc granio.TextTruncation
		switch s {
		case "Off":
			trunc = granio.TextTruncateOff
		case "Clip":
			trunc = granio.TextTruncateClip
		case "Ellipsis":
			trunc = granio.TextTruncateEllipsis
		}

		label.Truncation = trunc
		rich.Truncation = trunc

		label.Refresh()
		hyperlink.Refresh()
		entryLoremIpsum.Refresh()
		rich.Refresh()
	})
	radioTrunc.Horizontal = true
	radioTrunc.SetSelected("Off")

	fixed := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Text Alignment", radioAlign),
			widget.NewFormItem("Wrapping", radioWrap),
			widget.NewFormItem("Truncation", radioTrunc)),
		label,
		hyperlink,
	)

	grid := makeTextGrid()
	return container.NewBorder(fixed, grid, nil, nil,
		container.NewGridWithRows(2, rich, entryLoremIpsum))
}

func makeInputTab(_ granio.Window) granio.CanvasObject {
	selectEntry := widget.NewSelectEntry([]string{
		"Option A",
		"Option B",
		"Option C",
		"Option D",
		"Option E",
		"Option F",
		"Option G",
		"Option H",
		"Option I",
		"Option J",
		"Option K",
		"Option L",
		"Option M",
		"Option N",
		"Option O",
		"Option P",
		"Option Q",
		"Option R",
		"Option S",
		"Option T",
		"Option U",
		"Option V",
		"Option W",
		"Option X",
		"Option Y",
		"Option Z",
	})
	selectEntry.PlaceHolder = "Type or select"
	dateEntry := widget.NewDateEntry()
	dateEntry.PlaceHolder = "Choose a date"
	disabledCheck := widget.NewCheck("Disabled check", func(bool) {})
	disabledCheck.Disable()
	partialCheck := widget.NewCheck("Partial check", func(bool) {})
	partialCheck.Partial = true
	checkGroup := widget.NewCheckGroup([]string{"CheckGroup Item 1", "CheckGroup Item 2"}, func(s []string) { fmt.Println("selected", s) })
	checkGroup.Horizontal = true
	radio := widget.NewRadioGroup([]string{"Radio Item 1", "Radio Item 2"}, func(s string) { fmt.Println("selected", s) })
	radio.Horizontal = true
	disabledRadio := widget.NewRadioGroup([]string{"Disabled radio"}, func(string) {})
	disabledRadio.Disable()

	disabledSlider := widget.NewSlider(0, 1000)
	disabledSlider.Disable()
	return container.NewVBox(
		widget.NewSelect([]string{"Option 1", "Option 2", "Option 3"}, func(s string) { fmt.Println("selected", s) }),
		selectEntry,
		dateEntry,
		widget.NewCheck("Check", func(on bool) { fmt.Println("checked", on) }),
		partialCheck,
		checkGroup,
		radio,
		container.NewGridWithColumns(2, disabledCheck, disabledRadio),
		container.NewBorder(nil, nil, widget.NewLabel("Slider"), nil, widget.NewSlider(0, 1000)),
		container.NewBorder(nil, nil, widget.NewLabel("Disabled slider"), nil, disabledSlider),
	)
}

func makeProgressTab(_ granio.Window) granio.CanvasObject {
	progress := widget.NewProgressBar()

	fprogress := widget.NewProgressBar()
	fprogress.TextFormatter = func() string {
		return fmt.Sprintf("%.2f out of %.2f", fprogress.Value, fprogress.Max)
	}

	infProgress := widget.NewProgressBarInfinite()

	progressAnimation := granio.Animation{
		Curve:    granio.AnimationLinear,
		Duration: 10 * time.Second,
		Tick: func(percentage float32) {
			value := float64(percentage)
			progress.SetValue(value)
			fprogress.SetValue(value)
		},
	}

	progressAnimation.Start()

	OnChangeFuncs = append(OnChangeFuncs, progressAnimation.Stop, infProgress.Stop)

	return container.NewVBox(
		widget.NewLabel("Percent"), progress,
		widget.NewLabel("Formatted"), fprogress,
		widget.NewLabel("Infinite"), infProgress)
}

func makeFormTab(_ granio.Window) granio.CanvasObject {
	name := widget.NewEntry()
	name.SetPlaceHolder("John Smith")

	email := widget.NewEntry()
	email.SetPlaceHolder("test@example.com")
	email.Validator = validation.NewRegexp(`\w{1,}@\w{1,}\.\w{1,4}`, "not a valid email")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	disabled := widget.NewRadioGroup([]string{"Option 1", "Option 2"}, func(string) {})
	disabled.Horizontal = true
	disabled.Disable()
	largeText := widget.NewMultiLineEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: name, HintText: "Your full name"},
			{Text: "Email", Widget: email, HintText: "A valid email address"},
		},
		OnCancel: func() {
			fmt.Println("Cancelled")
		},
		OnSubmit: func() {
			fmt.Println("Form submitted")
			granio.CurrentApp().SendNotification(&granio.Notification{
				Title:   "Form for: " + name.Text,
				Content: largeText.Text,
			})
		},
	}
	form.Append("Password", password)
	form.Append("Disabled", disabled)
	form.Append("Message", largeText)

	form.Items[2].Required = true
	form.Items[3].Required = true
	form.Refresh()
	return form
}

func makeToolbarTab(_ granio.Window) granio.CanvasObject {
	t := widget.NewToolbar(widget.NewToolbarAction(theme.MailComposeIcon(), func() { fmt.Println("New") }),
		widget.NewToolbarSeparator(),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() { fmt.Println("Cut") }),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() { fmt.Println("Copy") }),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() { fmt.Println("Paste") }),
	)

	return container.NewBorder(t, nil, nil, nil)
}

// widgetScreen shows a panel containing widget demos
func widgetScreen(_ granio.Window) granio.CanvasObject {
	content := container.NewVBox(
		widget.NewLabel("Labels"),
		widget.NewButtonWithIcon("Icons", theme.HomeIcon(), func() {}),
		widget.NewSlider(0, 1))
	return container.NewCenter(content)
}

type numEntry struct {
	widget.Entry
}

func (n *numEntry) Keyboard() mobile.KeyboardType {
	return mobile.NumberKeyboard
}

func newNumEntry() *numEntry {
	e := &numEntry{}
	e.ExtendBaseWidget(e)
	e.Validator = validation.NewRegexp(`\d`, "Must contain a number")
	return e
}
