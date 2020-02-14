package engine

import (
	"strings"
	"sync"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

const (
	purgeCheckInterval = 5 * time.Second
)

type StatsQueue struct {
	Tenant         string                          `bson:"tenant"`
	Name           string                          `bson:"name"`
	TriggerRecords map[string]*ActionTriggerRecord `bson:"trigger_records"`
	triggers       ActionTriggers
	conf           *CdrStats
	metrics        map[string]Metric
	mux            sync.Mutex
	itemsCount     int
	lastSetupTime  time.Time
	accountingDB   AccountingStorage
	cdrDB          CdrStorage
	lastPurgeCheck time.Time // limit purge checks
}

// Simplified cdr structure containing only the necessary info
type QCDR struct {
	ID          bson.ObjectId `bson:"_id"`
	Tenant      string        `bson:"tenant"`
	Name        string        `bson:"name"`
	UniqueID    string        `bson:"unique_id"`
	RunID       string        `bson:"run_id"`
	SetupTime   time.Time     `bson:"setup_time"`
	AnswerTime  time.Time     `bson:"answer_time"`
	EventTime   time.Time     `bson:"event_time"` // moment of qcdr adding
	Pdd         time.Duration `bson:"pdd"`
	Usage       time.Duration `bson:"usage"`
	Cost        *dec.Dec      `bson:"cost"`
	Destination string        `bson:"destination"`
}

func (sq *StatsQueue) getTriggers() ActionTriggers {
	if sq.triggers != nil {
		return sq.triggers
	}
	for atgName := range sq.conf.TriggerIDs {
		atg, err := ratingStorage.GetActionTriggers(sq.Tenant, atgName, utils.CACHED)
		if err != nil || atg == nil {
			utils.Logger.Error("error getting triggers for ", zap.String("id", atgName), zap.Error(err))
			continue
		}
		atg.SetParentGroup()
		sq.triggers = append(sq.triggers, atg.ActionTriggers...)
	}
	sq.triggers.Sort()

	if len(sq.triggers) > 0 && sq.TriggerRecords == nil {
		sq.TriggerRecords = make(map[string]*ActionTriggerRecord)
	}
	for _, atr := range sq.triggers {
		if _, found := sq.TriggerRecords[atr.UniqueID]; !found {
			sq.TriggerRecords[atr.UniqueID] = &ActionTriggerRecord{
				UniqueID:       atr.UniqueID,
				Recurrent:      atr.Recurrent,
				ExpirationDate: atr.ExpirationDate,
				ActivationDate: atr.ActivationDate,
			}
		}
	}

	return sq.triggers
}

var METRIC_TRIGGER_MAP = map[string]string{
	"*min_asr": ASR,
	"*max_asr": ASR,
	"*min_pdd": PDD,
	"*max_pdd": PDD,
	"*min_acd": ACD,
	"*max_acd": ACD,
	"*min_tcd": TCD,
	"*max_tcd": TCD,
	"*min_acc": ACC,
	"*max_acc": ACC,
	"*min_tcc": TCC,
	"*max_tcc": TCC,
	"*min_ddc": DDC,
	"*max_ddc": DDC,
}

func NewStatsQueue(conf *CdrStats, accountingStorage AccountingStorage, cdrStorage CdrStorage) *StatsQueue {
	if conf == nil {
		return &StatsQueue{metrics: make(map[string]Metric)}
	}
	sq := &StatsQueue{Tenant: conf.Tenant, accountingDB: accountingStorage, cdrDB: cdrStorage}
	sq.updateConf(conf)
	return sq
}

func (sq *StatsQueue) updateConf(conf *CdrStats) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	// check if new conf asks for action trigger reset only
	if sq.conf != nil && (conf.hasGeneralConfigs() || sq.conf.equalExceptTriggers(conf)) {
		sq.conf.TriggerIDs = conf.TriggerIDs
		return
	}
	sq.Name = conf.Name
	sq.conf = conf
	sq.metrics = make(map[string]Metric, len(conf.Metrics))
	sq.save()
	for _, m := range conf.Metrics {
		if metric := CreateMetric(m); metric != nil {
			sq.metrics[m] = metric
		}
	}
}

