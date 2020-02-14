package engine

import (
	"log"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	dockertest "gopkg.in/ory-am/dockertest.v3"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/history"
	"github.com/accurateproject/accurate/utils"
)

func TestMain(m *testing.M) {
	var err error
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{
		"/tmp/accurate.log",
	}
	utils.Logger, err = cfg.Build()
	if err != nil {
		log.Print("Cannot initialize development logging!!!")
	}
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("mongo", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		port := resource.GetPort("27017/tcp")
		ratingStorage, err = NewMongoStorage("127.0.0.1", port, "tp_test", "", "", utils.TariffPlanDB, nil, &config.Cache{RatingPlans: &config.CacheParam{Precache: true}}, 10)
		if err != nil {
			log.Fatal(err)
		}
		accountingStorage, err = NewMongoStorage("127.0.0.1", port, "acc_test", "", "", utils.DataDB, nil, &config.Cache{RatingPlans: &config.CacheParam{Precache: true}}, 10)
		if err != nil {
			log.Fatal(err)
		}

		return accountingStorage.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	ratingStorage.Flush()
	accountingStorage.Flush()
	aliasService = NewAliasHandler(accountingStorage)
	historyScribe, _ = history.NewMockScribe()
	populateDB()
	loadFromJSON()

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func populateDB() {
	ats := []*Action{
		&Action{ActionType: "*topup", TOR: utils.MONETARY, Params: `{"Balance": {"Value":10}}`},
		&Action{ActionType: "*topup", TOR: utils.VOICE, Params: `{"Balance": {"Weight":20, "Value": 10, "DestinationIDs": {"NAT":true}}}`, Filter1: `{"Weight":20, "DestinationIDs": {"$in":["NAT"]}}`},
	}

	ats1 := []*Action{
		&Action{ActionType: "*topup", TOR: utils.MONETARY, Params: `{"Balance": {"Value":10}}`, Weight: 10},
		&Action{ActionType: "*reset_account", Weight: 20},
	}

	minu := &Account{
		Tenant: "test",
		Name:   "minu",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Value: dec.NewFloat(50)}},
			utils.VOICE: Balances{
				&Balance{Value: dec.NewFloat(200), DestinationIDs: utils.NewStringMap("NAT"), Weight: 10},
				&Balance{Value: dec.NewFloat(100), DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	broker := &Account{
		Tenant: "test",
		Name:   "broker",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{
				&Balance{Value: dec.NewFloat(20), DestinationIDs: utils.NewStringMap("NAT"), Weight: 10, RatingSubject: "rif"},
				&Balance{Value: dec.NewFloat(100), DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			}},
	}
	luna := &Account{
		Tenant: "test",
		Name:   "luna",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewFloat(0), Weight: 20},
			}},
	}
	// this is added to test if csv load tests account will not overwrite balances
	minitsboy := &Account{
		Tenant: "test",
		Name:   "minitsboy",
		BalanceMap: map[string]Balances{
			utils.VOICE: Balances{
				&Balance{Value: dec.NewFloat(20), DestinationIDs: utils.NewStringMap("NAT"), Weight: 10, RatingSubject: "rif"},
				&Balance{Value: dec.NewFloat(100), DestinationIDs: utils.NewStringMap("RET"), Weight: 20},
			},
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewFloat(100), Weight: 10},
			},
		},
	}
	max := &Account{
		Tenant: "test",
		Name:   "max",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewFloat(11), Weight: 20},
			}},
	}
	money := &Account{
		Tenant: "test",
		Name:   "money",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewFloat(10000), Weight: 10},
			}},
	}
	if accountingStorage != nil && ratingStorage != nil {
		ratingStorage.SetActionGroup(&ActionGroup{Tenant: "test", Name: "TEST_ACTIONS", Actions: ats})
		ratingStorage.SetActionGroup(&ActionGroup{Tenant: "test", Name: "TEST_ACTIONS_ORDER", Actions: ats1})
		accountingStorage.SetAccount(broker)
		accountingStorage.SetAccount(minu)
		accountingStorage.SetAccount(minitsboy)
		accountingStorage.SetAccount(luna)
		accountingStorage.SetAccount(max)
		accountingStorage.SetAccount(money)
	} else {
		log.Fatal("Could not connect to db!")
	}
}

func TestSplitSpans(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2, TOR: utils.VOICE}

	if err := cd.LoadRatingPlans(); err != nil {
		t.Fatal(err)
	}
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.RatingInfos)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestSplitSpansWeekend(t *testing.T) {
	cd := &CallDescriptor{Direction: utils.OUT,
		Category:        "postpaid",
		TOR:             utils.VOICE,
		Tenant:          "foehn",
		Subject:         "foehn",
		Account:         "foehn",
		Destination:     "0034678096720",
		TimeStart:       time.Date(2015, 4, 24, 7, 59, 4, 0, time.UTC),
		TimeEnd:         time.Date(2015, 4, 24, 8, 2, 0, 0, time.UTC),
		LoopIndex:       0,
		DurationIndex:   176 * time.Second,
		FallbackSubject: "",
		RatingInfos: RatingInfos{
			&RatingInfo{
				MatchedSubject: "*out:foehn:postpaid:foehn",
				MatchedPrefix:  "0034678",
				MatchedDestID:  "SPN_MOB",
				ActivationTime: time.Date(2015, 4, 23, 0, 0, 0, 0, time.UTC),
				RateIntervals: []*RateInterval{
					&RateInterval{
						Timing: &RITiming{
							WeekDays:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
							StartTime: "08:00:00",
						},
						Rating: &RIRate{
							ConnectFee: dec.NewVal(0, 0),
							Rates: RateGroups{
								&RateInfo{Value: dec.NewVal(1, 0), RateIncrement: 1 * time.Second, RateUnit: 1 * time.Second},
							},
						},
					},
					&RateInterval{
						Timing: &RITiming{
							WeekDays:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
							StartTime: "00:00:00",
						},
						Rating: &RIRate{
							ConnectFee: dec.NewVal(0, 0),
							Rates: RateGroups{
								&RateInfo{Value: dec.NewVal(1, 0), RateIncrement: 1 * time.Second, RateUnit: 1 * time.Second},
							},
						},
					},
					&RateInterval{
						Timing: &RITiming{
							WeekDays:  []time.Weekday{time.Saturday, time.Sunday},
							StartTime: "00:00:00",
						},
						Rating: &RIRate{
							ConnectFee: dec.NewVal(0, 0),
							Rates: RateGroups{
								&RateInfo{Value: dec.NewVal(1, 0), RateIncrement: 1 * time.Second, RateUnit: 1 * time.Second},
							},
						},
					},
				},
			},
		},
	}

	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.RatingInfos)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
	if timespans[0].RateInterval == nil ||
		timespans[0].RateInterval.Timing.StartTime != "00:00:00" ||
		timespans[1].RateInterval == nil ||
		timespans[1].RateInterval.Timing.StartTime != "08:00:00" {
		t.Errorf("Error setting rateinterval: %+v %+v", timespans[0].RateInterval.Timing.StartTime, timespans[1].RateInterval.Timing.StartTime)
	}
}

