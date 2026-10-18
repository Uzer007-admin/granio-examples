package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Temperature Converter")

	valueC := binding.NewFloat()
	valueF := celsiusToFarenheit(valueC)

	entryC := widget.NewEntryWithData(binding.FloatToString(valueC))
	entryC.Wrapping = granio.TextWrapOff
	entryF := widget.NewEntryWithData(binding.FloatToString(valueF))
	entryF.Wrapping = granio.TextWrapOff
	w.SetContent(container.NewGridWithColumns(4,
		entryC, widget.NewLabel("Celsius ="),
		entryF, widget.NewLabel("Fahrenheit")))

	w.ShowAndRun()
}

type cToF struct {
	binding.Float
}

func (c *cToF) Get() (float64, error) {
	cDeg, _ := c.Float.Get()
	fDeg := cDeg*(9.0/5.0) + 32
	return fDeg, nil
}

func (c *cToF) Set(f float64) error {
	cDeg := (f - 32) * (5.0 / 9.0)
	_ = c.Float.Set(cDeg)
	return nil
}

func celsiusToFarenheit(in binding.Float) binding.Float {
	return &cToF{in}
}
