package engine

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"
	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

var rsponder *Responder

// Test internal abilites of GetDerivedChargers
func TestResponderGetDerivedChargers(t *testing.T) {
	cfgedDC := &utils.DerivedChargerGroup{
		Direction: utils.OUT, Tenant: "hopa", Category: utils.ANY, Account: utils.ANY, Subject: utils.ANY,
		DestinationIDs: utils.StringMap{},
		Chargers: []*utils.DerivedCharger{
			&utils.DerivedCharger{
				RunID:  "responder1",
				Fields: `{Direction": {"$set":"test"}, "Tenant": {"$set":"test"}, "Category": {"$set":"test"}, "Account": {"$set":"test"}, "Subject": {"$set":"test"}, "Destination": {"$set":"test"}, "SetupTime": {"$set":"test"}, "AnswerTime": {"$set":"test"}, "Usage": {"$set":"test"}}`,
			},
		},
	}
	rsponder = &Responder{}
	if err := ratingStorage.SetDerivedChargers(cfgedDC); err != nil {
		t.Error(err)
	}
	attrs := &utils.AttrDerivedChargers{Tenant: "hopa", Category: "call", Direction: "*out", Account: "responder_test", Subject: "responder_test"}
	dcs := &utils.DerivedChargerGroup{}
	if err := rsponder.GetDerivedChargers(attrs, dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, cfgedDC) {
		t.Errorf("Expecting: %v, received: %v ", cfgedDC, dcs)
	}
}

