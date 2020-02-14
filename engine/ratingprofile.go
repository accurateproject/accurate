package engine

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/accurateproject/accurate/history"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

type RatingProfile struct {
	Direction             string                `bson:"direction"`
	Tenant                string                `bson:"tenant"`
	Category              string                `bson:"category"`
	Subject               string                `bson:"subject"`
	RatingPlanActivations RatingPlanActivations `bson:"rating_plan_activations"`
}

type RatingPlanActivation struct {
	ActivationTime  time.Time `bson:"activation_time"`
	RatingPlanID    string    `bson:"rating_plan_id"`
	FallbackKeys    []string  `bson:"fallback_keys"`
	CdrStatQueueIDs []string  `bson:"cdr_stat_queue_ids"`
}

func (rpf *RatingProfile) FullID() string {
	return utils.ConcatKey(rpf.Direction, rpf.Tenant, rpf.Category, rpf.Subject)
}

func (rpa *RatingPlanActivation) Equal(orpa *RatingPlanActivation) bool {
	return rpa.ActivationTime == orpa.ActivationTime && rpa.RatingPlanID == orpa.RatingPlanID
}

type RatingPlanActivations []*RatingPlanActivation

func (rpas RatingPlanActivations) Sort() {
	sort.Slice(rpas, func(i, j int) bool {
		return rpas[i].ActivationTime.Before(rpas[j].ActivationTime)
	})
}

func (rpas RatingPlanActivations) GetActiveForCall(cd *CallDescriptor) RatingPlanActivations {
	rpas.Sort()
	lastBeforeCallStart := 0
	firstAfterCallEnd := len(rpas)
	for index, rpa := range rpas {
		if rpa.ActivationTime.Before(cd.TimeStart) || rpa.ActivationTime.Equal(cd.TimeStart) {
			lastBeforeCallStart = index
		}
		if rpa.ActivationTime.After(cd.TimeEnd) {
			firstAfterCallEnd = index
			break
		}
	}
	return rpas[lastBeforeCallStart:firstAfterCallEnd]
}

type RatingInfo struct {
	MatchedSubject string
	RatingPlanID   string
	MatchedPrefix  string
	MatchedDestID  string
	ActivationTime time.Time
	RateIntervals  RateIntervalList
	FallbackKeys   []string
}

// SelectRatingIntevalsForTimespan orders rate intervals in time preserving only those which aply to the specified timestamp
func (ri RatingInfo) SelectRatingIntevalsForTimespan(ts *TimeSpan) (result RateIntervalList) {
	sorter := &RateIntervalTimeSorter{referenceTime: ts.TimeStart, ris: ri.RateIntervals}
	rateIntervals := sorter.Sort()
	// get the rating interval closest to begining of timespan
	var delta time.Duration = -1
	var bestRateIntervalIndex int
	var bestIntervalWeight float64
	for index, rateInterval := range rateIntervals {
		if !rateInterval.Contains(ts.TimeStart, false) {
			continue
		}
		if rateInterval.Weight < bestIntervalWeight {
			break // don't consider lower weights'
		}
		startTime := rateInterval.Timing.getLeftMargin(ts.TimeStart)
		tmpDelta := ts.TimeStart.Sub(startTime)
		if (startTime.Before(ts.TimeStart) ||
			startTime.Equal(ts.TimeStart)) &&
			(delta == -1 || tmpDelta < delta) {
			bestRateIntervalIndex = index
			bestIntervalWeight = rateInterval.Weight
			delta = tmpDelta
		}
	}
	result = append(result, rateIntervals[bestRateIntervalIndex])
	// check if later rating intervals influence this timespan
	//log.Print("RIS: ", utils.ToIJSON(rateIntervals))
	for i := bestRateIntervalIndex + 1; i < len(rateIntervals); i++ {
		if rateIntervals[i].Weight < bestIntervalWeight {
			break // don't consider lower weights'
		}
		startTime := rateIntervals[i].Timing.getLeftMargin(ts.TimeStart)
		if startTime.Before(ts.TimeEnd) {
			result = append(result, rateIntervals[i])
		}
	}
	return
}

