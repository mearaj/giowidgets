package calendar

import (
	"fmt"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"image"
	"image/color"
	"strings"
	"time"
)

type Dim = layout.Dimensions
type Gtx = layout.Context

type OnCalendarDateClick func(t time.Time)

var weekdays = [7]time.Weekday{
	time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday,
}

type monthButton struct {
	Month time.Month
	widget.Clickable
}

type yearButton struct {
	Year int
	widget.Clickable
}

type cellItem struct {
	widget.Clickable
	time.Time
}

var cellItemsArr = make([]cellItem, 35)

// space between months and years dropdown in the header
var spaceBetweenHeaderDropdowns = unit.Dp(32)

const dropdownWidth = unit.Dp(120)
const minCellHeight = unit.Dp(80)

var allMonthsButtonsArr = [12]monthButton{
	{Month: 1}, {Month: 2}, {Month: 3}, {Month: 4},
	{Month: 5}, {Month: 6}, {Month: 7}, {Month: 8},
	{Month: 9}, {Month: 10}, {Month: 11}, {Month: 12},
}
var allYearsButtonsSlice []yearButton
var allRows [5]layout.FlexChild
var monthsHeaderRowHeight = unit.Dp(64)
var viewHeaderHeight = unit.Dp(32)

func SetAllYearsButtonsSlice(startTime, endTime time.Time) {
	allYearsButtonsSlice = GetYearsRangeButtons(startTime.Year(), endTime.Year())
}

func init() {
	startTime := time.Now().AddDate(-100, 0, 0)
	endTime := time.Now().AddDate(101, 0, 0)
	SetAllYearsButtonsSlice(startTime, endTime)
}

type Calendar struct {
	Theme            *material.Theme
	time             time.Time
	btnDropdownMonth widget.Clickable
	monthsList       layout.List
	yearsList        layout.List
	btnDropdownYear  widget.Clickable
	initialized      bool
	showMonths       bool
	showYears        bool
	fullView         widget.Clickable
	viewList         layout.List
	MaxWidthHeight   image.Point
	OnCalendarDateClick
	layout.Inset
}

func (c *Calendar) SetTime(t time.Time) {
	c.time = t
}
func (c *Calendar) Time() time.Time {
	return c.time
}

func (c *Calendar) Layout(gtx Gtx) Dim {
	if !c.initialized {
		now := time.Now()
		if c.Time().IsZero() {
			c.SetTime(now)
		}
		c.initialized = true
	}
	if c.Theme == nil {
		c.Theme = material.NewTheme(gofont.Collection())
	}
	if c.fullView.Clicked() {
		if !c.btnDropdownMonth.Pressed() {
			c.showMonths = false
		}
		if !c.btnDropdownYear.Pressed() {
			c.showYears = false
		}
	}

	c.fullView.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return Dim{Size: image.Point{
			X: gtx.Constraints.Max.X,
			Y: gtx.Constraints.Max.Y,
		}}
	})

	d := c.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		flex := layout.Flex{Axis: layout.Vertical}
		return flex.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				return c.drawViewHeader(gtx)
			}),
			layout.Rigid(c.drawHeaderRow),
			layout.Rigid(c.drawBodyRows),
		)
	})
	if c.showMonths {
		c.drawMonthsDropdownItems(gtx)
	}
	if c.showYears {
		gtx.Constraints.Max.Y = d.Size.Y - gtx.Dp(viewHeaderHeight)
		c.drawYearsDropdownItems(gtx)
	}

	return d
}

