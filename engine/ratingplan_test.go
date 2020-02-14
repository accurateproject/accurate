package engine

import (
	"encoding/json"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"

	"reflect"
	"testing"
	"time"
)

func TestRPApRestoreFromStorage(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   utils.OUT,
		Category:    "0",
		Tenant:      "test",
		Subject:     "rif:from:tm",
		Destination: "49"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Errorf("Error restoring activation periods: %+v", cd.RatingInfos[0])
	}
}

func TestRPApStoreRestoreJson(t *testing.T) {
	i := &RateInterval{Timing: &RITiming{
		Months:    []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}}
	ap := &RatingPlan{Name: "test"}
	ap.AddRateInterval("0723", "NAT", i)
	result, _ := json.Marshal(ap)
	ap1 := &RatingPlan{}
	json.Unmarshal(result, ap1)
	if !reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestRPApStoreRestoreBlank(t *testing.T) {
	i := &RateInterval{}
	ap := &RatingPlan{Name: "test"}
	ap.AddRateInterval("0723", "NAT", i)
	result, _ := json.Marshal(ap)
	ap1 := RatingPlan{}
	json.Unmarshal(result, &ap1)
	if reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestRPFallbackDirect(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Category:    "0",
		Direction:   utils.OUT,
		Tenant:      "test",
		Subject:     "danb:87.139.12.167",
		Destination: "41"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestRPFallbackMultiple(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Category:    "0",
		Direction:   utils.OUT,
		Tenant:      "test",
		Subject:     "fall",
		Destination: "0723045"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Errorf("Error restoring rating plans: %+v", cd.RatingInfos)
	}
}

func TestRPFallbackWithBackTrace(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Category:    "0",
		Direction:   utils.OUT,
		Tenant:      "test",
		Subject:     "danb:87.139.12.167",
		Destination: "4123"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 1 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestRPFallbackNoDefault(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Category:    "0",
		Direction:   utils.OUT,
		Tenant:      "test",
		Subject:     "one",
		Destination: "0723"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestRPFallbackNoInfiniteLoop(t *testing.T) {
	cd := &CallDescriptor{Category: "0", Direction: utils.OUT, Tenant: "test", Subject: "rif", Destination: "0721"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestRPFallbackNoInfiniteLoopSelf(t *testing.T) {
	cd := &CallDescriptor{Category: "0", Direction: utils.OUT, Tenant: "test", Subject: "inf", Destination: "0721"}
	cd.LoadRatingPlans()
	if len(cd.RatingInfos) != 0 {
		t.Error("Error restoring activation periods: ", len(cd.RatingInfos))
	}
}

func TestRPApAddIntervalIfNotPresent(t *testing.T) {
	i1 := &RateInterval{
		Timing: &RITiming{
			Months:    utils.Months{time.February},
			MonthDays: utils.MonthDays{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	i2 := &RateInterval{Timing: &RITiming{
		Months:    utils.Months{time.February},
		MonthDays: utils.MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}}
	i3 := &RateInterval{Timing: &RITiming{
		Months:    utils.Months{time.February},
		MonthDays: utils.MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}}
	rp := &RatingPlan{}
	rp.AddRateInterval("0723", "NAT", i1)
	rp.AddRateInterval("0723", "NAT", i2)
	if len(rp.DestinationRates["0723"].DRateKeys) != 1 {
		t.Error("Wronfullyrppended interval ;)")
	}
	rp.AddRateInterval("0723", "NAT", i3)
	if len(rp.DestinationRates["0723"].DRateKeys) != 2 {
		t.Error("Wronfully not appended interval ;)", utils.ToIJSON(rp.DestinationRates))
	}
}

func TestRPApAddRateIntervalGroups(t *testing.T) {
	i1 := &RateInterval{
		Rating: &RIRate{Rates: RateGroups{&RateInfo{0, dec.NewFloat(1), 1 * time.Second, 1 * time.Second}}},
	}
	i2 := &RateInterval{
		Rating: &RIRate{Rates: RateGroups{&RateInfo{30 * time.Second, dec.NewFloat(2), 1 * time.Second, 1 * time.Second}}},
	}
	i3 := &RateInterval{
		Rating: &RIRate{Rates: RateGroups{&RateInfo{30 * time.Second, dec.NewFloat(2), 1 * time.Second, 1 * time.Second}}},
	}
	ap := &RatingPlan{}
	ap.AddRateInterval("0723", "NAT", i1)
	ap.AddRateInterval("0723", "NAT", i2)
	ap.AddRateInterval("0723", "NAT", i3)
	if len(ap.DestinationRates) != 1 {
		t.Error("Wronfully appended interval ;)", utils.ToIJSON(ap.DestinationRates))
	}
	if len(ap.RateIntervalList("0723")[0].Rating.Rates) != 1 {
		t.Errorf("Group prices not formed: %#v", ap.RateIntervalList("NAT")[0].Rating.Rates[0])
	}
}

func TestRPGetActiveForCall(t *testing.T) {
	rpas := RatingPlanActivations{
		&RatingPlanActivation{ActivationTime: time.Date(2013, 1, 1, 0, 0, 0, 0, time.UTC)},
		&RatingPlanActivation{ActivationTime: time.Date(2013, 11, 12, 11, 40, 0, 0, time.UTC)},
		&RatingPlanActivation{ActivationTime: time.Date(2013, 11, 13, 0, 0, 0, 0, time.UTC)},
	}
	cd := &CallDescriptor{
		TimeStart: time.Date(2013, 11, 12, 11, 39, 0, 0, time.UTC),
		TimeEnd:   time.Date(2013, 11, 12, 11, 45, 0, 0, time.UTC),
	}
	active := rpas.GetActiveForCall(cd)
	if len(active) != 2 {
		t.Errorf("Error getting active rating plans: %+v", active)
	}
}

func TestRPRatingPlanIsContinousEmpty(t *testing.T) {
	rpl := &RatingPlan{}
	if rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRPRatingPlanIsContinousBlank(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"blank": &RITiming{StartTime: "00:00:00"},
			"other": &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRPRatingPlanIsContinousGood(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"first":  &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second": &RITiming{WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":  &RITiming{WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRPRatingPlanisContinousBad(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"first":  &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second": &RITiming{WeekDays: utils.WeekDays{4, 5, 0}, StartTime: "00:00:00"},
		},
	}
	if rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRPRatingPlanIsContinousSpecial(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special": &RITiming{Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first":   &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"second":  &RITiming{WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":   &RITiming{WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRPRatingPlanIsContinousMultiple(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special":  &RITiming{Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first":    &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "00:00:00"},
			"first_08": &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "08:00:00"},
			"second":   &RITiming{WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":    &RITiming{WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if !rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRPRatingPlanIsContinousMissing(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"special":  &RITiming{Years: utils.Years{2015}, Months: utils.Months{5}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
			"first_08": &RITiming{WeekDays: utils.WeekDays{1, 2, 3}, StartTime: "08:00:00"},
			"second":   &RITiming{WeekDays: utils.WeekDays{4, 5, 6}, StartTime: "00:00:00"},
			"third":    &RITiming{WeekDays: utils.WeekDays{0}, StartTime: "00:00:00"},
		},
	}
	if rpl.isContinous() {
		t.Errorf("Error determining rating plan's valididty: %+v", rpl)
	}
}

func TestRPRatingPlanSaneTimingsBad(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"one": &RITiming{Years: utils.Years{2015}, WeekDays: utils.WeekDays{time.Monday}},
		},
	}
	if crazyTiming := rpl.getFirstUnsaneTiming(); crazyTiming == "" {
		t.Errorf("Error detecting bad timings in rating profile: %+v", rpl)
	}
}

func TestRPRatingPlanSaneTimingsGood(t *testing.T) {
	rpl := &RatingPlan{
		Timings: map[string]*RITiming{
			"one": &RITiming{Years: utils.Years{2015}},
			"two": &RITiming{WeekDays: utils.WeekDays{0, 1, 2, 3, 4}, StartTime: "00:00:00"},
		},
	}
	if crazyTiming := rpl.getFirstUnsaneTiming(); crazyTiming != "" {
		t.Errorf("Error detecting bad timings in rating profile: %+v", rpl)
	}
}

func TestRPRatingPlanSaneRatingsEqual(t *testing.T) {
	rpl := &RatingPlan{
		Ratings: map[string]*RIRate{
			"one": &RIRate{
				Rates: RateGroups{
					&RateInfo{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      30 * time.Second,
					},
					&RateInfo{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      30 * time.Second,
					},
				},
			},
		},
	}
	if crazyRating := rpl.getFirstUnsaneRating(); crazyRating == "" {
		t.Errorf("Error detecting bad rate groups in rating profile: %+v", rpl)
	}
}

func TestRPRatingPlanSaneRatingsNotMultiple(t *testing.T) {
	rpl := &RatingPlan{
		Ratings: map[string]*RIRate{
			"one": &RIRate{
				Rates: RateGroups{
					&RateInfo{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      30 * time.Second,
					},
					&RateInfo{
						GroupIntervalStart: 15 * time.Second,
						RateIncrement:      30 * time.Second,
					},
				},
			},
		},
	}
	if crazyRating := rpl.getFirstUnsaneRating(); crazyRating == "" {
		t.Errorf("Error detecting bad rate groups in rating profile: %+v", rpl)
	}
}

func TestRPRatingPlanSaneRatingsGoot(t *testing.T) {
	rpl := &RatingPlan{
		Ratings: map[string]*RIRate{
			"one": &RIRate{
				Rates: RateGroups{
					&RateInfo{
						GroupIntervalStart: 60 * time.Second,
						RateIncrement:      30 * time.Second,
						RateUnit:           1 * time.Second,
					},
					&RateInfo{
						GroupIntervalStart: 0 * time.Second,
						RateIncrement:      30 * time.Second,
						RateUnit:           1 * time.Second,
					},
				},
			},
		},
	}
	if crazyRating := rpl.getFirstUnsaneRating(); crazyRating != "" {
		t.Errorf("Error detecting bad rate groups in rating profile: %+v", rpl)
	}
}

/**************************** Benchmarks *************************************/

func BenchmarkRPMarshalJson(b *testing.B) {
	b.StopTimer()
	i := &RateInterval{
		Timing: &RITiming{
			Months:    []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	ap := &RatingPlan{Name: "test"}
	ap.AddRateInterval("0723", "NAT", i)

	ap1 := RatingPlan{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		result, _ := json.Marshal(ap)
		json.Unmarshal(result, &ap1)
	}
}

func BenchmarkRPRestore(b *testing.B) {
	i := &RateInterval{
		Timing: &RITiming{Months: []time.Month{time.February},
			MonthDays: []int{1},
			WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
			StartTime: "14:30:00",
			EndTime:   "15:00:00"}}
	rp := &RatingPlan{Tenant: "test", Name: "test"}
	rp.AddRateInterval("0723", "NAT", i)
	ratingStorage.SetRatingPlan(rp)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ratingStorage.GetRatingPlan(rp.Tenant, rp.Name, utils.CACHE_SKIP)
	}
}