func (sq *StatsQueue) save() {
	if err := sq.accountingDB.SetCdrStatsQueue(sq); err != nil {
		utils.Logger.Error("error saving cdr stats queue", zap.String("id", sq.Name), zap.Error(err))
		return
	}
}

func (sq *StatsQueue) load() error {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	qcdrIter := sq.accountingDB.Iterator(ColQcr, "$natural", map[string]interface{}{"tenant": sq.Tenant, "name": sq.Name})
	var qcdr QCDR
	for qcdrIter.Next(&qcdr) {
		sq.appendQcdr(&qcdr, false)
	}
	err := qcdrIter.Close()
	if err != nil {
		return err
	}
	sq.purgeObsoleteCdrs()
	return nil
}

func (sq *StatsQueue) saveStatsQueueQCDR(qcdr *QCDR) error {
	return sq.accountingDB.PushQCDR(qcdr)
}

func (sq *StatsQueue) appendCDR(cdr *CDR) {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	var qcdr *QCDR
	if sq.conf.acceptCdr(cdr) {
		qcdr = sq.simplifyCdr(cdr)
		sq.appendQcdr(qcdr, true)
	}
}

func (sq *StatsQueue) appendQcdr(qcdr *QCDR, normalAppend bool) { // normalAppend: apend during cdr rate (not reload)
	if qcdr.EventTime.IsZero() {
		qcdr.EventTime = time.Now() //used for TimeWindow
	}
	sq.lastSetupTime = qcdr.SetupTime
	sq.addToMetrics(qcdr)
	sq.itemsCount++
	// check for trigger
	if normalAppend {
		if err := sq.saveStatsQueueQCDR(qcdr); err != nil { // do not save items on load
			utils.Logger.Error("error saving stat queues qcdr: ", zap.Error(err))
		}
		sq.purgeObsoleteCdrs()
		values := sq.metricValues()
		executed := false
		for _, at := range sq.getTriggers() {
			// check is effective
			if at.IsExpired(time.Now()) || !at.IsActive(time.Now()) {
				continue
			}

			if sq.TriggerRecords[at.UniqueID].Executed {
				// trigger is marked as executed, so skipp it until
				// the next reset (see RESET_TRIGGERS action type)
				continue
			}

			if at.MinQueuedItems > 0 && sq.itemsCount < at.MinQueuedItems {
				continue
			}
			if strings.HasPrefix(at.ThresholdType, "*min_") {
				if value, ok := values[METRIC_TRIGGER_MAP[at.ThresholdType]]; ok {
					if value.Cmp(STATS_NA) > 0 && value.Cmp(at.ThresholdValue) <= 0 {
						if err := at.Execute(nil, sq.triggered(at)); err != nil {
							utils.Logger.Error("<cdr_stats> error executing trigger: ", zap.Error(err))
						}
						executed = true
					}
				}
			}
			if strings.HasPrefix(at.ThresholdType, "*max_") {
				if value, ok := values[METRIC_TRIGGER_MAP[at.ThresholdType]]; ok {
					if value.Cmp(STATS_NA) > 0 && value.Cmp(at.ThresholdValue) >= 0 {
						if err := at.Execute(nil, sq.triggered(at)); err != nil {
							utils.Logger.Error("<cdr_stats> error executing trigger: ", zap.Error(err))
						}
						executed = true
					}
				}
			}
		}
		if executed {
			sq.save()
		}
	}
}

func (sq *StatsQueue) getTriggerLastExecution(triggerID *string) time.Time {
	if trRec, found := sq.TriggerRecords[*triggerID]; found {
		return trRec.LastExecutionTime
	}
	return time.Time{}
}

func (sq *StatsQueue) addToMetrics(cdr *QCDR) {
	//log.Print("AddToMetrics: " + utils.ToIJSON(cdr))
	for _, metric := range sq.metrics {
		metric.AddCDR(cdr)
	}
}

func (sq *StatsQueue) removeFromMetrics(cdr *QCDR) {
	for _, metric := range sq.metrics {
		//log.Printf("Remove: %s, %+v", k, cdr)
		metric.RemoveCDR(cdr)
	}
}