func TestSplitSpansRoundToIncrements(t *testing.T) {
	t1 := time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC)
	t2 := time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "trp", Destination: "0256", TimeStart: t1, TimeEnd: t2, DurationIndex: 132 * time.Second}

	cd.LoadRatingPlans()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Logf("%+v", cd)
		t.Log(cd.RatingInfos)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
	var d time.Duration
	for _, ts := range timespans {
		d += ts.GetDuration()
	}
	if d != 132*time.Second {
		t.Error("Wrong duration for timespans: ", d)
	}
}

func TestCalldescHolliday(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart: time.Date(2015, time.May, 1, 13, 30, 0, 0, time.UTC),
		TimeEnd:   time.Date(2015, time.May, 1, 13, 35, 26, 0, time.UTC),
		RatingInfos: RatingInfos{
			&RatingInfo{
				RateIntervals: RateIntervalList{
					&RateInterval{
						Timing: &RITiming{WeekDays: utils.WeekDays{1, 2, 3, 4, 5}, StartTime: "00:00:00"},
						Weight: 10,
					},
					&RateInterval{
						Timing: &RITiming{WeekDays: utils.WeekDays{6, 7}, StartTime: "00:00:00"},
						Weight: 10,
					},
					&RateInterval{
						Timing: &RITiming{Months: utils.Months{time.May}, MonthDays: utils.MonthDays{1}, StartTime: "00:00:00"},
						Weight: 20,
					},
				},
			},
		},
	}
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 1 {
		t.Error("Error assiging holidy rate interval: ", timespans)
	}
	if timespans[0].RateInterval.Timing.MonthDays == nil {
		t.Errorf("Error setting holiday rate interval: %+v", timespans[0].RateInterval.Timing)
	}
}

func TestGetCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0256", Cost: dec.NewVal(2701, 0)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "1" {
		t.Log(result.Cost.Cmp(expected.Cost), expected.Cost.String(), result.Cost.String())
		//t.Log(result.Cost.Internal())
		//t.Log(expected.Cost.Internal())
		t.Errorf("Expected %v was %+v", expected.Cost.String(), result.Cost)
	}
}

func TestGetCostRounding(t *testing.T) {
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 33, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "test", Subject: "round", Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	result, _ := cd.GetCost()
	if result.Cost.Cmp(dec.NewFloat(0.300000001)) != 0 || result.GetConnectFee().String() != "0" {
		t.Error("bad cost", utils.ToIJSON(result))
	}
}

func TestDebitRounding(t *testing.T) {
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 33, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "test", Subject: "round", Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	result, _ := cd.Debit()
	if result.Cost.Cmp(dec.NewFloat(0.300000001)) != 0 || result.GetConnectFee().String() != "0" {
		t.Error("bad cost", utils.ToIJSON(result))
	}
}

func TestDebitPerformRounding(t *testing.T) {
	t1 := time.Date(2017, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2017, time.February, 2, 17, 33, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "call", Tenant: "test", Subject: "round", Destination: "49", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, PerformRounding: true}
	result, _ := cd.Debit()
	if result.Cost.Cmp(dec.NewFloat(0.300000001)) != 0 || result.GetConnectFee().String() != "0" {
		t.Error("bad cost", utils.ToIJSON(result))
	}
}

