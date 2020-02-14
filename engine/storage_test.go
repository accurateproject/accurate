package engine

import (
	"testing"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/utils"
)

func TestStorageDestinationContainsPrefixShort(t *testing.T) {
	dests, err := ratingStorage.GetDestinations("test", "", "NAT", utils.DestExact, utils.CACHE_SKIP)
	if err != nil || len(dests) != 4 {
		t.Error("Error finding prefix: ", err, dests)
	}
}

func TestStorageCacheRefresh(t *testing.T) {
	ratingStorage.SetDestination(&Destination{Tenant: "ts", Code: "0", Name: "T11"})
	ratingStorage.GetDestinations("ts", "", "T11", utils.DestExact, utils.CACHED)
	ratingStorage.SetDestination(&Destination{Tenant: "ts", Code: "1", Name: "T11"})
	err := ratingStorage.PreloadRatingCache()
	if err != nil {
		t.Error("Error cache rating: ", err)
	}
	d, err := ratingStorage.GetDestinations("ts", "", "T11", utils.DestExact, utils.CACHED)
	if err != nil || len(d) != 2 {
		t.Error("Error refreshing cache:", d)
	}
}

func TestStorageGetAliases(t *testing.T) {
	ala := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   utils.ALIAS_CONTEXT_RATING,
		Values: AliasValues{
			&AliasValue{
				Fields:        `{"Subject":{"$rpl":["b1", "aaa"]}}`,
				Weight:        10,
				DestinationID: utils.ANY,
			},
		},
	}
	alb := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   "*other",
		Values: AliasValues{
			&AliasValue{
				Fields:        `{"Account":{"$rpl":["b1", "aaa"]}}`,
				Weight:        10,
				DestinationID: utils.ANY,
			},
		},
	}
	accountingStorage.SetAlias(ala)
	accountingStorage.SetAlias(alb)
	foundAlias, err := accountingStorage.GetAlias("*out", "vdf", "0", "b1", "b1", utils.ALIAS_CONTEXT_RATING, utils.CACHE_SKIP)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = accountingStorage.GetAlias("*out", "vdf", "0", "b1", "b1", "*other", utils.CACHE_SKIP)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = accountingStorage.GetAlias("*out", "vdf", "0", "b1", "b1", utils.ALIAS_CONTEXT_RATING, utils.CACHED)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
	foundAlias, err = accountingStorage.GetAlias("*out", "vdf", "0", "b1", "b1", "*other", utils.CACHED)
	if err != nil || len(foundAlias.Values) != 1 {
		t.Errorf("Alias get error %+v, %v: ", foundAlias, err)
	}
}

func TestStorageCacheRemoveCachedAliases(t *testing.T) {
	ala := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   utils.ALIAS_CONTEXT_RATING,
	}
	alb := &Alias{
		Direction: "*out",
		Tenant:    "vdf",
		Category:  "0",
		Account:   "b1",
		Subject:   "b1",
		Context:   "*other",
	}
	accountingStorage.RemoveAlias("*out", "vdf", "0", "b1", "b1", utils.ALIAS_CONTEXT_RATING)
	accountingStorage.RemoveAlias("*out", "vdf", "0", "b1", "b1", "*other")

	if _, ok := cache2go.Get(ala.Tenant, utils.ALIASES_PREFIX+ala.FullID()); ok {
		t.Error("Error removing cached alias: ", ok)
	}
	if _, ok := cache2go.Get(ala.Tenant, utils.ALIASES_PREFIX+alb.FullID()); ok {
		t.Error("Error removing cached alias: ", ok)
	}
}

func TestStorageDisabledAccount(t *testing.T) {
	acc, err := accountingStorage.GetAccount("test", "alodis")
	if err != nil || acc == nil {
		t.Error("Error loading disabled user account: ", err, acc)
	}
	if acc.Disabled != true || acc.AllowNegative != true {
		t.Errorf("Error loading user account properties: %+v", acc)
	}
}

// Install fails to detect them and starting server will panic, these tests will fix this
func TestStoreInterfaces(t *testing.T) {
	mdb := new(MongoStorage)
	var _ RatingStorage = mdb
	var _ AccountingStorage = mdb
	var _ CdrStorage = mdb
}

func TestDifferentUuid(t *testing.T) {
	a1, err := accountingStorage.GetAccount("test", "12345")
	if err != nil {
		t.Error("Error getting account: ", err)
	}
	a2, err := accountingStorage.GetAccount("test", "123456")
	if err != nil {
		t.Error("Error getting account: ", err)
	}
	if a1.BalanceMap[utils.VOICE][0].UUID == a2.BalanceMap[utils.VOICE][0].UUID ||
		a1.BalanceMap[utils.MONETARY][0].UUID == a2.BalanceMap[utils.MONETARY][0].UUID {
		t.Errorf("Identical uuids in different accounts: %+v <-> %+v", a1.BalanceMap[utils.VOICE][0], a1.BalanceMap[utils.MONETARY][0])
	}
}

