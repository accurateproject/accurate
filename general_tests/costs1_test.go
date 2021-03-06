package general_tests

import (
	"testing"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func TestCosts1SetStorage(t *testing.T) {
	ratingDb, _ = engine.NewMapStorageJson()
	engine.SetRatingStorage(ratingDb)
	acntDb, _ = engine.NewMapStorageJson()
	engine.SetAccountingStorage(acntDb)
}

func TestCosts1LoadCsvTp(t *testing.T) {
	timings := `ALWAYS,*any,*any,*any,*any,00:00:00
ASAP,*any,*any,*any,*any,*asap`
	dests := `GERMANY,+49
GERMANY_MOBILE,+4915
GERMANY_MOBILE,+4916
GERMANY_MOBILE,+4917`
	rates := `RT_1CENT,0,1,1s,1s,0s
RT_DATA_2c,0,0.002,10,10,0
RT_SMS_5c,0,0.005,1,1,0`
	destinationRates := `DR_RETAIL,GERMANY,RT_1CENT,*up,4,0,
DR_RETAIL,GERMANY_MOBILE,RT_1CENT,*up,4,0,
DR_DATA_1,*any,RT_DATA_2c,*up,4,0,
DR_SMS_1,*any,RT_SMS_5c,*up,4,0,`
	ratingPlans := `RP_RETAIL,DR_RETAIL,ALWAYS,10
RP_DATA1,DR_DATA_1,ALWAYS,10
RP_SMS1,DR_SMS_1,ALWAYS,10`
	ratingProfiles := `*out,cgrates.org,call,*any,2012-01-01T00:00:00Z,RP_RETAIL,,
*out,cgrates.org,data,*any,2012-01-01T00:00:00Z,RP_DATA1,,
*out,cgrates.org,sms,*any,2012-01-01T00:00:00Z,RP_SMS1,,`
	csvr := engine.NewTpReader(ratingDb, acntDb, engine.NewStringCSVStorage(',', dests, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		"", "", "", "", "", "", "", "", "", "", ""), "", "")

	if err := csvr.LoadTimings(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadDestinations(); err != nil {
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

	if cachedRPlans := cache2go.CountEntries(utils.RATING_PLAN_PREFIX); cachedRPlans != 3 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := cache2go.CountEntries(utils.RATING_PROFILE_PREFIX); cachedRProfiles != 0 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
}

func TestCosts1GetCost1(t *testing.T) {
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:30Z", "")
	cd := &engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "+4986517174963",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 90 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}

func TestCosts1GetCostZeroDuration(t *testing.T) {
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	cd := &engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "+4986517174963",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}

/* FixMe
func TestCosts1GetCost2(t *testing.T) {
	tStart, _ := utils.ParseTimeDetectLayout("2004-06-04T00:00:01Z")
	tEnd, _ := utils.ParseTimeDetectLayout("2004-06-04T00:01:01Z")
	cd := &engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "+4986517174963",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	if cc, err := cd.GetCost(); err != nil {
		t.Error(err)
	} else if cc.Cost != 90 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
}
*/