func (c *Calendar) drawHeaderRow(gtx Gtx) Dim {
	var flexChildren []layout.FlexChild
	maxWidth := gtx.Constraints.Max.X
	columnWidth := maxWidth / 7
	for _, day := range weekdays {
		dayStr := strings.ToUpper(day.String()[0:3])
		flexChildren = append(flexChildren, c.drawHeaderColumn(gtx, dayStr, columnWidth))
	}
	flex := layout.Flex{}
	mac := op.Record(gtx.Ops)
	d := flex.Layout(gtx, flexChildren...)
	call := mac.Stop()
	rect := clip.Rect{Max: d.Size}
	paint.FillShape(gtx.Ops, c.Theme.ContrastBg, rect.Op())
	call.Add(gtx.Ops)
	return d
}
func (c *Calendar) drawHeaderColumn(gtx Gtx, day string, columnWidth int) layout.FlexChild {
	gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = gtx.Dp(monthsHeaderRowHeight), gtx.Dp(monthsHeaderRowHeight)
	return layout.Rigid(func(gtx Gtx) Dim {
		gtx.Constraints.Min.X, gtx.Constraints.Max.X = columnWidth, columnWidth
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx Gtx) Dim {
			return layout.Center.Layout(gtx, func(gtx Gtx) Dim {
				label := material.Label(c.Theme, c.Theme.TextSize, day)
				label.Color = c.Theme.ContrastFg
				label.MaxLines = 1
				d := label.Layout(gtx)
				return d
			})
		})
	})
}

func (c *Calendar) drawColumn(gtx Gtx, columnWidth int, btn *cellItem) layout.FlexChild {
	dayStr := fmt.Sprintf("%d", btn.Day())
	return layout.Rigid(func(gtx Gtx) Dim {
		bgColor := c.Theme.Bg
		txtColor := c.Theme.Fg
		if btn.Month() != c.Time().Month() {
			bgColor = color.NRGBA(colornames.BlueGrey50)
			txtColor.A = 100
		}
		if c.Time().Month() == btn.Month() {
			if btn.Clicked() {
				if c.OnCalendarDateClick != nil {
					c.OnCalendarDateClick(btn.Time)
				}
			}
			now := time.Now()
			isEqual := btn.Day() == now.Day() && now.Year() == btn.Year() && now.Month() == btn.Month()
			if btn.Hovered() || isEqual {
				bgColor = c.Theme.ContrastBg
				bgColor.A = 240
				txtColor = c.Theme.ContrastFg
				txtColor.A = 240
			}
		}
		return btn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			mac := op.Record(gtx.Ops)
			gtx.Constraints.Min.X, gtx.Constraints.Max.X = columnWidth, columnWidth
			d := layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx Gtx) Dim {
				txtSize := c.Theme.TextSize
				txtSize = txtSize * 1.5
				label := material.Label(c.Theme, txtSize, dayStr)
				label.MaxLines = 1
				label.Color = txtColor
				d := label.Layout(gtx)
				return d
			})
			call := mac.Stop()
			rect := clip.Rect{Max: d.Size}
			paint.FillShape(gtx.Ops, bgColor, rect.Op())
			call.Add(gtx.Ops)
			return d
		})
	})
}

func (c *Calendar) drawBodyRows(gtx Gtx) Dim {
	flex := layout.Flex{Axis: layout.Vertical}
	t := c.Time()
	maxWidth := gtx.Constraints.Max.X
	columnWidth := maxWidth / 7
	startDate := firstDayOfWeek(t, time.Monday)
	lastDate := lastDayOfWeek(t, time.Monday)
	day := startDate
	endDate := lastDate
	index := 0
	rowIndex := 0
	minMaxHeight := c.MaxWidthHeight.Y - gtx.Dp(viewHeaderHeight+monthsHeaderRowHeight)
	if unit.Dp(minMaxHeight) < minCellHeight {
		minMaxHeight = gtx.Dp(minCellHeight)
	}
	for ; day.Before(endDate) && rowIndex < 5; rowIndex++ {
		var flexChildren []layout.FlexChild
		for i := 0; i < 7; i++ {
			cellItemsArr[index].Time = day
			flexChildren = append(flexChildren, c.drawColumn(gtx, columnWidth, &cellItemsArr[index]))
			index++
			day = day.AddDate(0, 0, 1)
		}
		flexChild := layout.Rigid(func(gtx Gtx) Dim {
			flex := layout.Flex{}
			gtx.Constraints.Min.Y, gtx.Constraints.Max.Y = minMaxHeight, minMaxHeight
			return flex.Layout(gtx, flexChildren...)
		})
		allRows[rowIndex] = flexChild
	}
	c.viewList.Axis = layout.Vertical
	d := c.viewList.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
		return flex.Layout(gtx, allRows[:]...)
	})
	return d
}

