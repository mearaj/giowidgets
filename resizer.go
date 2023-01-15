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
	minLength          int
}

type Resizable struct {
	// ratio is only calculated during initialization, based on widget's natural size.
	//  It acts like minimum threshold ratio value beyond which widget size cannot be further reduced.
	ratio            float32
	Widget           layout.Widget
	DividerHandler   layout.Widget
	dividerThickness int
	float
	resize *Resize
	prev   *Resizable
	next   *Resizable
}

func NewResizeWidget(axis layout.Axis, resizables []*Resizable) *Resize {
	r := &Resize{axis: axis, resizables: resizables}
	for _, rz := range resizables {
		rz.resize = r
		if rz.DividerHandler == nil {
			rz.DividerHandler = r.CustomResizeHandleBar
		}
	}
	return r
}

// Layout displays w1 and w2 with handle in between.
//
// The widgets w1 and w2 must be able to gracefully resize their minimum and maximum dimensions
// in order for the resize to be smooth.
func (r *Resize) Layout(gtx layout.Context) layout.Dimensions {
	// Compute the first widget's max width/height.
	if len(r.resizables) == 0 {
		return layout.Dimensions{}
	}
	if len(r.resizables) == 1 {
		return r.resizables[0].Widget(gtx)
	}

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
		r.resizables[0].Layout(gtx)...,
	)
}

func (r *Resize) init(gtx layout.Context) {
	r.length = r.axis.Convert(gtx.Constraints.Max).X
	if r.minLength == 0 {
		r.minLength = int(0.1 * float32(r.length))
	}
	allowedMinLength := r.length / len(r.resizables)
	if r.minLength > allowedMinLength || r.minLength <= 0 {
		r.minLength = allowedMinLength
	}
	var totalRatio float32
	// Obtain the total ration to reset it between 0.0 - 1.00
	var totalHandlesLength int
	for i, rz := range r.resizables {
		if rz.DividerHandler == nil {
			rz.DividerHandler = r.CustomResizeHandleBar
		}
		m := op.Record(gtx.Ops)
		d := rz.DividerHandler(gtx)
		m.Stop()
		totalHandlesLength += r.axis.Convert(d.Size).X
		m = op.Record(gtx.Ops)
		d = rz.Widget(gtx)
		m.Stop()
		rz.ratio = float32(r.axis.Convert(d.Size).X) / float32(r.length)
		totalRatio += rz.ratio
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
	for _, rz := range r.resizables {
		rz.ratio /= totalRatio // reset the total ratio
		currTotalRatio += rz.ratio
		rz.float.pos = int(float32(r.length) * currTotalRatio)
	}
}

func (r *Resize) onWindowResize(gtx layout.Context) {
	currMinLength := r.minLength
	prevLength := r.length
	r.minLength = (currMinLength / prevLength) * r.length
	r.length = r.axis.Convert(gtx.Constraints.Max).X
	for _, rz := range r.resizables {
		rz.float.pos = int((float32(rz.float.pos) / float32(prevLength)) * float32(r.length))
	}
}

type float struct {
	pos  int // position in pixels of the handle
	drag gesture.Drag
}

func (r *Resizable) Layout(gtx layout.Context) []layout.FlexChild {
	m := op.Record(gtx.Ops)
	dims := r.handleDrag(gtx)
	c := m.Stop()
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			prePos := 0
			if r.prev != nil {
				prePos = r.prev.pos
			}
			gtx.Constraints.Max = image.Point{X: r.pos - prePos, Y: gtx.Constraints.Max.Y}
			if r.resize.axis == layout.Vertical {
				gtx.Constraints.Max = r.resize.axis.Convert(gtx.Constraints.Max)
			}

			d := r.Widget(gtx)
			d.Size = gtx.Constraints.Max
			return d
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

func (r *Resizable) handleDrag(gtx layout.Context) layout.Dimensions {
	if r.next == nil {
		return layout.Dimensions{}
	}
	gtx.Constraints.Min = image.Point{}
	dims := r.DividerHandler(gtx)

	var de *pointer.Event
	for _, e := range r.float.drag.Events(gtx.Metric, gtx, gesture.Axis(r.resize.axis)) {
		if e.Type == pointer.Drag {
			de = &e
		}
	}

	var posDifference float32
	if de != nil {
		posDifference = de.Position.X
		if r.resize.axis == layout.Vertical {
			posDifference = de.Position.Y
		}

		if posDifference < 0 {
			for curr := r; curr != nil; curr = curr.prev {
				curr.float.pos += int(posDifference)
				minPos := r.resize.minLength
				if curr.prev != nil {
					minPos = curr.prev.pos + curr.resize.minLength
				}
				if curr.float.pos < minPos {
					curr.float.pos = minPos
				} else {
					break
				}
			}
		}
		if posDifference > 0 {
			for curr := r; curr != nil; curr = curr.next {
				curr.float.pos += int(posDifference)
				maxPos := r.resize.length
				if curr.next != nil {
					maxPos = curr.next.pos - curr.resize.minLength
				}
				if curr.float.pos > maxPos {
					curr.float.pos = maxPos
				} else {
					break
				}
			}
		}
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

func (r *Resize) CustomResizeHandleBar(gtx Gtx) Dim {
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
