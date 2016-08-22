
package config

import (
	"time"
)

type CdrStatsConfig struct {
	Id               string        // Config id, unique per config instance
	QueueLength      int           // Number of items in the stats buffer
	TimeWindow       time.Duration // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
	SaveInterval     time.Duration
	Metrics          []string    // ASR, ACD, ACC
	SetupInterval    []time.Time // 2 or less items (>= start interval,< stop_interval)
	TORs             []string
	CdrHosts         []string
	CdrSources       []string
	ReqTypes         []string
	Directions       []string
	Tenants          []string
	Categories       []string
	Accounts         []string
	Subjects         []string
	DestinationIds   []string
	UsageInterval    []time.Duration // 2 or less items (>= Usage, <Usage)
	Suppliers        []string
	DisconnectCauses []string
	MediationRunIds  []string
	RatedAccounts    []string
	RatedSubjects    []string
	CostInterval     []float64 // 2 or less items, (>=Cost, <Cost)
}