func (c *Calendar) OnMonthButtonClick(gtx Gtx, month *monthButton) {
	t := c.Time()
	t = time.Date(t.Year(), month.Month, t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	c.SetTime(t)
	op.InvalidateOp{}.Add(gtx.Ops)
}

func (c *Calendar) OnYearButtonClick(gtx Gtx, year *yearButton) {
	t := c.Time()
	t = time.Date(year.Year, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	c.SetTime(t)
	op.InvalidateOp{}.Add(gtx.Ops)
}

func (c *Calendar) drawMonthsDropdownItems(gtx Gtx) Dim {
	gtx.Constraints.Max.Y = gtx.Constraints.Max.Y - gtx.Dp(viewHeaderHeight)
	op.Offset(image.Point{
		X: gtx.Dp(16),
		Y: gtx.Dp(viewHeaderHeight) + gtx.Dp(8),
	}).Add(gtx.Ops)
	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			mac := op.Record(gtx.Ops)
			gtx.Constraints.Min.X = gtx.Dp(dropdownWidth)
			c.monthsList.Axis = layout.Vertical

			border := widget.Border{
				Color:        c.Theme.ContrastBg,
				CornerRadius: 0,
				Width:        unit.Dp(1),
			}
			d := border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				d := c.monthsList.Layout(gtx, len(allMonthsButtonsArr), func(gtx layout.Context, index int) layout.Dimensions {
					bgColor := c.Theme.Bg
					txtColor := c.Theme.Fg
					isSelected := c.Time().Month().String() == allMonthsButtonsArr[index].Month.String()
					if allMonthsButtonsArr[index].Hovered() || isSelected {
						bgColor = c.Theme.Fg
						txtColor = c.Theme.Bg
					}
					if allMonthsButtonsArr[index].Clicked() {
						c.showMonths = false
						c.OnMonthButtonClick(gtx, &allMonthsButtonsArr[index])
					}
					mac := op.Record(gtx.Ops)
					d := allMonthsButtonsArr[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						inset := layout.Inset{
							Top:    unit.Dp(8),
							Bottom: unit.Dp(8),
							Left:   unit.Dp(16),
							Right:  unit.Dp(16),
						}
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							txt := allMonthsButtonsArr[index].Month.String()
							label := material.Label(c.Theme, c.Theme.TextSize, txt)
							label.Alignment = text.Start
							label.Color = txtColor
							return label.Layout(gtx)
						})
					})
					call := mac.Stop()
					rect := clip.Rect{Max: d.Size}
					paint.FillShape(gtx.Ops, bgColor, rect.Op())
					call.Add(gtx.Ops)
					return d
				})
				return d
			})
			call := mac.Stop()
			rect := clip.Rect{Max: d.Size}
			paint.FillShape(gtx.Ops, c.Theme.Bg, rect.Op())
			call.Add(gtx.Ops)
			return d
		}),
	)
	return Dim{}
}

func (c *Calendar) drawYearsDropdownItems(gtx Gtx) Dim {
	gtx.Constraints.Max.Y = gtx.Constraints.Max.Y - gtx.Dp(viewHeaderHeight)
	op.Offset(image.Point{
		X: gtx.Dp(16) + gtx.Dp(dropdownWidth) + gtx.Dp(spaceBetweenHeaderDropdowns),
		Y: gtx.Dp(viewHeaderHeight) + gtx.Dp(8),
	}).Add(gtx.Ops)
	layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			mac := op.Record(gtx.Ops)
			gtx.Constraints.Min.X = gtx.Dp(dropdownWidth)
			c.yearsList.Axis = layout.Vertical

			border := widget.Border{
				Color:        c.Theme.ContrastBg,
				CornerRadius: 0,
				Width:        unit.Dp(1),
			}
			d := border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				d := c.yearsList.Layout(gtx, len(allYearsButtonsSlice), func(gtx layout.Context, index int) layout.Dimensions {
					bgColor := c.Theme.Bg
					txtColor := c.Theme.Fg
					isSelected := c.Time().Year() == allYearsButtonsSlice[index].Year
					if allYearsButtonsSlice[index].Hovered() || isSelected {
						bgColor = c.Theme.Fg
						txtColor = c.Theme.Bg
					}
					if allYearsButtonsSlice[index].Clicked() {
						c.showYears = false
						c.OnYearButtonClick(gtx, &allYearsButtonsSlice[index])
					}
					mac := op.Record(gtx.Ops)
					d := allYearsButtonsSlice[index].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						inset := layout.Inset{
							Top:    unit.Dp(8),
							Bottom: unit.Dp(8),
							Left:   unit.Dp(16),
							Right:  unit.Dp(16),
						}
						return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							txt := fmt.Sprintf("%d", allYearsButtonsSlice[index].Year)
							label := material.Label(c.Theme, c.Theme.TextSize, txt)
							label.Alignment = text.Start
							label.Color = txtColor
							return label.Layout(gtx)
						})
					})
					call := mac.Stop()
					rect := clip.Rect{Max: d.Size}
					paint.FillShape(gtx.Ops, bgColor, rect.Op())
					call.Add(gtx.Ops)
					return d
				})
				return d
			})
			call := mac.Stop()
			rect := clip.Rect{Max: d.Size}
			paint.FillShape(gtx.Ops, c.Theme.Bg, rect.Op())
			call.Add(gtx.Ops)
			return d
		}),
	)
	return Dim{}
}