func (sq *StatsQueue) simplifyCdr(cdr *CDR) *QCDR {
	return &QCDR{
		Tenant:      sq.Tenant,
		Name:        sq.Name,
		UniqueID:    cdr.UniqueID,
		RunID:       cdr.RunID,
		SetupTime:   cdr.SetupTime,
		AnswerTime:  cdr.AnswerTime,
		Pdd:         cdr.PDD,
		Usage:       cdr.Usage,
		Cost:        cdr.GetCost(),
		Destination: cdr.Destination,
	}
}

func (sq *StatsQueue) purgeObsoleteCdrs() {
	if time.Since(sq.lastPurgeCheck) < purgeCheckInterval {
		return
	}
	sq.lastPurgeCheck = time.Now()
	if sq.conf.QueueLength > 0 {
		extraItems := sq.itemsCount - sq.conf.QueueLength
		if extraItems > 0 {
			qcdrs, err := sq.accountingDB.PopQCDR(sq.Tenant, sq.Name, nil, extraItems)
			if err != nil {
				utils.Logger.Error("error poping qcdrs: ", zap.Error(err))
				return
			}
			sq.itemsCount -= len(qcdrs)
			for _, qcdr := range qcdrs {
				sq.removeFromMetrics(qcdr)
			}
		}
	}
	if sq.conf.TimeWindow > 0 {
		obsoleteMarkerTime := time.Now().Add(-sq.conf.TimeWindow)
		qcdrs, err := sq.accountingDB.PopQCDR(sq.Tenant, sq.Name, map[string]interface{}{"event_time": bson.M{"$lte": obsoleteMarkerTime}}, -1)
		if err == utils.ErrNotFound {
			return
		}
		if err != nil {
			utils.Logger.Error("error poping qcdrs: ", zap.Error(err))
			return
		}
		sq.itemsCount -= len(qcdrs)
		for _, qcdr := range qcdrs {
			sq.removeFromMetrics(qcdr)
		}
	}
}

func (sq *StatsQueue) getMetricValues() map[string]*dec.Dec {
	sq.mux.Lock()
	defer sq.mux.Unlock()
	sq.purgeObsoleteCdrs()
	return sq.metricValues()
}

func (sq *StatsQueue) length() int {
	return sq.itemsCount
	/*length, err := sq.accountingDB.Count(ColQcr, map[string]interface{}{"tenant": sq.Tenant, "name": sq.Name})
	if err != nil {
		length = 0
		utils.Logger.Error("error getting qcdr count: ", zap.Error(err))
	}
	return length*/
}

func (sq *StatsQueue) resetQCDRs() error {
	return sq.accountingDB.RemoveQCDRs(sq.Tenant, sq.Name)
}
func (sq *StatsQueue) resetMetrics() {
	sq.metrics = make(map[string]Metric, len(sq.conf.Metrics))
	for _, m := range sq.conf.Metrics {
		if metric := CreateMetric(m); metric != nil {
			sq.metrics[m] = metric
		}
	}
}

func (sq *StatsQueue) getLastSetupTime() time.Time {
	return sq.lastSetupTime
}

func (sq *StatsQueue) metricValues() map[string]*dec.Dec {
	stat := make(map[string]*dec.Dec, len(sq.metrics))
	for key, metric := range sq.metrics {
		stat[key] = metric.GetValue()
	}
	return stat
}

// Convert data into a struct which can be used in actions based on triggers hit
func (sq *StatsQueue) triggered(at *ActionTrigger) *StatsQueueTriggered {
	return &StatsQueueTriggered{
		Tenant:         sq.Tenant,
		Name:           sq.Name,
		Metrics:        sq.metricValues(),
		Trigger:        at,
		TriggerRecords: sq.TriggerRecords,
	}
}

// Struct to be passed to triggered actions
type StatsQueueTriggered struct {
	Tenant         string
	Name           string // StatsQueueId
	Metrics        map[string]*dec.Dec
	Trigger        *ActionTrigger
	TriggerRecords map[string]*ActionTriggerRecord
}
