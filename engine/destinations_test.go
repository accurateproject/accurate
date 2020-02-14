package engine

import (
	"encoding/json"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/utils"

	"testing"
)

func TestDestinationStoreRestore(t *testing.T) {
	nationale := &Destination{Tenant: "test", Name: "nat", Code: "0257"}
	s, _ := json.Marshal(nationale)
	d1 := &Destination{Name: "nat"}
	json.Unmarshal(s, d1)
	s1, _ := json.Marshal(d1)
	if string(s1) != string(s) {
		t.Errorf("Expected %q was %q", s, s1)
	}
}

func TestDestinationStorageStore(t *testing.T) {
	nationale1 := &Destination{Tenant: "test", Name: "nat", Code: "0257"}
	nationale2 := &Destination{Tenant: "test", Name: "nat", Code: "0256"}
	nationale3 := &Destination{Tenant: "test", Name: "nat", Code: "0723"}

	if err := ratingStorage.SetDestination(nationale1); err != nil {
		t.Error("Error storing destination: ", err)
	}
	if err := ratingStorage.SetDestination(nationale2); err != nil {
		t.Error("Error storing destination: ", err)
	}
	if err := ratingStorage.SetDestination(nationale3); err != nil {
		t.Error("Error storing destination: ", err)
	}

	result, err := ratingStorage.GetDestinations("test", "", nationale1.Name, utils.DestExact, utils.CACHED)
	if err != nil {
		t.Error("error getting destinations: ", err)
	}
	if result[0].Code != "0257" || result[1].Code != "0256" || result[2].Code != "0723" {
		t.Errorf("bad destinations back: %v", result)
	}
}

/*
func TestDestinationContainsPrefix(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("0256")
	if precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}
}

func TestDestinationContainsPrefixLong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("0256723045")
	if precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}
}

func TestDestinationContainsPrefixWrong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("01234567")
	if precision != 0 {
		t.Error("Should not contain prefix: ", nationale)
	}
}
*/

func TestDestinationGetExists(t *testing.T) {
	d, err := ratingStorage.GetDestinations("test", "", "NAT", utils.DestExact, utils.CACHED)
	if err != nil || d == nil || len(d) == 0 {
		t.Error("Could not get destination: ", d)
	}
}

func TestDestinationGetExistsCache(t *testing.T) {
	ratingStorage.GetDestinations("test", "0256", "", utils.DestExact, utils.CACHED)
	if _, ok := cache2go.Get("test", utils.DESTINATION_PREFIX+utils.ConcatKey("0256", "", utils.DestExact)); !ok {
		t.Error("Destination not cached")
	}
}

func TestDestinationGetCodeNotExists(t *testing.T) {
	d, err := ratingStorage.GetDestinations("test", "not existing", "", utils.DestExact, utils.CACHED)
	if len(d) != 0 {
		t.Error("Got false destination: ", d, err)
	}
}

func TestDestinationGetNameNotExists(t *testing.T) {
	d, err := ratingStorage.GetDestinations("test", "", "not existing", utils.DestExact, utils.CACHED)
	if len(d) != 0 {
		t.Error("Got false destination: ", d, err)
	}
}

func TestDestinationGetNotExistsCache(t *testing.T) {
	ratingStorage.GetDestinations("test", "not existing", "", utils.DestExact, utils.CACHED)
	if d, ok := cache2go.Get("test", "not existing"); ok {
		t.Error("Bad destination cached: ", d)
	}
}

func TestDestinationCachedDestHasPrefix(t *testing.T) {
	if !CachedDestHasPrefix("test", "NAT", "0256") {
		t.Error("Could not find prefix in destination")
	}
}

func TestDestinationCachedDestHasWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("test", "NAT", "771") {
		t.Error("Prefix should not belong to destination")
	}
}

func TestDestinationNonCachedDestRightPrefix(t *testing.T) {
	if CachedDestHasPrefix("test", "FAKE", "0256") {
		t.Error("Destination should not belong to prefix")
	}
}

func TestDestinationNonCachedDestWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("test", "FAKE", "771") {
		t.Error("Both arguments should be fake")
	}
}

func TestDestinationStartegyMatching(t *testing.T) {
	dests, err := ratingStorage.GetDestinations("test", "0723045326", "", utils.DestMatching, utils.CACHED)
	if err != nil {
		t.Fatal(err)
	}
	if len(dests) != 4 {
		t.Error("error getting matching destinations: ", utils.ToIJSON(dests))
	}
	if dests.getBest().Code != "0723045" {
		t.Error("bad ordering in get destinations: ", utils.ToIJSON(dests))
	}
}

func TestDestinationStartegyMatchingMultiple(t *testing.T) {
	if err := ratingStorage.SetDestination(&Destination{Tenant: "test", Code: "0723085835", Name: "Liana"}); err != nil {
		t.Fatal(err)
	}
	if err := ratingStorage.SetDestination(&Destination{Tenant: "test", Code: "0723045326", Name: "Radu"}); err != nil {
		t.Fatal(err)
	}
	dests, err := ratingStorage.GetDestinations("test", "0723045326", "", utils.DestMatching, utils.CACHED)
	if err != nil {
		t.Fatal(err)
	}
	if len(dests) != 5 {
		t.Error("error getting matching destinations: ", utils.ToIJSON(dests))
	}
	if dests.getBest().Code != "0723045326" {
		t.Error("bad ordering in get destinations: ", utils.ToIJSON(dests))
	}
}

/********************************* Benchmarks **********************************/

func BenchmarkDestinationStorageStoreRestore(b *testing.B) {
	nationale1 := &Destination{Tenant: "test", Name: "nat", Code: "0257"}
	nationale2 := &Destination{Tenant: "test", Name: "nat", Code: "0256"}
	nationale3 := &Destination{Tenant: "test", Name: "nat", Code: "0723"}

	for i := 0; i < b.N; i++ {
		if err := ratingStorage.SetDestination(nationale1); err != nil {
			b.Error("Error storing destination: ", err)
		}
		if err := ratingStorage.SetDestination(nationale2); err != nil {
			b.Error("Error storing destination: ", err)
		}
		if err := ratingStorage.SetDestination(nationale3); err != nil {
			b.Error("Error storing destination: ", err)
		}
		ratingStorage.GetDestinations("test", "", "nat", utils.DestExact, utils.CACHE_SKIP)
		ratingStorage.GetDestinations("test", "0256", "nat", utils.DestExact, utils.CACHE_SKIP)
		ratingStorage.GetDestinations("test", "0257", "", utils.DestExact, utils.CACHE_SKIP)
	}
}
