
package general_tests

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/cache2go"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var ratingDbAcntActs engine.RatingStorage
var acntDbAcntActs engine.AccountingStorage

func TestAcntActsSetStorage(t *testing.T) {
	ratingDbAcntActs, _ = engine.NewMapStorageJson()
	engine.SetRatingStorage(ratingDbAcntActs)
	acntDbAcntActs, _ = engine.NewMapStorageJson()
	engine.SetAccountingStorage(acntDbAcntActs)
}

func TestAcntActsLoadCsv(t *testing.T) {
	timings := `ASAP,*any,*any,*any,*any,*asap`
	destinations := ``
	rates := ``
	destinationRates := ``
	ratingPlans := ``
	ratingProfiles := ``
	sharedGroups := ``
	lcrs := ``
	actions := `TOPUP10_AC,*topup_reset,,,,*voice,*out,,*any,,,*unlimited,,10,10,false,false,10
DISABLE_ACNT,*disable_account,,,,,,,,,,,,,,false,false,10
ENABLE_ACNT,*enable_account,,,,,,,,,,,,,,false,false,10`
	actionPlans := `TOPUP10_AT,TOPUP10_AC,ASAP,10`
	actionTriggers := ``
	accountActions := `cgrates.org,1,TOPUP10_AT,,,`
	derivedCharges := ``
	cdrStats := ``
	users := ``
	aliases := ``
	resLimits := ``
	csvr := engine.NewTpReader(ratingDbAcntActs, acntDbAcntActs, engine.NewStringCSVStorage(',', destinations, timings, rates, destinationRates, ratingPlans, ratingProfiles,
		sharedGroups, lcrs, actions, actionPlans, actionTriggers, accountActions, derivedCharges, cdrStats, users, aliases, resLimits), "", "")
	if err := csvr.LoadAll(); err != nil {
		t.Fatal(err)
	}
	csvr.WriteToDatabase(false, false, false)

	cache2go.Flush()
	ratingDbAcntActs.PreloadRatingCache()
	acntDbAcntActs.PreloadAccountingCache()

	expectAcnt := &engine.Account{ID: "cgrates.org:1"}
	if acnt, err := acntDbAcntActs.GetAccount("cgrates.org:1"); err != nil {
		t.Error(err)
	} else if acnt == nil {
		t.Error("No account created")
	} else if !reflect.DeepEqual(expectAcnt.ActionTriggers, acnt.ActionTriggers) {
		t.Errorf("Expecting: %+v, received: %+v", expectAcnt, acnt)
	}
}

func TestAcntActsDisableAcnt(t *testing.T) {
	acnt1Tag := "cgrates.org:1"
	at := &engine.ActionTiming{
		ActionsID: "DISABLE_ACNT",
	}
	at.SetAccountIDs(utils.StringMap{acnt1Tag: true})
	if err := at.Execute(); err != nil {
		t.Error(err)
	}
	expectAcnt := &engine.Account{ID: "cgrates.org:1", Disabled: true}
	if acnt, err := acntDbAcntActs.GetAccount(acnt1Tag); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectAcnt, acnt) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectAcnt), utils.ToJSON(acnt))
	}
}

func TestAcntActsEnableAcnt(t *testing.T) {
	acnt1Tag := "cgrates.org:1"
	at := &engine.ActionTiming{
		ActionsID: "ENABLE_ACNT",
	}
	at.SetAccountIDs(utils.StringMap{acnt1Tag: true})
	if err := at.Execute(); err != nil {
		t.Error(err)
	}
	expectAcnt := &engine.Account{ID: "cgrates.org:1", Disabled: false}
	if acnt, err := acntDbAcntActs.GetAccount(acnt1Tag); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectAcnt, acnt) {
		t.Errorf("Expecting: %+v, received: %+v", expectAcnt, acnt)
	}
}