func TestResponderGetDerivedMaxSessionTime(t *testing.T) {
	testTenant := "test"
	cdr := &CDR{
		UniqueID:    utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.META_RATED,
		Direction:   "*out",
		Tenant:      testTenant,
		Category:    "call",
		Account:     "dan",
		Subject:     "dan",
		Destination: "447956",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.DEFAULT_RUNID,
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{
			"field_extr1": "val_extr1",
			"fieldextr2":  "valextr2",
		},
		Cost: dec.NewFloat(1.01),
	}
	var maxSessionTime float64
	if err := accountingStorage.SetAccount(&Account{Tenant: testTenant, Name: "rif"}); err != nil {
		t.Error(err)
	}
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != 0 {
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
	if err := ratingStorage.SetDestination(&Destination{Tenant: "test", Code: "+49151", Name: "DE_TMOBILE"}); err != nil {
		t.Error(err)
	}
	if err := ratingStorage.SetDestination(&Destination{Tenant: "test", Code: "+49160", Name: "DE_TMOBILE"}); err != nil {
		t.Error(err)
	}
	if err := ratingStorage.SetDestination(&Destination{Tenant: "test", Code: "+49170", Name: "DE_TMOBILE"}); err != nil {
		t.Error(err)
	}
	if err := ratingStorage.SetDestination(&Destination{Tenant: "test", Code: "+49171", Name: "DE_TMOBILE"}); err != nil {
		t.Error(err)
	}
	if err := ratingStorage.SetDestination(&Destination{Tenant: "test", Code: "+49175", Name: "DE_TMOBILE"}); err != nil {
		t.Error(err)
	}
	testTenant = "test"
	b10 := &Balance{Value: dec.NewFloat(10), Weight: 10, DestinationIDs: utils.NewStringMap("DE_TMOBILE")}
	b20 := &Balance{Value: dec.NewFloat(20), Weight: 10, DestinationIDs: utils.NewStringMap("DE_TMOBILE")}
	rifsAccount := &Account{Tenant: testTenant, Name: "rif", BalanceMap: map[string]Balances{utils.VOICE: Balances{b10}}}
	dansAccount := &Account{Tenant: testTenant, Name: "dan", BalanceMap: map[string]Balances{utils.VOICE: Balances{b20}}}
	if err := accountingStorage.SetAccount(rifsAccount); err != nil {
		t.Error(err)
	}
	if err := accountingStorage.SetAccount(dansAccount); err != nil {
		t.Error(err)
	}
	charger1 := &utils.DerivedChargerGroup{
		Direction: "*out",
		Tenant:    testTenant,
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Chargers: []*utils.DerivedCharger{
			{RunID: "extra1", Fields: `{"RequestType": {"$set":"*prepaid"},  "Account": {"$set":"dan"}, "Subject": {"$set":"dan"}, "Destination": {"$set":"+49151708707"}}`},
			{RunID: "extra2", Fields: `{"Account":{"$set":"ivo"}, "Subject":{"$set":"ivo"}}`},
			{RunID: "extra3", Fields: `{"RequestType":{"$set":"*pseudoprepaid"}, "Account":{"$set":"rif"}, "Subject":{"$set":"rif"}, "Destination":{"$set":"+49151708707"}}`},
		}}
	if err := ratingStorage.SetDerivedChargers(charger1); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	if rifStoredAcnt, err := accountingStorage.GetAccount(testTenant, "rif"); err != nil {
		t.Error(err)
		//} else if rifStoredAcnt.BalanceMap[utils.VOICE].Equal(rifsAccount.BalanceMap[utils.VOICE]) {
		//	t.Errorf("Expected: %+v, received: %+v", rifsAccount.BalanceMap[utils.VOICE][0], rifStoredAcnt.BalanceMap[utils.VOICE][0])
	} else if rifStoredAcnt.BalanceMap[utils.VOICE][0].GetValue().String() != rifsAccount.BalanceMap[utils.VOICE][0].GetValue().String() {
		t.Error("BalanceValue: ", rifStoredAcnt.BalanceMap[utils.VOICE][0].GetValue(), rifsAccount.BalanceMap[utils.VOICE][0].GetValue())
	}
	if danStoredAcnt, err := accountingStorage.GetAccount(testTenant, "dan"); err != nil {
		t.Error(err)
	} else if danStoredAcnt.BalanceMap[utils.VOICE][0].GetValue().String() != dansAccount.BalanceMap[utils.VOICE][0].GetValue().String() {
		t.Error("BalanceValue: ", danStoredAcnt.BalanceMap[utils.VOICE][0].GetValue(), dansAccount.BalanceMap[utils.VOICE][0].GetValue())
	}
	dcs := &utils.DerivedChargerGroup{}
	attrs := &utils.AttrDerivedChargers{Tenant: testTenant, Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	if err := rsponder.GetDerivedChargers(attrs, dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs.Chargers, charger1.Chargers) {
		t.Errorf("Expecting: %+v, received: %+v ", charger1, dcs)
	}
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != float64(10*time.Second) { // Smallest one, 10 seconds
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
}

func TestResponderGetSessionRuns(t *testing.T) {
	testTenant := "test_x"
	cdr := &CDR{
		UniqueID:    utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.META_PREPAID,
		Direction:   "*out",
		Tenant:      testTenant,
		Category:    "call",
		Account:     "dan2",
		Subject:     "dan2",
		Destination: "1002",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), PDD: 3 * time.Second,
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Supplier:   "suppl1",
		RunID:      utils.DEFAULT_RUNID,
		Usage:      time.Duration(10) * time.Second,
		ExtraFields: map[string]string{
			"field_extr1": "val_extr1",
			"fieldextr2":  "valextr2"},
		Cost: dec.NewFloat(1.01),
	}
	dfDC := &utils.DerivedCharger{RunID: utils.DEFAULT_RUNID}
	extra1DC := &utils.DerivedCharger{RunID: "extra1", Fields: `{"RequestType":{"$set":"*prepaid"}, "Category":{"$set":"0"}, "Account":{"$set":"minitsboy"}, "Subject":{"$set":"rif"}, "Destination":{"$set":"0256"}}`}
	extra2DC := &utils.DerivedCharger{RunID: "extra2", Fields: `{"Account":{"$set":"ivo"}, "Subject":{"$set":"ivo"}}`}
	extra3DC := &utils.DerivedCharger{RunID: "extra3", Fields: `{"RequestType":{"$set":"*pseudoprepaid"}, "Category":{"$set":"0"}, "Account":{"$set":"minu"}, "Subject":{"$set":"rif"}, "Destination":{"$set":"0256"}}`}
	charger1 := &utils.DerivedChargerGroup{
		Direction: "*out",
		Tenant:    testTenant,
		Category:  "call",
		Account:   "dan2",
		Subject:   "dan2",
		Chargers:  []*utils.DerivedCharger{extra1DC, extra2DC, extra3DC}}
	if err := ratingStorage.SetDerivedChargers(charger1); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	sesRuns := make([]*SessionRun, 0)
	eSRuns := []*SessionRun{
		&SessionRun{DerivedCharger: extra1DC,
			CallDescriptor: &CallDescriptor{UniqueID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), RunID: "extra1", Direction: "*out", Category: "0",
				Tenant: testTenant, Subject: "rif", Account: "minitsboy", Destination: "0256", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), TimeEnd: time.Date(2013, 11, 7, 8, 42, 36, 0, time.UTC), TOR: utils.VOICE, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}},
		&SessionRun{DerivedCharger: extra2DC,
			CallDescriptor: &CallDescriptor{UniqueID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), RunID: "extra2", Direction: "*out", Category: "call",
				Tenant: testTenant, Subject: "ivo", Account: "ivo", Destination: "1002", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), TimeEnd: time.Date(2013, 11, 7, 8, 42, 36, 0, time.UTC), TOR: utils.VOICE, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}},
		&SessionRun{DerivedCharger: dfDC,
			CallDescriptor: &CallDescriptor{UniqueID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), RunID: "*default", Direction: "*out", Category: "call",
				Tenant: testTenant, Subject: "dan2", Account: "dan2", Destination: "1002", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), TimeEnd: time.Date(2013, 11, 7, 8, 42, 36, 0, time.UTC), TOR: utils.VOICE, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}}}
	if err := rsponder.GetSessionRuns(cdr, &sesRuns); err != nil {
		t.Error(err)
	} else if utils.ToJSON(eSRuns) != utils.ToJSON(sesRuns) {
		t.Errorf("Expecting: %s, received: %s", utils.ToIJSON(eSRuns), utils.ToIJSON(sesRuns))
	}
}