func TestStorageTask(t *testing.T) {
	// clean previous unused tasks
	for i := 0; i < 21; i++ {
		ratingStorage.PopTask()
	}

	if err := ratingStorage.PushTask(&Task{UUID: "1"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := ratingStorage.PushTask(&Task{UUID: "2"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := ratingStorage.PushTask(&Task{UUID: "3"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if err := ratingStorage.PushTask(&Task{UUID: "4"}); err != nil {
		t.Error("Error pushing task: ", err)
	}
	if task, err := ratingStorage.PopTask(); err != nil && task.UUID != "1" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := ratingStorage.PopTask(); err != nil && task.UUID != "2" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := ratingStorage.PopTask(); err != nil && task.UUID != "3" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := ratingStorage.PopTask(); err != nil && task.UUID != "4" {
		t.Error("Error poping task: ", task, err)
	}
	if task, err := ratingStorage.PopTask(); err == nil && task != nil {
		t.Errorf("Error poping task %+v, %v ", task, err)
	}
}

func TestFakeAPBIteratorAll(t *testing.T) {
	apbi := NewFakeAPBIterator("t1", "apl1", []string{"acc1", "acc2"})
	apbs := make([]*ActionPlanBinding, 0)
	apbi.All(&apbs)
	if len(apbs) != 2 ||
		apbs[0].Account != "acc1" ||
		apbs[1].Account != "acc2" {
		t.Error("failed to return all apbs from iterator: ", utils.ToIJSON(apbs))
	}
	if err := apbi.Close(); err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestFakeAPBIteratorAllEmpty(t *testing.T) {
	apbi := NewFakeAPBIterator("t1", "apl1", []string{})
	apbs := make([]*ActionPlanBinding, 0)
	apbi.All(&apbs)
	if len(apbs) != 0 {
		t.Error("failed to return all apbs from iterator: ", utils.ToIJSON(apbs))
	}
	if err := apbi.Close(); err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestFakeAPBIterator(t *testing.T) {
	apbi := NewFakeAPBIterator("t1", "apl1", []string{"acc1", "acc2"})
	var apb ActionPlanBinding
	x := apbi.Next(&apb)
	if !x || apb.Tenant != "t1" ||
		apb.Account != "acc1" ||
		apb.ActionPlan != "apl1" {
		t.Error("unexpected action plan binding: ", utils.ToIJSON(apb))
	}
	if apbi.Done() {
		t.Error("should not return Done!")
	}

	x = apbi.Next(&apb)
	if !x || apb.Tenant != "t1" ||
		apb.Account != "acc2" ||
		apb.ActionPlan != "apl1" {
		t.Error("unexpected action plan binding: ", utils.ToIJSON(apb))
	}
	if !apbi.Done() {
		t.Error("should return Done!")
	}
	x = apbi.Next(&apb)
	if x {
		t.Error("unexepected extra iterator next!")
	}

	if err := apbi.Close(); err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestFakeAPBIteratorEmpty(t *testing.T) {
	apbi := NewFakeAPBIterator("t1", "apl1", []string{})
	var apb ActionPlanBinding

	if !apbi.Done() {
		t.Error("should return Done!")
	}
	x := apbi.Next(&apb)
	if x {
		t.Error("unexepected extra iterator next!")
	}

	if err := apbi.Close(); err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestPushPopTask(t *testing.T) {
	var initialCount int
	var err error
	if initialCount, err = ratingStorage.Count(ColTsk); err != nil {
		t.Errorf("error counting tasks: %d, %v", initialCount, err)
	}
	t.Log("initial count: ", initialCount)
	if err := ratingStorage.PushTask(&Task{}); err != nil {
		t.Fatal(err)
	}
	if err := ratingStorage.PushTask(&Task{}); err != nil {
		t.Fatal(err)
	}
	if err := ratingStorage.PushTask(&Task{}); err != nil {
		t.Fatal(err)
	}
	if count, err := ratingStorage.Count(ColTsk); err != nil || count != initialCount+3 {
		t.Errorf("error pushing tasks: %d, %v", count, err)
	}

	if task, err := ratingStorage.PopTask(); err != nil || task == nil {
		t.Fatal(err)
	}
	if task, err := ratingStorage.PopTask(); err != nil || task == nil {
		t.Fatal(err)
	}
	if task, err := ratingStorage.PopTask(); err != nil || task == nil {
		t.Fatal(err)
	}
	if count, err := ratingStorage.Count(ColTsk); err != nil || count != initialCount {
		t.Errorf("error poping tasks: %d, %v", count, err)
	}
}