type RatingInfos []*RatingInfo

func (ris RatingInfos) Sort() {
	sort.Slice(ris, func(i, j int) bool {
		return ris[i].ActivationTime.Before(ris[j].ActivationTime)
	})
}

func (ris RatingInfos) String() string {
	b, _ := json.MarshalIndent(ris, "", " ")
	return string(b)
}

func (rpf *RatingProfile) GetRatingPlansForPrefix(cd *CallDescriptor) (err error) {
	//log.Print("RPF: ", utils.ToIJSON(rpf))
	var ris RatingInfos
	for index, rpa := range rpf.RatingPlanActivations.GetActiveForCall(cd) {
		rpl, err := ratingStorage.GetRatingPlan(rpf.Tenant, rpa.RatingPlanID, utils.CACHED)
		//log.Print("RPL: ", utils.ToIJSON(rpl))
		if err != nil || rpl == nil {
			utils.Logger.Error("error checking destination", zap.Error(err))
			continue
		}
		destinationCode := ""
		destinationName := ""
		var rps RateIntervalList
		if cd.Destination == utils.ANY || cd.Destination == "" {
			cd.Destination = utils.ANY
			if _, ok := rpl.DestinationRates[utils.ANY]; ok {
				rps = rpl.RateIntervalList(utils.ANY)
				destinationCode = utils.ANY
				destinationName = utils.ANY
			}
		} else {
			for _, p := range utils.SplitPrefix(cd.Destination, MIN_PREFIX_MATCH) {
				if helper, ok := rpl.DestinationRates[p]; ok {
					ril := rpl.RateIntervalList(p)
					rps = ril
					destinationCode = p
					destinationName = helper.CodeName
				}
				if rps != nil {
					break
				}
			}
			if rps == nil { // fallback on *any destination
				if _, ok := rpl.DestinationRates[utils.ANY]; ok {
					rps = rpl.RateIntervalList(utils.ANY)
					destinationCode = utils.ANY
					destinationName = utils.ANY
				}
			}
		}
		// check if it's the first ri and add a blank one for the initial part not covered
		if index == 0 && cd.TimeStart.Before(rpa.ActivationTime) {
			ris = append(ris, &RatingInfo{
				MatchedSubject: "",
				MatchedPrefix:  "",
				MatchedDestID:  "",
				ActivationTime: cd.TimeStart,
				RateIntervals:  nil,
				FallbackKeys:   []string{FALLBACK_SUBJECT}})
		}
		if len(destinationCode) > 0 {
			ris = append(ris, &RatingInfo{
				MatchedSubject: rpf.FullID(),
				RatingPlanID:   rpl.Name,
				MatchedPrefix:  destinationCode,
				MatchedDestID:  destinationName,
				ActivationTime: rpa.ActivationTime,
				RateIntervals:  rps,
				FallbackKeys:   rpa.FallbackKeys})
		} else {
			// add for fallback information
			if len(rpa.FallbackKeys) > 0 {
				ris = append(ris, &RatingInfo{
					MatchedSubject: "",
					MatchedPrefix:  "",
					MatchedDestID:  "",
					ActivationTime: rpa.ActivationTime,
					RateIntervals:  nil,
					FallbackKeys:   rpa.FallbackKeys,
				})
			}
		}
	}
	if len(ris) > 0 {
		cd.addRatingInfos(ris)
		return
	}

	return utils.ErrNotFound
}

// history record method
func (rpf *RatingProfile) GetHistoryRecord(deleted bool) history.Record {
	js, _ := json.Marshal(rpf)
	return history.Record{
		Id:       rpf.FullID(),
		Filename: history.RATING_PROFILES_FN,
		Payload:  js,
		Deleted:  deleted,
	}
}

/*
type TenantRatingSubject struct {
	Tenant, Subject string
}
*/
