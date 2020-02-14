package engine

import (
	//"fmt"

	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

/*
A unit in which a call will be split that has a specific price related interval attached to it.
*/
type TimeSpan struct {
	TimeStart, TimeEnd                                         time.Time
	Cost                                                       *dec.Dec
	RateInterval                                               *RateInterval
	DurationIndex                                              time.Duration // the call duration so far till TimeEnd
	Increments                                                 *Increments
	MatchedSubject, MatchedPrefix, MatchedDestID, RatingPlanID string
	CompressFactor                                             int
	ratingInfo                                                 *RatingInfo
}

func (ts *TimeSpan) getCost() *dec.Dec {
	if ts.Cost == nil {
		ts.Cost = dec.New()
	}
	return ts.Cost
}

type TimeSpans []*TimeSpan

// Will delete all timespans that are `under` the timespan at index
func (timespans *TimeSpans) RemoveOverlapedFromIndex(index int) {
	tsList := *timespans
	ts := tsList[index]
	endOverlapIndex := index
	for i := index + 1; i < len(tsList); i++ {
		if tsList[i].TimeEnd.Before(ts.TimeEnd) || tsList[i].TimeEnd.Equal(ts.TimeEnd) {
			endOverlapIndex = i
		} else if tsList[i].TimeStart.Before(ts.TimeEnd) {
			tsList[i].TimeStart = ts.TimeEnd
			break
		}
	}
	if endOverlapIndex > index {
		newSliceEnd := len(tsList) - (endOverlapIndex - index)
		// delete overlapped
		copy(tsList[index+1:], tsList[endOverlapIndex+1:])
		for i := newSliceEnd; i < len(tsList); i++ {
			tsList[i] = nil
		}
		*timespans = tsList[:newSliceEnd]
		return
	}
	*timespans = tsList
}

func (tss *TimeSpans) Compress() { // must be pointer receiver
	var cTss TimeSpans
	for _, ts := range *tss {
		if len(cTss) == 0 || !cTss[len(cTss)-1].Equal(ts) {
			if ts.CompressFactor == 0 {
				ts.CompressFactor = 1
			}
			cTss = append(cTss, ts)
		} else {
			cTs := cTss[len(cTss)-1]
			cTs.CompressFactor++
			cTs.getCost().AddS(ts.Cost)
			cTs.TimeEnd = ts.TimeEnd
			cTs.DurationIndex = ts.DurationIndex
		}
	}
	*tss = cTss
}

func (tss *TimeSpans) Decompress() { // must be pointer receiver
	var cTss TimeSpans
	for _, cTs := range *tss {
		var duration time.Duration
		if cTs.GetCompressFactor() > 1 {
			duration = cTs.GetUnitDuration()
		}
		for i := cTs.GetCompressFactor(); i > 1; i-- {
			uTs := &TimeSpan{}
			*uTs = *cTs          // cloned by copy
			uTs.Cost = dec.New() // avoid aliasing
			uTs.TimeEnd = cTs.TimeStart.Add(duration)
			uTs.DurationIndex = cTs.DurationIndex - time.Duration((i-1)*int(duration))
			uTs.CompressFactor = 1
			uTs.getCost().Quo(cTs.getCost(), dec.NewVal(int64(cTs.GetCompressFactor()), 0))
			cTs.TimeStart = uTs.TimeEnd
			cTss = append(cTss, uTs)
		}
		cTs.Cost = cTs.GetUnitCost()
		cTs.CompressFactor = 1
		cTss = append(cTss, cTs)
	}
	*tss = cTss
}

// Returns the duration of the timespan
func (ts *TimeSpan) GetDuration() time.Duration {
	if ts == nil {
		return 0
	}
	return ts.TimeEnd.Sub(ts.TimeStart)
}

//Returns the duration of a unitary timespan in a compressed set
func (ts *TimeSpan) GetUnitDuration() time.Duration {
	return time.Duration(int(ts.TimeEnd.Sub(ts.TimeStart)) / ts.GetCompressFactor())
}

func (ts *TimeSpan) GetUnitCost() *dec.Dec {
	cost := dec.New().Set(ts.getCost())
	return cost.QuoS(dec.NewVal(int64(ts.GetCompressFactor()), 0))
}

// Returns true if the given time is inside timespan range.
func (ts *TimeSpan) Contains(t time.Time) bool {
	return t.After(ts.TimeStart) && t.Before(ts.TimeEnd)
}

func (ts *TimeSpan) SetRateInterval(interval *RateInterval) {
	if interval == nil {
		return
	}
	if !ts.hasBetterRateIntervalThan(interval) {
		ts.RateInterval = interval
	}
}

// Returns the cost of the timespan according to the relevant cost interval.
// It also sets the Cost field of this timespan (used for refund on session
// manager debit loop where the cost cannot be recalculated)
func (ts *TimeSpan) CalculateCost() *dec.Dec {
	if ts.Increments == nil || ts.Increments.Len() == 0 {
		if ts.RateInterval == nil {
			return dec.New()
		}
		return ts.RateInterval.GetCost(ts.GetDuration(), ts.GetGroupStart())
	}
	return ts.Increments.GetTotalCost().MulS(dec.NewVal(int64(ts.GetCompressFactor()), 0))
}

func (ts *TimeSpan) setRatingInfo(rp *RatingInfo) {
	ts.ratingInfo = rp
	ts.MatchedSubject = rp.MatchedSubject
	ts.MatchedPrefix = rp.MatchedPrefix
	ts.MatchedDestID = rp.MatchedDestID
	ts.RatingPlanID = rp.RatingPlanID
}

func (ts *TimeSpan) createIncrementsSlice() {
	ts.Increments = &Increments{}
	if ts.RateInterval == nil {
		return
	}
	// create rated units series
	_, rateIncrement, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
	// we will use the calculated cost and devide by nb of increments
	// because ts cost is rounded
	//incrementCost := rate / rateUnit.Seconds() * rateIncrement.Seconds()
	nbIncrements := int(ts.GetDuration() / rateIncrement)
	nbIncrementsDec := dec.NewVal(int64(nbIncrements), 0)
	incrementCost := ts.CalculateCost().QuoS(nbIncrementsDec)
	ts.Increments.CompIncrement = &Increment{
		Duration:       rateIncrement,
		Cost:           dec.New().Set(incrementCost),
		BalanceInfo:    &DebitInfo{},
		CompressFactor: nbIncrements,
	}
	ts.Cost = incrementCost.MulS(nbIncrementsDec)
}

// returns whether the timespan has all increments marked as paid and if not
// it also returns the first unpaied increment
func (ts *TimeSpan) IsPaid() (bool, int) {
	if ts.Increments.Len() == 0 {
		return false, 0
	}
	if ts.Increments.CompIncrement.paid < ts.Increments.CompIncrement.CompressFactor {
		return false, ts.Increments.CompIncrement.paid + 1
	}
	return true, ts.Increments.Len()
}

/*
Splits the given timespan according to how it relates to the interval.
It will modify the endtime of the received timespan and it will return
a new timespan starting from the end of the received one.
The interval will attach itself to the timespan that overlaps the interval.
*/
func (ts *TimeSpan) SplitByRateInterval(i *RateInterval, data bool) (nts *TimeSpan) {
	// if the span is not in interval return nil
	//log.Printf("Checking: %+v (%v,%v)", i.Timing, ts.TimeStart, ts.TimeEnd)
	if !(i.Contains(ts.TimeStart, false) || i.Contains(ts.TimeEnd, true)) {
		//log.Print("Not in interval")
		return
	}
	//Logger.Debug(fmt.Sprintf("TS: %+v", ts))
	// split by GroupStart
	if i.Rating != nil {
		i.Rating.Rates.Sort()
		for _, rate := range i.Rating.Rates {
			//Logger.Debug(fmt.Sprintf("Rate: %+v", rate))
			if ts.GetGroupStart() < rate.GroupIntervalStart && ts.GetGroupEnd() > rate.GroupIntervalStart {
				//log.Print("Splitting")
				ts.SetRateInterval(i)
				splitTime := ts.TimeStart.Add(rate.GroupIntervalStart - ts.GetGroupStart())
				nts = &TimeSpan{
					TimeStart: splitTime,
					TimeEnd:   ts.TimeEnd,
				}
				nts.copyRatingInfo(ts)
				ts.TimeEnd = splitTime
				nts.SetRateInterval(i)
				nts.DurationIndex = ts.DurationIndex
				ts.SetNewDurationIndex(nts)
				// Logger.Debug(fmt.Sprintf("Group splitting: %+v %+v", ts, nts))
				return
			}
		}
	}
	if data {
		if i.Contains(ts.TimeStart, false) {
			ts.SetRateInterval(i)
		}
		return
	}
	// if the span is enclosed in the interval try to set as new interval and return nil
	//log.Printf("Timing: %+v", i.Timing)
	if i.Contains(ts.TimeStart, false) && i.Contains(ts.TimeEnd, true) {
		//log.Print("All in interval")
		ts.SetRateInterval(i)
		return
	}
	// if only the start time is in the interval split the interval to the right
	if i.Contains(ts.TimeStart, false) {
		//log.Print("Start in interval")
		splitTime := i.Timing.getRightMargin(ts.TimeStart)
		ts.SetRateInterval(i)
		if splitTime == ts.TimeStart || splitTime.Equal(ts.TimeEnd) {
			return
		}
		nts = &TimeSpan{
			TimeStart: splitTime,
			TimeEnd:   ts.TimeEnd,
		}
		nts.copyRatingInfo(ts)
		ts.TimeEnd = splitTime
		nts.DurationIndex = ts.DurationIndex
		ts.SetNewDurationIndex(nts)
		// Logger.Debug(fmt.Sprintf("right: %+v %+v", ts, nts))
		return
	}
	// if only the end time is in the interval split the interval to the left
	if i.Contains(ts.TimeEnd, true) {
		//log.Print("End in interval")
		//tmpTime := time.Date(ts.TimeStart.)
		splitTime := i.Timing.getLeftMargin(ts.TimeEnd)
		splitTime = utils.CopyHour(splitTime, ts.TimeStart)
		if splitTime.Equal(ts.TimeEnd) {
			return
		}
		nts = &TimeSpan{
			TimeStart: splitTime,
			TimeEnd:   ts.TimeEnd,
		}
		nts.copyRatingInfo(ts)
		ts.TimeEnd = splitTime
		nts.SetRateInterval(i)
		nts.DurationIndex = ts.DurationIndex
		ts.SetNewDurationIndex(nts)
		// Logger.Debug(fmt.Sprintf("left: %+v %+v", ts, nts))
		return
	}
	return
}

// Split the timespan at the given increment start
func (ts *TimeSpan) SplitByIncrement(index int) *TimeSpan {
	index-- // split before the current one
	if index <= 0 || index >= ts.Increments.Len() {
		return nil
	}
	timeStart := ts.GetTimeStartForIncrement(index)
	newTs := &TimeSpan{
		RateInterval: ts.RateInterval,
		TimeStart:    timeStart,
		TimeEnd:      ts.TimeEnd,
		Increments:   &Increments{CompIncrement: ts.Increments.CompIncrement.Clone()},
	}
	newTs.copyRatingInfo(ts)
	newTs.DurationIndex = ts.DurationIndex
	ts.TimeEnd = timeStart
	newTs.Increments.CompIncrement.CompressFactor = ts.Increments.CompIncrement.CompressFactor - index
	ts.Increments.CompIncrement.CompressFactor = index
	ts.SetNewDurationIndex(newTs)
	return newTs
}

// Splits the given timespan on activation period's activation time.
func (ts *TimeSpan) SplitByRatingPlan(rp *RatingInfo) (newTs *TimeSpan) {
	activationTime := rp.ActivationTime.In(ts.TimeStart.Location())
	if !ts.Contains(activationTime) {
		return nil
	}
	newTs = &TimeSpan{
		TimeStart: activationTime,
		TimeEnd:   ts.TimeEnd,
	}
	newTs.copyRatingInfo(ts)
	newTs.DurationIndex = ts.DurationIndex
	ts.TimeEnd = activationTime
	ts.SetNewDurationIndex(newTs)
	// Logger.Debug(fmt.Sprintf("RP SPLITTING: %+v %+v", ts, newTs))
	return
}

// Splits the given timespan on activation period's activation time.
func (ts *TimeSpan) SplitByDay() (newTs *TimeSpan) {
	if ts.TimeStart.Day() == ts.TimeEnd.Day() || utils.TimeIs0h(ts.TimeEnd) {
		return
	}

	splitDate := ts.TimeStart.AddDate(0, 0, 1)
	splitDate = time.Date(splitDate.Year(), splitDate.Month(), splitDate.Day(), 0, 0, 0, 0, splitDate.Location())
	newTs = &TimeSpan{
		TimeStart: splitDate,
		TimeEnd:   ts.TimeEnd,
	}
	newTs.copyRatingInfo(ts)
	newTs.DurationIndex = ts.DurationIndex
	ts.TimeEnd = splitDate
	ts.SetNewDurationIndex(newTs)
	// Logger.Debug(fmt.Sprintf("RP SPLITTING: %+v %+v", ts, newTs))
	return
}

// Returns the starting time of this timespan
func (ts *TimeSpan) GetGroupStart() time.Duration {
	s := ts.DurationIndex - ts.GetDuration()
	if s < 0 {
		s = 0
	}
	return s
}

func (ts *TimeSpan) GetGroupEnd() time.Duration {
	return ts.DurationIndex
}

// sets the DurationIndex attribute to reflect new timespan
func (ts *TimeSpan) SetNewDurationIndex(nts *TimeSpan) {
	d := ts.DurationIndex - nts.GetDuration()
	if d < 0 {
		d = 0
	}
	ts.DurationIndex = d
}

func (nts *TimeSpan) copyRatingInfo(ts *TimeSpan) {
	if ts.ratingInfo == nil {
		return
	}
	nts.setRatingInfo(ts.ratingInfo)
}

// returns a time for the specified second in the time span
func (ts *TimeSpan) GetTimeStartForIncrement(index int) time.Time {
	return ts.TimeStart.Add(time.Duration(int64(index) * ts.Increments.CompIncrement.Duration.Nanoseconds()))
}

func (ts *TimeSpan) RoundToDuration(duration time.Duration) {
	if duration < ts.GetDuration() {
		duration = utils.RoundDuration(duration, ts.GetDuration())
	}
	if duration > ts.GetDuration() {
		initialDuration := ts.GetDuration()
		ts.TimeEnd = ts.TimeStart.Add(duration)
		ts.DurationIndex = ts.DurationIndex + (duration - initialDuration)
	}
}

func (ts *TimeSpan) hasBetterRateIntervalThan(interval *RateInterval) bool {
	if interval.Timing == nil {
		return false
	}
	otherLeftMargin := interval.Timing.getLeftMargin(ts.TimeStart)
	otherDistance := ts.TimeStart.Sub(otherLeftMargin)
	//log.Print("OTHER LEFT: ", otherLeftMargin)
	//log.Print("OTHER DISTANCE: ", otherDistance)
	// if the distance is negative it's not usable
	if otherDistance < 0 {
		return true
	}
	//log.Print("RI: ", ts.RateInterval)
	if ts.RateInterval == nil {
		return false
	}

	// the higher the weight the better
	if ts.RateInterval != nil &&
		ts.RateInterval.Weight < interval.Weight {
		return false
	}
	// check interval is closer than the new one
	ownLeftMargin := ts.RateInterval.Timing.getLeftMargin(ts.TimeStart)
	ownDistance := ts.TimeStart.Sub(ownLeftMargin)

	//log.Print("OWN LEFT: ", otherLeftMargin)
	//log.Print("OWN DISTANCE: ", otherDistance)
	//endOtherDistance := ts.TimeEnd.Sub(otherLeftMargin)

	// if own interval is closer than its better
	//log.Print(ownDistance)
	if ownDistance > otherDistance {
		return false
	}
	ownPrice, _, _ := ts.RateInterval.GetRateParameters(ts.GetGroupStart())
	otherPrice, _, _ := interval.GetRateParameters(ts.GetGroupStart())
	// if own price is smaller than it's better
	//log.Print(ownPrice, otherPrice)
	if ownPrice.Cmp(otherPrice) < 0 {
		return true
	}
	return true
}

func (ts *TimeSpan) Equal(other *TimeSpan) bool {
	return ts.Increments.Equal(other.Increments) &&
		ts.RateInterval.Equal(other.RateInterval) &&
		ts.GetUnitCost().Cmp(other.GetUnitCost()) == 0 &&
		ts.GetUnitDuration() == other.GetUnitDuration() &&
		ts.MatchedSubject == other.MatchedSubject &&
		ts.MatchedPrefix == other.MatchedPrefix &&
		ts.MatchedDestID == other.MatchedDestID &&
		ts.RatingPlanID == other.RatingPlanID
}

func (ts *TimeSpan) GetCompressFactor() int {
	if ts.CompressFactor == 0 {
		ts.CompressFactor = 1
	}
	return ts.CompressFactor
}
