package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

var (
	//referenceDate = time.Date(2013, 7, 10, 10, 30, 0, 0, time.Local)
	//referenceDate = time.Date(2013, 12, 31, 23, 59, 59, 0, time.Local)
	//referenceDate = time.Date(2011, 1, 1, 0, 0, 0, 1, time.Local)
	referenceDate = time.Now()
	now           = referenceDate
)

func TestActionTimingAlways(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}}}
	st := at.GetNextStartTime(referenceDate)
	y, m, d := referenceDate.Date()
	expected := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanNothing(t *testing.T) {
	at := &ActionTiming{}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionTimingMidnight(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: "00:00:00"}}}
	y, m, d := referenceDate.Date()
	now := time.Date(y, m, d, 0, 0, 1, 0, time.Local)
	st := at.GetNextStartTime(now)
	expected := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanOnlyHour(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)

	y, m, d := now.Date()
	expected := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(0, 0, 1)
	}
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourYear(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Years: utils.Years{2022}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(2022, 1, 1, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanOnlyWeekdays(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Monday}}}}
	st := at.GetNextStartTime(referenceDate)

	y, m, d := now.Date()
	h, min, s := now.Clock()
	e := time.Date(y, m, d, h, min, s, 0, time.Local)
	day := e.Day()
	e = time.Date(e.Year(), e.Month(), day, 0, 0, 0, 0, e.Location())
	for i := 0; i < 8; i++ {
		n := e.AddDate(0, 0, i)
		if n.Weekday() == time.Monday && (n.Equal(now) || n.After(now)) {
			e = n
			break
		}
	}
	if !st.Equal(e) {
		t.Errorf("Expected %v was %v", e, st)
	}
}

func TestActionPlanHourWeekdays(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{WeekDays: []time.Weekday{time.Monday}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)

	y, m, d := now.Date()
	e := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	day := e.Day()
	for i := 0; i < 8; i++ {
		e = time.Date(e.Year(), e.Month(), day, e.Hour(), e.Minute(), e.Second(), e.Nanosecond(), e.Location())
		n := e.AddDate(0, 0, i)
		if n.Weekday() == time.Monday && (n.Equal(now) || n.After(now)) {
			e = n
			break
		}
	}
	if !st.Equal(e) {
		t.Errorf("Expected %v was %v", e, st)
	}
}