func TestGetCostZero(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2, LoopIndex: 0}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0256", Cost: dec.NewFloat(0)}
	if result.GetCost().Cmp(expected.GetCost()) != 0 || result.GetConnectFee().String() != "0" {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestGetCostTimespans(t *testing.T) {
	t1 := time.Date(2013, time.October, 8, 9, 23, 2, 0, time.UTC)
	t2 := time.Date(2013, time.October, 8, 9, 24, 27, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "trp", Destination: "0256", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: 85 * time.Second}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "trp", Destination: "0256", Cost: dec.NewFloat(85)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "0" || len(result.Timespans) != 2 {
		t.Errorf("Expected %+v was %+v", expected, result)
	}

}

func TestGetCostRatingPlansAndRatingIntervals(t *testing.T) {
	t1 := time.Date(2012, time.February, 27, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 28, 18, 10, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif:from:tm", Destination: "49178", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: t2.Sub(t1)}
	result, _ := cd.GetCost()
	if len(result.Timespans) != 3 ||
		!result.Timespans[0].TimeEnd.Equal(result.Timespans[1].TimeStart) ||
		!result.Timespans[1].TimeEnd.Equal(result.Timespans[2].TimeStart) {
		for _, ts := range result.Timespans {
			t.Logf("TS %+v", ts)
		}
		t.Errorf("Expected %+v was %+v", 3, len(result.Timespans))
	}
}

func TestGetCostRatingPlansAndRatingIntervalsMore(t *testing.T) {
	t1 := time.Date(2012, time.February, 27, 9, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 28, 18, 10, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif:from:tm", Destination: "49178", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: t2.Sub(t1)}
	result, _ := cd.GetCost()
	if len(result.Timespans) != 4 ||
		!result.Timespans[0].TimeEnd.Equal(result.Timespans[1].TimeStart) ||
		!result.Timespans[1].TimeEnd.Equal(result.Timespans[2].TimeStart) ||
		!result.Timespans[2].TimeEnd.Equal(result.Timespans[3].TimeStart) {
		for _, ts := range result.Timespans {
			t.Logf("TS %+v", ts)
		}
		t.Errorf("Expected %+v was %+v", 4, len(result.Timespans))
	}
}

func TestGetCostRatingPlansAndRatingIntervalsMoreDays(t *testing.T) {
	t1 := time.Date(2012, time.February, 20, 9, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 23, 18, 10, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif:from:tm", Destination: "49178", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: t2.Sub(t1)}
	result, _ := cd.GetCost()
	if len(result.Timespans) != 8 ||
		!result.Timespans[0].TimeEnd.Equal(result.Timespans[1].TimeStart) ||
		!result.Timespans[1].TimeEnd.Equal(result.Timespans[2].TimeStart) ||
		!result.Timespans[2].TimeEnd.Equal(result.Timespans[3].TimeStart) ||
		!result.Timespans[3].TimeEnd.Equal(result.Timespans[4].TimeStart) ||
		!result.Timespans[4].TimeEnd.Equal(result.Timespans[5].TimeStart) ||
		!result.Timespans[5].TimeEnd.Equal(result.Timespans[6].TimeStart) ||
		!result.Timespans[6].TimeEnd.Equal(result.Timespans[7].TimeStart) {
		for _, ts := range result.Timespans {
			t.Logf("TS %+v", ts)
		}
		t.Errorf("Expected %+v was %+v", 4, len(result.Timespans))
	}
}

func TestGetCostRatingPlansAndRatingIntervalsMoreDaysWeekend(t *testing.T) {
	t1 := time.Date(2012, time.February, 24, 9, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 27, 18, 10, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif:from:tm", Destination: "49178", TimeStart: t1, TimeEnd: t2, LoopIndex: 0, DurationIndex: t2.Sub(t1)}
	result, _ := cd.GetCost()
	expectedLength := 6
	if len(result.Timespans) != expectedLength ||
		!result.Timespans[0].TimeEnd.Equal(result.Timespans[1].TimeStart) ||
		!result.Timespans[1].TimeEnd.Equal(result.Timespans[2].TimeStart) ||
		!result.Timespans[2].TimeEnd.Equal(result.Timespans[3].TimeStart) ||
		!result.Timespans[3].TimeEnd.Equal(result.Timespans[4].TimeStart) ||
		!result.Timespans[4].TimeEnd.Equal(result.Timespans[5].TimeStart) {
		for _, ts := range result.Timespans {
			t.Logf("TS %v - %v", ts.TimeStart, ts.TimeEnd)
		}
		t.Errorf("Expected %+v was %+v", expectedLength, len(result.Timespans))
	}
}

func TestGetCostRateGroups(t *testing.T) {
	t1 := time.Date(2013, time.October, 7, 14, 50, 0, 0, time.UTC)
	t2 := time.Date(2013, time.October, 7, 14, 52, 12, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "trp", Destination: "0256", TimeStart: t1, TimeEnd: t2, DurationIndex: 132 * time.Second}

	result, err := cd.GetCost()
	if err != nil {
		t.Error("Error getting cost: ", err)
	}
	if result.Cost.String() != "132" {
		t.Error("Error calculating cost: ", result.Timespans)
	}
}

func TestGetCostNoConnectFee(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2, LoopIndex: 1}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0256", Cost: dec.NewFloat(2700)}
	// connect fee is not added because LoopIndex is 1
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "1" {
		t.Log(result.Cost.Cmp(expected.Cost), result.Cost.String(), expected.Cost.String())
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestGetCostAccount(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Account: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0256", Cost: dec.NewFloat(2701)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "1" {
		t.Log(result.Cost.Cmp(expected.Cost), result.Cost.String(), expected.Cost.String())
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestFullDestNotFound(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0256", Cost: dec.NewFloat(2701)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "1" {
		//t.Log(cd.RatingInfos)
		t.Log(result.Cost.Cmp(expected.Cost), result.Cost.String(), expected.Cost.String())
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSubjectNotFound(t *testing.T) {
	t1 := time.Date(2013, time.February, 1, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2013, time.February, 1, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "not_exiting", Destination: "025740532", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0257", Cost: dec.NewFloat(2701)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "1" {
		//t.Logf("%+v", result.Timespans[0].RateInterval)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSubjectNotFoundCostNegativeOne(t *testing.T) {
	t1 := time.Date(2013, time.February, 1, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2013, time.February, 1, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test_nix", Subject: "not_exiting", Destination: "025740532", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	if result.Cost.String() != "-1" || result.GetConnectFee().String() != "0" {
		//t.Logf("%+v", result.Timespans[0].RateInterval)
		t.Errorf("Expected -1 was %v", utils.ToIJSON(result))
	}
}

func TestMultipleRatingPlans(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0257", Cost: dec.NewFloat(2701)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "1" {
		t.Log(result.Timespans)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSpansMultipleRatingPlans(t *testing.T) {
	t1 := time.Date(2012, time.February, 7, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 0, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	cc, _ := cd.GetCost()
	if cc.Cost.String() != "2100" || cc.GetConnectFee().String() != "0" {
		utils.LogFull(cc)
		t.Errorf("Expected %v was %v (%v)", 2100, cc, cc.GetConnectFee())
	}
}

func TestLessThanAMinute(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 30, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0257", Cost: dec.NewFloat(15)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "0" {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestUniquePrice(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 21, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0723045326", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "rif", Destination: "0723", Cost: dec.NewFloat(1810.5)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "0" {
		t.Log(result.Cost.Cmp(expected.Cost), result.Cost.String(), expected.Cost.String())
		//t.Log(result.Cost.Internal())
		//t.Log(expected.Cost.Internal())
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMinutesCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 22, 51, 50, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0723", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "test", Subject: "minutosu", Destination: "0723", Cost: dec.NewFloat(55)}
	if result.Cost.Cmp(expected.Cost) != 0 || result.GetConnectFee().String() != "0" {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeNoAccount(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "ttttttt",
		Destination: "0723"}
	result, err := cd.GetMaxSessionDuration()
	if result != 0 || err == nil {
		t.Errorf("Expected %v was %v (%v)", 0, result, err)
	}
}

func TestMaxSessionTimeWithAccount(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "minu",
		Destination: "0723",
	}
	result, err := cd.GetMaxSessionDuration()
	expected := time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeWithMaxRate(t *testing.T) {
	ap, err := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	if err != nil {
		t.FailNow()
	}
	ap.SetParentActionPlan()
	//log.Print(ap)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	//acc, _ := accountingStorage.GetAccount("test:12345")
	//log.Print("ACC: ", utils.ToIJSON(acc))
	cd := &CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "test",
		Subject:     "12345",
		Account:     "12345",
		Destination: "447956",
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 1, 0, 0, time.UTC),
		MaxRate:     1.0,
		MaxRateUnit: time.Minute,
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 40 * time.Second
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeWithMaxCost(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 6, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewFloat(0),
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 10 * time.Second
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestGetMaxSessiontWithBlocker(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "BLOCK_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	acc, err := accountingStorage.GetAccount("test", "block")
	if err != nil {
		t.Error("error getting account: ", err)
	}
	if len(acc.BalanceMap[utils.MONETARY]) != 2 ||
		acc.BalanceMap[utils.MONETARY][0].Blocker != true {
		t.Error("Error executing action  plan on account: ", utils.ToIJSON(acc.BalanceMap[utils.MONETARY]))
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "block",
		Account:      "block",
		Destination:  "0723",
		TimeStart:    time.Date(2016, 1, 13, 14, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2016, 1, 13, 14, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 17 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v (%v)", expected, result, err)
	}
	cd = &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "block",
		Account:      "block",
		Destination:  "444",
		TimeStart:    time.Date(2016, 1, 13, 14, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2016, 1, 13, 14, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	result, err = cd.GetMaxSessionDuration()
	expected = 30 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v (%v)", expected, result, err)
	}
}

func TestGetMaxSessiontWithBlockerEmpty(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "BLOCK_EMPTY_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {

		at.Execute()
	}
	acc, err := accountingStorage.GetAccount("test", "block_empty")
	if err != nil {
		t.Error("error getting account: ", err)
	}
	if len(acc.BalanceMap[utils.MONETARY]) != 2 ||
		acc.BalanceMap[utils.MONETARY][0].Blocker != true {
		t.Error("Error executing action  plan on account: ", utils.ToIJSON(acc.BalanceMap))
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "block",
		Account:      "block_empty",
		Destination:  "0723",
		TimeStart:    time.Date(2016, 1, 13, 14, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2016, 1, 13, 14, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 0 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v (%v)", expected, result, err)
	}
	cd = &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "block",
		Account:      "block_empty",
		Destination:  "444",
		TimeStart:    time.Date(2016, 1, 13, 14, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2016, 1, 13, 14, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	result, err = cd.GetMaxSessionDuration()
	expected = 30 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v (%v)", expected, result, err)
	}
}

func TestGetCostWithMaxCost(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		at.Execute()
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 6, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 6, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	cc, err := cd.GetCost()
	expected := "1800"
	if cc.Cost.String() != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, cc.Cost)
	}
}

func TestGetCostRoundingIssue(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		at.Execute()
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 51, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	cc, err := cd.GetCost()
	expected := "0.17"
	if cc.Cost.String() != expected || err != nil {
		t.Log(utils.ToIJSON(cc))
		t.Errorf("Expected %v was %+v", expected, cc)
	}
}

func TestGetCostRatingInfoOnZeroTime(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		at.Execute()
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	cc, err := cd.GetCost()
	if err != nil ||
		len(cc.Timespans) != 1 ||
		cc.Timespans[0].MatchedDestID != "RET" ||
		cc.Timespans[0].MatchedSubject != "*out:test:call:dy" ||
		cc.Timespans[0].MatchedPrefix != "0723" ||
		cc.Timespans[0].RatingPlanID != "DY_PLAN" {
		t.Error("MatchedInfo not added:", utils.ToIJSON(cc))
	}
}

func TestDebitRatingInfoOnZeroTime(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		at.Execute()
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	cc, err := cd.Debit()
	if err != nil ||
		cc == nil ||
		len(cc.Timespans) != 1 ||
		cc.Timespans[0].MatchedDestID != "RET" ||
		cc.Timespans[0].MatchedSubject != "*out:test:call:dy" ||
		cc.Timespans[0].MatchedPrefix != "0723" ||
		cc.Timespans[0].RatingPlanID != "DY_PLAN" {
		t.Error("MatchedInfo not added:", utils.ToIJSON(cc))
	}
}

func TestMaxDebitRatingInfoOnZeroTime(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		at.Execute()
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	cc, err := cd.MaxDebit()
	if err != nil ||
		len(cc.Timespans) != 1 ||
		cc.Timespans[0].MatchedDestID != "RET" ||
		cc.Timespans[0].MatchedSubject != "*out:test:call:dy" ||
		cc.Timespans[0].MatchedPrefix != "0723" ||
		cc.Timespans[0].RatingPlanID != "DY_PLAN" {
		t.Error("MatchedInfo not added:", utils.ToIJSON(cc))
	}
}

func TestMaxDebitUnknowDest(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		at.Execute()
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "dy",
		Account:      "dy",
		Destination:  "9999999999",
		TimeStart:    time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:      time.Date(2015, 10, 26, 13, 29, 29, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	cc, err := cd.MaxDebit()
	if err == nil || err != utils.ErrUnauthorizedDestination {
		t.Errorf("Bad error reported %+v: %v", cc, err)
	}
}

func TestMaxDebitRoundingIssue(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd := &CallDescriptor{
		Direction:       "*out",
		Category:        "call",
		Tenant:          "test",
		Subject:         "dy",
		Account:         "dy",
		Destination:     "0723123113",
		TimeStart:       time.Date(2015, 10, 26, 13, 29, 27, 0, time.UTC),
		TimeEnd:         time.Date(2015, 10, 26, 13, 29, 51, 0, time.UTC),
		MaxCostSoFar:    dec.NewFloat(0),
		PerformRounding: true,
	}
	acc, err := accountingStorage.GetAccount("test", "dy")
	if err != nil || acc.BalanceMap[utils.MONETARY][0].Value.String() != "1" {
		t.Errorf("Error getting account: %+v (%v)", utils.ToIJSON(acc), err)
	}

	cc, err := cd.MaxDebit()
	if err != nil {
		t.Fatal(err)
	}
	expected := "0.17"
	if cc.GetCost().String() != expected || err != nil {
		t.Errorf("Expected %v was %+v (%v)", expected, utils.ToIJSON(cc), err)
	}
	acc, err = accountingStorage.GetAccount("test", "dy")
	if err != nil || acc.BalanceMap[utils.MONETARY][0].Value.String() != "0.83" {
		t.Errorf("Error getting account: %+v (%v)", utils.ToIJSON(acc), err)
	}
}

func TestDebitRoundingRefund(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd := &CallDescriptor{
		Direction:       "*out",
		Category:        "call",
		Tenant:          "test",
		Subject:         "dy",
		Account:         "dy",
		Destination:     "0723123113",
		TimeStart:       time.Date(2016, 3, 4, 13, 50, 00, 0, time.UTC),
		TimeEnd:         time.Date(2016, 3, 4, 13, 53, 00, 0, time.UTC),
		MaxCostSoFar:    dec.NewFloat(0),
		PerformRounding: true,
	}
	acc, err := accountingStorage.GetAccount("test", "dy")
	if err != nil || acc.BalanceMap[utils.MONETARY][0].Value.String() != "1" {
		t.Errorf("Error getting account: %+v (%v)", utils.ToIJSON(acc), err)
	}

	cc, err := cd.Debit()
	if err != nil {
		t.Fatal(err)
	}
	expected := "0.3"
	if cc.Cost.String() != expected || err != nil {
		t.Log(utils.ToIJSON(cc))
		t.Errorf("Expected %v was %+v (%v)", expected, cc, err)
	}
	acc, err = accountingStorage.GetAccount("test", "dy")
	if err != nil || acc.BalanceMap[utils.MONETARY][0].Value.String() != "0.7" {
		t.Errorf("Error getting account: %+v (%v)", utils.ToIJSON(acc), err)
	}
}

func TestMaxSessionTimeWithMaxCostFree(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 19, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 19, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	result, err := cd.GetMaxSessionDuration()
	expected := 30 * time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxDebitWithMaxCostFree(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 19, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 19, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}
	cc, err := cd.MaxDebit()
	expected := "10"
	if cc.Cost.String() != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, utils.ToIJSON(cc))
	}
}

func TestGetCostWithMaxCostFree(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd := &CallDescriptor{
		Direction:    "*out",
		Category:     "call",
		Tenant:       "test",
		Subject:      "max",
		Account:      "max",
		Destination:  "0723123113",
		TimeStart:    time.Date(2015, 3, 23, 19, 0, 0, 0, time.UTC),
		TimeEnd:      time.Date(2015, 3, 23, 19, 30, 0, 0, time.UTC),
		MaxCostSoFar: dec.NewVal(0, 0),
	}

	cc, err := cd.GetCost()
	expected := "10"
	if cc.Cost.String() != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, cc.Cost)
	}
}

func TestMaxSessionTimeWithAccountAlias(t *testing.T) {
	aliasService = NewAliasHandler(accountingStorage)
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "a1",
		Account:     "a1",
		Destination: "0723",
	}
	LoadAlias(&AttrAlias{
		Destination: cd.Destination,
		Direction:   cd.Direction,
		Tenant:      cd.Tenant,
		Category:    cd.Category,
		Account:     cd.Account,
		Subject:     cd.Subject,
		Context:     utils.ALIAS_CONTEXT_RATING,
	}, cd, utils.EXTRA_FIELDS)

	result, err := cd.GetMaxSessionDuration()
	expected := time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v, %v", expected, result, err)
	}
}

func TestMaxSessionTimeWithAccountShared(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP_SHARED0_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	ap, _ = ratingStorage.GetActionPlan("test", "TOPUP_SHARED10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}

	cd0 := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "rif",
		Account:     "empty0",
		Destination: "0723",
	}

	cd1 := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "rif",
		Account:     "empty10",
		Destination: "0723",
	}

	result0, err := cd0.GetMaxSessionDuration()
	result1, err := cd1.GetMaxSessionDuration()
	if result0 != result1/2 || err != nil {
		t.Errorf("Expected %v was %v, %v", result1/2, result0, err)
	}
}

func TestMaxDebitWithAccountShared(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP_SHARED0_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	ap, _ = ratingStorage.GetActionPlan("test", "TOPUP_SHARED10_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}

	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 5, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "minu",
		Account:     "empty0",
		Destination: "0723",
	}

	cc, err := cd.MaxDebit()
	if err != nil || cc.Cost.String() != "2.5" {
		t.Errorf("Wrong callcost in shared debit: %s, %v", utils.ToIJSON(cc), err)
	}
	acc, _ := cd.getAccount()
	balanceMap := acc.BalanceMap[utils.MONETARY]
	if len(balanceMap) != 1 || balanceMap[0].GetValue().String() != "0" {
		t.Errorf("Wrong shared balance debited: %+v", balanceMap[0])
	}
	other, err := accountingStorage.GetAccount("test", "empty10")
	if err != nil || other.BalanceMap[utils.MONETARY][0].GetValue().String() != "7.5" {
		t.Errorf("Error debiting shared balance: %+v", other.BalanceMap[utils.MONETARY][0])
	}
}

func TestMaxSessionTimeWithAccountAccount(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "minu_from_tm",
		Account:     "minu",
		Destination: "0723",
	}
	result, err := cd.GetMaxSessionDuration()
	expected := time.Minute
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeNoCredit(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "broker",
		Destination: "0723",
		TOR:         utils.VOICE,
	}
	result, err := cd.GetMaxSessionDuration()
	if result != time.Minute || err != nil {
		t.Errorf("Expected %v was %v", time.Minute, result)
	}
}

func TestMaxSessionModifiesCallDesc(t *testing.T) {
	t1 := time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC)
	t2 := time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC)
	cd := &CallDescriptor{
		TimeStart:     t1,
		TimeEnd:       t2,
		Direction:     "*out",
		Category:      "0",
		Tenant:        "test",
		Subject:       "minu_from_tm",
		Account:       "minu",
		Destination:   "0723",
		DurationIndex: t2.Sub(t1),
		TOR:           utils.VOICE,
	}
	initial := cd.Clone()
	_, err := cd.GetMaxSessionDuration()
	if err != nil {
		t.Error("Got error from max duration: ", err)
	}
	cd.account = nil // it's OK to cache the account
	if !reflect.DeepEqual(cd, initial) {
		t.Errorf("GetMaxSessionDuration is changing the call descriptor %+v != %+v", cd, initial)
	}
}

func TestMaxDebitDurationNoGreatherThanInitialDuration(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "minu_from_tm",
		Account:     "minu",
		Destination: "0723",
	}
	initialDuration := cd.TimeEnd.Sub(cd.TimeStart)
	result, err := cd.GetMaxSessionDuration()
	if err != nil {
		t.Error("Got error from max duration: ", err)
	}
	if result > initialDuration {
		t.Error("max session duration greather than initial duration", initialDuration, result)
	}
}

func TestDebitAndMaxDebit(t *testing.T) {
	cd1 := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 10, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "minu_from_tm",
		Account:     "minu",
		Destination: "0723",
	}
	cd2 := cd1.Clone()
	cc1, err1 := cd1.Debit()
	cc2, err2 := cd2.MaxDebit()
	if err1 != nil || err2 != nil {
		t.Error("Error debiting and/or maxdebiting: ", err1, err2)
	}
	if cc1.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.getValue().String() != "90" ||
		cc2.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.getValue().String() != "80" {
		t.Error("Error setting the Unit.Value: ", cc1.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.Value, cc2.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.Value)
	}
	// make Unit.Values have the same value
	cc1.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.Value = dec.New()
	cc2.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.Value = dec.New()
	if !reflect.DeepEqual(cc1, cc2) {
		t.Log("CC1: ", utils.ToIJSON(cc1))
		t.Log("CC2: ", utils.ToIJSON(cc2))
		t.Error("Debit and MaxDebit differ")
	}
}

func TestMaxSesionTimeEmptyBalance(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "minu_from_tm",
		Account:     "luna",
		Destination: "0723",
	}
	acc, _ := accountingStorage.GetAccount("test", "luna")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	if err != nil || allowedTime != 0 {
		t.Error("Error get max session for 0 acount", err)
	}
}

func TestMaxSesionTimeEmptyBalanceAndNoCost(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 35, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "one",
		Account:     "luna",
		Destination: "112",
	}
	acc, _ := accountingStorage.GetAccount("test", "luna")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	if err != nil || allowedTime == 0 {
		t.Error("Error get max session for 0 acount", err)
	}
}

