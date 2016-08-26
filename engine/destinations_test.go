package engine

import (
	"encoding/json"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/utils"

	"testing"
)

func TestDestinationStoreRestore(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	s, _ := json.Marshal(nationale)
	d1 := &Destination{Id: "nat"}
	json.Unmarshal(s, d1)
	s1, _ := json.Marshal(d1)
	if string(s1) != string(s) {
		t.Errorf("Expected %q was %q", s, s1)
	}
}

func TestDestinationStorageStore(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	err := ratingStorage.SetDestination(nationale)
	if err != nil {
		t.Error("Error storing destination: ", err)
	}
	result, err := ratingStorage.GetDestination(nationale.Id, utils.CACHED)
	if nationale.containsPrefix("0257") == 0 || nationale.containsPrefix("0256") == 0 || nationale.containsPrefix("0723") == 0 {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

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

func TestDestinationGetExists(t *testing.T) {
	d, err := ratingStorage.GetDestination("NAT", utils.CACHED)
	if err != nil || d == nil {
		t.Error("Could not get destination: ", d)
	}
}

func TestDestinationReverseGetExistsCache(t *testing.T) {
	ratingStorage.GetReverseDestination("0256", utils.CACHED)
	if _, ok := cache2go.Get(utils.REVERSE_DESTINATION_PREFIX + "0256"); !ok {
		t.Error("Destination not cached:", err)
	}
}

func TestDestinationGetNotExists(t *testing.T) {
	d, err := ratingStorage.GetDestination("not existing", utils.CACHED)
	if d != nil {
		t.Error("Got false destination: ", d, err)
	}
}

func TestDestinationGetNotExistsCache(t *testing.T) {
	ratingStorage.GetDestination("not existing", utils.CACHED)
	if d, ok := cache2go.Get("not existing"); ok {
		t.Error("Bad destination cached: ", d)
	}
}

func TestCachedDestHasPrefix(t *testing.T) {
	if !CachedDestHasPrefix("NAT", "0256") {
		t.Error("Could not find prefix in destination")
	}
}

func TestCachedDestHasWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("NAT", "771") {
		t.Error("Prefix should not belong to destination")
	}
}

func TestNonCachedDestRightPrefix(t *testing.T) {
	if CachedDestHasPrefix("FAKE", "0256") {
		t.Error("Destination should not belong to prefix")
	}
}

func TestNonCachedDestWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("FAKE", "771") {
		t.Error("Both arguments should be fake")
	}
}

/*
func TestCleanStalePrefixes(t *testing.T) {
	x := struct{}{}
	cache2go.Set(utils.DESTINATION_PREFIX+"1", map[string]struct{}{"D1": x, "D2": x})
	cache2go.Set(utils.DESTINATION_PREFIX+"2", map[string]struct{}{"D1": x})
	cache2go.Set(utils.DESTINATION_PREFIX+"3", map[string]struct{}{"D2": x})
	CleanStalePrefixes([]string{"D1"})
	if r, ok := cache2go.Get(utils.DESTINATION_PREFIX + "1"); !ok || len(r.(map[string]struct{})) != 1 {
		t.Error("Error cleaning stale destination ids", r)
	}
	if r, ok := cache2go.Get(utils.DESTINATION_PREFIX + "2"); ok {
		t.Error("Error removing stale prefix: ", r)
	}
	if r, ok := cache2go.Get(utils.DESTINATION_PREFIX + "3"); !ok || len(r.(map[string]struct{})) != 1 {
		t.Error("Error performing stale cleaning: ", r)
	}
}*/

/********************************* Benchmarks **********************************/

func BenchmarkDestinationStorageStoreRestore(b *testing.B) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		ratingStorage.SetDestination(nationale)
		ratingStorage.GetDestination(nationale.Id, utils.CACHE_SKIP)
	}
}
