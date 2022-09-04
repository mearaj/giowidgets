package calendar

import "time"

// Ref https://stackoverflow.com/questions/36830212/get-the-first-and-last-day-of-current-month-in-go-golang
func beginningOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -date.Day()+1)
}
func endOfMonth(date time.Time) time.Time {
	return date.AddDate(0, 1, -date.Day())
}

func firstDayOfWeek(tm time.Time, weekStartDay time.Weekday) time.Time {
	//tm = time.Date(tm.Year(), tm.Month(), 1, 0, 0, 0, 0, tm.Location())
	tm = beginningOfMonth(tm)
	for tm.Weekday() != weekStartDay {
		tm = tm.AddDate(0, 0, -1)
	}
	return tm
}
func lastDayOfWeek(tm time.Time, weekStartDay time.Weekday) time.Time {
	tm = endOfMonth(tm)
	weekEndDay := (weekStartDay + 6) % 7
	for tm.Weekday() != weekEndDay {
		tm = tm.AddDate(0, 0, 1)
	}
	return tm
}

// GetYearsRangeButtons returns slice of yearButton with year range between startYear and upto but not including lastYear
func GetYearsRangeButtons(startYear, endYear int) []yearButton {
	yearsRange := make([]yearButton, 0)
	for currentYear := startYear; currentYear < endYear; currentYear++ {
		yearsRange = append(yearsRange, yearButton{Year: currentYear})
	}
	return yearsRange
}
