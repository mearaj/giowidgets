package main

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/mearaj/giowidgets"
	"log"
	"os"
	"time"
)

func main() {
	go func() {
		w := app.NewWindow()
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	c := giowidgets.Calendar{Theme: th}
	c.Inset = layout.UniformInset(unit.Dp(16))
	c.FirstDayOfWeek = time.Monday
	var ops op.Ops

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				c.Layout(gtx)
				e.Frame(gtx.Ops)
			}
		}
	}
}
