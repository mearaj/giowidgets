package main

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/mearaj/giowidgets"
	"log"
	"os"
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

type CustomView struct {
	Title string
	*material.Theme
}

func (c *CustomView) Layout(gtx giowidgets.Gtx) layout.Dimensions {
	if c.Theme == nil {
		c.Theme = material.NewTheme(gofont.Collection())
	}
	return func(gtx giowidgets.Gtx) giowidgets.Dim {
		return material.Body1(c.Theme, c.Title).Layout(gtx)
	}(gtx)
}

func loop(w *app.Window) error {
	//resizer := giowidgets.Resize{}
	cust1 := CustomView{Title: "Widget One Widget One Widget One Widget One Widget One Widget One Widget One Widget One"}
	cust2 := CustomView{Title: "Widget Two Widget Two Widget Two Widget Two Widget Two Widget Two Widget Two Widget Two "}
	cust3 := CustomView{Title: "Widget Three Widget Three Widget Three Widget Three Widget Three Widget Three Widget Three"}
	cust4 := CustomView{Title: "Widget Four Widget Four Widget Four Widget Four Widget Four Widget Four Widget Four "}
	r1 := giowidgets.Resizable{Widget: cust1.Layout}
	r2 := giowidgets.Resizable{Widget: cust2.Layout}
	r3 := giowidgets.Resizable{Widget: cust3.Layout}
	r4 := giowidgets.Resizable{Widget: cust4.Layout}
	var ops op.Ops
	resizables := []*giowidgets.Resizable{&r1, &r2, &r3, &r4}
	resizer := giowidgets.NewResizeWidget(layout.Horizontal, resizables)

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				resizer.Layout(gtx)
				//resizer.Layout(gtx, cust2.Layout, nil)
				//resizer.Layout(gtx, cust3.Layout, nil)
				//resizer.Layout(gtx, cust4.Layout, nil)
				e.Frame(gtx.Ops)
			}
		}
	}
}
