package engine

import (
	"reflect"
	"strconv"
	"time"

	"github.com/accurateproject/accurate/dec"
)

type Increment struct {
	Duration       time.Duration
	Cost           *dec.Dec
	BalanceInfo    *DebitInfo // need more than one for units with cost
	CompressFactor int
	PostATIDs      map[string][]string // compressed index where the postAT are attached, using string for bson
	paid           int                 // the amount of the compressed that is paid
}

func (i *Increment) getCost() *dec.Dec {
	if i.Cost == nil {
		i.Cost = dec.New()
	}
	return i.Cost
}

// Holds information about the balance that made a specific payment
type DebitInfo struct {
	Unit      *UnitInfo
	Monetary  *MonetaryInfo
	AccountID string // used when debited from shared balance
}

func (di *DebitInfo) Equal(other *DebitInfo) bool {
	return di.Unit.Equal(other.Unit) &&
		di.Monetary.Equal(other.Monetary) &&
		di.AccountID == other.AccountID
}

func (di *DebitInfo) Clone() *DebitInfo {
	nDi := &DebitInfo{
		AccountID: di.AccountID,
	}
	if di.Unit != nil {
		nDi.Unit = di.Unit.Clone()
	}
	if di.Monetary != nil {
		nDi.Monetary = di.Monetary.Clone()
	}
	return nDi
}

type MonetaryInfo struct {
	UUID         string
	ID           string
	Value        *dec.Dec
	RateInterval *RateInterval
}

func (mi *MonetaryInfo) getValue() *dec.Dec {
	if mi.Value == nil {
		mi.Value = dec.New()
	}
	return mi.Value
}

func (mi *MonetaryInfo) Clone() *MonetaryInfo {
	newMi := *mi
	return &newMi
}

func (mi *MonetaryInfo) Equal(other *MonetaryInfo) bool {
	if mi == nil && other == nil {
		return true
	}
	if mi == nil || other == nil {
		return false
	}
	return mi.UUID == other.UUID &&
		reflect.DeepEqual(mi.RateInterval, other.RateInterval)
}

type UnitInfo struct {
	UUID          string
	ID            string
	Value         *dec.Dec
	DestinationID string
	Consumed      string
	TOR           string
	RateInterval  *RateInterval
}

func (ui *UnitInfo) getValue() *dec.Dec {
	if ui.Value == nil {
		ui.Value = dec.New()
	}
	return ui.Value
}

func (ui *UnitInfo) Clone() *UnitInfo {
	newUi := *ui
	return &newUi
}

func (ui *UnitInfo) Equal(other *UnitInfo) bool {
	if ui == nil && other == nil {
		return true
	}
	if ui == nil || other == nil {
		return false
	}
	return ui.UUID == other.UUID &&
		ui.DestinationID == other.DestinationID &&
		ui.Consumed == other.Consumed &&
		ui.TOR == other.TOR &&
		reflect.DeepEqual(ui.RateInterval, other.RateInterval)
}

func (inc *Increment) AddPostATIDs(index int, pats []string) {
	if len(pats) == 0 {
		return
	}
	if inc.PostATIDs == nil {
		inc.PostATIDs = make(map[string][]string)
	}
	i := strconv.Itoa(index)
	postATIDs := inc.PostATIDs[i]
	postATIDs = append(postATIDs, pats...)
	inc.PostATIDs[i] = postATIDs
}

func (incr *Increment) Clone() *Increment {
	nInc := &Increment{
		Duration:  incr.Duration,
		Cost:      dec.New().Set(incr.getCost()),
		PostATIDs: incr.PostATIDs,
	}
	if incr.BalanceInfo != nil {
		nInc.BalanceInfo = incr.BalanceInfo.Clone()
	}
	return nInc
}

func (incr *Increment) Equal(other *Increment) bool {
	if len(incr.PostATIDs) != len(other.PostATIDs) {
		return false
	}
	if incr.PostATIDs != nil {
		for index, postATIDs := range incr.PostATIDs {
			otherPostATIDs, found := other.PostATIDs[index]
			if !found {
				return false
			}
			if len(postATIDs) != len(otherPostATIDs) {
				return false
			}
			for i := 0; i < len(postATIDs); i++ {
				if postATIDs[i] != otherPostATIDs[i] {
					return false
				}
			}
		}
	}
	return incr.Duration == other.Duration &&
		incr.getCost().Cmp(other.getCost()) == 0 &&
		((incr.BalanceInfo == nil && other.BalanceInfo == nil) || incr.BalanceInfo.Equal(other.BalanceInfo))
}