func TestMaxSesionTimeLong(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2015, 07, 24, 13, 37, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 07, 24, 15, 37, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "call",
		Tenant:      "test",
		Subject:     "money",
		Destination: "0723",
	}
	acc, _ := accountingStorage.GetAccount("test", "money")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	if err != nil || allowedTime != cd.TimeEnd.Sub(cd.TimeStart) {
		t.Error("Error get max session for acount:", allowedTime, err)
	}
}

func TestMaxSesionTimeLongerThanMoney(t *testing.T) {
	cd := &CallDescriptor{
		TimeStart:   time.Date(2015, 07, 24, 13, 37, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 07, 24, 16, 37, 0, 0, time.UTC),
		Direction:   "*out",
		Category:    "call",
		Tenant:      "test",
		Subject:     "money",
		Destination: "0723",
	}
	acc, _ := accountingStorage.GetAccount("test", "money")
	allowedTime, err := cd.getMaxSessionDuration(acc)
	expected, err := time.ParseDuration("9999s") // 1 is the connect fee
	if err != nil || allowedTime != expected {
		t.Log(utils.ToIJSON(acc))
		t.Errorf("Expected: %v got %v", expected, allowedTime)
	}
}

func TestDebitFromShareAndNormal(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP_SHARED10_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		at.Execute()
	}

	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 5, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "rif",
		Account:     "empty10",
		Destination: "0723",
	}
	cc, err := cd.MaxDebit()
	acc, _ := cd.getAccount()
	balanceMap := acc.BalanceMap[utils.MONETARY]
	if err != nil || cc.Cost.String() != "2.5" {
		t.Errorf("Debit from share and normal error: %+v, %v", cc, err)
	}

	if balanceMap[0].GetValue().String() != "10" || balanceMap[1].GetValue().String() != "27.5" || len(balanceMap) != 2 {
		t.Errorf("Error debiting from right balance: %s", utils.ToIJSON(balanceMap))
	}
}

