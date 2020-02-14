package engine

import (
	"reflect"
	"strings"
	"time"

	"github.com/accurateproject/accurate/utils"
)

type CdrStats struct {
	Tenant      string
	Name        string        // Config id, unique per config instance
	QueueLength int           // Number of items in the stats buffer
	TimeWindow  time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	Metrics     []string      // ASR, ACD, ACC
	Filter      string
	TriggerIDs  utils.StringMap
	Disabled    bool
	filter      *utils.StructQ
}

func (cs *CdrStats) getFilter() *utils.StructQ {
	if cs.filter != nil {
		return cs.filter
	}
	cs.filter, _ = utils.NewStructQ(cs.Filter) // the error should be check at load
	return cs.filter
}

func (cs *CdrStats) acceptCdr(cdr *CDR) bool {
	if cdr == nil || cs.Disabled || cdr.Tenant != cs.Tenant {
		return false
	}
	if len(cs.Filter) == 0 {
		return true
	}
	var destinationIDs []string
	if strings.Contains(cs.Filter, "DestinationIDs") {
		if dests, err := ratingStorage.GetDestinations(cdr.Tenant, cdr.Destination, "", utils.DestMatching, utils.CACHED); err == nil {
			for _, dest := range dests {
				destinationIDs = append(destinationIDs, dest.Name)
			}
		}
	}
	match, err := cs.getFilter().Query(&struct {
		DestinationIDs []string
		*CDR
	}{
		DestinationIDs: destinationIDs,
		CDR:            cdr,
	}, false)
	if err != nil {
		match = false
	}
	return match
}

func (cs *CdrStats) hasGeneralConfigs() bool {
	return cs.QueueLength == 0 &&
		cs.TimeWindow == 0 &&
		len(cs.Metrics) == 0
}

func (cs *CdrStats) equalExceptTriggers(other *CdrStats) bool {
	return cs.QueueLength == other.QueueLength &&
		cs.TimeWindow == other.TimeWindow &&
		reflect.DeepEqual(cs.Metrics, other.Metrics) &&
		cs.Filter == other.Filter
}
