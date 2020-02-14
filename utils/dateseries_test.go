package utils

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestDateseriesMonthStoreRestoreJson(t *testing.T) {
	m := Months{5, 6, 7, 8}
	r, _ := json.Marshal(m)
	if string(r) != "[5,6,7,8]" {
		t.Errorf("Error serializing months: %v", string(r))
	}
	o := Months{}
	json.Unmarshal(r, &o)
	if !reflect.DeepEqual(o, m) {
		t.Errorf("Expected %v was  %v", m, o)
	}
}

func TestDateseriesMonthDayStoreRestoreJson(t *testing.T) {
	md := MonthDays{24, 25, 26}
	r, _ := json.Marshal(md)
	if string(r) != "[24,25,26]" {
		t.Errorf("Error serializing month days: %v", string(r))
	}
	o := MonthDays{}
	json.Unmarshal(r, &o)
	if !reflect.DeepEqual(o, md) {
		t.Errorf("Expected %v was  %v", md, o)
	}
}

func TestDateseriesWeekDayStoreRestoreJson(t *testing.T) {
	wd := WeekDays{time.Saturday, time.Sunday}
	r, _ := json.Marshal(wd)
	if string(r) != "[6,0]" {
		t.Errorf("Error serializing week days: %v", string(r))
	}
	o := WeekDays{}
	json.Unmarshal(r, &o)
	if !reflect.DeepEqual(o, wd) {
		t.Errorf("Expected %v was  %v", wd, o)
	}
}

func TestDateseriesMonthsIsCompleteNot(t *testing.T) {
	months := Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November}
	if months.IsComplete() {
		t.Error("Error months IsComplete: ", months)
	}
}

func TestDateseriesMonthsIsCompleteYes(t *testing.T) {
	months := Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December}
	if !months.IsComplete() {
		t.Error("Error months IsComplete: ", months)
	}
}

func TestDateseriesDaysInMonth(t *testing.T) {
	if n := DaysInMonth(2016, 4); n != 30 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInMonth(2016, 2); n != 29 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInMonth(2016, 1); n != 31 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInMonth(2016, 12); n != 31 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInMonth(2015, 2); n != 28 {
		t.Error("error calculating days: ", n)
	}
}

func TestDateseriesDaysInYear(t *testing.T) {
	if n := DaysInYear(2016); n != 366 {
		t.Error("error calculating days: ", n)
	}
	if n := DaysInYear(2015); n != 365 {
		t.Error("error calculating days: ", n)
	}
}
