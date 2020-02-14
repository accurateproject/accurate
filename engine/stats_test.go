package engine

import (
	"testing"
	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

func TestStatsQueueInit(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACC}}, accountingStorage, cdrStorage)
	if len(sq.metrics) != 2 {
		t.Error("Expected 2 metrics got ", len(sq.metrics))
	}
}

func TestStatsValue(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}}, accountingStorage, cdrStorage)
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(1),
	}
	sq.appendCDR(cdr)
	cdr.Cost = dec.NewFloat(2)
	sq.appendCDR(cdr)
	cdr.Cost = dec.NewFloat(3)
	sq.appendCDR(cdr)
	s := sq.getMetricValues()
	if s[ASR].Cmp(dec.NewVal(100, 0)) != 0 ||
		s[ACD].Cmp(dec.NewVal(10, 0)) != 0 ||
		s[TCD].Cmp(dec.NewVal(30, 0)) != 0 ||
		s[ACC].Cmp(dec.NewVal(2, 0)) != 0 ||
		s[TCC].Cmp(dec.NewVal(6, 0)) != 0 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsSimplifyCDR(t *testing.T) {
	cdr := &CDR{
		ToR:         "tor",
		OriginID:    "accid",
		OriginHost:  "cdrhost",
		Source:      "cdrsource",
		RequestType: "reqtype",
		Direction:   "direction",
		Tenant:      "tenant",
		Category:    "category",
		Account:     "account",
		Subject:     "subject",
		Destination: "12345678",
		SetupTime:   time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
		Usage:       10 * time.Second,
		RunID:       "mri",
		Cost:        dec.NewFloat(10),
	}
	sq := &StatsQueue{}
	qcdr := sq.simplifyCdr(cdr)
	if cdr.SetupTime != qcdr.SetupTime ||
		cdr.AnswerTime != qcdr.AnswerTime ||
		cdr.Usage != qcdr.Usage ||
		cdr.Cost.Cmp(qcdr.Cost) != 0 {
		t.Errorf("Failed to simplify cdr: %+v", qcdr)
	}
}

func TestStatsAcceptCdr(t *testing.T) {
	sq := NewStatsQueue(nil, accountingStorage, cdrStorage)
	cdr := &CDR{
		ToR:             "tor",
		OriginID:        "accid",
		OriginHost:      "cdrhost",
		Source:          "cdrsource",
		RequestType:     "reqtype",
		Direction:       "direction",
		Tenant:          "test",
		Category:        "category",
		Account:         "account",
		Subject:         "subject",
		Destination:     "0723045326",
		SetupTime:       time.Date(2014, 7, 3, 13, 43, 0, 0, time.UTC),
		Usage:           10 * time.Second,
		PDD:             7 * time.Second,
		Supplier:        "supplier1",
		DisconnectCause: "normal",
		RunID:           "mri",
		Cost:            dec.NewFloat(10),
	}
	sq.conf = &CdrStats{Tenant: "test"}
	if sq.conf.acceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"TOR": {"$in":["test"]}}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"CdrHost": "test"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"CdrSource": "test"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"Direction": "test"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"Tenant": "tenant"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"Category": "test"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"Account": "test"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"Subject": "test"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"Supplier": "test"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"DisconnectCause": "test"}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"DestinationIDs": {"$has":["test"]}}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"DestinationIDs": {"$has": ["NAT", "RET"]}}`}
	if sq.conf.acceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"SetupTime": {"$gt":"2014-07-03T13:43:00Z"}}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"SetupTime": {"$btw":["2014-07-03T13:42:00Z", "2014-07-03T13:43:00Z"]}}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"SetupTime": {"$gte":"2014-07-03T13:42:00Z"}}`}
	if sq.conf.acceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"SetupTime": {"$btw":["2014-07-03T13:42:00Z", "2014-07-03T13:43:01Z"]}}`}
	if sq.conf.acceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"Usage": {"$gte":"11s"}}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"Usage": {"$btw":["1s","10s"]}}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"PDD": {"$gte":"8s"}}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"PDD": {"$btw":["3s","7s"]}}`}
	if sq.conf.acceptCdr(cdr) == true {
		t.Errorf("Should have NOT accepted this CDR: %+v", cdr)
	}
	sq.conf = &CdrStats{Tenant: "test", Filter: `{"PDD": {"$btw":["3s","18s"]}}`}
	if sq.conf.acceptCdr(cdr) != true {
		t.Errorf("Should have accepted this CDR: %+v", cdr)
	}
}