func TestDebitFromEmptyShare(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP_EMPTY_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}

	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 5, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "rif",
		Account:     "emptyX",
		Destination: "0723",
	}

	cc, err := cd.MaxDebit()
	if err != nil || cc.Cost.String() != "2.5" {
		t.Errorf("Debit from empty share error: %+v, %v", cc, err)
	}
	acc, _ := cd.getAccount()
	balanceMap := acc.BalanceMap[utils.MONETARY]
	if len(balanceMap) != 2 || balanceMap[0].GetValue().String() != "0" || balanceMap[1].GetValue().String() != "-2.5" {
		t.Errorf("Error debiting from empty share: %+v", balanceMap[1].GetValue())
	}
}

func TestDebitNegatve(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "POST_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		at.Execute()
	}

	cd := &CallDescriptor{
		TimeStart:   time.Date(2013, 10, 21, 18, 34, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 21, 18, 34, 5, 0, time.UTC),
		Direction:   "*out",
		Category:    "0",
		Tenant:      "test",
		Subject:     "rif",
		Account:     "post",
		Destination: "0723",
	}
	cc, err := cd.MaxDebit()
	//utils.PrintFull(cc)
	if err != nil || cc.Cost.String() != "2.5" {
		t.Errorf("Debit from empty share error: %+v, %v", cc, err)
	}
	acc, _ := cd.getAccount()
	//utils.PrintFull(acc)
	balanceMap := acc.BalanceMap[utils.MONETARY]
	if len(balanceMap) != 1 || balanceMap[0].GetValue().String() != "-2.5" {
		t.Errorf("Error debiting from empty share: %+v", balanceMap[0].GetValue())
	}
	cc, err = cd.MaxDebit()
	acc, _ = cd.getAccount()
	balanceMap = acc.BalanceMap[utils.MONETARY]
	//utils.LogFull(balanceMap)
	if err != nil || cc.Cost.String() != "2.5" {
		t.Errorf("Debit from empty share error: %+v, %v", cc, err)
	}
	if len(balanceMap) != 1 || balanceMap[0].GetValue().String() != "-5" {
		t.Errorf("Error debiting from empty share: %+v", balanceMap[0].GetValue())
	}
}

