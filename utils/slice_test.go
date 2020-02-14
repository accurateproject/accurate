package utils

import (
	"sort"
	"testing"

	"github.com/accurateproject/accurate/dec"
)

func TestAvg(t *testing.T) {
	values := dec.DecSlice{dec.NewVal(1, 0), dec.NewVal(2, 0), dec.NewVal(3, 0)}
	result := Avg(values)
	expected := dec.NewVal(2, 0)
	if expected.Cmp(result) != 0 {
		t.Errorf("Wrong Avg: expected %v got %v", expected, result)
	}
}

func TestAvgEmpty(t *testing.T) {
	values := dec.DecSlice{}
	result := Avg(values)
	expected := dec.New()
	if expected.Cmp(result) != 0 {
		t.Errorf("Wrong Avg: expected %v got %v", expected, result)
	}
}

// ********************* bench ********************

func BenchmarkSearchMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := StringMap{"some": true, "values": true, "to": true, "search": true, "not": true, "too": true, "many": true}
		if a["test"] {

		}
	}
}

func BenchmarkSearchSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ss := []string{"some", "values", "to", "search", "not", "too", "many"}
		s := "test"
		sort.Strings(ss)
		if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {

		}
	}
}
