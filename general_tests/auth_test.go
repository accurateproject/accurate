package general_tests

import (
	"testing"
	"time"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

var ratingDbAuth engine.RatingStorage
var acntDbAuth engine.AccountingStorage
var rsponder *engine.Responder

func TestAuthSetStorage(t *testing.T) {
	ratingDbAuth, _ = engine.NewMapStorageJson()
	engine.SetRatingStorage(ratingDbAuth)
	acntDbAuth, _ = engine.NewMapStorageJson()
	engine.SetAccountingStorage(acntDbAuth)
	config.Reset()
	//cfg := config.Get()
	rsponder = new(engine.Responder)

}

func TestAuthLoadCsv(t *testing.T) {
	timings := ``
	destinations := `DST_GERMANY_LANDLINE,49`
	rates := `RT_1CENTWITHCF,0.02,0.01,60s,60s,0s`
	destinationRates := `DR_GERMANY,DST_GERMANY_LANDLINE,RT_1CENTWITHCF,*up,8,,
DR_ANY_1CNT,*any,RT_1CENTWITHCF,*up,8,,`
	ratingPlans := `RP_1,DR_GERMANY,*any,10
RP_ANY,DR_ANY_1CNT,*any,10`
	ratingProfiles := `*out,cgrates.org,call,testauthpostpaid1,2013-01-06T00:00:00Z,RP_1,,
*out,cgrates.org,call,testauthpostpaid2,2013-01-06T00:00:00Z,RP_1,*any,
*out,cgrates.org,call,*any,2013-01-06T00:00:00Z,RP_ANY,,`
	sharedGroups := ``
	lcrs := ``
	actions := `TOPUP10_AC,*topup_reset,,,,*monetary,*out,,*any,,,*unlimited,,0,10,false,false,10`
	actionPlans := `TOPUP10_AT,TOPUP10_AC,*asap,10`
	actionTriggers := ``
	accountActions := `cgrates.org,testauthpostpaid1,TOPUP10_AT,,,`
	derivedCharges := ``
	cdrStats := ``
	users := ``
	aliases := ``
	resLimits := ``
	csvr := engine.NewTpReader(ratingDbAuth, acntDbAuth, engine.NewStringCSVStorage(',', destinations, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		sharedGroups, lcrs, actions, actionPlans, actionTriggers, accountActions, derivedCharges, cdrStats, users, aliases, resLimits), "", "")
	if err := csvr.LoadAll(); err != nil {
		t.Fatal(err)
	}
	csvr.WriteToDatabase(false, false, false)
	if acnt, err := acntDbAuth.GetAccount("cgrates.org:testauthpostpaid1"); err != nil {
		t.Error(err)
	} else if acnt == nil {
		t.Error("No account saved")
	}

	cache2go.Flush()
	ratingDbAuth.PreloadRatingCache()
	acntDbAuth.PreloadAccountingCache()

	if cachedDests := cache2go.CountEntries(utils.DESTINATION_PREFIX); cachedDests != 0 {
		t.Error("Wrong number of cached destinations found", cachedDests)
	}
	if cachedRPlans := cache2go.CountEntries(utils.RATING_PLAN_PREFIX); cachedRPlans != 2 {
		t.Error("Wrong number of cached rating plans found", cachedRPlans)
	}
	if cachedRProfiles := cache2go.CountEntries(utils.RATING_PROFILE_PREFIX); cachedRProfiles != 0 {
		t.Error("Wrong number of cached rating profiles found", cachedRProfiles)
	}
	if cachedActions := cache2go.CountEntries(utils.ACTION_PREFIX); cachedActions != 0 {
		t.Error("Wrong number of cached actions found", cachedActions)
	}
}

func TestAuthPostpaidNoAcnt(t *testing.T) {
	cdr := &engine.CDR{ToR: utils.VOICE, RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "nonexistent", Subject: "testauthpostpaid1",
		Destination: "4986517174963", SetupTime: time.Date(2015, 8, 27, 11, 26, 0, 0, time.UTC)}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err == nil || err != utils.ErrAccountNotFound {
		t.Error(err)
	}
}

func TestAuthPostpaidNoDestination(t *testing.T) {
	// Test subject which does not have destination attached
	cdr := &engine.CDR{ToR: utils.VOICE, RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "testauthpostpaid1", Subject: "testauthpostpaid1",
		Destination: "441231234", SetupTime: time.Date(2015, 8, 27, 11, 26, 0, 0, time.UTC)}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err == nil {
		t.Error("Expecting error for destination not allowed to subject")
	}
}

func TestAuthPostpaidFallbackDest(t *testing.T) {
	// Test subject which has fallback for destination
	cdr := &engine.CDR{ToR: utils.VOICE, RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "testauthpostpaid1", Subject: "testauthpostpaid2",
		Destination: "441231234", SetupTime: time.Date(2015, 8, 27, 11, 26, 0, 0, time.UTC)}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != -1 {
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
}

func TestAuthPostpaidWithDestination(t *testing.T) {
	// Test subject which does not have destination attached
	cdr := &engine.CDR{ToR: utils.VOICE, RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "testauthpostpaid1", Subject: "testauthpostpaid1",
		Destination: "4986517174963", SetupTime: time.Date(2015, 8, 27, 11, 26, 0, 0, time.UTC)}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != -1 {
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
}