func TestMaxDebitZeroDefinedRate(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	ap.SetParentActionPlan()
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd1 := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "test",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "447956",
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 1, 0, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 0}
	cc, err := cd1.MaxDebit()
	if err != nil {
		t.Error("Error maxdebiting: ", err)
	}
	if cc.GetDuration() != 49*time.Second {
		t.Error("Error obtaining max debit duration: ", cc.GetDuration())
	}
	if cc.Cost.String() != "0.91" {
		t.Error("Error in max debit cost: ", utils.ToIJSON(cc))
	}
}

func TestMaxDebitForceDuration(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd1 := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "test",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "447956",
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 1, 40, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 0,
		ForceDuration: true,
	}
	_, err := cd1.MaxDebit()
	if err != utils.ErrInsufficientCredit {
		t.Fatal("Error forcing duration: ", err)
	}
}

func TestMaxDebitZeroDefinedRateOnlyMinutes(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd1 := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "test",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "447956",
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 0, 40, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 0}
	cc, err := cd1.MaxDebit()
	if err != nil {
		t.Fatal("Error maxdebiting: ", err)
	}
	if cc.GetDuration() != 40*time.Second {
		t.Error("Error obtaining max debit duration: ", cc.GetDuration())
	}
	if cc.Cost.String() != "0.01" {
		t.Error("Error in max debit cost: ", cc.Cost)
	}
}

func TestMaxDebitConsumesMinutes(t *testing.T) {
	ap, _ := ratingStorage.GetActionPlan("test", "TOPUP10_AT", utils.CACHED)
	for _, at := range ap.ActionTimings {
		if err := at.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	cd1 := &CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "test",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "447956",
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 0, 5, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 0}
	cd1.MaxDebit()
	if cd1.account.BalanceMap[utils.VOICE][0].GetValue().String() != "20" {
		t.Error("Error using minutes: ", cd1.account.BalanceMap[utils.VOICE][0].GetValue())
	}
}

func TestCDGetCostANY(t *testing.T) {
	cd1 := &CallDescriptor{
		Direction:   "*out",
		Category:    "data",
		Tenant:      "test",
		Subject:     "rif",
		Destination: utils.ANY,
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 0, 1, 0, time.UTC),
		TOR:         utils.DATA,
	}
	cc, err := cd1.GetCost()
	if err != nil || cc.Cost.String() != "60" {
		t.Errorf("Error getting *any dest: %+v %v", cc, err)
	}
}

