package engine

import (
	"errors"
	"strconv"
	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

// The output structure that will be returned with the call cost information.
type CallCost struct {
	Direction, Category, Tenant, Subject, Account, Destination, TOR string
	Cost                                                            *dec.Dec
	Timespans                                                       TimeSpans
	RatedUsage                                                      float64
	deductConnectFee                                                bool
	negativeConnectFee                                              bool // the connect fee went negative on default balance
	maxCostDisconect                                                bool
	postActionTrigger                                               bool
}

func (cc *CallCost) GetCost() *dec.Dec {
	if cc.Cost == nil {
		cc.Cost = dec.New()
	}
	return cc.Cost
}

// Merges the received timespan if they are similar (same activation period, same interval, same minute info.
func (cc *CallCost) Merge(other *CallCost) {
	cc.Timespans = append(cc.Timespans, other.Timespans...)
	cc.GetCost().AddS(other.GetCost())
}

func (cc *CallCost) GetStartTime() time.Time {
	if len(cc.Timespans) == 0 {
		return time.Now()
	}
	return cc.Timespans[0].TimeStart
}

func (cc *CallCost) GetEndTime() time.Time {
	if len(cc.Timespans) == 0 {
		return time.Now()
	}
	return cc.Timespans[len(cc.Timespans)-1].TimeEnd
}

func (cc *CallCost) GetDuration() (td time.Duration) {
	for _, ts := range cc.Timespans {
		td += ts.GetDuration()
	}
	return
}

func (cc *CallCost) UpdateRatedUsage() time.Duration {
	if cc == nil {
		return 0
	}
	totalDuration := cc.GetDuration()
	cc.RatedUsage = totalDuration.Seconds()
	return totalDuration
}

func (cc *CallCost) GetConnectFee() *dec.Dec {
	if len(cc.Timespans) == 0 ||
		cc.Timespans[0].RateInterval == nil ||
		cc.Timespans[0].RateInterval.Rating == nil {
		return dec.Zero
	}
	if cc.Timespans[0].RateInterval.Rating.ConnectFee == nil {
		cc.Timespans[0].RateInterval.Rating.ConnectFee = dec.New()
	}

	return cc.Timespans[0].RateInterval.Rating.ConnectFee
}

func (cc *CallCost) GetFirstIncrement() *Increment {
	if len(cc.Timespans) == 0 || cc.Timespans[0].Increments == nil {
		return nil
	}
	return cc.Timespans[0].Increments.CompIncrement
}

// Creates a CallDescriptor structure copying related data from CallCost
func (cc *CallCost) CreateCallDescriptor() *CallDescriptor {
	return &CallDescriptor{
		Direction:   cc.Direction,
		Category:    cc.Category,
		Tenant:      cc.Tenant,
		Subject:     cc.Subject,
		Account:     cc.Account,
		Destination: cc.Destination,
		TOR:         cc.TOR,
	}
}

func (cc *CallCost) IsPaid() bool {
	for _, ts := range cc.Timespans {
		if paid, _ := ts.IsPaid(); !paid {
			return false
		}
	}
	return true
}

func (cc *CallCost) ToDataCost() (*DataCost, error) {
	if cc.TOR == utils.VOICE {
		return nil, errors.New("Not a data call!")
	}
	dc := &DataCost{
		Direction:        cc.Direction,
		Category:         cc.Category,
		Tenant:           cc.Tenant,
		Subject:          cc.Subject,
		Account:          cc.Account,
		Destination:      cc.Destination,
		TOR:              cc.TOR,
		Cost:             dec.New().Set(cc.Cost),
		deductConnectFee: cc.deductConnectFee,
	}
	dc.DataSpans = make([]*DataSpan, len(cc.Timespans))
	for i, ts := range cc.Timespans {
		length := ts.TimeEnd.Sub(ts.TimeStart).Seconds()
		callDuration := ts.DurationIndex.Seconds()
		dc.DataSpans[i] = &DataSpan{
			DataStart:      callDuration - length,
			DataEnd:        callDuration,
			Cost:           ts.Cost,
			ratingInfo:     ts.ratingInfo,
			RateInterval:   ts.RateInterval,
			DataIndex:      callDuration,
			MatchedSubject: ts.MatchedSubject,
			MatchedPrefix:  ts.MatchedPrefix,
			MatchedDestID:  ts.MatchedDestID,
			RatingPlanID:   ts.RatingPlanID,
		}
		incr := ts.Increments.CompIncrement
		dc.DataSpans[i].Increments = &DataIncrements{
			CompIncrement: &DataIncrement{
				Amount:         dec.NewFloat(incr.Duration.Seconds()),
				Cost:           dec.New().Set(incr.Cost),
				BalanceInfo:    incr.BalanceInfo,
				CompressFactor: incr.CompressFactor,
				paid:           incr.paid,
			},
		}
	}
	return dc, nil
}

func (cc *CallCost) AsJSON() string {
	return utils.ToJSON(cc)
}

// public function to update final (merged) callcost
func (cc *CallCost) UpdateCost() {
	cc.deductConnectFee = true
	cc.updateCost()
}

func (cc *CallCost) updateCost() {
	cost := dec.New()
	if cc.deductConnectFee { // add back the connectFee
		cost.AddS(cc.GetConnectFee())
	}
	for _, ts := range cc.Timespans {
		ts.Cost = ts.CalculateCost()
		cost.AddS(ts.Cost)
	}
	cost.Round(globalRoundingDecimals)
	cc.Cost = cost
}

func (cc *CallCost) TruncateTimespansAtDuration(truncateDuration time.Duration) []*Increment {
	cc.Timespans.Decompress()
	var refundIncrements []*Increment
	for i := len(cc.Timespans) - 1; i >= 0; i-- {
		ts := cc.Timespans[i]
		tsDuration := ts.GetDuration()
		compInc := ts.Increments.CompIncrement.Clone()
		refundIncrements = append(refundIncrements, compInc)
		if truncateDuration <= tsDuration {
			lastRefundedIncrementIndex := 0
			for incIndex, increment := ts.Increments.Next(); increment != nil; incIndex, increment = ts.Increments.Next() {
				if increment.Duration <= truncateDuration {
					compInc.CompressFactor++
					truncateDuration -= increment.Duration
					lastRefundedIncrementIndex = incIndex
				} else {
					break //increment duration is larger, cannot refund increment
				}
			}
			if lastRefundedIncrementIndex == ts.Increments.CompIncrement.CompressFactor {
				cc.Timespans[i] = nil
				cc.Timespans = cc.Timespans[:i]
			} else {
				ts.SplitByIncrement(ts.Increments.CompIncrement.CompressFactor - lastRefundedIncrementIndex + 1)
				ts.Cost = ts.CalculateCost()
			}
			break // do not go to other timespans
		} else {
			compInc.CompressFactor = ts.Increments.CompIncrement.CompressFactor
			// remove the timespan entirely
			cc.Timespans[i] = nil
			cc.Timespans = cc.Timespans[:i]
			// continue to the next timespan with what is left to refund
			truncateDuration -= tsDuration
		}
	}
	return refundIncrements
}

func (cc *CallCost) GetPostActionTriggers(splitDuration time.Duration) (exe, unexe map[string][]string) {
	exe = make(map[string][]string)
	unexe = make(map[string][]string)
	cc.Timespans.Decompress()
	var durationSoFar time.Duration
	for _, ts := range cc.Timespans {
		for index, incr := ts.Increments.Next(); incr != nil; index, incr = ts.Increments.Next() {
			if incr.PostATIDs != nil {
				if postATIDs, ok := incr.PostATIDs[strconv.Itoa(index)]; ok {
					if durationSoFar < splitDuration {
						exe[incr.BalanceInfo.AccountID] = append(exe[incr.BalanceInfo.AccountID], postATIDs...)
					} else {
						unexe[incr.BalanceInfo.AccountID] = append(unexe[incr.BalanceInfo.AccountID], postATIDs...)
					}
				}
			}
			durationSoFar += incr.Duration
		}
	}
	return
}

func (cc *CallCost) MatchCCFilter(filter *utils.StructQ) (bool, error) {
	if filter == nil {
		return true, nil
	}
	a := &struct {
		Categories     utils.StringMap
		Directions     utils.StringMap
		DestinationIDs utils.StringMap
	}{
		Categories:     utils.StringMap{},
		Directions:     utils.StringMap{},
		DestinationIDs: utils.StringMap{},
	}
	if cc.Category != "" {
		a.Categories[cc.Category] = true
	}
	if cc.Direction != "" {
		a.Directions[cc.Direction] = true
	}

	// match destination ids
	if cc.Destination != "" {
		if dests, err := ratingStorage.GetDestinations(cc.Tenant, cc.Destination, "", utils.DestMatching, utils.CACHED); err == nil {
			for _, dest := range dests {
				a.DestinationIDs[dest.Name] = true
			}
		}
	}
	return filter.Query(a, false)
}