func TestActionPlanOnlyMonthdays(t *testing.T) {

	y, m, d := now.Date()
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{MonthDays: utils.MonthDays{1, 25, 2, tomorrow.Day()}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := tomorrow
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonthdays(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	if now.After(testTime) {
		y, m, d = tomorrow.Date()
	}
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{MonthDays: utils.MonthDays{now.Day(), tomorrow.Day()}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanOnlyMonths(t *testing.T) {

	y, m, _ := now.Date()
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Months: utils.Months{time.February, time.May, nextMonth.Month()}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Log("NextMonth: ", nextMonth)
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonths(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	if now.After(testTime) {
		testTime = testTime.AddDate(0, 0, 1)
		y, m, d = testTime.Date()
	}
	if now.After(testTime) {
		m = nextMonth.Month()
		y = nextMonth.Year()

	}
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{
		Months:    utils.Months{now.Month(), nextMonth.Month()},
		StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(y, m, 1, 10, 1, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(0, 1, 0)
	}
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonthdaysMonths(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)

	if now.After(testTime) {
		y, m, d = tomorrow.Date()
	}
	nextDay := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	month := nextDay.Month()
	if nextDay.Before(now) {
		if now.After(testTime) {
			month = nextMonth.Month()
		}
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{now.Month(), nextMonth.Month()},
			MonthDays: utils.MonthDays{now.Day(), tomorrow.Day()},
			StartTime: "10:01:00",
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(y, month, d, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanFirstOfTheMonth(t *testing.T) {

	y, m, _ := now.Date()
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			MonthDays: utils.MonthDays{1},
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	expected := nextMonth
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanOnlyYears(t *testing.T) {
	y, _, _ := referenceDate.Date()
	nextYear := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Years: utils.Years{now.Year(), nextYear.Year()}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := nextYear
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanPast(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Years: utils.Years{2023}}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourYears(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{Years: utils.Years{referenceDate.Year(), referenceDate.Year() + 1}, StartTime: "10:01:00"}}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(referenceDate.Year(), 1, 1, 10, 1, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(1, 0, 0)
	}
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonthdaysYear(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	tomorrow := time.Date(y, m, d, 10, 1, 0, 0, time.Local).AddDate(0, 0, 1)
	nextYear := time.Date(y, 1, d, 10, 1, 0, 0, time.Local).AddDate(1, 0, 0)
	expected := testTime
	if referenceDate.After(testTime) {
		if referenceDate.After(tomorrow) {
			expected = nextYear
		} else {
			expected = tomorrow
		}
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years:     utils.Years{now.Year(), nextYear.Year()},
			MonthDays: utils.MonthDays{now.Day(), tomorrow.Day()},
			StartTime: "10:01:00",
		},
	}}
	at.Timing.Timing.CronString() // side effect!
	st := at.GetNextStartTime(referenceDate)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanHourMonthdaysMonthYear(t *testing.T) {

	y, m, d := now.Date()
	testTime := time.Date(y, m, d, 10, 1, 0, 0, time.Local)
	nextYear := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	nextMonth := time.Date(y, m, 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, 0)
	tomorrow := time.Date(y, m, d, 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
	day := now.Day()
	if now.After(testTime) {
		day = tomorrow.Day()
	}
	nextDay := time.Date(y, m, day, 10, 1, 0, 0, time.Local)
	month := now.Month()
	if nextDay.Before(now) {
		if now.After(testTime) {
			month = nextMonth.Month()
		}
	}
	nextDay = time.Date(y, month, day, 10, 1, 0, 0, time.Local)
	year := now.Year()
	if nextDay.Before(now) {
		if now.After(testTime) {
			year = nextYear.Year()
		}
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years:     utils.Years{now.Year(), nextYear.Year()},
			Months:    utils.Months{now.Month(), nextMonth.Month()},
			MonthDays: utils.MonthDays{now.Day(), tomorrow.Day()},
			StartTime: "10:01:00",
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	expected := time.Date(year, month, day, 10, 1, 0, 0, time.Local)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanFirstOfTheYear(t *testing.T) {
	y, _, _ := now.Date()
	nextYear := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local).AddDate(1, 0, 0)
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years:     utils.Years{nextYear.Year()},
			Months:    utils.Months{time.January},
			MonthDays: utils.MonthDays{1},
			StartTime: "00:00:00",
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	expected := nextYear
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanFirstMonthOfTheYear(t *testing.T) {
	y, _, _ := now.Date()
	expected := time.Date(y, 1, 1, 0, 0, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(1, 0, 0)
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Months: utils.Months{time.January},
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanFirstMonthOfTheYearSecondDay(t *testing.T) {
	y, _, _ := now.Date()
	expected := time.Date(y, 1, 2, 0, 0, 0, 0, time.Local)
	if referenceDate.After(expected) {
		expected = expected.AddDate(1, 0, 0)
	}
	at := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{time.January},
			MonthDays: utils.MonthDays{2},
		},
	}}
	st := at.GetNextStartTime(referenceDate)
	if !st.Equal(expected) {
		t.Errorf("Expected %v was %v", expected, st)
	}
}

func TestActionPlanCheckForASAP(t *testing.T) {
	at := &ActionTiming{Timing: &RateInterval{Timing: &RITiming{StartTime: utils.ASAP}}}
	if !at.IsASAP() {
		t.Errorf("%v should be asap!", at)
	}
}

func TestActionPlanLogFunction(t *testing.T) {
	a := &Action{
		ActionType: "*log",
		//x
		/*Balance: &BalancePointer{
			Type:  utils.StringPointer("test"),
			Value: &utils.ValueFormula{Static: 1.1},
		},*/
	}
	at := &ActionTiming{
		actions:    []*Action{a},
		actionPlan: &ActionPlan{Tenant: "test"},
	}
	err := at.Execute()
	if err != nil {
		t.Errorf("Could not execute LOG action: %v", err)
	}
}

func TestActionPlanFunctionNotAvailable(t *testing.T) {
	a := &Action{
		ActionType: "VALID_FUNCTION_TYPE",
		Filter1:    `{"Type":"test", "Value":1.1}`,
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"dy": true},
		Timing:     &RateInterval{},
		actions:    []*Action{a},
		actionPlan: &ActionPlan{Tenant: "test"},
	}
	err := at.Execute()
	if err != nil {
		t.Errorf("Faild to detect wrong function type: %v", err)
	}
}

func TestActionTimingPriorityListSortByWeight(t *testing.T) {
	at1 := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years:     utils.Years{2020},
			Months:    utils.Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
			MonthDays: utils.MonthDays{1},
			StartTime: "00:00:00",
		},
		Weight: 20,
	}}
	at2 := &ActionTiming{Timing: &RateInterval{
		Timing: &RITiming{
			Years:     utils.Years{2020},
			Months:    utils.Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
			MonthDays: utils.MonthDays{2},
			StartTime: "00:00:00",
		},
		Weight: 10,
	}}
	var atpl ActionTimingPriorityList
	atpl = append(atpl, at2, at1)
	atpl.Sort()
	if atpl[0] != at1 || atpl[1] != at2 {
		t.Error("Timing list not sorted correctly: ", at1, at2, atpl)
	}
}

func TestActionTimingPriorityListWeight(t *testing.T) {
	at1 := &ActionTiming{
		Timing: &RateInterval{
			Timing: &RITiming{
				Months:    utils.Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
				MonthDays: utils.MonthDays{1},
				StartTime: "00:00:00",
			},
		},
		Weight: 20,
	}
	at2 := &ActionTiming{
		Timing: &RateInterval{
			Timing: &RITiming{
				Months:    utils.Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
				MonthDays: utils.MonthDays{1},
				StartTime: "00:00:00",
			},
		},
		Weight: 10,
	}
	var atpl ActionTimingPriorityList
	atpl = append(atpl, at2, at1)
	atpl.Sort()
	if atpl[0] != at1 || atpl[1] != at2 {
		t.Error("Timing list not sorted correctly: ", atpl)
	}
}

/*
func TestActionPlansRemoveMember(t *testing.T) {
	at1 := &ActionPlan{
		UUID:       "some uuid",
		Id:         "test",
		AccountIDs: []string{"one", "two", "three"},
		ActionsID:  "TEST_ACTIONS",
	}
	at2 := &ActionPlan{
		UUID:       "SOME uuid22",
		Id:         "test2",
		AccountIDs: []string{"three", "four"},
		ActionsID:  "TEST_ACTIONS2",
	}
	ats := ActionPlans{at1, at2}
	if outAts := RemActionPlan(ats, "", "four"); len(outAts[1].AccountIds) != 1 {
		t.Error("Expecting fewer balance ids", outAts[1].AccountIds)
	}
	if ats = RemActionPlan(ats, "", "three"); len(ats) != 1 {
		t.Error("Expecting fewer actionTimings", ats)
	}
	if ats = RemActionPlan(ats, "some_uuid22", ""); len(ats) != 1 {
		t.Error("Expecting fewer actionTimings members", ats)
	}
	ats2 := ActionPlans{at1, at2}
	if ats2 = RemActionPlan(ats2, "", ""); len(ats2) != 0 {
		t.Error("Should have no members anymore", ats2)
	}
}*/

func TestActionTriggerMatchAllBlank(t *testing.T) {
	at := &ActionTrigger{
		TOR:            utils.MONETARY,
		Filter:         `{"Directions":{"$has":["*out"]}}`,
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewVal(2, 0),
	}
	a := &Action{}
	match, err := a.getFilter().Query(at, false)
	if err != nil {
		t.Fatalf("matcher failed: %v", err)
	}
	if !match {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchMinuteBucketBlank(t *testing.T) {
	at := &ActionTrigger{
		TOR:            utils.MONETARY,
		Filter:         `{"Directions":{"$has":["*out"]}}`,
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewVal(2, 0),
	}
	a := &Action{TOR: utils.MONETARY}
	match, err := a.getFilter().Query(at, false)
	if err != nil {
		t.Fatalf("matcher failed: %v", err)
	}
	if !match {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchMinuteBucketFull(t *testing.T) {
	at := &ActionTrigger{
		TOR:            utils.MONETARY,
		Filter:         `{"Directions":{"$has":["*out"]}}`,
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewVal(2, 0),
	}
	a := &Action{Filter1: fmt.Sprintf(`{"ThresholdType":"%v", "ThresholdValue": %v}`, utils.TRIGGER_MAX_BALANCE, 2)}
	match, err := a.getFilter().Query(at, false)
	if err != nil {
		t.Fatalf("matcher failed: %v", err)
	}
	if !match {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchAllFull(t *testing.T) {
	at := &ActionTrigger{
		TOR:            utils.MONETARY,
		Filter:         `{"Directions":{"$has":["*out"]}}`,
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewVal(2, 0),
	}
	a := &Action{Filter1: fmt.Sprintf(`{"ThresholdType":"%v", "ThresholdValue": %v}`, utils.TRIGGER_MAX_BALANCE, 2)}
	match, err := a.getFilter().Query(at, false)
	if err != nil {
		t.Fatalf("matcher failed: %v", err)
	}
	if !match {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchSomeFalse(t *testing.T) {
	at := &ActionTrigger{
		TOR:            utils.MONETARY,
		Filter:         `{"Directions":{"$has":["*out"]}}`,
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewVal(2, 0),
	}
	a := &Action{Filter1: fmt.Sprintf(`{"ThresholdType":"%s"}`, utils.TRIGGER_MAX_BALANCE_COUNTER)}
	match, err := a.getFilter().Query(at, false)
	if err != nil {
		t.Fatalf("matcher failed: %v", err)
	}
	if match {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatcBalanceFalse(t *testing.T) {
	at := &ActionTrigger{
		TOR:            utils.MONETARY,
		Filter:         `{"Directions":{"$has":["*out"]}}`,
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewVal(2, 0),
	}
	a := &Action{Filter1: fmt.Sprintf(`{"ThresholdType":"%s"}`, utils.TRIGGER_MIN_BALANCE)}
	match, err := a.getFilter().Query(at, false)
	if err != nil {
		t.Fatalf("matcher failed: %v", err)
	}
	if match {
		t.Errorf("Action trigger [%+v] does not match action [%+v]", at, a)
	}
}

func TestActionTriggerMatcAllFalse(t *testing.T) {
	at := &ActionTrigger{
		TOR:            utils.MONETARY,
		Filter:         `{"Directions":{"$has":["*out"]}}`,
		ThresholdType:  utils.TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewVal(2, 0),
	}
	a := &Action{Filter1: fmt.Sprintf(`{"UniqueID":"ZIP", "ThresholdType":"%s"}`, utils.TRIGGER_MAX_BALANCE)}
	match, err := a.getFilter().Query(at, false)
	if err != nil {
		t.Fatalf("matcher failed: %v", err)
	}
	if match {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggerMatchAll(t *testing.T) {
	at := &ActionTrigger{
		UniqueID:      "ZIP",
		ThresholdType: "TT",
	}
	//x
	a := &Action{ /*Balance: &BalancePointer{
			Type:           utils.StringPointer(utils.MONETARY),
			RatingSubject:  utils.StringPointer("test1"),
			Directions:     utils.StringMapPointer(utils.NewStringMap(utils.OUT)),
			Value:          &utils.ValueFormula{Static: 2},
			Weight:         utils.Float64Pointer(1.0),
			DestinationIDs: utils.StringMapPointer(utils.NewStringMap("NAT")),
			SharedGroups:   utils.StringMapPointer(utils.NewStringMap("test2")),
		},*/Params: fmt.Sprintf(`{"UniqueID":"ZIP", "ThresholdType":"TT"}`)}
	match, err := a.getFilter().Query(at, false)
	if err != nil {
		t.Fatalf("matcher failed: %v", err)
	}
	if !match {
		t.Errorf("Action trigger [%v] does not match action [%v]", at, a)
	}
}

func TestActionTriggers(t *testing.T) {
	at1 := &ActionTrigger{Weight: 30}
	at2 := &ActionTrigger{Weight: 20}
	at3 := &ActionTrigger{Weight: 10}
	var atpl ActionTriggers
	atpl = append(atpl, at2, at1, at3)
	atpl.Sort()
	if atpl[0] != at1 || atpl[2] != at3 || atpl[1] != at2 {
		t.Error("List not sorted: ", atpl)
	}
}

func TestActionResetTriggres(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(10, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "y76", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "y77", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"y76": &ActionTriggerRecord{Executed: true}, "y77": &ActionTriggerRecord{Executed: true}},
	}
	resetTriggersAction(ub, nil, nil, nil)
	if ub.TriggerRecords["y76"].Executed == true || ub.TriggerRecords["y77"].Executed == true {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresExecutesThem(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(10, 0)}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "y78", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"y78": &ActionTriggerRecord{Executed: true}},
	}
	resetTriggersAction(ub, nil, nil, nil)
	if ub.TriggerRecords["y78"].Executed == true || ub.BalanceMap[utils.MONETARY][0].GetValue().String() == "12" {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionResetTriggresActionFilter(t *testing.T) {
	ub := &Account{
		Name:         "TEST_UB",
		BalanceMap:   map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(10, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers: ActionTriggers{
			&ActionTrigger{UniqueID: "y79", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"},
			&ActionTrigger{UniqueID: "y80", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{
			"y79": &ActionTriggerRecord{UniqueID: "y79", Executed: true},
			"y80": &ActionTriggerRecord{UniqueID: "y80", Executed: true},
		},
	}
	resetTriggersAction(ub, nil, &Action{Filter1: `{"UniqueID":"y81"}`}, nil)
	if ub.TriggerRecords["y79"].Executed == false || ub.TriggerRecords["y80"].Executed == false {
		t.Error("Reset triggers action failed!")
	}
}

func TestActionSetPostpaid(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "y81", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "y82", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"y81": &ActionTriggerRecord{Executed: true}, "y82": &ActionTriggerRecord{Executed: true}},
	}
	allowNegativeAction(ub, nil, nil, nil)
	if !ub.AllowNegative {
		t.Error("Set postpaid action failed!")
	}
}

func TestActionSetPrepaid(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		AllowNegative:  true,
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "y83", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "y84", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"y83": &ActionTriggerRecord{Executed: true}, "y84": &ActionTriggerRecord{Executed: true}},
	}
	denyNegativeAction(ub, nil, nil, nil)
	if ub.AllowNegative {
		t.Error("Set prepaid action failed!")
	}
}

func TestActionResetPrepaid(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		AllowNegative:  true,
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "y85", TOR: utils.SMS, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "y86", TOR: utils.SMS, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"y85": &ActionTriggerRecord{Executed: true}, "y86": &ActionTriggerRecord{Executed: true}},
	}
	resetAccountAction(ub, nil, nil, nil)
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "0" ||
		len(ub.UnitCounters) != 0 ||
		ub.BalanceMap[utils.VOICE][0].GetValue().String() != "0" ||
		ub.TriggerRecords["y85"].Executed == true || ub.TriggerRecords["y86"].Executed == true {
		t.Log(ub.BalanceMap)
		t.Error("Reset account action failed!")
	}
}

func TestActionResetPostpaid(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "y87", TOR: utils.SMS, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "y88", TOR: utils.SMS, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"y87": &ActionTriggerRecord{Executed: true}, "y88": &ActionTriggerRecord{Executed: true}},
	}
	resetAccountAction(ub, nil, nil, nil)
	if ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "0" ||
		len(ub.UnitCounters) != 0 ||
		ub.BalanceMap[utils.VOICE][0].GetValue().String() != "0" ||
		ub.TriggerRecords["y87"].Executed == true || ub.TriggerRecords["y88"].Executed == true {
		t.Error("Reset account action failed!")
	}
}

func TestActionTopupResetCredit(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0), Filter: `{"Directions":{"$in":["*out"]}}`}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "y89", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "y90", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"y89": &ActionTriggerRecord{Executed: true}, "y90": &ActionTriggerRecord{Executed: true}},
	}
	a := &Action{TOR: utils.MONETARY, Params: `{"Balance": {"Value":10}}`, Filter1: `{"Directions":{"$in":["*out"]}}`}
	if err := topupResetAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "10" ||
		len(ub.UnitCounters) != 0 || // InitCounters finds no counters
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.TriggerRecords["y89"].Executed != true || ub.TriggerRecords["y90"].Executed != true {
		t.Errorf("Topup reset action failed: %+s", utils.ToIJSON(ub))
	}
}

func TestActionTopupValueFactor(t *testing.T) {
	ub := &Account{
		Name:       "TEST_UB",
		BalanceMap: map[string]Balances{},
	}
	a := &Action{
		TOR:    utils.MONETARY,
		Params: `{"Balance":{"Value":10, "Directions":["*out"]}, "ValueFactor":{"*monetary":2.0}}`,
	}
	if err := topupResetAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if len(ub.BalanceMap) != 1 || ub.BalanceMap[utils.MONETARY][0].Factor[utils.MONETARY] != 2.0 {
		t.Errorf("Topup reset action failed to set Factor: %+v", ub.BalanceMap[utils.MONETARY][0].Factor)
	}
}

func TestActionTopupResetCreditId(t *testing.T) {
	ub := &Account{
		Name: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewVal(100, 0)},
				&Balance{ID: "TEST_B", Value: dec.NewVal(15, 0)},
			},
		},
	}
	a := &Action{TOR: utils.MONETARY, Params: `{"Balance":{"Value":10}}`, Filter1: `{"ID": "TEST_B", "Directions":{"$in":["*out"]}}`}
	if err := topupResetAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "110" ||
		len(ub.BalanceMap[utils.MONETARY]) != 2 {
		t.Errorf("Topup reset action failed: %s", utils.ToIJSON(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestActionTopupResetCreditNoId(t *testing.T) {
	ub := &Account{
		Name: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewVal(100, 0), Directions: utils.NewStringMap(utils.OUT)},
				&Balance{ID: "TEST_B", Value: dec.NewVal(15, 0), Directions: utils.NewStringMap(utils.OUT)},
			},
		},
	}
	a := &Action{TOR: utils.MONETARY, Params: `{"Balance":{"Value":10}}`, Filter1: `{"Directions":{"$in":["*out"]}}`}
	if err := topupResetAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "20" ||
		len(ub.BalanceMap[utils.MONETARY]) != 2 {
		t.Errorf("Topup reset action failed: %s", utils.ToIJSON(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestActionTopupResetMinutes(t *testing.T) {
	ub := &Account{
		Name: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}},
			utils.VOICE:    Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT"), Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0), Filter: `{"Directions":{"$in":["*out"]}}`}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "y91", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "y92", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"y91": &ActionTriggerRecord{Executed: true}, "y92": &ActionTriggerRecord{Executed: true}},
	}
	a := &Action{TOR: utils.VOICE, Params: `{"Balance":{"Value":5}}`, Filter1: `{"Weight":20, "DestinationIDs":{"$in":["NAT"]}, "Directions":{"$in":["*out"]}}`}
	if err := topupResetAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if ub.AllowNegative ||
		ub.BalanceMap[utils.VOICE].GetTotalValue().String() != "5" ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "100" ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.TriggerRecords["y91"].Executed != true || ub.TriggerRecords["y92"].Executed != true {
		t.Errorf("Topup reset minutes action failed: %s", utils.ToIJSON(ub.BalanceMap[utils.VOICE]))
	}
}

func TestActionTopupCredit(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT"), Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0), Filter: `{"Directions":{"$in":["*out"]}}`}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "yx1", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "yx2", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"yx1": &ActionTriggerRecord{Executed: true}, "yx2": &ActionTriggerRecord{Executed: true}},
	}
	a := &Action{TOR: utils.MONETARY, Params: `{"Balance":{"Value":10}}`, Filter1: `{"Directions":{"$in":["*out"]}}`}
	if err := topupAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "110" ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.TriggerRecords["yx1"].Executed != true || ub.TriggerRecords["yx2"].Executed != true {
		t.Error("Topup action failed!", ub.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestActionTopupMinutes(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT"), Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "yx1", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "yx2", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"yx1": &ActionTriggerRecord{Executed: true}, "yx2": &ActionTriggerRecord{Executed: true}},
	}
	a := &Action{TOR: utils.VOICE, Params: `{"Balance":{"Value":5}}`, Filter1: `{"Weight":20, "DestinationIDs":{"$in":["NAT"]}, "Directions":{"$in":["*out"]}}`}
	if err := topupAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if ub.AllowNegative ||
		ub.BalanceMap[utils.VOICE].GetTotalValue().String() != "15" ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "100" ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.TriggerRecords["yx1"].Executed != true || ub.TriggerRecords["yx2"].Executed != true {
		t.Error("Topup minutes action failed!", ub.BalanceMap[utils.VOICE])
	}
}

func TestActionDebitCredit(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0), Filter: `{"Directions":{"$in":["*out"]}}`}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "yx1", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "yx2", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"yx1": &ActionTriggerRecord{Executed: true}, "yx2": &ActionTriggerRecord{Executed: true}},
	}
	a := &Action{TOR: utils.MONETARY, Params: `{"Balance":{"Value":10}}`, Filter1: `{"Directions":{"$in":["*out"]}}`}
	if err := debitAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "90" ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.TriggerRecords["yx1"].Executed != true || ub.TriggerRecords["yx2"].Executed != true {
		t.Error("Debit action failed!", utils.ToIJSON(ub))
	}
}

func TestActionDebitMinutes(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT"), Directions: utils.NewStringMap(utils.OUT)}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0)}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "yx1", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}, &ActionTrigger{UniqueID: "yx2", TOR: utils.MONETARY, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"yx1": &ActionTriggerRecord{Executed: true}, "yx2": &ActionTriggerRecord{Executed: true}},
	}
	a := &Action{TOR: utils.VOICE, Params: `{"Balance":{"Value":5}}`, Filter1: `{"Weight":20, "DestinationIDs":{"$in":["NAT"]}, "Directions":{"$in":["*out"]}}`}
	if err := debitAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if ub.AllowNegative ||
		ub.BalanceMap[utils.VOICE][0].GetValue().String() != "5" ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "100" ||
		len(ub.UnitCounters) != 0 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.TriggerRecords["yx1"].Executed != true || ub.TriggerRecords["yx2"].Executed != true {
		t.Error("Debit minutes action failed!", ub.BalanceMap[utils.VOICE][0])
	}
}

func TestActionResetAllCounters(t *testing.T) {
	ub := &Account{
		Name:          "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}},
			utils.VOICE: Balances{
				&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT"), Directions: utils.NewStringMap(utils.OUT)},
				&Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET"), Directions: utils.NewStringMap(utils.OUT)}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "yx1", ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ThresholdValue: dec.NewVal(2, 0), TOR: utils.MONETARY, Filter: `{"DestinationIDs":{"$has":["NAT"]}, "Weight":20}`, ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"yx1": &ActionTriggerRecord{Executed: true}},
	}
	ub.InitCounters()
	if err := resetCountersAction(ub, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "100" ||
		len(ub.UnitCounters) != 1 ||
		len(ub.UnitCounters[utils.MONETARY][0].Counters) != 1 ||
		len(ub.BalanceMap[utils.MONETARY]) != 1 ||
		ub.TriggerRecords["yx1"].Executed != true {
		t.Errorf("Reset counters action failed: %+v %+v %+v", ub.UnitCounters, ub.UnitCounters[utils.MONETARY][0], ub.UnitCounters[utils.MONETARY][0].Counters[0])
	}
	if len(ub.UnitCounters) < 1 {
		t.FailNow()
	}
	c := ub.UnitCounters[utils.MONETARY][0].Counters[0]
	if c.UniqueID != "yx1" || c.Filter == "" {
		t.Errorf("bad counter: %s", utils.ToIJSON(c))
	}
}

func TestActionResetCounterOnlyDefault(t *testing.T) {
	ub := &Account{
		Name:          "TEST_UB",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}},
			utils.VOICE:    Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "yx1", TOR: utils.MONETARY, ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"yx1": &ActionTriggerRecord{Executed: true}},
	}
	a := &Action{ /*xBalance: &BalancePointer{Type: utils.StringPointer(utils.MONETARY)}*/ }
	ub.InitCounters()
	if err := resetCountersAction(ub, nil, a, nil); err != nil {
		t.Fatal(err)
	}
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "100" ||
		len(ub.UnitCounters) != 1 ||
		len(ub.UnitCounters[utils.MONETARY][0].Counters) != 1 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.TriggerRecords["yx1"].Executed != true {
		for _, b := range ub.UnitCounters[utils.MONETARY][0].Counters {
			t.Logf("B: %+v", b)
		}
		t.Errorf("Reset counters action failed: %+v", ub.UnitCounters)
	}
	if len(ub.UnitCounters) < 1 || len(ub.UnitCounters[utils.MONETARY][0].Counters) < 1 {
		t.FailNow()
	}
	c := ub.UnitCounters[utils.MONETARY][0].Counters[0]
	if c.UniqueID != "yx1" || c.Filter != "" {
		t.Errorf("Balance cloned incorrectly: %s", utils.ToIJSON(ub.UnitCounters[utils.MONETARY]))
	}
}

func TestActionResetCounterCredit(t *testing.T) {
	ub := &Account{
		Name:           "TEST_UB",
		AllowNegative:  true,
		BalanceMap:     map[string]Balances{utils.MONETARY: Balances{&Balance{Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.NewStringMap("NAT")}, &Balance{Weight: 10, DestinationIDs: utils.NewStringMap("RET")}}},
		UnitCounters:   UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0), Filter: `{"Directions":{"$in":["*out"]}}`}}}}, utils.SMS: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0), Filter: `{"Directions":{"$in":["*out"]}}`}}}}},
		triggers:       ActionTriggers{&ActionTrigger{UniqueID: "yx1", TOR: utils.MONETARY, Filter: `{"Directions":{"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ActionsID: "TEST_ACTIONS"}},
		TriggerRecords: map[string]*ActionTriggerRecord{"yx1": &ActionTriggerRecord{Executed: true}},
	}
	a := &Action{ /*xBalance: &BalancePointer{Type: utils.StringPointer(utils.MONETARY)}*/ }
	resetCountersAction(ub, nil, a, nil)
	if !ub.AllowNegative ||
		ub.BalanceMap[utils.MONETARY].GetTotalValue().String() != "100" ||
		len(ub.UnitCounters) != 2 ||
		len(ub.BalanceMap[utils.VOICE]) != 2 ||
		ub.TriggerRecords["yx1"].Executed != true {
		t.Error("Reset counters action failed!", ub.UnitCounters)
	}
}

func TestActionTriggerLogging(t *testing.T) {
	at := &ActionTrigger{
		UniqueID:       "some_uuid",
		TOR:            utils.MONETARY,
		Filter:         `{"Directions":{"$in":["*out"]}}, "DestinationIDs":{"$has":["NAT"]}}`,
		ThresholdValue: dec.NewVal(100, 0),
		Weight:         10.0,
		ActionsID:      "TEST_ACTIONS",
		parentGroup:    &ActionTriggerGroup{Tenant: "test"},
	}
	as, err := ratingStorage.GetActionGroup(at.parentGroup.Tenant, at.ActionsID, utils.CACHED)
	if err != nil {
		t.Error("Error getting actions for the action timing: ", as, err)
	}
	Publish(CgrEvent{
		"EventName": utils.EVT_ACTION_TRIGGER_FIRED,
		"Uuid":      at.UniqueID,
		"Tenant":    at.parentGroup.Tenant,
		"Id":        at.parentGroup.Name,
		"ActionIds": at.ActionsID,
	})
	//expected := "rif*some_uuid;MONETARY;OUT;NAT;TEST_ACTIONS;100;10;false*|TOPUP|MONETARY|OUT|10|0"
	var key string
	aplIter := ratingStorage.Iterator(ColApl, "", nil)
	var apl ActionPlan
	for aplIter.Next(&apl) {
		//_ = k
		//_ = v
		/*if strings.Contains(k, LOG_ACTION_utils.TRIGGER_PREFIX) && strings.Contains(v, expected) {
		    key = k
		    break
		}*/
	}
	if key != "" {
		t.Error("Action timing was not logged")
	}
}

func TestActionPlanLogging(t *testing.T) {
	i := &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{time.January, time.February, time.March, time.April, time.May, time.June, time.July, time.August, time.September, time.October, time.November, time.December},
			MonthDays: utils.MonthDays{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			WeekDays:  utils.WeekDays{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
			StartTime: "18:00:00",
			EndTime:   "00:00:00",
		},
		Weight: 10.0,
		Rating: &RIRate{
			ConnectFee: dec.NewVal(0, 0),
			Rates:      RateGroups{&RateInfo{0, dec.NewVal(1, 0), 1 * time.Second, 60 * time.Second}},
		},
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"one": true, "two": true, "three": true},
		Timing:     i,
		Weight:     10.0,
		ActionsID:  "TEST_ACTIONS",
		actionPlan: &ActionPlan{Tenant: "test", Name: "APL1"},
	}
	/*if err != nil {
		t.Error("Error getting actions for the action trigger: ", err)
	}*/
	Publish(CgrEvent{
		"EventName": utils.EVT_ACTION_TIMING_FIRED,
		"Uuid":      at.UUID,
		"Tenant":    at.actionPlan.Tenant,
		"Id":        at.actionPlan.Name,
		"ActionIds": at.ActionsID,
	})
	//expected := "some uuid|test|one,two,three|;1,2,3,4,5,6,7,8,9,10,11,12;1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31;1,2,3,4,5;18:00:00;00:00:00;10;0;1;60;1|10|TEST_ACTIONS*|TOPUP|MONETARY|OUT|10|0"
	var key string
	aplIter := ratingStorage.Iterator(ColApl, "", nil)
	var apl ActionPlan
	for aplIter.Next(&apl) {
		//_ = k
		//_ = v
		/*if strings.Contains(k, LOG_ACTION_TIMMING_PREFIX) && strings.Contains(string(v), expected) {
		    key = k
		}*/
	}
	if key != "" {
		t.Error("Action trigger was not logged")
	}
}

func TestActionMakeNegative(t *testing.T) {
	a := &Action{Params: `{"Balance":{"Value":10}}`}
	genericMakeNegative(a)
	b, err := a.getBalance(nil)
	if err != nil {
		t.Fatal(err)
	}
	if b.GetValue().GtZero() {
		t.Error("Failed to make negative: ", a)
	}
	genericMakeNegative(a)
	if b.GetValue().GtZero() {
		t.Error("Failed to preserve negative: ", a)
	}
}

func TestRemoveAction(t *testing.T) {
	if _, err := accountingStorage.GetAccount("test", "remo"); err != nil {
		t.Errorf("account to be removed not found: %v", err)
	}
	a := &Action{
		ActionType: REMOVE_ACCOUNT,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"remo": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "TOPUP10_AT"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}
	afterUb, err := accountingStorage.GetAccount("test", "remo")
	if err == nil || afterUb != nil {
		t.Error("error removing account: ", err, afterUb)
	}
}

func TestTopupAction(t *testing.T) {
	initialUb, _ := accountingStorage.GetAccount("test", "minu")
	a := &Action{
		ActionType: TOPUP,
		TOR:        utils.MONETARY,
		Params:     `{"Balance":{"Value":25}}`,
		Filter1:    `{"Weight":20, "DestinationIDs":{"$in":["RET"]}, "Directions":{"$in":["*out"]}}`,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"minu": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "TOPUP10_AT"},
	}

	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}
	afterUb, _ := accountingStorage.GetAccount("test", "minu")
	initialValue := initialUb.BalanceMap[utils.MONETARY].GetTotalValue()
	afterValue := afterUb.BalanceMap[utils.MONETARY].GetTotalValue()
	if afterValue.Cmp(initialValue.AddS(dec.NewVal(25, 0))) != 0 {
		t.Error("Bad topup before and after: ", initialValue, afterValue)
	}
}

func TestTopupActionLoaded(t *testing.T) {
	initialUb, _ := accountingStorage.GetAccount("test", "minitsboy")
	a := &Action{
		ActionType: TOPUP,
		TOR:        utils.MONETARY,
		Params:     `{"Balance":{"Value":25}}`,
		Filter1:    `{"Weight":20, "DestinationIDs":{"$in":["RET"]}, "Directions":{"$in":["*out"]}}`,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"minitsboy": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "MORE_MINUTES"},
	}

	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}
	afterUb, _ := accountingStorage.GetAccount("test", "minitsboy")
	initialValue := initialUb.BalanceMap[utils.MONETARY].GetTotalValue()
	afterValue := afterUb.BalanceMap[utils.MONETARY].GetTotalValue()
	if afterValue.Cmp(initialValue.AddS(dec.NewVal(25, 0))) != 0 {
		t.Logf("Initial: %+v", initialUb)
		t.Logf("After: %+v", afterUb)
		t.Error("Bad topup before and after: ", initialValue, afterValue)
	}
}

func TestActionCdrlogEmpty(t *testing.T) {
	acnt := &Account{Tenant: "test", Name: "dan2904"}
	cdrlog := &Action{
		ActionType: CDRLOG,
	}
	err := cdrLogAction(acnt, nil, cdrlog, Actions{
		&Action{
			ActionType: DEBIT,
			Params:     `{"Balance":{"Value":25}}`,
			Filter1:    `{"Weight":20, "DestinationIDs":{"$in":["RET"]}}`,
		},
	})
	if err != nil {
		t.Error("Error performing cdrlog action: ", err)
	}
	cdrs := make([]*CDR, 0)
	if err := json.Unmarshal([]byte(cdrlog.ExecFilter), &cdrs); err != nil {
		t.Fatal(err)
	}
	if len(cdrs) != 1 || cdrs[0].Source != CDRLOG {
		t.Errorf("Wrong cdrlogs: %+v", cdrs[0])
	}
}

func TestActionCdrlogWithParams(t *testing.T) {
	acnt := &Account{Tenant: "test", Name: "dan2904"}
	cdrlog := &Action{
		ActionType: CDRLOG,
		Params:     `{"CdrLogTemplate":{"ReqType":"^*pseudoprepaid","Subject":"^rif", "TOR":"~action_type:s/^\\*(.*)$/did_$1/"}}`,
	}
	err := cdrLogAction(acnt, nil, cdrlog, Actions{
		&Action{
			ActionType: DEBIT,
			Params:     `{"Balance":{"Value":25}}`,
			Filter1:    `{"Weight":20, "DestinationIDs":{"$in":["RET"]}}`,
		},
		&Action{
			ActionType: DEBIT_RESET,
			Params:     `{"Balance":{"Value":25}}`,
			Filter1:    `{"Weight":20, "DestinationIDs":{"$in":["RET"]}}`,
		},
	})
	if err != nil {
		t.Error("Error performing cdrlog action: ", err)
	}
	cdrs := make([]*CDR, 0)
	if err := json.Unmarshal([]byte(cdrlog.ExecFilter), &cdrs); err != nil {
		t.Fatal(err)
	}
	if len(cdrs) != 2 ||
		cdrs[0].Subject != "rif" {
		t.Errorf("Wrong cdrlogs: %+v", cdrs[0])
	}
}

func TestActionCdrLogParamsWithOverload(t *testing.T) {
	acnt := &Account{Tenant: "test", Name: "dan2904"}
	cdrlog := &Action{
		ActionType: CDRLOG,
		Params:     `{"CdrLogTemplate":{"Subject":"^rif","Destination":"^1234","ToR":"~ActionTag:s/^at(.)$/0$1/","AccountID":"~AccountID:s/^\\*(.*)$/$1/"}}`,
	}
	err := cdrLogAction(acnt, nil, cdrlog, Actions{
		&Action{
			ActionType: DEBIT,
			Params:     `{"Balance":{"Value":25}}`,
			Filter1:    `{"Weight":20, "DestinationIDs":{"$in":["RET"]}}`,
		},
		&Action{
			ActionType: DEBIT_RESET,
			Params:     `{"Balance":{"Value":25}}`,
			Filter1:    `{"Weight":20, "DestinationIDs":{"$in":["RET"]}}`,
		},
	})
	if err != nil {
		t.Error("Error performing cdrlog action: ", err)
	}
	cdrs := make([]*CDR, 0)
	json.Unmarshal([]byte(cdrlog.ExecFilter), &cdrs)
	expectedExtraFields := map[string]string{
		"AccountID": "test:dan2904",
	}
	if len(cdrs) != 2 ||
		cdrs[0].Subject != "rif" {
		t.Errorf("Wrong cdrlogs: %+v", cdrs[0])
	}
	if !reflect.DeepEqual(cdrs[0].ExtraFields, expectedExtraFields) {
		t.Errorf("Expecting extra fields: %+v, received: %+v", expectedExtraFields, cdrs[0].ExtraFields)
	}
}

func TestActionSetDDestination(t *testing.T) {
	acc := &Account{Tenant: "test", BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{DestinationIDs: utils.NewStringMap("*ddc_test")}}}}
	origD1 := &Destination{Tenant: "test", Name: "*ddc_test", Code: "111"}
	origD2 := &Destination{Tenant: "test", Name: "*ddc_test", Code: "222"}
	if err := ratingStorage.SetDestination(origD1); err != nil {
		t.Fatal(err)
	}
	if err := ratingStorage.SetDestination(origD2); err != nil {
		t.Fatal(err)
	}
	// check redis and cache
	if d, err := ratingStorage.GetDestinations("test", "", "*ddc_test", utils.DestExact, utils.CACHED); err != nil || !reflect.DeepEqual(d, Destinations{origD1, origD2}) {
		t.Error("Error storing destination: ", utils.ToIJSON(d), err)
	}
	if _, err := ratingStorage.GetDestinations("test", "111", "", utils.DestExact, utils.CACHED); err != nil {
		t.Fatal(err)
	}
	x1, found := cache2go.Get("test", utils.DESTINATION_PREFIX+utils.ConcatKey("111", "", utils.DestExact))
	if !found || len(x1.(Destinations)) != 1 {
		t.Error("Error cacheing destination: ", x1)
	}
	if _, err := ratingStorage.GetDestinations("test", "222", "", utils.DestExact, utils.CACHED); err != nil {
		t.Fatal(err)
	}
	x1, found = cache2go.Get("test", utils.DESTINATION_PREFIX+utils.ConcatKey("222", "", utils.DestExact))
	if !found || len(x1.(Destinations)) != 1 {
		t.Error("Error cacheing destination: ", x1)
	}
	if err := setddestinations(acc, &StatsQueueTriggered{Tenant: "test", Metrics: map[string]*dec.Dec{"333": dec.NewVal(1, 0), "666": dec.NewVal(1, 0)}}, nil, nil); err != nil {
		t.Fatal(err)
	}
	d, err := ratingStorage.GetDestinations("test", "", "*ddc_test", utils.DestExact, utils.CACHED)
	if err != nil ||
		d[0].Name != origD1.Name ||
		len(d) != 2 { //333 and 666
		t.Errorf("Error storing destination: %s %v", utils.ToIJSON(d), err)
	}

	var ok bool
	x1, ok = cache2go.Get("test", utils.DESTINATION_PREFIX+utils.ConcatKey("111", "", utils.DestExact))
	if ok {
		t.Error("Error cacheing destination: ", x1)
	}
	x1, ok = cache2go.Get("test", utils.DESTINATION_PREFIX+utils.ConcatKey("222", "", utils.DestExact))
	if ok {
		t.Error("Error cacheing destination: ", x1)
	}
	if _, err = ratingStorage.GetDestinations("test", "333", "", utils.DestExact, utils.CACHED); err != nil {
		t.Fatal(err)
	}
	x1, found = cache2go.Get("test", utils.DESTINATION_PREFIX+utils.ConcatKey("333", "", utils.DestExact))
	if !found || len(x1.(Destinations)) != 1 {
		t.Error("Error cacheing destination: ", x1)
	}
	if _, err := ratingStorage.GetDestinations("test", "666", "", utils.DestExact, utils.CACHED); err != nil {
		t.Fatal(err)
	}
	x1, found = cache2go.Get("test", utils.DESTINATION_PREFIX+utils.ConcatKey("333", "", utils.DestExact))
	if !found || len(x1.(Destinations)) != 1 {
		t.Error("Error cacheing destination: ", x1)
	}
}

func TestActionTransactionFuncType(t *testing.T) {
	err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "trans",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
				Value: dec.NewVal(10, 0),
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"trans": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			&Action{
				ActionType: TOPUP,
				TOR:        utils.MONETARY,
				Params:     `{"Balance":{"Value":1.1}}`,
			},
			&Action{
				ActionType: "VALID_FUNCTION_TYPE",
				TOR:        "test",
				Params:     `{"Balance":{"Value":1.1}}`,
			},
		},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	err = at.Execute()
	acc, err := accountingStorage.GetAccount("test", "trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MONETARY][0].Value.String() != "10" {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MONETARY][0].Value)
	}
}

func TestActionTransactionBalanceType(t *testing.T) {
	err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "trans",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
				Value: dec.NewVal(10, 0),
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"trans": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			&Action{
				ActionType: TOPUP,
				TOR:        utils.MONETARY,
				Params:     `{"Balance":{"Value":1.1}}`,
			},
			&Action{
				ActionType: TOPUP,
				TOR:        "test",
				Params:     `{"Balance":{}}`,
			},
		},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	err = at.Execute()
	if err != nil {
		t.Error(err)
	}
	acc, err := accountingStorage.GetAccount("test", "trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MONETARY][0].Value.String() != "11.1" {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MONETARY][0].Value)
	}
}

func TestActionTransactionBalanceNotType(t *testing.T) {
	err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "trans",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
				Value: dec.NewVal(10, 0),
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"trans": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			&Action{
				ActionType: TOPUP,
				TOR:        utils.VOICE,
				Params:     `{"Balance":{"Value":1.1}}`,
			},
			&Action{
				ActionType: TOPUP,
				TOR:        "test",
				Params:     `{"Balance":{}}`,
			},
		},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	err = at.Execute()
	acc, err := accountingStorage.GetAccount("test", "trans")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MONETARY][0].Value.String() != "10" {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MONETARY][0].Value)
	}
}

func TestActionWithExpireWithoutExpire(t *testing.T) {
	err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "exp",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
				Value: dec.NewVal(10, 0),
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"exp": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			&Action{
				ActionType: TOPUP,
				TOR:        utils.VOICE,
				Params:     `{"Balance":{"Value":15}}`,
			},
			&Action{
				ActionType: TOPUP,
				TOR:        utils.VOICE,
				Filter1:    `{"ExpirationDate":"2025-11-11T22:39:00Z"}`,
				Params:     `{"Balance":{"Value":30}, "ExpirationDate":"2025-11-11T22:39:00Z"}`,
			},
		},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	err = at.Execute()
	acc, err := accountingStorage.GetAccount("test", "exp")
	if err != nil || acc == nil {
		t.Errorf("Error getting account: %+v: %v", acc, err)
	}
	if len(acc.BalanceMap) != 2 ||
		len(acc.BalanceMap[utils.VOICE]) != 2 {
		t.Errorf("Error debiting expir and unexpire: %s", utils.ToIJSON(acc.BalanceMap[utils.VOICE]))
	}
}

func TestActionRemoveBalance(t *testing.T) {
	err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "rembal",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value: dec.NewVal(10, 0),
				},
				&Balance{
					Value:          dec.NewVal(10, 0),
					DestinationIDs: utils.NewStringMap("NAT", "RET"),
					ExpirationDate: time.Date(2025, time.November, 11, 22, 39, 0, 0, time.UTC),
				},
				&Balance{
					Value:          dec.NewVal(10, 0),
					DestinationIDs: utils.NewStringMap("NAT", "RET"),
				},
			},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"rembal": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			&Action{
				ActionType: REMOVE_BALANCE,
				TOR:        utils.MONETARY,
				Filter1:    `{"DestinationIDs":{"$has":["NAT", "RET"]}}`,
			},
		},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	err = at.Execute()
	acc, err := accountingStorage.GetAccount("test", "rembal")
	if err != nil || acc == nil {
		t.Errorf("Error getting account: %+v: %v", acc, err)
	}
	if len(acc.BalanceMap) != 1 ||
		len(acc.BalanceMap[utils.MONETARY]) != 1 {
		t.Errorf("Error removing balance: %s", utils.ToIJSON(acc.BalanceMap[utils.MONETARY]))
	}
}

func TestActionTransferMonetaryDefault(t *testing.T) {
	err := accountingStorage.SetAccount(
		&Account{
			Tenant: "test",
			Name:   "trans",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						UUID:  utils.GenUUID(),
						ID:    utils.META_DEFAULT,
						Value: dec.NewVal(10, 0),
					},
					&Balance{
						UUID:  utils.GenUUID(),
						Value: dec.NewVal(3, 0),
					},
					&Balance{
						UUID:  utils.GenUUID(),
						Value: dec.NewVal(6, 0),
					},
					&Balance{
						UUID:  utils.GenUUID(),
						Value: dec.NewVal(-2, 0),
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: TRANSFER_MONETARY_DEFAULT,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"trans": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}

	afterUb, err := accountingStorage.GetAccount("test", "trans")
	if err != nil {
		t.Fatal("account not found: ", err, afterUb)
	}
	if afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String() != "17" ||
		afterUb.BalanceMap[utils.MONETARY][0].Value.String() != "19" ||
		afterUb.BalanceMap[utils.MONETARY][1].Value.String() != "0" ||
		afterUb.BalanceMap[utils.MONETARY][2].Value.String() != "0" ||
		afterUb.BalanceMap[utils.MONETARY][3].Value.String() != "-2" {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String())
	}
}

func TestActionTransferMonetaryDefaultFilter(t *testing.T) {
	err := accountingStorage.SetAccount(
		&Account{
			Tenant: "test",
			Name:   "trans",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						UUID:   utils.GenUUID(),
						ID:     utils.META_DEFAULT,
						Value:  dec.NewVal(10, 0),
						Weight: 20,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(3, 0),
						Weight: 20,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(1, 0),
						Weight: 10,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(6, 0),
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: TRANSFER_MONETARY_DEFAULT,
		/*xBalance:    &BalancePointer{Weight: utils.Float64Pointer(20)},*/
		Filter1: `{"Weight":20}`,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"trans": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}

	afterUb, err := accountingStorage.GetAccount("test", "trans")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String() != "20" ||
		afterUb.BalanceMap[utils.MONETARY][0].Value.String() != "19" ||
		afterUb.BalanceMap[utils.MONETARY][1].Value.String() != "0" ||
		afterUb.BalanceMap[utils.MONETARY][2].Value.String() != "1" ||
		afterUb.BalanceMap[utils.MONETARY][3].Value.String() != "0" {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String())
	}
}

func TestActionConditionalTopup(t *testing.T) {
	err := accountingStorage.SetAccount(
		&Account{
			Tenant: "test",
			Name:   "cond",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						UUID:   utils.GenUUID(),
						ID:     utils.META_DEFAULT,
						Value:  dec.NewVal(10, 0),
						Weight: 20,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(3, 0),
						Weight: 20,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(1, 0),
						Weight: 10,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(6, 0),
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: TOPUP,
		ExecFilter: `{"Type":"*monetary","Value":1,"Weight":10}`,
		TOR:        utils.MONETARY,
		Filter1:    `{"Weight":30}`,
		Params:     `{"Balance":{"Weight":30, "Value":11}}`,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cond": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}

	afterUb, err := accountingStorage.GetAccount("test", "cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MONETARY]) != 5 ||
		afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String() != "31" ||
		afterUb.BalanceMap[utils.MONETARY][4].Value.String() != "11" {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String())
	}
}

func TestActionConditionalTopupNoQuery(t *testing.T) {
	err := accountingStorage.SetAccount(
		&Account{
			Tenant: "test",
			Name:   "cond",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						UUID:   utils.GenUUID(),
						ID:     utils.META_DEFAULT,
						Value:  dec.NewVal(10, 0),
						Weight: 20,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(3, 0),
						Weight: 20,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(1, 0),
						Weight: 10,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(6, 0),
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: TOPUP,
		ExecFilter: `{"Type":"*monetary","Value":2,"Weight":10}`,
		TOR:        utils.MONETARY,
		Filter1:    `{"Weight":30}`,
		Params:     `{"Balance":{"Weight":30, "Value":11}}`,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cond": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}

	afterUb, err := accountingStorage.GetAccount("test", "cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MONETARY]) != 4 ||
		afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String() != "20" {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String())
	}
}

func TestActionConditionalTopupExistingBalance(t *testing.T) {
	err := accountingStorage.SetAccount(
		&Account{
			Tenant: "test",
			Name:   "cond",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(1, 0),
						Weight: 10,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(6, 0),
						Weight: 20,
					},
				},
				utils.VOICE: Balances{
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(10, 0),
						Weight: 10,
					},
					&Balance{
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(100, 0),
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: TOPUP,
		ExecFilter: `{"Type":"*voice", "Value":{"$gte":100}}`,
		TOR:        utils.MONETARY,
		Filter1:    `{"Weight":10}`,
		Params:     `{"Balance":{"Weight":10, "Value":11}}`,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"cond": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}

	afterUb, err := accountingStorage.GetAccount("test", "cond")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MONETARY]) != 2 ||
		afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String() != "18" {
		for _, b := range afterUb.BalanceMap[utils.MONETARY] {
			t.Logf("B: %+v", b)
		}
		t.Error("ransfer balance value: ", afterUb.BalanceMap[utils.MONETARY].GetTotalValue().String())
	}
}

func TestActionConditionalDisabledIfNegative(t *testing.T) {
	err := accountingStorage.SetAccount(
		&Account{
			Tenant: "test",
			Name:   "af",
			BalanceMap: map[string]Balances{
				"*data": Balances{
					&Balance{
						UUID:          "fc927edb-1bd6-425e-a2a3-9fd8bafaa524",
						ID:            "for_v3hsillmilld500m_data_500_m",
						Value:         dec.NewFloat(5.242),
						Weight:        10,
						RatingSubject: "for_v3hsillmilld500m_data_forfait",
						Categories: utils.StringMap{
							"data_france": true,
						},
					},
				},
				"*monetary": Balances{
					&Balance{
						UUID:  "9fa1847a-f36a-41a7-8ec0-dfaab370141e",
						ID:    "*default",
						Value: dec.NewFloat(-1.95001),
					},
				},
				"*sms": Balances{
					&Balance{
						UUID:   "d348d15d-2988-4ee4-b847-6a552f94e2ec",
						ID:     "for_v3hsillmilld500m_mms_ill",
						Value:  dec.NewVal(20000, 0),
						Weight: 10,
						DestinationIDs: utils.StringMap{
							"FRANCE_NATIONAL": true,
						},
						Categories: utils.StringMap{
							"mms_france":  true,
							"tmms_france": true,
							"vmms_france": true,
						},
					},
					&Balance{
						UUID:   "f4643517-31f6-4199-980f-04cf535471ed",
						ID:     "for_v3hsillmilld500m_sms_ill",
						Value:  dec.NewVal(20000, 0),
						Weight: 10,
						DestinationIDs: utils.StringMap{
							"FRANCE_NATIONAL": true,
						},
						Categories: utils.StringMap{
							"ms_france": true,
						},
					},
				},
				"*voice": Balances{
					&Balance{
						UUID:   "079ab190-77f4-44f3-9c6f-3a0dd1a59dfd",
						ID:     "for_v3hsillmilld500m_voice_3_h",
						Value:  dec.NewVal(10800, 0),
						Weight: 10,
						DestinationIDs: utils.StringMap{
							"FRANCE_NATIONAL": true,
						},
						Categories: utils.StringMap{
							"call_france": true,
						},
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a1 := &Action{
		ActionType: "*set_balance",
		ExecFilter: `{"$and":[{"Value":{"$lt":0}},{"ID":{"$eq":"*default"}}]}`,
		TOR:        utils.SMS,
		Filter1:    `{"ID":"for_v3hsillmilld500m_sms_ill"}`,
		Params:     `{"Balance":{"Disabled":true}}`,
		Weight:     9,
	}
	a2 := &Action{
		ActionType: "*set_balance",
		ExecFilter: `{"$and":[{"Value":{"$lt":0}},{"ID":{"$eq":"*default"}}]}`,
		TOR:        utils.SMS,
		Filter1:    `{"ID":"for_v3hsillmilld500m_mms_ill"}`,
		Params:     `{"Balance":{"Disabled":true, "DestinationIDs":["FRANCE_NATIONAL"], "Weight":10}}`,
		Weight:     8,
	}
	a3 := &Action{
		ActionType: "*set_balance",
		ExecFilter: `{"$and":[{"Value":{"$lt":0}},{"ID":{"$eq":"*default"}}]}`,
		TOR:        utils.SMS,
		Filter1:    `{"ID":"for_v3hsillmilld500m_sms_ill"}`,
		Params:     `{"Balance":{"Disabled":true, "DestinationIDs":["FRANCE_NATIONAL"], "Weight":10}}`,
		Weight:     8,
	}
	a4 := &Action{
		ActionType: "*set_balance",
		ExecFilter: `{"$and":[{"Value":{"$lt":0}},{"ID":{"$eq":"*default"}}]}`,
		TOR:        utils.DATA,
		Filter1:    `{"UUID":"fc927edb-1bd6-425e-a2a3-9fd8bafaa524"}`,
		Params:     `{"Balance":{"Disabled":true, "RatingSubject":"for_v3hsillmilld500m_data_forfait", "Weight":10}}`,
		Weight:     7,
	}
	a5 := &Action{
		ActionType: "*set_balance",
		ExecFilter: `{"$and":[{"Value":{"$lt":0}},{"ID":{"$eq":"*default"}}]}`,
		TOR:        utils.VOICE,
		Filter1:    `{"ID":"for_v3hsillmilld500m_voice_3_h"}`,
		Params:     `{"Balance":{"Disabled":true, "DestinationIDs":["FRANCE_NATIONAL"], "Weight":10}}`,
		Weight:     6,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"af": true},
		actions:    Actions{a1, a2, a3, a4, a5},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}

	afterUb, err := accountingStorage.GetAccount("test", "af")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}

	for btype, chain := range afterUb.BalanceMap {
		if btype != utils.MONETARY {
			for _, b := range chain {
				if b.Disabled != true {
					t.Errorf("Failed to disabled balance (%s): %s", btype, utils.ToIJSON(b))
				}
			}
		}
	}
}

func TestActionSetBalance(t *testing.T) {
	err := accountingStorage.SetAccount(
		&Account{
			Tenant: "test",
			Name:   "setb",
			BalanceMap: map[string]Balances{
				utils.MONETARY: Balances{
					&Balance{
						ID:     "m1",
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(1, 0),
						Weight: 10,
					},
					&Balance{
						ID:     "m2",
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(6, 0),
						Weight: 20,
					},
				},
				utils.VOICE: Balances{
					&Balance{
						ID:     "v1",
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(10, 0),
						Weight: 10,
					},
					&Balance{
						ID:     "v2",
						UUID:   utils.GenUUID(),
						Value:  dec.NewVal(100, 0),
						Weight: 20,
					},
				},
			},
		})
	if err != nil {
		t.Errorf("error setting account: %v", err)
	}

	a := &Action{
		ActionType: SET_BALANCE,
		TOR:        utils.MONETARY,
		Filter1:    `{"ID":"m2"}`,
		Params:     `{"Balance":{"Value":11,"Weight":10}}`,
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"setb": true},
		actions:    Actions{a},
		actionPlan: &ActionPlan{Tenant: "test", Name: "XXX"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}

	afterUb, err := accountingStorage.GetAccount("test", "setb")
	if err != nil {
		t.Error("account not found: ", err, afterUb)
	}
	if len(afterUb.BalanceMap[utils.MONETARY]) != 2 ||
		afterUb.BalanceMap[utils.MONETARY][1].Value.String() != "11" ||
		afterUb.BalanceMap[utils.MONETARY][1].Weight != 10 {
		t.Errorf("Balance: %s", utils.ToIJSON(afterUb.BalanceMap[utils.MONETARY]))
	}
}

func TestActionCSVFilter(t *testing.T) {
	ag, err := ratingStorage.GetActionGroup("test", "FILTER", utils.CACHED)
	if err != nil {
		t.Error("error getting actions: ", err)
	}
	if len(ag.Actions) != 1 || ag.Actions[0].ExecFilter != `{"$and":[{"Value":{"$lt":0}},{"ID":{"$eq":"*default"}}]}` {
		t.Error("Error loading actions: ", ag.Actions[0].ExecFilter)
	}
}

func TestActionExpirationTime(t *testing.T) {
	ag, err := ratingStorage.GetActionGroup("test", "EXP", utils.CACHED)
	if err != nil || ag == nil {
		t.Error("Error getting actions: ", err)
	}
	ag.SetParentGroup()

	at := &ActionTiming{
		accountIDs: utils.StringMap{"expo": true},
		actions:    ag.Actions,
		actionPlan: &ActionPlan{Tenant: "test"},
	}
	for rep := 0; rep < 5; rep++ {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
		afterUb, err := accountingStorage.GetAccount("test", "expo")
		if err != nil ||
			len(afterUb.BalanceMap[utils.VOICE]) != 1 ||
			afterUb.BalanceMap[utils.VOICE][0].ExpirationDate.IsZero() {
			t.Error("error topuping expiration balance: ", utils.ToIJSON(afterUb))
		}
	}
}

func TestActionExpNoExp(t *testing.T) {
	exp, err := ratingStorage.GetActionGroup("test", "EXP", utils.CACHED)
	if err != nil || exp == nil {
		t.Error("Error getting actions: ", err)
	}
	exp.SetParentGroup()
	noexp, err := ratingStorage.GetActionGroup("test", "NOEXP", utils.CACHED)
	if err != nil || noexp == nil {
		t.Error("Error getting actions: ", err)
	}
	noexp.SetParentGroup()
	exp.Actions = append(exp.Actions, noexp.Actions...)
	at := &ActionTiming{
		accountIDs: utils.StringMap{"expnoexp": true},
		actions:    exp.Actions,
		actionPlan: &ActionPlan{Tenant: "test"},
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}
	afterUb, err := accountingStorage.GetAccount("test", "expnoexp")
	if err != nil ||
		len(afterUb.BalanceMap[utils.VOICE]) != 2 {
		t.Error("error topuping expiration balance: ", utils.ToIJSON(afterUb))
	}
}

func TestActionCdrlogBalanceValue(t *testing.T) {
	err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "bv",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{
				ID:    "*default",
				UUID:  "25a02c82-f09f-4c6e-bacf-8ed4b076475a",
				Value: dec.NewVal(10, 0),
			}},
		},
	})
	if err != nil {
		t.Error("Error setting account: ", err)
	}
	at := &ActionTiming{
		accountIDs: utils.StringMap{"bv": true},
		Timing:     &RateInterval{},
		actions: []*Action{
			&Action{
				//Id:         "RECUR_FOR_V3HSILLMILLD1G",
				ActionType:  TOPUP,
				TOR:         utils.MONETARY,
				Filter1:     `{"ID":"*default", "UUID":"25a02c82-f09f-4c6e-bacf-8ed4b076475a"}`,
				Params:      `{"Balance":{"Value":1.1}}`,
				parentGroup: &ActionGroup{Name: "AG1"},
			},
			&Action{
				//Id:         "RECUR_FOR_V3HSILLMILLD5G",
				ActionType:  DEBIT,
				TOR:         utils.MONETARY,
				Filter1:     `{"ID":"*default", "UUID":"25a02c82-f09f-4c6e-bacf-8ed4b076475a"}`,
				Params:      `{"Balance":{"Value":2.1}}`,
				parentGroup: &ActionGroup{Name: "AG1"},
			},
			&Action{
				//Id:              "c",
				ActionType:  CDRLOG,
				Params:      `{"CdrLogTemplate":{"BalanceID":"BalanceID","BalanceUUID":"BalanceUUID","ActionID":"ActionID","BalanceValue":"BalanceValue"}}`,
				parentGroup: &ActionGroup{Name: "AG1"},
			},
		},
		actionPlan: &ActionPlan{Tenant: "test"},
	}
	err = at.Execute()
	acc, err := accountingStorage.GetAccount("test", "bv")
	if err != nil || acc == nil {
		t.Error("Error getting account: ", acc, err)
	}
	if acc.BalanceMap[utils.MONETARY][0].Value.String() != "9" {
		t.Errorf("Transaction didn't work: %v", acc.BalanceMap[utils.MONETARY][0].Value)
	}
	cdrs := make([]*CDR, 0)
	json.Unmarshal([]byte(at.actions[2].ExecFilter), &cdrs)
	if len(cdrs) != 2 ||
		cdrs[0].ExtraFields["BalanceValue"] != "11.1" ||
		cdrs[1].ExtraFields["BalanceValue"] != "9" {
		t.Errorf("Wrong cdrlogs: %s", utils.ToIJSON(cdrs))
	}
}

type TestRPCParameters struct {
	status string
}

type Attr struct {
	Name    string
	Surname string
	Age     float64
}

func (trpcp *TestRPCParameters) Hopa(in Attr, out *float64) error {
	trpcp.status = utils.OK
	return nil
}

func (trpcp *TestRPCParameters) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(trpcp).MethodByName(parts[1])
	if !method.IsValid() {
		return utils.ErrNotImplemented
	}

	// construct the params
	params := []reflect.Value{reflect.ValueOf(args).Elem(), reflect.ValueOf(reply)}

	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

func TestCgrRpcAction(t *testing.T) {
	trpcp := &TestRPCParameters{}
	utils.RegisterRpcParams("", trpcp)
	a := &Action{
		Params: `{"RpcRequest":{"Address": "*internal",
	"Transport": "*gob",
	"Method": "TestRPCParameters.Hopa",
	"Attempts":1,
	"Async" :false,
	"Params": {"Name":"n", "Surname":"s", "Age":10.2}}}`,
	}
	if err := cgrRPCAction(nil, nil, a, nil); err != nil {
		t.Error("error executing cgr action: ", err)
	}
	if trpcp.status != utils.OK {
		t.Error("RPC not called!")
	}
}

func TestValueFormulaDebit(t *testing.T) {
	if _, err := accountingStorage.GetAccount("test", "vf"); err != nil {
		t.Errorf("account to be removed not found: %v", err)
	}

	at := &ActionTiming{
		accountIDs: utils.StringMap{"vf": true},
		actionPlan: &ActionPlan{Tenant: "test"},
		ActionsID:  "VF",
	}
	if err := at.Execute(); err != nil {
		t.Fatal(err)
	}
	afterUb, err := accountingStorage.GetAccount("test", "vf")
	// not an exact value, depends of month
	v := afterUb.BalanceMap[utils.MONETARY].GetTotalValue()
	t.Log(dec.NewFloat(-0.30), v, dec.NewFloat(-0.36))
	if err != nil || v.Cmp(dec.NewFloat(-0.30)) > 0 || v.Cmp(dec.NewFloat(-0.36)) < 0 {
		t.Error("error debiting account: ", err, utils.ToIJSON(afterUb))
	}
}

/**************** Benchmarks ********************************/

func BenchmarkUUID(b *testing.B) {
	m := make(map[string]int, 1000)
	for i := 0; i < b.N; i++ {
		uuid := utils.GenUUID()
		if len(uuid) == 0 {
			b.Fatalf("GenUUID error %s", uuid)
		}
		b.StopTimer()
		c := m[uuid]
		if c > 0 {
			b.Fatalf("duplicate uuid[%s] count %d", uuid, c)
		}
		m[uuid] = c + 1
		b.StartTimer()
	}
}