func TestStatsQueueIds(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	ids := []string{}
	if err := cdrStats.GetQueueIDs("test", &ids); err != nil {
		t.Error("Errorf getting queue ids: ", err)
	}
	result := len(ids)
	expected := 5
	if result != expected {
		t.Errorf("Errorf loading stats queues. Expected %v was %v (%v)", expected, result, ids)
	}
}

func TestStatsAppendCdr(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdr := &CDR{
		Tenant:          "test",
		Category:        "call",
		AnswerTime:      time.Now(),
		SetupTime:       time.Now(),
		Usage:           10 * time.Second,
		Cost:            dec.NewFloat(10),
		Supplier:        "suppl1",
		DisconnectCause: "NORMAL_CLEARING",
	}
	err := cdrStats.AppendCDR(cdr, nil)
	if err != nil {
		t.Error("Error appending cdr to stats: ", err)
	}
	if len(cdrStats.queues) != 5 ||
		cdrStats.queues["test:CDRST1"].length() != 0 ||
		cdrStats.queues["test:CDRST2"].length() != 1 {
		t.Logf("%+v", cdrStats.queues["test:CDRST2"])
		t.Error("Error appending cdr to queue: ", utils.ToIJSON(cdrStats.queues))
	}
}

func TestStatsGetMetrics(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdr := &CDR{
		Tenant:     "test",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(10),
	}
	cdrStats.AppendCDR(cdr, nil)
	cdr = &CDR{
		Tenant:     "test",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      2 * time.Second,
		Cost:       dec.NewFloat(4),
	}
	cdrStats.AppendCDR(cdr, nil)
	valMap := make(map[string]*dec.Dec)
	if err := cdrStats.GetMetrics(utils.AttrStatsQueueID{Tenant: "test", ID: "CDRST2"}, &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"].Cmp(dec.NewVal(6, 0)) != 0 || valMap["ASR"].Cmp(dec.NewVal(100, 0)) != 0 {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsPurgeBySize(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Tenant: "statstest", Name: "sqX", QueueLength: 2, Metrics: []string{ASR, ACC}}, accountingStorage, cdrStorage)
	sq.appendCDR(&CDR{
		Tenant:     "statstest",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      2 * time.Second,
		Cost:       dec.NewFloat(4),
	})
	sq.appendCDR(&CDR{
		Tenant:     "statstest",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      2 * time.Second,
		Cost:       dec.NewFloat(4),
	})
	if sq.length() != 2 {
		t.Error("error adding cdr: ", sq.length())
	}

	sq.appendCDR(&CDR{
		Tenant:     "statstest",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      2 * time.Second,
		Cost:       dec.NewFloat(4),
	})
	if sq.length() != 2 {
		t.Error("error adding cdr: ", sq.length())
	}
}

func TestStatsReloadQueues(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdr := &CDR{
		Tenant:     "test",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(10),
	}
	cdrStats.AppendCDR(cdr, nil)
	if err := cdrStats.ReloadQueues(utils.AttrStatsQueueIDs{Tenant: "test"}, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIDs("test", &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 5
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v: %v", expected, result, ids)
	}
	valMap := make(map[string]*dec.Dec)
	if err := cdrStats.GetMetrics(utils.AttrStatsQueueID{Tenant: "test", ID: "CDRST2"}, &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"].Cmp(dec.NewVal(10, 0)) != 0 || valMap["ASR"].Cmp(dec.NewVal(100, 0)) != 0 {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsReloadQueuesWithDefault(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdrStats.AddQueue(&CdrStats{
		Tenant: "test",
		Name:   utils.META_DEFAULT,
	}, nil)
	cdr := &CDR{
		Tenant:     "test",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(10),
	}
	cdrStats.AppendCDR(cdr, nil)

	if err := cdrStats.ReloadQueues(utils.AttrStatsQueueIDs{Tenant: "test"}, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIDs("test", &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 6
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]*dec.Dec)
	if err := cdrStats.GetMetrics(utils.AttrStatsQueueID{Tenant: "test", ID: "CDRST2"}, &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"].Cmp(dec.NewVal(10, 0)) != 0 || valMap["ASR"].Cmp(dec.NewVal(100, 0)) != 0 {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsReloadQueuesWithIds(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdr := &CDR{
		Tenant:     "test",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(10),
	}
	cdrStats.AppendCDR(cdr, nil)
	if err := cdrStats.ReloadQueues(utils.AttrStatsQueueIDs{Tenant: "test", IDs: []string{"CDRST1"}}, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIDs("test", &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 6
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]*dec.Dec)
	if err := cdrStats.GetMetrics(utils.AttrStatsQueueID{Tenant: "test", ID: "CDRST2"}, &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"].Cmp(dec.NewVal(10, 0)) != 0 || valMap["ASR"].Cmp(dec.NewVal(100, 0)) != 0 {
		t.Error("Error on metric map: ", valMap)
	}
}

/*
func TestStatsSaveQueues(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdr := &CDR{
		Tenant:     "test",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(10),
	}
	cdrStats.AppendCDR(cdr, nil)
	ids := []string{}
	cdrStats.GetQueueIDs("test", &ids)
	if _, found := cdrStats.queueSavers["test:CDRST1"]; !found {
		t.Error("Error creating queue savers: ", cdrStats.queueSavers)
	}
}
*/

func TestStatsResetQueues(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdr := &CDR{
		Tenant:     "test",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(10),
	}
	cdrStats.AppendCDR(cdr, nil)
	if err := cdrStats.ResetQueues(utils.AttrStatsQueueIDs{Tenant: "test"}, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIDs("test", &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 6
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]*dec.Dec)
	if err := cdrStats.GetMetrics(utils.AttrStatsQueueID{Tenant: "test", ID: "CDRST2"}, &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"].Cmp(STATS_NA) != 0 || valMap["ASR"].Cmp(STATS_NA) != 0 {
		t.Error("Error on metric map: ", valMap)
	}
}

func TestStatsResetQueuesWithIds(t *testing.T) {
	cdrStats := NewStats(ratingStorage, accountingStorage, cdrStorage)
	cdr := &CDR{
		Tenant:     "test",
		Category:   "call",
		AnswerTime: time.Now(),
		SetupTime:  time.Now(),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(10),
	}
	cdrStats.AppendCDR(cdr, nil)
	if err := cdrStats.ResetQueues(utils.AttrStatsQueueIDs{Tenant: "test", IDs: []string{"CDRST1"}}, nil); err != nil {
		t.Error("Error reloading queues: ", err)
	}
	ids := []string{}
	if err := cdrStats.GetQueueIDs("test", &ids); err != nil {
		t.Error("Error getting queue ids: ", err)
	}
	result := len(ids)
	expected := 6
	if result != expected {
		t.Errorf("Error loading stats queues. Expected %v was %v", expected, result)
	}
	valMap := make(map[string]*dec.Dec)
	if err := cdrStats.GetMetrics(utils.AttrStatsQueueID{Tenant: "test", ID: "CDRST2"}, &valMap); err != nil {
		t.Error("Error getting metric values: ", err)
	}
	if len(valMap) != 2 || valMap["ACD"].Cmp(dec.NewVal(10, 0)) != 0 || valMap["ASR"].Cmp(dec.NewVal(100, 0)) != 0 {
		t.Error("Error on metric map: ", valMap)
	}
}

/*
func TestStatsSaveRestoreQueue(t *testing.T) {
	sq := &StatsQueue{
		conf: &CdrStats{Tenant: "test", Name: "TTT"},
		Cdrs: []*QCdr{&QCdr{Cost: 9.0}},
	}
	if err := accountingStorage.SetCdrStatsQueue(sq); err != nil {
		t.Error("Error saving metric: ", err)
	}
	recovered, err := accountingStorage.GetCdrStatsQueue(sq.Tenant, sq.Name)
	if err != nil {
		t.Error("Error loading metric: ", err)
	}
	if len(recovered.Cdrs) != 1 || recovered.Cdrs[0].Cost != sq.Cdrs[0].Cost {
		t.Errorf("Expecting %+v got: %+v", sq.Cdrs[0], recovered.Cdrs[0])
	}
}

func TestStatsPurgeTimeOne(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}, TimeWindow: 30 * time.Minute}, accountingStorage, cdrStorage)
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(1),
	}
	qcdr := sq.appendCDR(cdr)
	qcdr.EventTime = qcdr.SetupTime
	s := sq.GetMetricValues()
	if s[ASR] != -1 ||
		s[ACD] != -1 ||
		s[TCD] != -1 ||
		s[ACC] != -1 ||
		s[TCC] != -1 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsPurgeTime(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}, TimeWindow: 30 * time.Minute})
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(1),
	}
	qcdr := sq.appendCDR(cdr)
	qcdr.EventTime = qcdr.SetupTime
	cdr.Cost = dec.NewFloat(2)
	qcdr = sq.appendCDR(cdr)
	qcdr.EventTime = qcdr.SetupTime
	cdr.Cost = dec.NewFloat(3)
	qcdr = sq.appendCDR(cdr)
	qcdr.EventTime = qcdr.SetupTime
	s := sq.GetMetricValues()
	if s[ASR] != -1 ||
		s[ACD] != -1 ||
		s[TCD] != -1 ||
		s[ACC] != -1 ||
		s[TCC] != -1 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsPurgeTimeFirst(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}, TimeWindow: 30 * time.Minute})
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(1),
	}
	qcdr := sq.appendCDR(cdr)
	cdr.Cost = dec.NewFloat(2)
	cdr.SetupTime = time.Date(2024, 7, 14, 14, 25, 0, 0, time.UTC)
	cdr.AnswerTime = time.Date(2024, 7, 14, 14, 25, 0, 0, time.UTC)
	qcdr.EventTime = qcdr.SetupTime
	sq.appendCDR(cdr)
	cdr.Cost = dec.NewFloat(3)
	sq.appendCDR(cdr)
	s := sq.GetMetricValues()
	if s[ASR] != 100 ||
		s[ACD] != 10 ||
		s[TCD] != 20 ||
		s[ACC] != 2.5 ||
		s[TCC] != 5 {
		t.Errorf("Error getting stats: %+v", s)
	}
}

func TestStatsPurgeLength(t *testing.T) {
	sq := NewStatsQueue(&CdrStats{Metrics: []string{ASR, ACD, TCD, ACC, TCC}, QueueLength: 1})
	cdr := &CDR{
		SetupTime:  time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		Usage:      10 * time.Second,
		Cost:       dec.NewFloat(1),
	}
	sq.appendCDR(cdr)
	cdr.Cost = dec.NewFloat(2)
	sq.appendCDR(cdr)
	cdr.Cost = dec.NewFloat(3)
	sq.appendCDR(cdr)
	s := sq.GetMetricValues()
	if s[ASR] != 100 ||
		s[ACD] != 10 ||
		s[TCD] != 10 ||
		s[ACC] != 3 ||
		s[TCC] != 3 {
		t.Errorf("Error getting stats: %+v", s)
	}
}
*/
