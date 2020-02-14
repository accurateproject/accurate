package utils

import (
	"reflect"
	"sort"
	"time"
)

// Defines years days series
type Years []int

func (ys Years) Sort() {
	sort.Slice(ys, func(j, i int) bool { return ys[j] < ys[i] })
}

// Return true if the specified date is inside the series
func (ys Years) Contains(year int) (result bool) {
	result = false
	for _, yss := range ys {
		if yss == year {
			result = true
			break
		}
	}
	return
}

// Defines months series
type Months []time.Month

func (m Months) Sort() {
	sort.Slice(m, func(j, i int) bool { return m[j] < m[i] })
}

// Return true if the specified date is inside the series
func (m Months) Contains(month time.Month) (result bool) {
	for _, ms := range m {
		if ms == month {
			result = true
			break
		}
	}
	return
}
func (m Months) IsComplete() bool {
	allMonths := Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December}
	m.Sort()
	return reflect.DeepEqual(m, allMonths)
}

// Defines month days series
type MonthDays []int

func (md MonthDays) Sort() {
	sort.Slice(md, func(j, i int) bool { return md[j] < md[i] })
}

// Return true if the specified date is inside the series
func (md MonthDays) Contains(monthDay int) (result bool) {
	result = false
	for _, mds := range md {
		if mds == monthDay {
			result = true
			break
		}
	}
	return
}

// Defines week days series
type WeekDays []time.Weekday

func (wd WeekDays) Sort() {
	sort.Slice(wd, func(j, i int) bool { return wd[j] < wd[i] })
}

// Return true if the specified date is inside the series
func (wd WeekDays) Contains(weekDay time.Weekday) (result bool) {
	result = false
	for _, wds := range wd {
		if wds == weekDay {
			result = true
			break
		}
	}
	return
}

func DaysInMonth(year int, month time.Month) float64 {
	return float64(time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1).Day())
}

func DaysInYear(year int) float64 {
	first := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(1, 0, 0)
	return float64(last.Sub(first).Hours() / 24)
}
