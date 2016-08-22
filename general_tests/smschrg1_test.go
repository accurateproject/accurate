package general_tests

import (
	"testing"
	"time"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func TestSMSSetStorageSmsChrg1(t *testing.T) {
	ratingDb, _ = engine.NewMapStorageJson()
	engine.SetRatingStorage(ratingDb)
	acntDb, _ = engine.NewMapStorageJson()
	engine.SetAccountingStorage(acntDb)
}

func TestSMSLoadCsvTpSmsChrg1(t *testing.T) {
	timings := `ALWAYS,*any,*any,*any,*any,00:00:00`
	rates := `RT_SMS_5c,0,0.005,1,1,0`
	destinationRates := `DR_SMS_1,*any,RT_SMS_5c,*up,4,0,`
	ratingPlans := `RP_SMS1,DR_SMS_1,ALWAYS,10`
	ratingProfiles := `*out,cgrates.org,sms,*any,2012-01-01T00:00:00Z,RP_SMS1,,`
	csvr := engine.NewTpReader(ratingDb, acntDb, engine.NewStringCSVStorage(',', "", timings, rates, destinationRates, ratingPlans, ratingProfiles,
		"", "", "", "", "", "", "", "", "", "", ""), "", "")
	if err := csvr.LoadTimings(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadRates(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadDestinationRates(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadRatingPlans(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadRatingProfiles(); err != nil {
		t.Fatal(err)
	}
	csvr.WriteToDatabase(false, false, false)
	cache2go.Flush()
	ratingDb.PreloadRatingCache()
	acntDb.PreloadAccountingCache()

	if cachedRPlans := cache2go.CountEntries(utils.RATING_PLAN_PREFIX); cachedRPlans != 1 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := cache2go.CountEntries(utils.RATING_PROFILE_PREFIX); cachedRProfiles != 0 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
}

func TestSMSGetDataCostSmsChrg1(t *testing.T) {
	usageDur := time.Second
	timeStart := time.Date(2014, 3, 4, 0, 0, 0, 0, time.Local)
	cd := &engine.CallDescriptor{
		Direction:     "*out",
		Category:      "sms",
		Tenant:        "cgrates.org",
		Subject:       "12345",
		Account:       "12345",
		Destination:   "+4917621621391",
		TimeStart:     timeStart,
		TimeEnd:       timeStart.Add(usageDur),
		DurationIndex: usageDur,
		TOR:           utils.SMS,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.005 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}
