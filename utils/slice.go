package utils

import (
	"sort"
	"strings"

	"github.com/accurateproject/accurate/dec"
)

// Binary string search in slice
func IsSliceMember(ss []string, s string) bool {
	sort.Strings(ss)
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		return true
	}
	return false
}

func SliceWithoutMember(ss []string, s string) []string {
	sort.Strings(ss)
	if i := sort.SearchStrings(ss, s); i < len(ss) && ss[i] == s {
		ss[i], ss = ss[len(ss)-1], ss[:len(ss)-1]
	}
	return ss
}

//Iterates over slice members and returns true if one starts with prefix
func SliceMemberHasPrefix(ss []string, prfx string) bool {
	for _, mbr := range ss {
		if strings.HasPrefix(mbr, prfx) {
			return true
		}
	}
	return false
}

func Avg(values dec.DecSlice) *dec.Dec {
	if len(values) == 0 {
		return dec.New()
	}
	sum := dec.New()
	for _, val := range values {
		sum.AddS(val)
	}
	return dec.New().Quo(sum, dec.NewVal(int64(len(values)), 0))
}

func AvgNegative(values dec.DecSlice) *dec.Dec {
	if len(values) == 0 {
		return dec.NewVal(-1, 0) // return -1 if no data
	}
	return Avg(values)
}
