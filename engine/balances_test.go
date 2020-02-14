package engine

import (
	"testing"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

func TestBalanceSortPrecision(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 2}
	mb2 := &Balance{Weight: 2, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by weight!")
	}
}

func TestBalanceSortPrecisionWeightEqual(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 2}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortPrecisionWeightGreater(t *testing.T) {
	mb1 := &Balance{Weight: 2, precision: 2}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeight(t *testing.T) {
	mb1 := &Balance{Weight: 2, precision: 1}
	mb2 := &Balance{Weight: 1, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb1 || bs[1] != mb2 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceSortWeightLess(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1}
	mb2 := &Balance{Weight: 2, precision: 1}
	var bs Balances
	bs = append(bs, mb2, mb1)
	bs.Sort()
	if bs[0] != mb2 || bs[1] != mb1 {
		t.Error("Buckets not sorted by precision!")
	}
}

func TestBalanceEqual(t *testing.T) {
	mb1 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb2 := &Balance{Weight: 1, precision: 1, RatingSubject: "1", DestinationIDs: utils.StringMap{}}
	mb3 := &Balance{Weight: 1, precision: 1, RatingSubject: "2", DestinationIDs: utils.StringMap{}}
	if !mb1.Equal(mb2) || mb2.Equal(mb3) {
		t.Error("Equal failure!", mb1 == mb2, mb3)
	}
}

func TestBalanceClone(t *testing.T) {
	mb1 := &Balance{Value: dec.NewVal(1, 0), Weight: 2, RatingSubject: "test", DestinationIDs: utils.NewStringMap("5")}
	mb2 := mb1.Clone()
	if mb1 == mb2 || !mb1.Equal(mb2) {
		t.Errorf("Cloning failure: \n%+v\n%+v", mb1, mb2)
	}
}

func TestBalanceMatchActionTriggerId(t *testing.T) {
	at := &ActionTrigger{Filter: `{"ID":"test"}`}
	b := &Balance{ID: "test"}
	match, _ := b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = "test1"
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = ""
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.ID = "test"
	at.Filter = ``
	match, _ = b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerDestination(t *testing.T) {
	at := &ActionTrigger{Filter: `{"DestinationIDs":{"$has":["test"]}}`}
	b := &Balance{DestinationIDs: utils.NewStringMap("test")}
	match, _ := b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("test1")
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("")
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.DestinationIDs = utils.NewStringMap("test")
	at.Filter = ``
	match, _ = b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerWeight(t *testing.T) {
	at := &ActionTrigger{Filter: `{"Weight":1}`}
	b := &Balance{Weight: 1}
	match, _ := b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 2
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 0
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.Weight = 1
	at.Filter = ``
	match, _ = b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerRatingSubject(t *testing.T) {
	at := &ActionTrigger{Filter: `{"RatingSubject":"test"}`}
	b := &Balance{RatingSubject: "test"}
	match, _ := b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = "test1"
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = ""
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.RatingSubject = "test"
	at.Filter = ``
	match, _ = b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceMatchActionTriggerSharedGroup(t *testing.T) {
	at := &ActionTrigger{Filter: `{"SharedGroups":{"$has":["test"]}}`}
	b := &Balance{SharedGroups: utils.NewStringMap("test")}
	match, _ := b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("test1")
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("")
	match, _ = b.MatchActionTrigger(at)
	if match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
	b.SharedGroups = utils.NewStringMap("test")
	at.Filter = ``
	match, _ = b.MatchActionTrigger(at)
	if !match {
		t.Errorf("Error matching action trigger: %+v %+v", b, at)
	}
}

func TestBalanceIsDefault(t *testing.T) {
	b := &Balance{Weight: 0}
	if b.IsDefault() {
		t.Errorf("Balance should not be default: %+v", b)
	}
	b = &Balance{ID: utils.META_DEFAULT}
	if !b.IsDefault() {
		t.Errorf("Balance should be default: %+v", b)
	}
}