func (c *Calendar) drawViewHeader(gtx Gtx) Dim {
	month := c.Time().Month().String()
	year := fmt.Sprintf("%d", c.Time().Year())
	gtx.Constraints.Max.Y, gtx.Constraints.Min.Y = gtx.Dp(viewHeaderHeight), gtx.Dp(viewHeaderHeight)
	inset := layout.Inset{Left: unit.Dp(16)}
	d := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		flex := layout.Flex{Spacing: layout.SpaceSides, Alignment: layout.Middle}
		return flex.Layout(gtx,
			layout.Rigid(func(gtx Gtx) Dim {
				if c.btnDropdownMonth.Clicked() {
					c.showMonths = !c.showMonths
					c.showYears = false
					if c.showMonths {
						for i, eachButton := range allMonthsButtonsArr {
							if eachButton.Month.String() == c.Time().Month().String() {
								c.monthsList.Position.First = i
								c.monthsList.Position.Offset = -32
								break
							}
						}
					}
				}
				gtx.Constraints.Min.X = gtx.Dp(dropdownWidth)
				d := c.btnDropdownMonth.Layout(gtx, func(gtx Gtx) Dim {
					flex := layout.Flex{Spacing: layout.SpaceBetween}
					return flex.Layout(gtx,
						layout.Rigid(func(gtx Gtx) Dim {
							label := material.Label(c.Theme, c.Theme.TextSize, month)
							return label.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx Gtx) Dim {
							downIcon, _ := widget.NewIcon(icons.NavigationArrowDropDown)
							return downIcon.Layout(gtx, c.Theme.ContrastBg)
						}),
					)
				})
				return d
			}),
			layout.Rigid(layout.Spacer{Width: spaceBetweenHeaderDropdowns}.Layout),
			layout.Rigid(func(gtx Gtx) Dim {
				if c.btnDropdownYear.Clicked() {
					c.showMonths = false
					c.showYears = !c.showYears
					if c.showYears {
						c.scrollToSelectedYear()
					}
				}
				d := c.btnDropdownYear.Layout(gtx, func(gtx Gtx) Dim {
					gtx.Constraints.Min.X = gtx.Dp(dropdownWidth)
					flex := layout.Flex{Spacing: layout.SpaceBetween}
					return flex.Layout(gtx,
						layout.Rigid(func(gtx Gtx) Dim {
							label := material.Label(c.Theme, c.Theme.TextSize, year)
							return label.Layout(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx Gtx) Dim {
							downIcon, _ := widget.NewIcon(icons.NavigationArrowDropDown)
							return downIcon.Layout(gtx, c.Theme.ContrastBg)
						}),
					)
				})
				return d
			}),
		)
	})
	return d
}

func (c *Calendar) scrollToSelectedYear() {
	for i, eachYear := range allYearsButtonsSlice {
		if eachYear.Year == c.Time().Year() {
			c.yearsList.Position.First = i
			c.yearsList.Position.Offset = -32
			break
		}
	}
}
