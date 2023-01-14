package giowidgets

import (
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"image"
	"image/color"
)

// Resize provides a draggable handle in between two widgets for resizing their area.
type Resize struct {
	// axis defines how the widgets and the handle are laid out.
	axis               layout.Axis
	initialized        bool
	length             int
	totalHandlesLength int
	resizables         []*Resizable
}

type Resizable struct {
	Ratio  float32
	Widget layout.Widget
	Handle layout.Widget
	float
	resize *Resize
	prev   *Resizable
	next   *Resizable
}

// Layout displays w1 and w2 with handle in between.
//
// The widgets w1 and w2 must be able to gracefully resize their minimum and maximum dimensions
// in order for the resize to be smooth.
func (r *Resize) Layout(gtx layout.Context, axis layout.Axis, resizables []*Resizable) layout.Dimensions {
	// Compute the first widget's max width/height.
	r.resizables = resizables
	if len(resizables) == 0 {
		return layout.Dimensions{}
	}
	if len(resizables) == 1 {
		return resizables[0].Widget(gtx)
	}
	r.axis = axis

	if !r.initialized {
		r.init(gtx)
		r.initialized = true
	}

	// On Window Resize
	if r.length != r.axis.Convert(gtx.Constraints.Max).X {
		r.onWindowResize(gtx)
	}
	gtx.Constraints.Min = gtx.Constraints.Max

	flex := layout.Flex{Axis: r.axis}
	return flex.Layout(gtx,
		resizables[0].Layout(gtx)...,
	)
}

func (r *Resize) init(gtx layout.Context) {
	r.length = r.axis.Convert(gtx.Constraints.Max).X
	var totalRatio float32
	// Obtain the total ration to reset it between 0.0 - 1.00
	var totalHandlesLength int
	for i, rz := range r.resizables {
		if rz.Handle == nil {
			rz.Handle = r.CustomResizeHandle
		}
		m := op.Record(gtx.Ops)
		d := rz.Handle(gtx)
		m.Stop()
		totalHandlesLength += r.axis.Convert(d.Size).X
		totalRatio += rz.Ratio
		var prevResizable *Resizable
		var nextResizable *Resizable
		if i != 0 {
			prevResizable = r.resizables[i-1]
		}
		if i < len(r.resizables)-1 {
			nextResizable = r.resizables[i+1]
		}
		rz.prev = prevResizable
		rz.next = nextResizable
		rz.resize = r
	}
	r.totalHandlesLength = totalHandlesLength
	// Reset the ratio between 0.0 - 1.00
	var currTotalRatio float32
	for i, rz := range r.resizables {
		rz.Ratio /= totalRatio // reset the total ratio
		currTotalRatio += rz.Ratio
		if i == len(r.resizables)-1 {
			currTotalRatio = 1.0
		}
		rz.float.Pos = int(float32(r.length) * currTotalRatio)
	}
}

func (r *Resize) onWindowResize(gtx layout.Context) {
	maxLength := r.axis.Convert(gtx.Constraints.Max).X
	for _, rz := range r.resizables {
		rz.Ratio = float32(rz.Pos) / float32(r.length)
	}
	r.length = maxLength
	var totalHandlesLength int
	for _, rz := range r.resizables {
		rz.float.Pos = int(float32(r.length) * rz.Ratio)
		m := op.Record(gtx.Ops)
		d := rz.Handle(gtx)
		m.Stop()
		totalHandlesLength += r.axis.Convert(d.Size).X
	}
	r.totalHandlesLength = totalHandlesLength
}

type float struct {
	Pos  int // position in pixels of the handle
	drag gesture.Drag
}

func (r *Resizable) Layout(gtx layout.Context) []layout.FlexChild {
	m := op.Record(gtx.Ops)
	dims := r.handleResize(gtx)
	c := m.Stop()
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			prevPos := 0
			if r.prev != nil {
				prevPos = r.prev.Pos
			}
			gtx.Constraints.Min = image.Point{X: r.Pos - prevPos, Y: 20}
			if r.resize.axis == layout.Vertical {
				gtx.Constraints.Min = r.resize.axis.Convert(gtx.Constraints.Min)
			}
			r.Widget(gtx)
			return Dim{Size: gtx.Constraints.Min}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			c.Add(gtx.Ops)
			return dims
		}),
	}
	if r.next != nil {
		children = append(children, r.next.Layout(gtx)...)
	}
	return children
}

func (r *Resizable) handleResize(gtx layout.Context) layout.Dimensions {
	if r.next == nil {
		return layout.Dimensions{}
	}
	gtx.Constraints.Min = image.Point{}
	dims := r.Handle(gtx)

	var de *pointer.Event
	for _, e := range r.float.drag.Events(gtx.Metric, gtx, gesture.Axis(r.resize.axis)) {
		if e.Type == pointer.Drag {
			de = &e
		}
	}

	prevWidgetPos := 0
	nextWidgetPos := r.next.Pos
	currentWidgetSize := 0
	nextWidgetSize := 0
	if r.prev != nil {
		prevWidgetPos = r.prev.Pos
	}
	mac := op.Record(gtx.Ops)
	d := r.Widget(gtx)
	mac.Stop()
	currentWidgetSize = r.resize.axis.Convert(d.Size).X
	mac = op.Record(gtx.Ops)
	d = r.next.Widget(gtx)
	mac.Stop()
	nextWidgetSize = r.resize.axis.Convert(d.Size).X

	minPos := prevWidgetPos + currentWidgetSize
	maxPos := nextWidgetPos - nextWidgetSize

	// referencing the last element, accounting for all the handle width
	if r.next.next == nil {
		maxPos -= r.resize.totalHandlesLength
	}

	if de != nil {
		xy := de.Position.X
		if r.resize.axis == layout.Vertical {
			xy = de.Position.Y
		}
		pos := r.float.Pos + int(xy)
		r.float.Pos = pos
	}

	if r.float.Pos < minPos {
		r.float.Pos = minPos
	} else if r.float.Pos > maxPos {
		r.float.Pos = maxPos
	}

	rect := image.Rectangle{Max: dims.Size}
	defer clip.Rect(rect).Push(gtx.Ops).Pop()
	r.float.drag.Add(gtx.Ops)
	cursor := pointer.CursorRowResize
	if r.resize.axis == layout.Horizontal {
		cursor = pointer.CursorColResize
	}
	cursor.Add(gtx.Ops)

	return layout.Dimensions{Size: dims.Size}
}

func (r *Resize) CustomResizeHandle(gtx Gtx) Dim {
	x := gtx.Dp(unit.Dp(4))
	y := gtx.Constraints.Max.Y
	if r.axis == layout.Vertical {
		x = gtx.Constraints.Max.X
		y = gtx.Dp(unit.Dp(4))
	}
	rect := image.Rectangle{
		Max: image.Point{
			X: x,
			Y: y,
		},
	}
	paint.FillShape(gtx.Ops, color.NRGBA{A: 200}, clip.Rect(rect).Op())
	return Dim{Size: rect.Max}
}
