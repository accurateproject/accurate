package engine

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/accurateproject/accurate/history"
	"github.com/accurateproject/accurate/utils"
)

/*
The struture that is saved to storage.
*/
type RatingPlan struct {
	Tenant           string                  `bson:"tenant"`
	Name             string                  `bson:"name"`
	Timings          map[string]*RITiming    `bson:"timings"`
	Ratings          map[string]*RIRate      `bson:"ratings"`
	DRates           map[string]*DRate       `bson:"d_rates"`
	DestinationRates map[string]*DRateHelper `bson:"destination_rates"`
}

type DRate struct {
	Timing string  `bson:"timing"`
	Rating string  `bson:"rating"`
	Weight float64 `bson:"weight"`
}

type DRateHelper struct {
	DRateKeys map[string]struct{}
	CodeName  string
}

func (rpr *DRate) hash() string {
	return utils.Sha1(fmt.Sprintf("%v", rpr))[:6]
}

func (rpr *DRate) Equal(orpr *DRate) bool {
	return rpr.Timing == orpr.Timing && rpr.Rating == orpr.Rating && rpr.Weight == orpr.Weight
}

func (rp *RatingPlan) RateIntervalList(code string) RateIntervalList {
	ril := make(RateIntervalList, 0)
	for rprTag := range rp.DestinationRates[code].DRateKeys {
		rpr := rp.DRates[rprTag]
		ril = append(ril, &RateInterval{
			Timing: rp.Timings[rpr.Timing],
			Rating: rp.Ratings[rpr.Rating],
			Weight: rpr.Weight,
		})
	}
	return ril
}

// no sorter because it's sorted with RateIntervalTimeSorter

/*
Adds one ore more intervals to the internal interval list only if it is not allready in the list.
*/
func (rp *RatingPlan) AddRateInterval(code, name string, ris ...*RateInterval) {
	//log.Printf("Code %v with RIS: %+v", code, utils.ToIJSON(ris))
	if rp.DestinationRates == nil {
		rp.Timings = make(map[string]*RITiming)
		rp.Ratings = make(map[string]*RIRate)
		rp.DRates = make(map[string]*DRate)
		rp.DestinationRates = make(map[string]*DRateHelper)
	}
	for _, ri := range ris {
		rpr := &DRate{Weight: ri.Weight}
		if ri.Timing != nil {
			timingTag := ri.Timing.hash()
			rp.Timings[timingTag] = ri.Timing
			rpr.Timing = timingTag
		}
		if ri.Rating != nil {
			ratingTag := ri.Rating.hash()
			rp.Ratings[ratingTag] = ri.Rating
			rpr.Rating = ratingTag
		}
		drateHash := rpr.hash()
		rp.DRates[drateHash] = rpr
		if rp.DestinationRates[code] == nil {
			rp.DestinationRates[code] = &DRateHelper{DRateKeys: make(map[string]struct{}), CodeName: name}
		}
		rp.DestinationRates[code].DRateKeys[drateHash] = struct{}{}
	}
}

func (rp *RatingPlan) Equal(o *RatingPlan) bool {
	return rp.Tenant == o.Tenant && rp.Name == o.Name
}

// history record method
func (rp *RatingPlan) GetHistoryRecord() history.Record {
	js, _ := json.Marshal(rp)
	return history.Record{
		Id:       utils.ConcatKey(rp.Tenant, rp.Name),
		Filename: history.RATING_PLANS_FN,
		Payload:  js,
	}
}

// IsValid determines if the rating plan covers a continous period of time
func (rp *RatingPlan) isContinous() bool {
	weekdays := make([]int, 7)
	for _, tm := range rp.Timings {
		// if it is a blank timing than it will match all
		if tm.IsBlank() {
			return true
		}
		// skip the special timings (for specific dates)
		if len(tm.Years) != 0 || len(tm.Months) != 0 || len(tm.MonthDays) != 0 {
			continue
		}
		// if the startime is not midnight than is an extra time
		if tm.StartTime != "00:00:00" {
			continue
		}
		//check if all weekdays are covered
		for _, wd := range tm.WeekDays {
			weekdays[wd] = 1
		}
		allWeekdaysCovered := true
		for _, wd := range weekdays {
			if wd != 1 {
				allWeekdaysCovered = false
				break
			}
		}
		if allWeekdaysCovered {
			return true
		}
	}
	return false
}

func (rp *RatingPlan) getFirstUnsaneRating() string {
	for _, rating := range rp.Ratings {
		rating.Rates.Sort()
		for i, rate := range rating.Rates {
			if i < (len(rating.Rates) - 1) {
				nextRate := rating.Rates[i+1]
				if nextRate.GroupIntervalStart <= rate.GroupIntervalStart {
					return utils.ToJSON(rating)
				}
				if math.Mod(nextRate.GroupIntervalStart.Seconds(), rate.RateIncrement.Seconds()) != 0 {
					return utils.ToJSON(rating)
				}
				if rate.RateUnit == 0 || rate.RateIncrement == 0 {
					return utils.ToJSON(rating)
				}
			}
		}
	}
	return ""
}

func (rp *RatingPlan) getFirstUnsaneTiming() string {
	for _, timing := range rp.Timings {
		if (len(timing.Years) != 0 || len(timing.Months) != 0 || len(timing.MonthDays) != 0) &&
			len(timing.WeekDays) != 0 {
			return utils.ToJSON(timing)
		}
	}
	return ""
}
