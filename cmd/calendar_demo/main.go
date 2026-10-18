package main

import (
	"time"

	xwidget "github.com/Uzer007-admin/granio-x/widget"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Calendar")

	i := widget.NewLabel("Please Choose a Date")
	i.Alignment = granio.TextAlignCenter
	l := widget.NewLabel("")
	l.Alignment = granio.TextAlignCenter
	d := &date{instruction: i, dateChosen: l}

	// Defines which date you would like the calendar to start
	startingDate := time.Now()
	calendar := xwidget.NewCalendar(startingDate, d.onSelected)

	c := container.NewVBox(i, l, calendar)

	w.SetContent(c)
	w.ShowAndRun()
}

type date struct {
	instruction *widget.Label
	dateChosen  *widget.Label
}

func (d *date) onSelected(t time.Time) {
	// use time object to set text on label with given format
	d.instruction.SetText("Date Selected:")
	d.dateChosen.SetText(t.Format("Mon 02 Jan 2006"))
}
