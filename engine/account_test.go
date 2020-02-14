package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

var (
//NAT = &Destination{Id: "NAT", Prefixes: []string{"0257", "0256", "0723"}}
//RET = &Destination{Id: "RET", Prefixes: []string{"0723", "0724"}}
)

func TestAccountStorageStoreRestore(t *testing.T) {
	b1 := &Balance{Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: dec.NewVal(100, 0), Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	rifsBalance := &Account{Tenant: "x", Name: "other", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}}}}
	if err := accountingStorage.SetAccount(rifsBalance); err != nil {
		t.Fatal(err)
	}
	ub1, err := accountingStorage.GetAccount("x", "other")
	if err != nil || !ub1.BalanceMap[utils.MONETARY].Equal(rifsBalance.BalanceMap[utils.MONETARY]) {
		t.Log("UB: ", ub1)
		t.Errorf("Expected %v was %v", rifsBalance, ub1)
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &Balance{Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: dec.NewVal(100, 0), Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	ub1 := &Account{Tenant: "test", Name: "rif", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: dec.NewVal(200, 0)}}}}
	cd := &CallDescriptor{
		Category:      "0",
		Tenant:        "test",
		TimeStart:     time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:       time.Date(2013, 10, 4, 15, 46, 10, 0, time.UTC),
		LoopIndex:     0,
		DurationIndex: 10 * time.Second,

		Direction:   utils.OUT,
		Destination: "0723",
		TOR:         utils.VOICE,
	}
	seconds, credit, bucketList := ub1.getCreditForPrefix(cd)
	expected := 110 * time.Second
	if credit.String() != "200" || seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Log(seconds, credit, bucketList)
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetSpecialPricedSeconds(t *testing.T) {
	b1 := &Balance{Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "minu"}
	b2 := &Balance{Value: dec.NewVal(100, 0), Weight: 20, DestinationIDs: utils.StringMap{"RET": true}, RatingSubject: "minu"}

	ub1 := &Account{
		Tenant: "test",
		Name:   "rif",
		BalanceMap: map[string]Balances{
			utils.VOICE:    Balances{b1, b2},
			utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}},
		},
	}
	cd := &CallDescriptor{
		Category:    "0",
		Tenant:      "test",
		TimeStart:   time.Date(2013, 10, 4, 15, 46, 0, 0, time.UTC),
		TimeEnd:     time.Date(2013, 10, 4, 15, 46, 60, 0, time.UTC),
		LoopIndex:   0,
		Direction:   utils.OUT,
		Destination: "0723",
		TOR:         utils.VOICE,
	}
	seconds, credit, bucketList := ub1.getCreditForPrefix(cd)
	expected := 20 * time.Second
	if credit.String() != "0" || seconds != expected || len(bucketList) != 2 || bucketList[0].Weight < bucketList[1].Weight {
		t.Errorf("Expected %v was %v, \n %v \n, %s", expected, seconds, credit, utils.ToIJSON(bucketList))
	}
}

func TestAccountStorageStore(t *testing.T) {
	//FIXME: mongo will have a problem with null and {} so the Equal will not work
	b1 := &Balance{Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: dec.NewVal(100, 0), Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}}}}
	accountingStorage.SetAccount(rifsBalance)
	result, err := accountingStorage.GetAccount(rifsBalance.Tenant, rifsBalance.Name)
	if err != nil || rifsBalance.Name != result.Name ||
		len(rifsBalance.BalanceMap[utils.VOICE]) < 2 || len(result.BalanceMap[utils.VOICE]) < 2 ||
		!(rifsBalance.BalanceMap[utils.VOICE][0].Value.String() == result.BalanceMap[utils.VOICE][0].Value.String()) ||
		!(rifsBalance.BalanceMap[utils.VOICE][1].Value.String() == result.BalanceMap[utils.VOICE][1].Value.String()) ||
		!(rifsBalance.BalanceMap[utils.MONETARY][0].Value.String() == result.BalanceMap[utils.MONETARY][0].Value.String()) {
		t.Errorf("Expected %s was %s", utils.ToIJSON(rifsBalance), utils.ToIJSON(result))
	}
}

func TestDebitCreditZeroSecond(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Direction:    utils.OUT,
		Destination:  "0723045326",
		Category:     "0",
		TOR:          utils.VOICE,
		testCallcost: cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1}, utils.MONETARY: Balances{&Balance{Categories: utils.NewStringMap("0"), Value: dec.NewVal(21, 0)}}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" {
		t.Logf("%+v", cc.Timespans[0])
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "0" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "21" {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.VOICE][0])
	}
}