func (incr *Increment) GetTotalCost() *dec.Dec {
	return dec.NewVal(int64(incr.CompressFactor), 0).MulS(incr.Cost)
}

type Increments struct {
	CompIncrement    *Increment
	MaxCostFreeIndex int
	index            int // the index in the compressed list
}

func (ics *Increments) Len() (length int) {
	if ics.CompIncrement == nil {
		return 0
	}
	return ics.CompIncrement.CompressFactor
}

func (ics *Increments) Next() (int, *Increment) {
	if ics.CompIncrement == nil {
		return 0, nil
	}
	if ics.index < ics.CompIncrement.CompressFactor {
		ics.index++
		return ics.index, ics.CompIncrement
	}
	ics.index = 0 // reset
	return ics.index, nil
}

func (ics *Increments) Reset() {
	ics.index = 0
}

func (ics *Increments) Equal(other *Increments) bool {
	if ics.CompIncrement == nil && other.CompIncrement != nil ||
		ics.CompIncrement != nil && other.CompIncrement == nil {
		return false
	}
	if ics.CompIncrement == nil && other.CompIncrement == nil {
		return true
	}
	return ics.CompIncrement.Equal(other.CompIncrement) &&
		ics.CompIncrement.CompressFactor != other.CompIncrement.CompressFactor

}

func (ics *Increments) GetTotalCost() *dec.Dec {
	mulFactor := ics.CompIncrement.CompressFactor
	if ics.MaxCostFreeIndex != 0 {
		mulFactor = ics.MaxCostFreeIndex
	}
	return dec.New().Set(ics.CompIncrement.Cost).MulS(dec.NewVal(int64(mulFactor), 0))
}

/*
func (incs *Increments) Compress() { // must be pointer receiver
	var cIncrs Increments
	for _, incr := range *incs {
		if len(cIncrs) == 0 || !cIncrs[len(cIncrs)-1].Equal(incr) {
			if incr.CompressFactor == 0 {
				incr.CompressFactor = 1
			}
			cIncrs = append(cIncrs, incr)
		} else {
			lastCIncr := cIncrs[len(cIncrs)-1]
			lastCIncr.CompressFactor++
			if lastCIncr.BalanceInfo != nil && incr.BalanceInfo != nil {
				if lastCIncr.BalanceInfo.Monetary != nil && incr.BalanceInfo.Monetary != nil {
					lastCIncr.BalanceInfo.Monetary.Value = incr.BalanceInfo.Monetary.Value
				}
				if lastCIncr.BalanceInfo.Unit != nil && incr.BalanceInfo.Unit != nil {
					lastCIncr.BalanceInfo.Unit.Value = incr.BalanceInfo.Unit.Value
				}
			}
		}
	}
	*incs = cIncrs
}

func (incs *Increments) Decompress() { // must be pointer receiver
	var cIncrs Increments
	for _, cIncr := range *incs {
		cf := cIncr.GetCompressFactor()
		for i := 0; i < cf; i++ {
			incr := cIncr.Clone()
			// set right Values
			if incr.BalanceInfo != nil {
				if incr.BalanceInfo.Monetary != nil {
					monetaryValue, _ := dec.New().SetString(incr.BalanceInfo.Monetary.Value)
					monetaryValue.AddS(dec.NewVal(int64(cf-(i+1)), 0).MulS(incr.getCost()))
					incr.BalanceInfo.Monetary.Value = monetaryValue.String()
				}
				if incr.BalanceInfo.Unit != nil {
					consumed, _ := dec.New().SetString(incr.BalanceInfo.Unit.Consumed)
					unitValue, _ := dec.New().SetString(incr.BalanceInfo.Unit.Value)
					unitValue.AddS(dec.NewVal(int64(cf-(i+1)), 0).MulS(consumed))
					incr.BalanceInfo.Unit.Value = unitValue.String()
				}
			}
			cIncrs = append(cIncrs, incr)
		}
	}
	*incs = cIncrs
}
*/