func TestResponderGetLCR(t *testing.T) {
	rsponder.Stats = NewStats(ratingStorage, accountingStorage, cdrStorage) // Load stats instance
	if err := ratingStorage.SetDestination(&Destination{Tenant: "test", Code: "+49", Name: "GERMANY"}); err != nil {
		t.Error(err)
	}

	rp1 := &RatingPlan{
		Tenant: "test",
		Name:   "RP1",
		Timings: map[string]*RITiming{
			"30eab300": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": &RIRate{
				ConnectFee: dec.NewFloat(0),
				Rates: []*RateInfo{
					&RateInfo{
						GroupIntervalStart: 0,
						Value:              dec.NewFloat(0.01),
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
			},
		},
		DRates: map[string]*DRate{
			"30eab300_b457f86d_10": &DRate{
				Timing: "30eab300",
				Rating: "b457f86d",
				Weight: 10,
			},
		},
		DestinationRates: map[string]*DRateHelper{
			"49": &DRateHelper{
				DRateKeys: map[string]struct{}{"30b410": struct{}{}},
				CodeName:  "GERMANY",
			},
		},
	}
	rp2 := &RatingPlan{
		Tenant: "test",
		Name:   "RP2",
		Timings: map[string]*RITiming{
			"30eab300": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": &RIRate{
				ConnectFee: dec.NewFloat(0),
				Rates: []*RateInfo{
					&RateInfo{
						GroupIntervalStart: 0,
						Value:              dec.NewFloat(0.02),
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
			},
		},
		DRates: map[string]*DRate{
			"30eab300_b457f86d_10": &DRate{
				Timing: "30eab300",
				Rating: "b457f86d",
				Weight: 10,
			},
		},
		DestinationRates: map[string]*DRateHelper{
			"49": &DRateHelper{
				DRateKeys: map[string]struct{}{"30b410": struct{}{}},
				CodeName:  "GERMANY",
			},
		},
	}
	rp3 := &RatingPlan{
		Tenant: "test",
		Name:   "RP3",
		Timings: map[string]*RITiming{
			"30eab300": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": &RIRate{
				ConnectFee: dec.NewFloat(0),
				Rates: []*RateInfo{
					&RateInfo{
						GroupIntervalStart: 0,
						Value:              dec.NewFloat(0.03),
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
			},
		},
		DRates: map[string]*DRate{
			"30eab300_b457f86d_10": &DRate{
				Timing: "30eab300",
				Rating: "b457f86d",
				Weight: 10,
			},
		},
		DestinationRates: map[string]*DRateHelper{
			"49": &DRateHelper{
				DRateKeys: map[string]struct{}{"30b410": struct{}{}},
				CodeName:  "GERMANY",
			},
		},
	}
	for _, rpf := range []*RatingPlan{rp1, rp2, rp3} {
		if err := ratingStorage.SetRatingPlan(rpf); err != nil {
			t.Error(err)
		}
	}
	danStatsID := "dan12_stats"
	var r int
	rsponder.Stats.Call("CDRStatsV1.AddQueue", &CdrStats{Tenant: "test", Name: danStatsID, Metrics: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}, Filter: `{"Supplier":{"$in":["dan12"]}}`}, &r)
	danRpfl := &RatingProfile{
		Direction: utils.OUT,
		Tenant:    "test",
		Category:  "call",
		Subject:   "dan12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanID:    rp1.Name,
			FallbackKeys:    []string{},
			CdrStatQueueIDs: []string{danStatsID},
		}},
	}
	rifStatsID := "rif12_stats"
	rsponder.Stats.Call("CDRStatsV1.AddQueue", &CdrStats{Tenant: "test", Name: rifStatsID, Metrics: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}, Filter: `{"Supplier":"rif12"}`}, &r)
	rifRpfl := &RatingProfile{
		Direction: utils.OUT,
		Tenant:    "test",
		Category:  "call",
		Subject:   "rif12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanID:    rp2.Name,
			FallbackKeys:    []string{},
			CdrStatQueueIDs: []string{rifStatsID},
		}},
	}
	ivoStatsID := "ivo12_stats"
	rsponder.Stats.Call("CDRStatsV1.AddQueue", &CdrStats{Tenant: "test", Name: ivoStatsID, Filter: `"Supplier":"ivo12"}`, Metrics: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}}, &r)
	ivoRpfl := &RatingProfile{
		Direction: "*out",
		Tenant:    "test",
		Category:  "call",
		Subject:   "ivo12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanID:    rp3.Name,
			FallbackKeys:    []string{},
			CdrStatQueueIDs: []string{ivoStatsID},
		}},
	}
	for _, rpfl := range []*RatingProfile{danRpfl, rifRpfl, ivoRpfl} {
		if err := ratingStorage.SetRatingProfile(rpfl); err != nil {
			t.Error(err)
		}
	}
	lcrStatic := &LCR{Direction: utils.OUT, Tenant: "test", Category: "call_static", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_STATIC, StrategyParams: "ivo12;dan12;rif12", Weight: 10.0}},
			},
		},
	}
	lcrLowestCost := &LCR{Direction: utils.OUT, Tenant: "test", Category: "call_least_cost", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOWEST, Weight: 10.0}},
			},
		},
	}
	lcrQosThreshold := &LCR{Direction: utils.OUT, Tenant: "test", Category: "call_qos_threshold", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;;;4m;;;;;;;;;", Weight: 10.0}},
			},
		},
	}
	lcrQos := &LCR{Direction: utils.OUT, Tenant: "test", Category: "call_qos", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS, Weight: 10.0}},
			},
		},
	}
	lcrLoad := &LCR{Direction: utils.OUT, Tenant: "test", Category: "call_load", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOAD, StrategyParams: "ivo12:10;dan12:3", Weight: 10.0}},
			},
		},
	}
	for _, lcr := range []*LCR{lcrStatic, lcrLowestCost, lcrQosThreshold, lcrQos, lcrLoad} {
		if err := ratingStorage.SetLCR(lcr); err != nil {
			t.Error(err)
		}
	}
	cdStatic := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "test",
		Direction:   utils.OUT,
		Category:    "call_static",
		Destination: "+4986517174963",
		Account:     "dan",
		Subject:     "dan",
	}
	eStLcr := &LCRCost{
		Entry: &LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_STATIC, StrategyParams: "ivo12;dan12;rif12", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:test:call:ivo12", Cost: dec.NewFloat(1.8), Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:test:call:dan12", Cost: dec.NewFloat(0.6), Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:test:call:rif12", Cost: dec.NewFloat(1.2), Duration: 60 * time.Second},
		},
	}
	var lcr LCRCost
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdStatic}, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		//t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eStLcr.SupplierCosts), utils.ToJSON(lcr.SupplierCosts))
	}
	// Test *least_cost strategy here
	cdLowestCost := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "test",
		Direction:   utils.OUT,
		Category:    "call_least_cost",
		Destination: "+4986517174963",
		Account:     "dan",
		Subject:     "dan",
	}
	eLcLcr := &LCRCost{
		Entry: &LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOWEST, Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:test:call:dan12", Cost: dec.NewFloat(0.6), Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:test:call:rif12", Cost: dec.NewFloat(1.2), Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:test:call:ivo12", Cost: dec.NewFloat(1.8), Duration: 60 * time.Second},
		},
	}
	var lcrLc LCRCost
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdLowestCost}, &lcrLc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcLcr.Entry, lcrLc.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.Entry, lcrLc.Entry)

	} else if !reflect.DeepEqual(eLcLcr.SupplierCosts, lcrLc.SupplierCosts) {
		//t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eLcLcr.SupplierCosts), utils.ToJSON(lcrLc.SupplierCosts))
	}
	bRif12 := &Balance{Value: dec.NewFloat(40), Weight: 10, DestinationIDs: utils.NewStringMap("GERMANY")}
	bIvo12 := &Balance{Value: dec.NewFloat(60), Weight: 10, DestinationIDs: utils.NewStringMap("GERMANY")}
	rif12sAccount := &Account{Tenant: "test", Name: "rif12", BalanceMap: map[string]Balances{utils.VOICE: Balances{bRif12}}, AllowNegative: true}
	ivo12sAccount := &Account{Tenant: "test", Name: "ivo12", BalanceMap: map[string]Balances{utils.VOICE: Balances{bIvo12}}, AllowNegative: true}
	for _, acnt := range []*Account{rif12sAccount, ivo12sAccount} {
		if err := accountingStorage.SetAccount(acnt); err != nil {
			t.Error(err)
		}
	}
	eLcLcr = &LCRCost{
		Entry: &LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOWEST, Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:test:call:ivo12", Cost: dec.NewFloat(0), Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:test:call:rif12", Cost: dec.NewFloat(0.4), Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:test:call:dan12", Cost: dec.NewFloat(0.6), Duration: 60 * time.Second},
		},
	}
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdLowestCost}, &lcrLc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcLcr.Entry, lcrLc.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.Entry, lcrLc.Entry)

	} else if !reflect.DeepEqual(eLcLcr.SupplierCosts, lcrLc.SupplierCosts) {
		//t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eLcLcr.SupplierCosts), utils.ToJSON(lcrLc.SupplierCosts))
	}

	// Test *qos_threshold strategy here
	cdQosThreshold := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "test",
		Direction:   utils.OUT,
		Category:    "call_qos_threshold",
		Destination: "+4986517174963",
		Account:     "dan",
		Subject:     "dan",
	}
	eQTLcr := &LCRCost{
		Entry: &LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;;;4m;;;;;;;;;", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:test:call:ivo12", Cost: dec.NewFloat(0), Duration: 60 * time.Second, QOS: map[string]*dec.Dec{TCD: dec.NewVal(-1, 0), ACC: dec.NewVal(-1, 0), TCC: dec.NewVal(-1, 0), ASR: dec.NewVal(-1, 0), ACD: dec.NewVal(-1, 0), PDD: dec.NewVal(-1, 0), DDC: dec.NewVal(-1, 0)}, qosSortParams: []string{"35", "4m"}},
			&LCRSupplierCost{Supplier: "*out:test:call:rif12", Cost: dec.NewFloat(0.4), Duration: 60 * time.Second, QOS: map[string]*dec.Dec{TCD: dec.NewVal(-1, 0), ACC: dec.NewVal(-1, 0), TCC: dec.NewVal(-1, 0), ASR: dec.NewVal(-1, 0), ACD: dec.NewVal(-1, 0), PDD: dec.NewVal(-1, 0), DDC: dec.NewVal(-1, 0)}, qosSortParams: []string{"35", "4m"}},
			&LCRSupplierCost{Supplier: "*out:test:call:dan12", Cost: dec.NewFloat(0.6), Duration: 60 * time.Second, QOS: map[string]*dec.Dec{TCD: dec.NewVal(-1, 0), ACC: dec.NewVal(-1, 0), TCC: dec.NewVal(-1, 0), ASR: dec.NewVal(-1, 0), ACD: dec.NewVal(-1, 0), PDD: dec.NewVal(-1, 0), DDC: dec.NewVal(-1, 0)}, qosSortParams: []string{"35", "4m"}},
		},
	}
	var lcrQT LCRCost
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdQosThreshold}, &lcrQT); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eQTLcr.Entry, lcrQT.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eQTLcr.Entry, lcrQT.Entry)

	} else if !reflect.DeepEqual(eQTLcr.SupplierCosts, lcrQT.SupplierCosts) {
		//t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eQTLcr.SupplierCosts), utils.ToJSON(lcrQT.SupplierCosts))
	}

	cdr := &CDR{Supplier: "rif12", AnswerTime: time.Now(), Usage: 3 * time.Minute, Cost: dec.NewFloat(1)}
	rsponder.Stats.Call("CDRStatsV1.AppendCDR", cdr, &r)
	cdr = &CDR{Supplier: "dan12", AnswerTime: time.Now(), Usage: 5 * time.Minute, Cost: dec.NewFloat(2)}
	rsponder.Stats.Call("CDRStatsV1.AppendCDR", cdr, &r)

	eQTLcr = &LCRCost{
		Entry: &LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;;;4m;;;;;;;;;", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:test:call:ivo12", Cost: dec.NewFloat(0), Duration: 60 * time.Second, QOS: map[string]*dec.Dec{PDD: dec.NewVal(-1, 0), TCD: dec.NewVal(-1, 0), ACC: dec.NewVal(-1, 0), TCC: dec.NewVal(-1, 0), ASR: dec.NewVal(-1, 0), ACD: dec.NewVal(-1, 0), DDC: dec.NewVal(-1, 0)}, qosSortParams: []string{"35", "4m"}},
			&LCRSupplierCost{Supplier: "*out:test:call:dan12", Cost: dec.NewFloat(0.6), Duration: 60 * time.Second, QOS: map[string]*dec.Dec{PDD: dec.NewVal(-1, 0), ACD: dec.NewVal(300, 0), TCD: dec.NewVal(300, 0), ASR: dec.NewVal(100, 0), ACC: dec.NewVal(2, 0), TCC: dec.NewVal(2, 0), DDC: dec.NewVal(2, 0)}, qosSortParams: []string{"35", "4m"}},
		},
	}
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdQosThreshold}, &lcrQT); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eQTLcr.Entry, lcrQT.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eQTLcr.Entry, lcrQT.Entry)

	} else if !reflect.DeepEqual(eQTLcr.SupplierCosts, lcrQT.SupplierCosts) {
		//t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eQTLcr.SupplierCosts), utils.ToJSON(lcrQT.SupplierCosts))
	}

	// Test *qos strategy here
	cdQos := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "test",
		Direction:   utils.OUT,
		Category:    "call_qos",
		Destination: "+4986517174963",
		Account:     "dan",
		Subject:     "dan",
	}
	eQosLcr := &LCRCost{
		Entry: &LCREntry{DestinationID: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS, Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:test:call:ivo12", Cost: dec.NewFloat(0), Duration: 60 * time.Second, QOS: map[string]*dec.Dec{ACD: dec.NewVal(-1, 0), PDD: dec.NewVal(-1, 0), TCD: dec.NewVal(-1, 0), ASR: dec.NewVal(-1, 0), ACC: dec.NewVal(-1, 0), TCC: dec.NewVal(-1, 0), DDC: dec.NewVal(-1, 0)}, qosSortParams: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}},
			&LCRSupplierCost{Supplier: "*out:test:call:dan12", Cost: dec.NewFloat(0.6), Duration: 60 * time.Second, QOS: map[string]*dec.Dec{ACD: dec.NewVal(300, 0), PDD: dec.NewVal(-1, 0), TCD: dec.NewVal(300, 0), ASR: dec.NewVal(100, 0), ACC: dec.NewVal(2, 0), TCC: dec.NewVal(2, 0), DDC: dec.NewVal(2, 0)}, qosSortParams: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}},
			&LCRSupplierCost{Supplier: "*out:test:call:rif12", Cost: dec.NewFloat(0.4), Duration: 60 * time.Second, QOS: map[string]*dec.Dec{ACD: dec.NewVal(180, 0), PDD: dec.NewVal(-1, 0), TCD: dec.NewVal(180, 0), ASR: dec.NewVal(100, 0), ACC: dec.NewVal(1, 0), TCC: dec.NewVal(1, 0), DDC: dec.NewVal(1, 0)}, qosSortParams: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}},
		},
	}
	var lcrQ LCRCost
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdQos}, &lcrQ); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eQosLcr.Entry, lcrQ.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eQosLcr.Entry, lcrQ.Entry)

	} else if !reflect.DeepEqual(eQosLcr.SupplierCosts, lcrQ.SupplierCosts) {
		//t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eQosLcr.SupplierCosts), utils.ToJSON(lcrQ.SupplierCosts))
	}
}

