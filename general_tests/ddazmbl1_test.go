package general_tests

import (
	"testing"
	"time"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/scheduler"
	"github.com/accurateproject/accurate/utils"
)

var ratingDb engine.RatingStorage
var acntDb engine.AccountingStorage

func TestSetStorage(t *testing.T) {
	ratingDb, _ = engine.NewMapStorageJson()
	engine.SetRatingStorage(ratingDb)
	acntDb, _ = engine.NewMapStorageJson()
	engine.SetAccountingStorage(acntDb)
}

func TestLoadCsvTp(t *testing.T) {
	timings := `ALWAYS,*any,*any,*any,*any,00:00:00
ASAP,*any,*any,*any,*any,*asap`
	destinations := `DST_UK_Mobile_BIG5,447596
DST_UK_Mobile_BIG5,447956`
	rates := `RT_UK_Mobile_BIG5_PKG,0.01,0,20s,20s,0s
RT_UK_Mobile_BIG5,0.01,0.10,1s,1s,0s`
	destinationRates := `DR_UK_Mobile_BIG5_PKG,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5_PKG,*up,8,0,
DR_UK_Mobile_BIG5,DST_UK_Mobile_BIG5,RT_UK_Mobile_BIG5,*up,8,0,`
	ratingPlans := `RP_UK_Mobile_BIG5_PKG,DR_UK_Mobile_BIG5_PKG,ALWAYS,10
RP_UK,DR_UK_Mobile_BIG5,ALWAYS,10`
	ratingProfiles := `*out,cgrates.org,call,*any,2013-01-06T00:00:00Z,RP_UK,,
*out,cgrates.org,call,discounted_minutes,2013-01-06T00:00:00Z,RP_UK_Mobile_BIG5_PKG,,`
	sharedGroups := ``
	lcrs := ``
	actions := `TOPUP10_AC,*topup_reset,,,,*monetary,*out,,*any,,,*unlimited,,10,10,false,false,10
TOPUP10_AC1,*topup_reset,,,,*voice,*out,,DST_UK_Mobile_BIG5,discounted_minutes,,*unlimited,,40,10,false,false,10`
	actionPlans := `TOPUP10_AT,TOPUP10_AC,ASAP,10
TOPUP10_AT,TOPUP10_AC1,ASAP,10`
	actionTriggers := ``
	accountActions := `cgrates.org,12344,TOPUP10_AT,,,`
	derivedCharges := ``
	cdrStats := ``
	users := ``
	aliases := ``
	resLimits := ``
	csvr := engine.NewTpReader(ratingDb, acntDb, engine.NewStringCSVStorage(',', destinations, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		sharedGroups, lcrs, actions, actionPlans, actionTriggers, accountActions, derivedCharges, cdrStats, users, aliases, resLimits), "", "")
	if err := csvr.LoadDestinations(); err != nil {
		t.Fatal(err)
	}
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
	if err := csvr.LoadSharedGroups(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadLCRs(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadActions(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadActionPlans(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadActionTriggers(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadAccountActions(); err != nil {
		t.Fatal(err)
	}
	if err := csvr.LoadDerivedChargers(); err != nil {
		t.Fatal(err)
	}
	csvr.WriteToDatabase(false, false, false)
	if acnt, err := acntDb.GetAccount("cgrates.org:12344"); err != nil {
		t.Error(err)
	} else if acnt == nil {
		t.Error("No account saved")
	}
	cache2go.Flush()
	ratingDb.PreloadRatingCache()
	acntDb.PreloadAccountingCache()

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

func TestExecuteActions(t *testing.T) {
	scheduler.NewScheduler(ratingDb).Reload(false)
	time.Sleep(10 * time.Millisecond) // Give time to scheduler to topup the account
	if acnt, err := acntDb.GetAccount("cgrates.org:12344"); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap) != 2 {
		t.Error("Account does not have enough balances: ", acnt.BalanceMap)
	} else if acnt.BalanceMap[utils.VOICE][0].Value != 40 {
		t.Error("Account does not have enough minutes in balance", acnt.BalanceMap[utils.VOICE][0].Value)
	} else if acnt.BalanceMap[utils.MONETARY][0].Value != 10 {
		t.Error("Account does not have enough monetary balance", acnt.BalanceMap[utils.MONETARY][0].Value)
	}
}

func TestDebit(t *testing.T) {
	cd := &engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "12344",
		Account:     "12344",
		Destination: "447956933443",
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 0, 10, 0, time.UTC),
	}
	if cc, err := cd.Debit(); err != nil {
		t.Error(err)
	} else if cc.Cost != 0.01 {
		t.Error("Wrong cost returned: ", cc.Cost)
	}
	acnt, err := acntDb.GetAccount("cgrates.org:12344")
	if err != nil {
		t.Error(err)
	}
	if acnt.BalanceMap[utils.VOICE][0].Value != 20 {
		t.Error("Account does not have expected minutes in balance", acnt.BalanceMap[utils.VOICE][0].Value)
	}
	if acnt.BalanceMap[utils.MONETARY][0].Value != 9.99 {
		t.Error("Account does not have expected monetary balance", acnt.BalanceMap[utils.MONETARY][0].Value)
	}
}