func TestCDSplitInDataSlots(t *testing.T) {
	cd := &CallDescriptor{
		Direction:     "*out",
		Category:      "data",
		Tenant:        "test",
		Subject:       "rif",
		Destination:   utils.ANY,
		TimeStart:     time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, 3, 4, 6, 1, 5, 0, time.UTC),
		TOR:           utils.DATA,
		DurationIndex: 65 * time.Second,
	}
	if err := cd.LoadRatingPlans(); err != nil {
		t.Fatal(err)
	}
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.RatingInfos[0])
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestCDDataGetCost(t *testing.T) {
	cd := &CallDescriptor{
		Direction:   "*out",
		Category:    "data",
		Tenant:      "test",
		Subject:     "rif",
		Destination: utils.ANY,
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 1, 5, 0, time.UTC),
		TOR:         utils.DATA,
	}
	cc, err := cd.GetCost()
	if err != nil || cc.Cost.String() != "65" {
		t.Errorf("Error getting *any dest: %+v %v", cc, err)
	}
}

func TestCDRefundIncrements(t *testing.T) {
	ub := &Account{
		Tenant: "test",
		Name:   "ref",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{UUID: "moneya", Value: dec.NewFloat(100)},
			},
			utils.VOICE: Balances{
				&Balance{UUID: "minutea", Value: dec.NewFloat(10), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{UUID: "minuteb", Value: dec.NewFloat(10), DestinationIDs: utils.StringMap{"RET": true}},
			},
		},
	}
	if err := accountingStorage.SetAccount(ub); err != nil {
		t.Fatal(err)
	}
	increments := []*Increment{
		&Increment{CompressFactor: 1, Cost: dec.NewFloat(2), BalanceInfo: &DebitInfo{Monetary: &MonetaryInfo{UUID: "moneya"}, AccountID: ub.Name}},
		&Increment{CompressFactor: 1, Cost: dec.NewFloat(2), Duration: 3 * time.Second, BalanceInfo: &DebitInfo{Unit: &UnitInfo{UUID: "minutea"}, Monetary: &MonetaryInfo{UUID: "moneya"}, AccountID: ub.Name}},
		&Increment{CompressFactor: 1, Duration: 4 * time.Second, BalanceInfo: &DebitInfo{Unit: &UnitInfo{UUID: "minuteb"}, AccountID: ub.Name}},
	}
	cd := &CallDescriptor{Tenant: "test", TOR: utils.VOICE, Increments: increments}
	if err := cd.RefundIncrements(); err != nil {
		t.Fatal(err)
	}
	ub, _ = accountingStorage.GetAccount(ub.Tenant, ub.Name)
	if ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "104" ||
		ub.BalanceMap[utils.VOICE][0].GetValue().String() != "13" ||
		ub.BalanceMap[utils.VOICE][1].GetValue().String() != "14" {
		t.Error("Error refunding money: ", utils.ToIJSON(ub.BalanceMap))
	}
}

func TestCDRefundIncrementsZeroValue(t *testing.T) {
	ub := &Account{
		Tenant: "test",
		Name:   "ref",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{UUID: "moneya", Value: dec.NewFloat(100)},
			},
			utils.VOICE: Balances{
				&Balance{UUID: "minutea", Value: dec.NewFloat(10), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{UUID: "minuteb", Value: dec.NewFloat(10), DestinationIDs: utils.StringMap{"RET": true}},
			},
		},
	}
	accountingStorage.SetAccount(ub)
	increments := []*Increment{
		&Increment{Cost: dec.NewFloat(0), BalanceInfo: &DebitInfo{AccountID: ub.Name}},
		&Increment{Cost: dec.NewFloat(0), Duration: 3 * time.Second, BalanceInfo: &DebitInfo{AccountID: ub.Name}},
		&Increment{Cost: dec.NewFloat(0), Duration: 4 * time.Second, BalanceInfo: &DebitInfo{AccountID: ub.Name}},
	}
	cd := &CallDescriptor{TOR: utils.VOICE, Increments: increments}
	if err := cd.RefundIncrements(); err != nil {
		t.Fatal(err)
	}
	ub, _ = accountingStorage.GetAccount(ub.Tenant, ub.Name)
	if ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "100" ||
		ub.BalanceMap[utils.VOICE][0].GetValue().String() != "10" ||
		ub.BalanceMap[utils.VOICE][1].GetValue().String() != "10" {
		t.Error("Error refunding money: ", utils.ToIJSON(ub.BalanceMap))
	}
}

func TestCDMaxDebitPostAT(t *testing.T) {
	acc := &Account{
		Tenant: "test",
		Name:   "rif_uni",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					ID:    utils.META_DEFAULT,
					Value: dec.NewFloat(10.0),
				},
			},
		},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{
				&UnitCounter{
					//CounterType: utils.MONETARY,
					Counters: CounterFilters{
						&CounterFilter{
							Value:  dec.NewFloat(6),
							Filter: `{"Directions":{"$has":["*out"]}}`,
						},
					},
				},
			},
		},
		TriggerIDs: utils.NewStringMap("AT111"),
	}
	if err := ratingStorage.SetActionTriggers(&ActionTriggerGroup{
		Tenant: "test",
		Name:   "AT111",
		ActionTriggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:       "AT1-uuid",
				ThresholdType:  utils.TRIGGER_MAX_EVENT_COUNTER,
				ThresholdValue: dec.NewFloat(5),
				Filter:         `{"Directions":{"$has":["*out"]}}`,
				Weight:         10,
				ActionsID:      "TOPUP10_AC",
				parentGroup:    &ActionTriggerGroup{Tenant: "test"},
			},
		}}); err != nil {
		t.Fatal(err)
	}
	if err := accountingStorage.SetAccount(acc); err != nil {
		t.Fatal(err)
	}
	cd := &CallDescriptor{
		Direction:         utils.OUT,
		Category:          "call",
		Tenant:            "test",
		Subject:           "rif_uni",
		Destination:       "447956",
		TimeStart:         time.Date(2016, 8, 1, 19, 34, 0, 0, time.UTC),
		TimeEnd:           time.Date(2016, 8, 1, 19, 35, 0, 0, time.UTC),
		TOR:               utils.VOICE,
		PostActionTrigger: true,
	}
	cc, err := cd.MaxDebit()
	if err != nil {
		t.Fatal("Error on max debit: ", err)
	}
	if cc != nil &&
		len(cc.Timespans) != 1 ||
		cc.Timespans[0].Increments.Len() != 60 ||
		len(cc.Timespans[0].Increments.CompIncrement.PostATIDs) != 1 ||
		cc.Timespans[0].Increments.CompIncrement.PostATIDs["1"][0] != "AT1-uuid" {
		t.Error("Error max debiting with post ats: ", utils.ToIJSON(cc))
	}

	accAfter, err := accountingStorage.GetAccount(acc.Tenant, acc.Name)
	if err != nil {
		t.Error("Error getting after account: ", err)
	}
	if accAfter != nil &&
		len(accAfter.BalanceMap[utils.MONETARY]) != 1 ||
		len(accAfter.TriggerRecords) != 1 ||
		accAfter.TriggerRecords["AT1-uuid"].Executed != true {
		t.Error("Error action triggers executed: ", utils.ToIJSON(accAfter))
	}
	// execute again with atids set
	cd.ExeATIDs = map[string][]string{"rif_uni": []string{"AT1-uuid"}}
	cd.TimeEnd = time.Date(2016, 8, 1, 19, 34, 10, 0, time.UTC)
	cc, err = cd.MaxDebit()
	if err != nil {
		t.Error("Error on max debit: ", err)
	}
	if cc != nil &&
		len(cc.Timespans) != 1 ||
		cc.Timespans[0].Increments.Len() != 1 ||
		len(cc.Timespans[0].Increments.CompIncrement.PostATIDs) != 0 {
		t.Error("Error max debiting with post ats: ", utils.ToIJSON(cc))
	}

	accAfter, err = accountingStorage.GetAccount(acc.Tenant, acc.Name)
	if err != nil {
		t.Error("Error getting after account: ", err)
	}
	if accAfter != nil &&
		len(accAfter.BalanceMap[utils.MONETARY]) != 2 ||
		len(accAfter.TriggerRecords) != 1 ||
		accAfter.TriggerRecords["AT1-uuid"].Executed != true {
		t.Error("Error action triggers executed: ", utils.ToIJSON(accAfter))
	}

}

