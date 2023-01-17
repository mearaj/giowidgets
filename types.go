package giowidgets

import "gioui.org/layout"

type (
	Gtx       = layout.Context
	Dim       = layout.Dimensions
	Inset     = layout.Inset
	Flex      = layout.Flex
	FlexChild = layout.FlexChild
)

type View interface {
	Layout(gtx Gtx) Dim
}
