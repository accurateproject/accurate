package engine

import "github.com/accurateproject/accurate/utils"

// Amount of a trafic of a certain type
type UnitCounter struct {
	CounterType string         // *event or *balance
	Counters    CounterFilters // first balance is the general one (no destination)
}

type CounterFilter struct {
	Value  float64
	Filter *BalanceFilter
}

type CounterFilters []*CounterFilter

func (cfs CounterFilters) HasCounter(cf *CounterFilter) bool {
	for _, c := range cfs {
		if c.Filter.Equal(cf.Filter) {
			return true
		}
	}
	return false
}

// Returns true if the counters were of the same type
// Copies the value from old balances
func (uc *UnitCounter) CopyCounterValues(oldUc *UnitCounter) bool {
	if uc.CounterType != oldUc.CounterType { // type check
		return false
	}
	for _, c := range uc.Counters {
		for _, oldC := range oldUc.Counters {
			if c.Filter.Equal(oldC.Filter) {
				c.Value = oldC.Value
				break
			}
		}
	}
	return true
}

type UnitCounters map[string][]*UnitCounter

func (ucs UnitCounters) addUnits(amount float64, kind string, cc *CallCost, b *Balance) {
	counters, found := ucs[kind]
	if !found {
		return
	}
	for _, uc := range counters {
		if uc == nil { // safeguard
			continue
		}
		if uc.CounterType == "" {
			uc.CounterType = utils.COUNTER_EVENT
		}
		for _, c := range uc.Counters {
			if uc.CounterType == utils.COUNTER_EVENT && cc != nil && cc.MatchCCFilter(c.Filter) {
				c.Value += amount
				continue
			}

			if uc.CounterType == utils.COUNTER_BALANCE && b != nil && b.MatchFilter(c.Filter, true) {
				c.Value += amount
				continue
			}
		}

	}
}

func (ucs UnitCounters) resetCounters(a *Action) {
	for key, counters := range ucs {
		if a != nil && a.Balance.Type != nil && a.Balance.GetType() != key {
			continue
		}
		for _, c := range counters {
			if c == nil { // safeguard
				continue
			}
			for _, cf := range c.Counters {
				if a == nil || a.Balance == nil || cf.Filter.Equal(a.Balance) {
					cf.Value = 0
				}
			}
		}
	}
}
