
package engine

import (
	"sort"
)

type GroupLink struct {
	Id     string
	Weight float64
}

type GroupLinks []*GroupLink

func (gls GroupLinks) Len() int {
	return len(gls)
}

func (gls GroupLinks) Swap(i, j int) {
	gls[i], gls[j] = gls[j], gls[i]
}

func (gls GroupLinks) Less(j, i int) bool {
	return gls[i].Weight < gls[j].Weight
}

func (gls GroupLinks) Sort() {
	sort.Sort(gls)
}
