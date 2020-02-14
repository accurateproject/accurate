package engine

import (
	"time"

	"github.com/accurateproject/accurate/dec"
)

type Metric interface {
	AddCDR(*QCDR)
	RemoveCDR(*QCDR)
	GetValue() *dec.Dec
}

const ASR = "ASR"
const ACD = "ACD"
const TCD = "TCD"
const ACC = "ACC"
const TCC = "TCC"
const PDD = "PDD"
const DDC = "DDC"

var STATS_NA = dec.MinusOne

func CreateMetric(metric string) Metric {
	switch metric {
	case ASR:
		return &ASRMetric{}
	case PDD:
		return &PDDMetric{}
	case ACD:
		return &ACDMetric{}
	case TCD:
		return &TCDMetric{}
	case ACC:
		return &ACCMetric{sum: dec.New()}
	case TCC:
		return &TCCMetric{sum: dec.New()}
	case DDC:
		return NewDccMetric()
	}
	return nil
}

// ASR - Answer-Seizure Ratio
// successfully answered Calls divided by the total number of Calls attempted and multiplied by 100
type ASRMetric struct {
	answered int64
	count    int64
}

func (asr *ASRMetric) AddCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() {
		asr.answered++
	}
	asr.count++
}

func (asr *ASRMetric) RemoveCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() {
		asr.answered--
	}
	asr.count--
}

func (asr *ASRMetric) GetValue() *dec.Dec {
	if asr.count == 0 {
		return STATS_NA
	}
	return dec.New().Quo(dec.NewVal(asr.answered, -2), dec.NewVal(asr.count, 0))
}

// PDD – Post Dial Delay (average)
// the sum of PDD seconds of total calls divided by the number of these calls.
type PDDMetric struct {
	sum   time.Duration
	count int64
}

func (pdd *PDDMetric) AddCDR(cdr *QCDR) {
	if cdr.Pdd == 0 { // Pdd not defined
		return
	}
	pdd.sum += cdr.Pdd
	pdd.count++
}

func (pdd *PDDMetric) RemoveCDR(cdr *QCDR) {
	if cdr.Pdd == 0 { // Pdd not defined
		return
	}
	pdd.sum -= cdr.Pdd
	pdd.count--
}

func (pdd *PDDMetric) GetValue() *dec.Dec {
	if pdd.count == 0 {
		return STATS_NA
	}
	return dec.New().Quo(dec.NewFloat(pdd.sum.Seconds()), dec.NewVal(pdd.count, 0))
}

// ACD – Average Call Duration
// the sum of billable seconds (billsec) of answered calls divided by the number of these answered calls.
type ACDMetric struct {
	sum   time.Duration
	count int64
}

func (acd *ACDMetric) AddCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() {
		acd.sum += cdr.Usage
		acd.count++
	}
}

func (acd *ACDMetric) RemoveCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() {
		acd.sum -= cdr.Usage
		acd.count--
	}
}

func (acd *ACDMetric) GetValue() *dec.Dec {
	if acd.count == 0 {
		return STATS_NA
	}
	return dec.New().Quo(dec.NewFloat(acd.sum.Seconds()), dec.NewVal(acd.count, 0))
}

// TCD – Total Call Duration
// the sum of billable seconds (billsec) of answered calls
type TCDMetric struct {
	sum   time.Duration
	count int64
}

func (tcd *TCDMetric) AddCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() {
		tcd.sum += cdr.Usage
		tcd.count++
	}
}

func (tcd *TCDMetric) RemoveCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() {
		tcd.sum -= cdr.Usage
		tcd.count--
	}
}

func (tcd *TCDMetric) GetValue() *dec.Dec {
	if tcd.count == 0 {
		return STATS_NA
	}
	return dec.NewFloat(tcd.sum.Seconds())
}

// ACC – Average Call Cost
// the sum of cost of answered calls divided by the number of these answered calls.
type ACCMetric struct {
	sum   *dec.Dec
	count int64
}

func (acc *ACCMetric) AddCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() && cdr.Cost.GtZero() {
		acc.sum.AddS(cdr.Cost)
		acc.count++
	}
}

func (acc *ACCMetric) RemoveCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() && cdr.Cost.GtZero() {
		acc.sum.SubS(cdr.Cost)
		acc.count--
	}
}

func (acc *ACCMetric) GetValue() *dec.Dec {
	if acc.count == 0 {
		return STATS_NA
	}
	return dec.New().Quo(acc.sum, dec.NewVal(acc.count, 0))
}

// TCC – Total Call Cost
// the sum of cost of answered calls
type TCCMetric struct {
	sum   *dec.Dec
	count int64
}

func (tcc *TCCMetric) AddCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() && cdr.Cost.GtZero() {
		tcc.sum.AddS(cdr.Cost)
		tcc.count++
	}
}

func (tcc *TCCMetric) RemoveCDR(cdr *QCDR) {
	if !cdr.AnswerTime.IsZero() && cdr.Cost.GtZero() {
		tcc.sum.SubS(cdr.Cost)
		tcc.count--
	}
}

func (tcc *TCCMetric) GetValue() *dec.Dec {
	if tcc.count == 0 {
		return STATS_NA
	}
	return tcc.sum
}

// DDC - Destination Distinct Count
//
type DCCMetric struct {
	destinations map[string]int64
}

func NewDccMetric() *DCCMetric {
	return &DCCMetric{
		destinations: make(map[string]int64),
	}
}

func (dcc *DCCMetric) AddCDR(cdr *QCDR) {
	if count, exists := dcc.destinations[cdr.Destination]; exists {
		dcc.destinations[cdr.Destination] = count + 1
	} else {
		dcc.destinations[cdr.Destination] = 0
	}
}

func (dcc *DCCMetric) RemoveCDR(cdr *QCDR) {
	if count, exists := dcc.destinations[cdr.Destination]; exists && count > 1 {
		dcc.destinations[cdr.Destination] = count - 1
	} else {
		dcc.destinations[cdr.Destination] = 0
	}
}

func (dcc *DCCMetric) GetValue() *dec.Dec {
	if len(dcc.destinations) == 0 {
		return STATS_NA
	}
	return dec.NewVal(int64(len(dcc.destinations)), 0)
}
