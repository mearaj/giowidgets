package calendar

import "gioui.org/layout"

type (
	Gtx = layout.Context
	Dim = layout.Dimensions
)

type View interface {
	Layout(gtx Gtx) Dim
}
