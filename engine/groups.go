package engine

import (
	"sort"
)

type GroupLink struct {
	Id     string
	Weight float64
}

type GroupLinks []*GroupLink

func (gls GroupLinks) Sort() {
	sort.Slice(gls, func(j, i int) bool {
		return gls[i].Weight < gls[j].Weight
	})
}