func TestResponderGobSMCost(t *testing.T) {
	attr := AttrCDRSStoreSMCost{
		Cost: &SMCost{
			UniqueID:   "b783a8bcaa356570436983cd8a0e6de4993f9ba6",
			RunID:      "*default",
			OriginHost: "",
			OriginID:   "testdatagrp_grp1",
			CostSource: "SMR",
			Usage:      1536,
			CostDetails: &CallCost{
				Direction:   "*out",
				Category:    "generic",
				Tenant:      "test",
				Subject:     "1001",
				Account:     "1001",
				Destination: "data",
				TOR:         "*data",
				Cost:        dec.NewFloat(0),
				Timespans: TimeSpans{&TimeSpan{
					TimeStart: time.Date(2016, 1, 5, 12, 30, 10, 0, time.UTC),
					TimeEnd:   time.Date(2016, 1, 5, 12, 55, 46, 0, time.UTC),
					Cost:      dec.NewFloat(0),
					RateInterval: &RateInterval{
						Timing: nil,
						Rating: &RIRate{
							ConnectFee:      dec.NewFloat(0),
							MaxCost:         dec.NewFloat(0),
							MaxCostStrategy: "",
							Rates: RateGroups{&RateInfo{
								GroupIntervalStart: 0,
								Value:              dec.NewFloat(0),
								RateIncrement:      1 * time.Second,
								RateUnit:           1 * time.Second,
							},
							},
						},
						Weight: 0,
					},
					DurationIndex: 0,
					Increments: &Increments{CompIncrement: &Increment{
						Duration: 1 * time.Second,
						Cost:     dec.NewFloat(0),
						BalanceInfo: &DebitInfo{
							Unit: &UnitInfo{
								UUID:          "fa0aa280-2b76-4b5b-bb06-174f84b8c321",
								ID:            "",
								Value:         dec.NewFloat(100864),
								DestinationID: "data",
								Consumed:      "1",
								TOR:           "*data",
								RateInterval:  nil,
							},
							Monetary:  nil,
							AccountID: "test:1001",
						},
						CompressFactor: 1536,
					},
					},
					MatchedSubject: "fa0aa280-2b76-4b5b-bb06-174f84b8c321",
					MatchedPrefix:  "data",
					MatchedDestID:  "*any",
					RatingPlanID:   "*none",
					CompressFactor: 1,
				},
				},
				RatedUsage: 1536,
			},
		},
		CheckDuplicate: false,
	}

	var network bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&network) // Will write to network.
	dec := gob.NewDecoder(&network) // Will read from network.
	err := enc.Encode(attr)
	if err != nil {
		t.Error("encode error: ", err)
	}

	// Decode (receive) and print the values.
	var q AttrCDRSStoreSMCost
	err = dec.Decode(&q)
	if err != nil {
		t.Error("decode error: ", err)
	}
	if !reflect.DeepEqual(attr, q) {
		t.Error("wrong transmission")
	}
}