func TestParallelDebit(t *testing.T) {
	if err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "parallel",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewFloat(10000), Weight: 10},
			}},
	}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	iterations := 50
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cd := &CallDescriptor{
				Direction:   utils.OUT,
				Category:    "call",
				Tenant:      "test",
				Account:     "parallel",
				Subject:     "nt",
				Destination: "49",
				TimeStart:   time.Date(2017, time.February, 9, 9, 0, 0, 0, time.UTC),
				TimeEnd:     time.Date(2017, time.February, 9, 9, 0, 59, 0, time.UTC),
				LoopIndex:   0,
			}
			_, err := cd.Debit()
			if err != nil {
				t.Errorf("Error debiting balance: %s", err)
			}
		}()
	}
	wg.Wait()

	acc, err := accountingStorage.GetAccount("test", "parallel")
	if err != nil {
		t.Errorf("Error getting acc: %v", err)
	}

	expected := dec.NewFloat(float64(10000 - (iterations * 60)))
	if acc.BalanceMap[utils.MONETARY][0].GetValue().Cmp(expected) != 0 {
		t.Log("Account: ", utils.ToIJSON(acc))
		t.Errorf("got: %v, expected %v", acc.BalanceMap[utils.MONETARY][0].GetValue(), expected)
	}

}

func TestParallelMaxDebit(t *testing.T) {
	if err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "parallel",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewFloat(10000), Weight: 10},
			}},
	}); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	iterations := 50
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cd := &CallDescriptor{
				Direction:   utils.OUT,
				Category:    "call",
				Tenant:      "test",
				Account:     "parallel",
				Subject:     "nt",
				Destination: "49",
				TimeStart:   time.Date(2017, time.February, 9, 9, 0, 0, 0, time.UTC),
				TimeEnd:     time.Date(2017, time.February, 9, 9, 0, 59, 0, time.UTC),
				LoopIndex:   0,
			}
			_, err := cd.MaxDebit()
			if err != nil {
				t.Errorf("Error debiting balance: %s", err)
			}
		}()
	}
	wg.Wait()

	acc, err := accountingStorage.GetAccount("test", "parallel")
	if err != nil {
		t.Errorf("Error getting acc: %v", err)
	}

	expected := dec.NewFloat(float64(10000 - (iterations * 60)))
	if acc.BalanceMap[utils.MONETARY][0].GetValue().Cmp(expected) != 0 {
		t.Log("Account: ", utils.ToIJSON(acc))
		t.Errorf("got: %v, expected %v", acc.BalanceMap[utils.MONETARY][0].GetValue(), expected)
	}

}

func TestSerialDebit(t *testing.T) {
	if err := accountingStorage.SetAccount(&Account{
		Tenant: "test",
		Name:   "parallel",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Value: dec.NewFloat(10000), Weight: 10},
			}},
	}); err != nil {
		t.Fatal(err)
	}

	iterations := 50
	for i := 0; i < iterations; i++ {
		cd := &CallDescriptor{
			Direction:   utils.OUT,
			Category:    "call",
			Tenant:      "test",
			Account:     "parallel",
			Subject:     "nt",
			Destination: "49",
			TimeStart:   time.Date(2017, time.February, 9, 17, 30, 0, 0, time.UTC),
			TimeEnd:     time.Date(2017, time.February, 9, 17, 30, 59, 0, time.UTC),
			LoopIndex:   0,
		}
		_, err := cd.Debit()
		if err != nil {
			t.Errorf("Error debiting balance: %s", err)
		}
	}

	acc, err := accountingStorage.GetAccount("test", "parallel")
	if err != nil {
		t.Errorf("Error debiting balance: %v", err)
	}
	expected := dec.NewFloat(float64(10000 - (iterations * 60)))
	if acc.BalanceMap[utils.MONETARY][0].GetValue().Cmp(expected) != 0 {
		t.Log("Account: ", utils.ToIJSON(acc))
		t.Errorf("got: %v, expected %v", acc.BalanceMap[utils.MONETARY][0].GetValue(), expected)
	}
}

/*************** BENCHMARKS ********************/
func BenchmarkStorageGetting(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ratingStorage.GetRatingProfile(cd.Direction, cd.Tenant, cd.Category, cd.Subject, false, utils.CACHED)
	}
}

func BenchmarkStorageRestoring(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.LoadRatingPlans()
	}
}

func BenchmarkStorageGetCost(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost()
	}
}

func BenchmarkSplitting(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cd.LoadRatingPlans()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.splitInTimeSpans()
	}
}

func BenchmarkStorageSingleGetSessionTime(b *testing.B) {
	b.StopTimer()
	cd := &CallDescriptor{Tenant: "test", Subject: "minutosu", Destination: "0723"}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionDuration()
	}
}

func BenchmarkStorageMultipleGetSessionTime(b *testing.B) {
	b.StopTimer()
	cd := &CallDescriptor{Direction: "*out", Category: "0", Tenant: "test", Subject: "minutosu", Destination: "0723"}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionDuration()
	}
}
