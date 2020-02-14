package engine

import (
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

// Amount of a trafic of a certain type
type UnitCounter struct {
	CounterType string         // *event or *balance
	Counters    CounterFilters // first balance is the general one (no destination)
}

type CounterFilter struct {
	UniqueID string
	Value    *dec.Dec
	Filter   string
	filter   *utils.StructQ
}

func (cf *CounterFilter) getValue() *dec.Dec {
	if cf.Value == nil {
		cf.Value = dec.New()
	}
	return cf.Value
}

func (cf *CounterFilter) getFilter() *utils.StructQ {
	if cf.filter != nil {
		return cf.filter
	}
	// ignore error as its hould be checked at load time
	cf.filter, _ = utils.NewStructQ(cf.Filter)
	return cf.filter
}

type CounterFilters []*CounterFilter

func (cfs CounterFilters) HasCounter(cf *CounterFilter) bool {
	for _, c := range cfs {
		if c.UniqueID == cf.UniqueID {
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
			if c.Filter == oldC.Filter {
				c.Value = oldC.Value
				break
			}
		}
	}
	return true
}

type UnitCounters map[string][]*UnitCounter

func (ucs UnitCounters) addUnits(amount *dec.Dec, kind string, cc *CallCost, b *Balance) {
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
			if uc.CounterType == utils.COUNTER_EVENT && cc != nil {
				match, err := cc.MatchCCFilter(c.getFilter())
				if err != nil {
					utils.Logger.Error("<addUnits> counter filter error", zap.String("filter", c.Filter), zap.Error(err))
				}
				if match {
					c.getValue().AddS(amount)
					continue
				}
			}

			if uc.CounterType == utils.COUNTER_BALANCE && b != nil {
				match, err := c.getFilter().Query(b, false)
				if err != nil {
					utils.Logger.Error("<addUnits> counter filter error", zap.String("filter", c.Filter), zap.Error(err))
				}
				if match {
					c.getValue().AddS(amount)
					continue
				}
			}
		}

	}
}

func (ucs UnitCounters) resetCounters(a *Action) {
	for key, counters := range ucs {
		if a != nil && a.TOR != "" && a.TOR != key {
			continue
		}
		for _, c := range counters {
			if c == nil { // safeguard
				continue
			}
			for _, cf := range c.Counters {
				var match bool
				var err error
				if a != nil && a.Filter1 != "" {
					match, err = a.getFilter().Query(cf, false)
					if err != nil {
						utils.Logger.Warn("<resetCounters> action ", zap.String("filter", a.Filter1), zap.Error(err))
					}
				}
				if a == nil || a.Filter1 == "" || match {
					cf.getValue().Set(dec.Zero)
				}
			}
		}
	}
}