func TestDebitCreditBlocker(t *testing.T) {
	b1 := &Balance{UUID: "testa", Value: dec.NewVal(1152, 4), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "passmonde", Blocker: true}
	b2 := &Balance{UUID: "*default", Value: dec.NewFloat(1.5), Weight: 0}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{ConnectFee: dec.NewFloat(0.15), Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewFloat(0.1), RateIncrement: time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
		TOR:              utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Direction:    utils.OUT,
		Destination:  "0723045326",
		Category:     "0",
		TOR:          utils.VOICE,
		testCallcost: cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{utils.MONETARY: Balances{b1, b2}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, true, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if len(cc.Timespans) != 0 {
		t.Error("Wrong call cost: ", utils.ToIJSON(cc))
	}
	if rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "0.1152" ||
		rifsBalance.BalanceMap[utils.MONETARY][1].GetValue().String() != "1.5" {
		t.Error("should not have touched the balances: ", utils.ToIJSON(rifsBalance.BalanceMap[utils.MONETARY]))
	}
}

func TestDebitFreeEmpty(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "112",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{ConnectFee: dec.NewVal(0, 0), Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(0, 0), RateIncrement: time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
		TOR:              utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Direction:    utils.OUT,
		Tenant:       "test",
		Subject:      "rif:from:tm",
		Destination:  "112",
		Category:     "0",
		TOR:          utils.VOICE,
		testCallcost: cc,
	}
	// empty account
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{utils.MONETARY: Balances{}}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, true, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if len(cc.Timespans) == 0 || cc.GetCost().String() != "0" {
		t.Error("Wrong call cost: ", utils.ToIJSON(cc))
	}
	if len(rifsBalance.BalanceMap[utils.MONETARY]) != 0 {
		t.Error("should not have touched the balances: ", utils.ToIJSON(rifsBalance.BalanceMap[utils.MONETARY]))
	}
}

func TestDebitCreditZeroMinute(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: dec.NewVal(70, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:    time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
		Direction:    utils.OUT,
		Destination:  "0723045326",
		Category:     "0",
		TOR:          utils.VOICE,
		testCallcost: cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE:    Balances{b1},
		utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	//t.Logf("%+v", cc.Timespans)
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "10" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "21" {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][0])
	}
}

func TestDebitCreditZeroMixedMinute(t *testing.T) {
	b1 := &Balance{UUID: "testm", Value: dec.NewVal(70, 0), Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	b2 := &Balance{UUID: "tests", Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.Timespans[0].GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE:    Balances{b1, b2},
		utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "tests" ||
		cc.Timespans[1].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans)
	}
	if rifsBalance.BalanceMap[utils.VOICE][1].GetValue().String() != "0" ||
		rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "10" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "21" {
		t.Logf("TS0: %+v", cc.Timespans[0])
		t.Logf("TS1: %+v", cc.Timespans[1])
		t.Errorf("Error extracting minutes from balance: %+v", rifsBalance.BalanceMap[utils.VOICE][1])
	}
}

func TestDebitCreditNoCredit(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: dec.NewVal(70, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		Tenant:        "test",
		Subject:       "",
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE: Balances{b1},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Showing no enough credit error: ", utils.ToIJSON(cc))
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "10" {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][0])
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != time.Minute {
		t.Error("Error truncating extra timespans: ", utils.ToIJSON(cc.Timespans))
	}
}

func TestDebitCreditHasCredit(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: dec.NewVal(70, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(1, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(1, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE:    Balances{b1},
		utils.MONETARY: Balances{&Balance{UUID: "moneya", Value: dec.NewVal(110, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "10" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "30" {
		t.Errorf("Error extracting minutes from balance: %v, %v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 3 || cc.Timespans[0].GetDuration() != time.Minute {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSplitMinutesMoney(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 20, 0, time.UTC),
				DurationIndex: 0,
				ratingInfo:    &RatingInfo{},
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(1, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE:    Balances{b1},
		utils.MONETARY: Balances{&Balance{UUID: "moneya", Value: dec.NewVal(50, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != 1*time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement.Duration)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "0" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "30" {
		t.Errorf("Error extracting minutes from balance: %v, %v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != 10*time.Second || cc.Timespans[1].GetDuration() != 20*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans[1].GetDuration())
	}
}

func TestDebitCreditMoreTimespans(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: dec.NewVal(150, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE: Balances{b1},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "30" {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][0])
	}
}

func TestDebitCreditMoreTimespansMixed(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: dec.NewVal(70, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	b2 := &Balance{UUID: "testa", Value: dec.NewVal(150, 0), Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1s"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
				Increments:    &Increments{},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
				Increments:    &Increments{},
			},
		},

		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE: Balances{b1, b2},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != time.Minute {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "10" ||
		rifsBalance.BalanceMap[utils.VOICE][1].GetValue().String() != "130" {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.VOICE][1], cc.Timespans[1])
	}
}

func TestDebitCreditNoConectFeeCredit(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: dec.NewVal(70, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "*zero1m"}
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{ConnectFee: dec.NewVal(10, 0.0), Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(1, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR:              utils.VOICE,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE: Balances{b1},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Error showing debiting balance error: ", err)
	}

	if len(cc.Timespans) != 1 || rifsBalance.BalanceMap[utils.MONETARY].GetTotalValue().String() != "0" {
		t.Error("Error cutting at no connect fee: ", rifsBalance.BalanceMap[utils.MONETARY])
	}
}

func TestDebitCreditMoneyOnly(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(1, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 10, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				DurationIndex: 10 * time.Second,
				ratingInfo:    &RatingInfo{},
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(1, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[1].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{UUID: "money", Value: dec.NewVal(50, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err == nil {
		t.Error("Missing noy enough credit error ")
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Monetary.UUID != "money" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != 10*time.Second {
		t.Logf("%+v", cc.Timespans[0].Increments)
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement.BalanceInfo)
	}
	if rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "0" {
		t.Error("Error extracting minutes from balance: ",
			rifsBalance.BalanceMap[utils.MONETARY][0])
	}
	if len(cc.Timespans) != 2 ||
		cc.Timespans[0].GetDuration() != 10*time.Second ||
		cc.Timespans[1].GetDuration() != 40*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSubjectMinutes(t *testing.T) {
	b1 := &Balance{UUID: "testb", Categories: utils.NewStringMap("0"), Value: dec.NewVal(250, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "minu"}
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(1, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR:              utils.VOICE,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      "0",
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE:    Balances{b1},
		utils.MONETARY: Balances{&Balance{UUID: "moneya", Value: dec.NewVal(350, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Monetary.UUID != "moneya" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != 10*time.Second {
		t.Errorf("Error setting balance id to increment: %+v", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue().String() != "180" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "280" {
		t.Errorf("Error extracting minutes from balance: %v, %v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != 70*time.Second {
		for _, ts := range cc.Timespans {
			t.Log(ts)
		}
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

func TestDebitCreditSubjectMoney(t *testing.T) {
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 10, 0, time.UTC),
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(1, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR:              utils.VOICE,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{UUID: "moneya", Value: dec.NewVal(75, 0), DestinationIDs: utils.StringMap{"NAT": true}, RatingSubject: "minu"}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Monetary.UUID != "moneya" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != 10*time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "5" {
		t.Errorf("Error extracting minutes from balance: %v",
			rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 1 || cc.Timespans[0].GetDuration() != 70*time.Second {
		t.Error("Error truncating extra timespans: ", cc.Timespans)
	}
}

/*func TestDebitCreditSubjectMixed(t *testing.T) {
	b1 := &Balance{UUID: "testb", Value: 40, Weight: 10, DestinationId: "NAT", RatingSubject: "minu"}
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 55, 0, time.UTC),
				DurationIndex: 55 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: 1, RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR:              utils.VOICE,
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost: cc,
	}
	rifsBalance := &Account{Tenant:"test", Name: "other", BalanceMap: map[string]Balances{
		utils.VOICE: Balances{b1},
		utils.MONETARY:  Balances{&Balance{Uuid: "moneya", Value: 19500, RatingSubject: "minu"}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testb" ||
		cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Money.UUID != "moneya" ||
		cc.Timespans[0].Increments.CompIncrement.Duration != 10*time.Second {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.VOICE][0].GetValue() != 0 ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue() != 7 {
		t.Errorf("Error extracting minutes from balance: %+v, %+v",
			rifsBalance.BalanceMap[utils.VOICE][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
	if len(cc.Timespans) != 2 || cc.Timespans[0].GetDuration() != 40*time.Second {
		for _, ts := range cc.Timespans {
			t.Log(ts)
		}
		t.Error("Error truncating extra timespans: ", len(cc.Timespans), cc.Timespans[0].GetDuration())
	}
}*/

func TestAccountdebitBalance(t *testing.T) {
	ub := &Account{
		Name:          "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.SMS:  Balances{&Balance{Value: dec.NewVal(14, 0)}},
			utils.DATA: Balances{&Balance{Value: dec.NewVal(1204, 0)}},
			utils.VOICE: Balances{
				&Balance{Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}},
				&Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}},
			},
		},
	}
	a := &Action{
		TOR:     utils.VOICE,
		Filter1: `{"DestinationIDs":{"$in":["NEW"]}, "Directions":{"$in":["*out"]}}`,
		Params:  `{"Balance":{"Value":"20", "DestinationIDs":["NEW"], "Directions":["*out"]}}`,
	}
	if err := ub.debitBalanceAction(a, false); err != nil {
		t.Error(err)
	}
	if len(ub.BalanceMap[utils.VOICE]) != 3 ||
		ub.BalanceMap[utils.VOICE][2].DestinationIDs["NEW"] != true ||
		ub.BalanceMap[utils.VOICE][2].Directions["*out"] != true ||
		ub.BalanceMap[utils.VOICE][2].Value.String() != "-20" {
		t.Errorf("Error adding minute bucket! %s", utils.ToIJSON(ub.BalanceMap[utils.VOICE]))
	}
}

func TestAccountdebitBalanceExists(t *testing.T) {

	ub := &Account{
		Name:          "rif",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.SMS:  Balances{&Balance{Value: dec.NewVal(14, 0)}},
			utils.DATA: Balances{&Balance{Value: dec.NewVal(1024, 0)}},
			utils.VOICE: Balances{
				&Balance{Value: dec.NewVal(15, 0), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.NewStringMap(utils.OUT)},
				&Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}},
			},
		},
	}
	a := &Action{
		TOR:     utils.VOICE,
		Filter1: `{"DestinationIDs":{"$in":["NAT"]}, "Directions":{"$in":["*out"]}, "Weight":20}`,
		Params:  `{"Balance":{"Value":-10, "DestinationIDs":["NAT"], "Directions":["*out"], "Weight":20}}`,
	}
	if err := ub.debitBalanceAction(a, false); err != nil {
		t.Fatal(err)
	}
	if len(ub.BalanceMap[utils.VOICE]) != 2 || ub.BalanceMap[utils.VOICE][0].GetValue().String() != "25" {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinuteNil(t *testing.T) {
	ub := &Account{
		Name:          "rif",
		AllowNegative: true,
		BalanceMap:    map[string]Balances{utils.SMS: Balances{&Balance{Value: dec.NewVal(14, 0)}}, utils.DATA: Balances{&Balance{Value: dec.NewVal(1024, 0)}}, utils.VOICE: Balances{&Balance{Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}}, &Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
	}
	ub.debitBalanceAction(nil, false)
	if len(ub.BalanceMap[utils.VOICE]) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestAccountAddMinutBucketEmpty(t *testing.T) {
	ub := &Account{}
	a := &Action{
		TOR:     utils.VOICE,
		Filter1: `{"DestinationIDs":{"$in":["NAT"]}, "Directions":{"$in":["*out"]}}`,
		Params:  `{"Balance": {"Value":-10, "DestinationIDs":["NAT"], "Directions":["*out"]}}`,
	}
	if err := ub.debitBalanceAction(a, false); err != nil {
		t.Fatal(err)
	}
	if len(ub.BalanceMap[utils.VOICE]) != 1 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
	a = &Action{
		TOR:     utils.VOICE,
		Filter1: `{"DestinationIDs":{"$in":["NAT"]}, "Directions":{"$in":["*out"]}}`,
		Params:  `{"Balance": {"Value":-10, "DestinationIDs":["NAT"], "Directions":["*out"]}}`,
	}
	ub.debitBalanceAction(a, false)
	if len(ub.BalanceMap[utils.VOICE]) != 1 || ub.BalanceMap[utils.VOICE][0].GetValue().String() != "20" {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
	a = &Action{
		TOR:     utils.VOICE,
		Filter1: `{"DestinationIDs":{"$in":["OTHER"]}, "Directions":{"$in":["*out"]}}`,
		Params:  `{"Balance": {"Value":-10, "DestinationIDs":["OTHER"], "Directions":["*out"]}}`,
	}
	ub.debitBalanceAction(a, false)
	if len(ub.BalanceMap[utils.VOICE]) != 2 {
		t.Error("Error adding minute bucket: ", ub.BalanceMap[utils.VOICE])
	}
}

func TestAccountExecuteTriggeredActions(t *testing.T) {
	ub := &Account{
		Name: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(100, 0)}},
			utils.VOICE: Balances{
				&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}},
				&Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		UnitCounters: UnitCounters{
			utils.MONETARY: []*UnitCounter{
				&UnitCounter{
					Counters: CounterFilters{
						&CounterFilter{Value: dec.NewVal(1, 0), Filter: `{"Directions": {"$in":["*out"]}}`},
					},
				},
			},
		},
		triggers: ActionTriggers{&ActionTrigger{TOR: utils.MONETARY, Filter: `{"Directions": {"$in":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ActionsID: "TEST_ACTIONS", parentGroup: &ActionTriggerGroup{Tenant: "test"}}},
	}
	ub.InitTriggerRecords()
	ub.countUnits(dec.NewVal(1, 0), utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "110" || ub.BalanceMap[utils.VOICE][0].GetValue().String() != "20" {
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue())
	}
	// are set to executed
	ub.countUnits(dec.NewVal(1, 0), utils.MONETARY, nil, nil)
	if ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "110" || ub.BalanceMap[utils.VOICE][0].GetValue().String() != "20" {
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue())
	}
	// we can reset them
	ub.ResetActionTriggers(nil)
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, nil, nil)
	if ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "120" || ub.BalanceMap[utils.VOICE][0].GetValue().String() != "30" {
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue())
	}
}

func TestAccountExecuteTriggeredActionsBalance(t *testing.T) {
	ub := &Account{
		Name:         "TEST_UB",
		BalanceMap:   map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}}, &Balance{Directions: utils.NewStringMap(utils.OUT), Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Filter: `{"Directions": {"$has":["*out"]}}`, Value: dec.NewVal(1, 0)}}}}},
		triggers:     ActionTriggers{&ActionTrigger{TOR: utils.MONETARY, Filter: `{"Directions": {"$has":["*out"]}}`, ThresholdValue: dec.NewVal(100, 0), ThresholdType: utils.TRIGGER_MIN_EVENT_COUNTER, ActionsID: "TEST_ACTIONS", parentGroup: &ActionTriggerGroup{Tenant: "test"}}},
	}
	ub.InitTriggerRecords()
	ub.countUnits(dec.NewVal(1, 0), utils.MONETARY, nil, nil)
	if ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "110" || ub.BalanceMap[utils.VOICE][0].GetValue().String() != "20" {
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue(), len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestAccountExecuteTriggeredActionsOrder(t *testing.T) {
	ub := &Account{
		Name:         "TEST_UB_OREDER",
		BalanceMap:   map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(100, 0)}}},
		UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(1, 0), Filter: `{"Directions": {"$has":["*out"]}}`}}}}},
		triggers:     ActionTriggers{&ActionTrigger{TOR: utils.MONETARY, Filter: `{"Directions": {"$has":["*out"]}}`, ThresholdValue: dec.NewVal(2, 0), ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ActionsID: "TEST_ACTIONS_ORDER", parentGroup: &ActionTriggerGroup{Tenant: "test"}}},
	}
	ub.InitTriggerRecords()
	ub.countUnits(dec.NewVal(1, 0), utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.BalanceMap[utils.MONETARY]) != 1 || ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "10" {

		t.Errorf("Error executing triggered actions in order %v", ub.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestAccountExecuteTriggeredDayWeek(t *testing.T) {
	ub := &Account{
		Name:       "TEST_UB",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}, &Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		triggers: ActionTriggers{
			&ActionTrigger{UniqueID: "day_trigger", TOR: utils.MONETARY, Filter: `{"Directions": {"$has":["*out"]}}`, ThresholdValue: dec.NewVal(10, 0), ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ActionsID: "TEST_ACTIONS", parentGroup: &ActionTriggerGroup{Tenant: "test"}},
			&ActionTrigger{UniqueID: "week_trigger", TOR: utils.MONETARY, Filter: `{"Directions": {"$has":["*out"]}}`, ThresholdValue: dec.NewVal(100, 0), ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER, ActionsID: "TEST_ACTIONS", parentGroup: &ActionTriggerGroup{Tenant: "test"}},
		},
	}
	ub.InitTriggerRecords()
	ub.InitCounters()
	if len(ub.UnitCounters) != 1 || len(ub.UnitCounters[utils.MONETARY][0].Counters) != 2 {
		t.Error("Error initializing counters: ", ub.UnitCounters[utils.MONETARY][0].Counters[0])
	}

	ub.countUnits(dec.NewVal(1, 0), utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "1" ||
		ub.UnitCounters[utils.MONETARY][0].Counters[1].Value.String() != "1" {
		t.Error("Error incrementing both counters", ub.UnitCounters[utils.MONETARY][0].Counters[0].Value, ub.UnitCounters[utils.MONETARY][0].Counters[1].Value)
	}

	// we can reset them
	if err := resetCountersAction(ub, nil, &Action{TOR: utils.MONETARY, Filter1: `{"UniqueID":"day_trigger"}`}, nil); err != nil {
		t.Fatal(err)
	}
	if ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "0" ||
		ub.UnitCounters[utils.MONETARY][0].Counters[1].Value.String() != "1" {
		t.Error("Error reseting both counters", ub.UnitCounters[utils.MONETARY][0].Counters[0].Value, ub.UnitCounters[utils.MONETARY][0].Counters[1].Value)
	}
}

func TestAccountExpActionTrigger(t *testing.T) {
	ub := &Account{
		Name:       "TEST_UB",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(100, 0), ExpirationDate: time.Date(2015, time.November, 9, 9, 48, 0, 0, time.UTC)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}, &Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		triggers: ActionTriggers{
			&ActionTrigger{UniqueID: "check expired balances", TOR: utils.MONETARY, Filter: `{"Directions": {"$has":["*out"]}}`, ThresholdValue: dec.NewVal(10, 0), ThresholdType: utils.TRIGGER_BALANCE_EXPIRED, ActionsID: "TEST_ACTIONS", parentGroup: &ActionTriggerGroup{Tenant: "test"}},
		},
	}
	ub.InitTriggerRecords()
	ub.ExecuteActionTriggers(nil, false)
	if ub.BalanceMap[utils.MONETARY][0].IsExpired() ||
		ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "10" || // expired was cleaned
		ub.BalanceMap[utils.VOICE][0].GetValue().String() != "20" ||
		ub.TriggerRecords["check expired balances"].Executed != true {
		t.Log(ub.BalanceMap[utils.MONETARY][0].IsExpired())
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue(), len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestAccountExpActionTriggerNotActivated(t *testing.T) {
	ub := &Account{
		Name:       "TEST_UB",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(100, 0)}}, utils.VOICE: Balances{&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}}, &Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		triggers: ActionTriggers{
			&ActionTrigger{UniqueID: "check expired balances", ActivationDate: time.Date(2116, 2, 5, 18, 0, 0, 0, time.UTC), TOR: utils.MONETARY, Filter: `{"Directions": {"$has":["*out"]}}`, ThresholdValue: dec.NewVal(10, 0), ThresholdType: utils.TRIGGER_BALANCE_EXPIRED, ActionsID: "TEST_ACTIONS", parentGroup: &ActionTriggerGroup{Tenant: "test"}},
		},
	}
	ub.InitTriggerRecords()
	ub.ExecuteActionTriggers(nil, false)
	if ub.BalanceMap[utils.MONETARY][0].IsExpired() ||
		ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "100" ||
		ub.BalanceMap[utils.VOICE][0].GetValue().String() != "10" ||
		ub.TriggerRecords["check expired balances"].Executed != false {
		t.Log(ub.BalanceMap[utils.MONETARY][0].IsExpired())
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue(), len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestAccountExpActionTriggerExpired(t *testing.T) {
	ub := &Account{
		Name: "TEST_UB",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{Directions: utils.NewStringMap(utils.OUT), Value: dec.NewVal(100, 0)}},
			utils.VOICE: Balances{
				&Balance{Value: dec.NewVal(10, 0), Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Directions: utils.StringMap{utils.OUT: true}},
				&Balance{Weight: 10, DestinationIDs: utils.StringMap{"RET": true}}}},
		triggers: ActionTriggers{
			&ActionTrigger{UniqueID: "check expired balances", ExpirationDate: time.Date(2016, 2, 4, 18, 0, 0, 0, time.UTC), TOR: utils.MONETARY, Filter: `{"Directions": {"$has":["*out"]}}`, ThresholdValue: dec.NewVal(10, 0), ThresholdType: utils.TRIGGER_BALANCE_EXPIRED, ActionsID: "TEST_ACTIONS"},
		},
	}
	ub.InitTriggerRecords()
	ub.ExecuteActionTriggers(nil, false)
	if ub.BalanceMap[utils.MONETARY][0].IsExpired() ||
		ub.BalanceMap[utils.MONETARY][0].GetValue().String() != "100" ||
		ub.BalanceMap[utils.VOICE][0].GetValue().String() != "10" {
		t.Log(ub.BalanceMap[utils.MONETARY][0].IsExpired())
		t.Error("Error executing triggered actions", ub.BalanceMap[utils.MONETARY][0].GetValue(), ub.BalanceMap[utils.VOICE][0].GetValue(), len(ub.BalanceMap[utils.MONETARY]))
	}
}

func TestCleanExpired(t *testing.T) {
	ub := &Account{
		Name: "TEST_UB_OREDER",
		BalanceMap: map[string]Balances{utils.MONETARY: Balances{
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)}}, utils.VOICE: Balances{
			&Balance{ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC)},
			&Balance{ExpirationDate: time.Now().Add(10 * time.Second)},
		}},
		triggers: ActionTriggers{
			&ActionTrigger{
				ExpirationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC),
			},
			&ActionTrigger{
				ActivationDate: time.Date(2013, 7, 18, 14, 33, 0, 0, time.UTC),
			},
		},
	}
	ub.CleanExpiredStuff()
	if len(ub.BalanceMap[utils.MONETARY]) != 2 {
		t.Error("Error cleaning expired balances!")
	}
	if len(ub.BalanceMap[utils.VOICE]) != 1 {
		t.Error("Error cleaning expired minute buckets!")
	}
}

func TestAccountUnitCounting(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(0, 0)}}}}}}
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, &CallCost{}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "10" {
		t.Error("Error counting units")
	}
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, &CallCost{}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "20" {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutbound(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(0, 0), Filter: `{"Directions": {"$has":["*out"]}}`}}}}}}
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "10" {
		t.Error("Error counting units")
	}
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "20" {
		t.Error("Error counting units")
	}
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "30" {
		t.Error("Error counting units")
	}
}

func TestAccountUnitCountingOutboundInbound(t *testing.T) {
	ub := &Account{UnitCounters: UnitCounters{utils.MONETARY: []*UnitCounter{&UnitCounter{Counters: CounterFilters{&CounterFilter{Value: dec.NewVal(0, 0), Filter: `{"Directions": {"$has":["*out"]}}`}}}}}}
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "10" {
		t.Errorf("Error counting units: %+v", ub.UnitCounters[utils.MONETARY][0].Counters[0])
	}
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, &CallCost{Direction: utils.OUT}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "20" {
		t.Error("Error counting units")
	}
	ub.countUnits(dec.NewVal(10, 0), utils.MONETARY, &CallCost{Direction: utils.IN}, nil)
	if len(ub.UnitCounters[utils.MONETARY]) != 1 || ub.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "20" {
		t.Error("Error counting units")
	}
}

func TestDebitShared(t *testing.T) {
	cc := &CallCost{
		Tenant:      "vdf",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 0, 0, time.UTC),
				DurationIndex: 55 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(2, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rif := &Account{Tenant: "test", Name: "rif", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{UUID: "moneya", Value: dec.NewVal(0, 0), SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}
	groupie := &Account{Tenant: "test", Name: "groupie", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{UUID: "moneyc", Value: dec.NewVal(130, 0), SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}

	sg := &SharedGroup{Tenant: "test", Name: "SG_TEST", MemberIDs: utils.NewStringMap(rif.Name, groupie.Name), AccountParameters: map[string]*SharingParam{"*any": &SharingParam{Strategy: STRATEGY_MINE_RANDOM}}}

	if err := accountingStorage.SetAccount(groupie); err != nil {
		t.Fatal(err)
	}
	if err := ratingStorage.SetSharedGroup(sg); err != nil {
		t.Fatal(err)
	}
	cc, err := rif.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if rif.BalanceMap[utils.MONETARY][0].GetValue().String() != "0" {
		t.Errorf("Error debiting from shared group: %+v", rif.BalanceMap[utils.MONETARY][0])
	}
	groupie, err = accountingStorage.GetAccount("test", "groupie")
	if err != nil {
		t.Fatal(err)
	}
	if groupie.BalanceMap[utils.MONETARY][0].GetValue().String() != "10" {
		t.Errorf("Error debiting from shared group: %+v", groupie.BalanceMap[utils.MONETARY][0])
	}

	if len(cc.Timespans) != 1 {
		t.Errorf("Wrong number of timespans: %v", cc.Timespans)
	}
	if cc.Timespans[0].Increments.Len() != 6 {
		t.Errorf("Wrong number of increments: %v", cc.Timespans[0].Increments)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.AccountID != "groupie" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
}

func TestMaxDurationShared(t *testing.T) {
	cc := &CallCost{
		Tenant:      "test",
		Category:    "0",
		Direction:   utils.OUT,
		Destination: "4978",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 0, 0, time.UTC),
				DurationIndex: 55 * time.Second,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(2, 0), RateIncrement: 10 * time.Second, RateUnit: time.Second}}}},
			},
		},
		deductConnectFee: true,
	}
	cd := &CallDescriptor{
		Tenant:        cc.Tenant,
		Category:      cc.Category,
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rif := &Account{Tenant: "test", Name: "rif", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{UUID: "moneya", Value: dec.NewVal(0, 0), SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}
	groupie := &Account{Tenant: "test", Name: "groupie", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{UUID: "moneyc", Value: dec.NewVal(130, 0), SharedGroups: utils.NewStringMap("SG_TEST")}},
	}}

	sg := &SharedGroup{Tenant: "test", Name: "SG_TEST", MemberIDs: utils.NewStringMap(rif.Name, groupie.Name), AccountParameters: map[string]*SharingParam{"*any": &SharingParam{Strategy: STRATEGY_MINE_RANDOM}}}

	accountingStorage.SetAccount(groupie)
	ratingStorage.SetSharedGroup(sg)
	duration, err := cd.getMaxSessionDuration(rif)
	if err != nil {
		t.Error("Error getting max session duration from shared group: ", err)
	}
	if duration != 1*time.Minute {
		t.Error("Wrong max session from shared group: ", duration)
	}

}

func TestMaxDurationConnectFeeOnly(t *testing.T) {
	cd := &CallDescriptor{
		Tenant:        "test",
		Category:      "call",
		TimeStart:     time.Date(2015, 9, 24, 10, 48, 0, 0, time.UTC),
		TimeEnd:       time.Date(2015, 9, 24, 10, 58, 1, 0, time.UTC),
		Direction:     utils.OUT,
		Destination:   "4444",
		Subject:       "dy",
		Account:       "dy",
		TOR:           utils.VOICE,
		DurationIndex: 600,
	}
	rif := &Account{Tenant: "test", Name: "rif", BalanceMap: map[string]Balances{
		utils.MONETARY: Balances{&Balance{UUID: "moneya", Value: dec.NewFloat(0.2)}},
	}}

	duration, err := cd.getMaxSessionDuration(rif)
	if err != nil {
		t.Error("Error getting max session duration: ", err)
	}
	if duration != 0 {
		t.Error("Wrong max session: ", duration)
	}

}

func TestDebitSMS(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 1, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 1 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.SMS,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.SMS:      Balances{&Balance{UUID: "testm", Value: dec.NewVal(100, 0), Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}}},
		utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.SMS][0].GetValue().String() != "99" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "21" {
		t.Log(cc.Timespans[0].Increments)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.SMS][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitGeneric(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 1, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 1 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.GENERIC,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.GENERIC:  Balances{&Balance{UUID: "testm", Value: dec.NewVal(100, 0), Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}}},
		utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.GENERIC][0].GetValue().String() != "99" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "21" {
		t.Log(cc.Timespans[0].Increments)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.GENERIC][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitGenericBalance(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 30, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(100, 0), RateIncrement: 1 * time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.GENERIC:  Balances{&Balance{UUID: "testm", Value: dec.NewVal(100, 0), Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}, Factor: ValueFactor{utils.VOICE: 60.0}}},
		utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0].Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.GENERIC][0].GetValue().String() != "99.5" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "21" {
		t.Logf("%+v", cc.Timespans[0].Increments.CompIncrement)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.GENERIC][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitGenericBalanceWithRatingSubject(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 48, 30, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval:  &RateInterval{Rating: &RIRate{Rates: RateGroups{&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(0, 0), RateIncrement: time.Second, RateUnit: time.Second}}}},
			},
		},
		TOR: utils.VOICE,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.GENERIC:  Balances{&Balance{UUID: "testm", Value: dec.NewVal(100, 0), Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}, Factor: ValueFactor{utils.VOICE: 60.0}, RatingSubject: "free"}},
		utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if cc.Timespans[0].Increments.CompIncrement.BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", cc.Timespans[0])
	}
	if rifsBalance.BalanceMap[utils.GENERIC][0].GetValue().String() != "99.5" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "21" {
		t.Logf("%+v", cc.Timespans[0].Increments.CompIncrement)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.GENERIC][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitDataUnits(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(2, 0), RateIncrement: 1 * time.Second, RateUnit: time.Minute},
							&RateInfo{GroupIntervalStart: 60, Value: dec.NewVal(1, 0), RateIncrement: 1 * time.Second, RateUnit: time.Second},
						},
					},
				},
			},
		},
		TOR: utils.DATA,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.DATA:     Balances{&Balance{UUID: "testm", Value: dec.NewVal(100, 0), Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}}},
		utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	// test rating information
	ts := cc.Timespans[0]
	if ts.MatchedSubject != "testm" || ts.MatchedPrefix != "0723" || ts.MatchedDestID != "NAT" || ts.RatingPlanID != utils.META_NONE {
		t.Error("Error setting rating info: ", utils.ToIJSON(ts.ratingInfo))
	}
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if ts.Increments.CompIncrement.BalanceInfo.Unit.UUID != "testm" {
		t.Error("Error setting balance id to increment: ", ts.Increments.CompIncrement)
	}
	if rifsBalance.BalanceMap[utils.DATA][0].GetValue().String() != "20" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "21" {
		t.Log(ts.Increments)
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.DATA][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestDebitDataMoney(t *testing.T) {
	cc := &CallCost{
		Direction:   utils.OUT,
		Destination: "0723045326",
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart:     time.Date(2013, 9, 24, 10, 48, 0, 0, time.UTC),
				TimeEnd:       time.Date(2013, 9, 24, 10, 49, 20, 0, time.UTC),
				ratingInfo:    &RatingInfo{},
				DurationIndex: 0,
				RateInterval: &RateInterval{
					Rating: &RIRate{
						Rates: RateGroups{
							&RateInfo{GroupIntervalStart: 0, Value: dec.NewVal(2, 0), RateIncrement: time.Minute, RateUnit: time.Second},
						},
					},
				},
			},
		},
		TOR: utils.DATA,
	}
	cd := &CallDescriptor{
		TimeStart:     cc.Timespans[0].TimeStart,
		TimeEnd:       cc.Timespans[0].TimeEnd,
		Direction:     cc.Direction,
		Destination:   cc.Destination,
		TOR:           cc.TOR,
		DurationIndex: cc.GetDuration(),
		testCallcost:  cc,
	}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{
		utils.DATA:     Balances{&Balance{UUID: "testm", Value: dec.NewVal(0, 0), Weight: 5, DestinationIDs: utils.StringMap{"NAT": true}}},
		utils.MONETARY: Balances{&Balance{Value: dec.NewVal(160, 0)}},
	}}
	var err error
	cc, err = rifsBalance.debitCreditBalance(cd, false, false, true)
	if err != nil {
		t.Error("Error debiting balance: ", err)
	}
	if rifsBalance.BalanceMap[utils.DATA][0].GetValue().String() != "0" ||
		rifsBalance.BalanceMap[utils.MONETARY][0].GetValue().String() != "0" {
		t.Error("Error extracting minutes from balance: ", rifsBalance.BalanceMap[utils.DATA][0].GetValue(), rifsBalance.BalanceMap[utils.MONETARY][0].GetValue())
	}
}

func TestAccountGetDefaultMoneyBalanceEmpty(t *testing.T) {
	acc := &Account{}
	defBal := acc.GetDefaultMoneyBalance()
	if defBal == nil || len(acc.BalanceMap) != 1 || !defBal.IsDefault() {
		t.Errorf("Bad default money balance: %+v", defBal)
	}
}

func TestAccountGetDefaultMoneyBalance(t *testing.T) {
	acc := &Account{}
	acc.BalanceMap = make(map[string]Balances)
	tag := utils.MONETARY
	acc.BalanceMap[tag] = append(acc.BalanceMap[tag], &Balance{Weight: 10})
	defBal := acc.GetDefaultMoneyBalance()
	if defBal == nil || len(acc.BalanceMap[tag]) != 2 || !defBal.IsDefault() {
		t.Errorf("Bad default money balance: %+v", defBal)
	}
}

func TestAccountInitCounters(t *testing.T) {
	a := &Account{
		triggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
		},
	}
	a.InitCounters()
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MONETARY][0].Counters) != 2 ||
		len(a.UnitCounters[utils.VOICE][0].Counters) != 1 ||
		len(a.UnitCounters[utils.VOICE][1].Counters) != 1 ||
		len(a.UnitCounters[utils.SMS][0].Counters) != 1 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, c := range uc.Counters {
					t.Logf("B: %+v", c)
				}
			}
		}
		t.Errorf("Error Initializing unit counters: %v", len(a.UnitCounters))
	}
}

func TestAccountDoubleInitCounters(t *testing.T) {
	a := &Account{
		triggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$has":["*out", "*in"]}, "Weight":10}`,
			},
		},
	}
	a.InitCounters()
	a.InitCounters()
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MONETARY][0].Counters) != 2 ||
		len(a.UnitCounters[utils.VOICE][0].Counters) != 1 ||
		len(a.UnitCounters[utils.VOICE][1].Counters) != 1 ||
		len(a.UnitCounters[utils.SMS][0].Counters) != 1 {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, c := range uc.Counters {
					t.Logf("B: %+v", c)
				}
			}
		}
		t.Errorf("Error Initializing unit counters: %v", len(a.UnitCounters))
	}
}

func TestAccountGetBalancesForPrefixMixed(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value:          dec.NewVal(10, 0),
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.OUT, utils.MONETARY, "")
	if len(bcs) != 0 {
		t.Error("error excluding on mixed balances")
	}
}

func TestAccountGetBalancesForPrefixAllExcl(t *testing.T) {
	acc := &Account{
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value:          dec.NewVal(10, 0),
					DestinationIDs: utils.StringMap{"NAT": false, "RET": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.OUT, utils.MONETARY, "")
	if len(bcs) == 0 {
		t.Error("error finding balance on all excluded")
	}
}

func TestAccountGetBalancesForPrefixMixedGood(t *testing.T) {
	acc := &Account{
		Tenant: "test",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value:          dec.NewVal(10, 0),
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false, "EXOTIC": true},
				},
			},
		},
	}

	bcs := acc.getBalancesForPrefix("999123", "", utils.OUT, utils.MONETARY, "")
	if len(bcs) == 0 {
		t.Error("error finding on mixed balances good: ", utils.ToIJSON(bcs))
	}
}

func TestAccountGetBalancesForPrefixMixedBad(t *testing.T) {
	acc := &Account{
		Tenant: "test",
		BalanceMap: map[string]Balances{
			utils.MONETARY: Balances{
				&Balance{
					Value:          dec.NewVal(10, 0),
					DestinationIDs: utils.StringMap{"NAT": true, "RET": false, "EXOTIC": false},
				},
			},
		},
	}
	bcs := acc.getBalancesForPrefix("999123", "", utils.OUT, utils.MONETARY, "")
	if len(bcs) != 0 {
		t.Error("error excluding on mixed balances bad")
	}
}

func TestAccountNewAccountSummaryFromJSON(t *testing.T) {
	if acnt, err := NewAccountSummaryFromJSON("null"); err != nil {
		t.Error(err)
	} else if acnt != nil {
		t.Errorf("Expecting nil, received: %+v", acnt)
	}
}

func TestAccountAsAccountDigest(t *testing.T) {
	acnt1 := &Account{
		Tenant:        "test",
		Name:          "account1",
		AllowNegative: true,
		BalanceMap: map[string]Balances{
			utils.SMS:  Balances{&Balance{ID: "sms1", Value: dec.NewVal(14, 0)}},
			utils.DATA: Balances{&Balance{ID: "data1", Value: dec.NewVal(1204, 0)}},
			utils.VOICE: Balances{
				&Balance{ID: "voice1", Weight: 20, DestinationIDs: utils.StringMap{"NAT": true}, Value: dec.NewVal(3600, 0)},
				&Balance{ID: "voice2", Weight: 10, DestinationIDs: utils.StringMap{"RET": true}, Value: dec.NewVal(1200, 0)},
			},
		},
	}
	expectacntSummary := &AccountSummary{
		Tenant: "test",
		ID:     "account1",
		BalanceSummaries: []*BalanceSummary{
			&BalanceSummary{ID: "sms1", Type: utils.SMS, Value: "14", Disabled: false},
			&BalanceSummary{ID: "data1", Type: utils.DATA, Value: "1204", Disabled: false},
			&BalanceSummary{ID: "voice1", Type: utils.VOICE, Value: "1204", Disabled: false},
			&BalanceSummary{ID: "voice2", Type: utils.VOICE, Value: "1200", Disabled: false},
		},
		AllowNegative: true,
		Disabled:      false,
	}
	acntSummary := acnt1.AsAccountSummary()
	if expectacntSummary.Tenant != acntSummary.Tenant ||
		expectacntSummary.ID != acntSummary.ID ||
		expectacntSummary.AllowNegative != acntSummary.AllowNegative ||
		expectacntSummary.Disabled != acntSummary.Disabled ||
		len(expectacntSummary.BalanceSummaries) != len(acntSummary.BalanceSummaries) {
		t.Errorf("Expecting: %+v, received: %+v", expectacntSummary, acntSummary)
	}
	// Since maps are unordered, slices will be too so we need to find element to compare
	for _, bd := range acntSummary.BalanceSummaries {
		if bd.ID == "sms1" && !reflect.DeepEqual(expectacntSummary.BalanceSummaries[0], bd) {
			t.Errorf("Expecting: %+v, received: %+v", expectacntSummary, acntSummary)
		}
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &Balance{Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: dec.NewVal(100, 0), Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}

	ub1 := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}

func BenchmarkAccountStorageStoreRestore(b *testing.B) {
	b1 := &Balance{Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: dec.NewVal(100, 0), Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	rifsBalance := &Account{Tenant: "test", Name: "other", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}}}}
	for i := 0; i < b.N; i++ {
		accountingStorage.SetAccount(rifsBalance)
		accountingStorage.GetAccount(rifsBalance.Tenant, rifsBalance.Name)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &Balance{Value: dec.NewVal(10, 0), Weight: 10, DestinationIDs: utils.StringMap{"NAT": true}}
	b2 := &Balance{Value: dec.NewVal(100, 0), Weight: 20, DestinationIDs: utils.StringMap{"RET": true}}
	ub1 := &Account{Tenant: "CUSTOMER_1", Name: "rif", BalanceMap: map[string]Balances{utils.VOICE: Balances{b1, b2}, utils.MONETARY: Balances{&Balance{Value: dec.NewVal(21, 0)}}}}
	cd := &CallDescriptor{
		Destination: "0723",
	}
	for i := 0; i < b.N; i++ {
		ub1.getCreditForPrefix(cd)
	}
}
