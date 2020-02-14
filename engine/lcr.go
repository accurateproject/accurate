package engine

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

const (
	LCR_STRATEGY_STATIC        = "*static"
	LCR_STRATEGY_LOWEST        = "*lowest_cost"
	LCR_STRATEGY_HIGHEST       = "*highest_cost"
	LCR_STRATEGY_QOS_THRESHOLD = "*qos_threshold"
	LCR_STRATEGY_QOS           = "*qos"
	LCR_STRATEGY_LOAD          = "*load_distribution"

	// used for load distribution sorting
	RAND_LIMIT          = 99
	LOW_PRIORITY_LIMIT  = 100
	MED_PRIORITY_LIMIT  = 200
	HIGH_PRIORITY_LIMIT = 300
)

// A request for LCR, used in APIer and SM where we need to expose it
type LcrRequest struct {
	Direction    string
	Tenant       string
	Category     string
	Account      string
	Subject      string
	Destination  string
	SetupTime    string
	Duration     string
	IgnoreErrors bool
	ExtraFields  map[string]string
	*LCRFilter
	*utils.Paginator
}

type LCRFilter struct {
	MinCost *float64
	MaxCost *float64
}

func (self *LcrRequest) AsCallDescriptor(timezone string) (*CallDescriptor, error) {
	if len(self.Account) == 0 || len(self.Destination) == 0 {
		return nil, utils.ErrMandatoryIeMissing
	}
	// Set defaults
	if len(self.Direction) == 0 {
		self.Direction = utils.OUT
	}
	if len(self.Tenant) == 0 {
		self.Tenant = *config.Get().General.DefaultTenant
	}
	if len(self.Category) == 0 {
		self.Category = *config.Get().General.DefaultCategory
	}
	if len(self.Subject) == 0 {
		self.Subject = self.Account
	}
	var timeStart time.Time
	var err error
	if len(self.SetupTime) == 0 {
		timeStart = time.Now()
	} else if timeStart, err = utils.ParseTimeDetectLayout(self.SetupTime, timezone); err != nil {
		return nil, err
	}
	var callDur time.Duration
	if len(self.Duration) == 0 {
		callDur = time.Duration(1) * time.Minute
	} else if callDur, err = utils.ParseDurationWithSecs(self.Duration); err != nil {
		return nil, err
	}
	cd := &CallDescriptor{
		Direction:   self.Direction,
		Tenant:      self.Tenant,
		Category:    self.Category,
		Account:     self.Account,
		Subject:     self.Subject,
		Destination: self.Destination,
		TimeStart:   timeStart,
		TimeEnd:     timeStart.Add(callDur),
	}
	if self.ExtraFields != nil {
		cd.ExtraFields = make(map[string]string)
	}
	for key, val := range self.ExtraFields {
		cd.ExtraFields[key] = val
	}
	return cd, nil
}

// A LCR reply, used in APIer and SM where we need to expose it
type LcrReply struct {
	DestinationID string
	RPCategory    string
	Strategy      string
	Suppliers     []*LcrSupplier
}

// One supplier out of LCR reply
type LcrSupplier struct {
	Supplier string
	Cost     *dec.Dec
	QOS      map[string]*dec.Dec
}

type LCR struct {
	Direction   string           `bson:"direction"`
	Tenant      string           `bson:"tenant"`
	Category    string           `bson:"category"`
	Account     string           `bson:"account"`
	Subject     string           `bson:"subject"`
	Activations []*LCRActivation `bson:"activations"`
}
type LCRActivation struct {
	ActivationTime time.Time   `bson:"activation_time"`
	Entries        []*LCREntry `bson:"entries"`
	parentLCR      *LCR
}
type LCREntry struct {
	DestinationID  string  `bson:"destination_id"`
	RPCategory     string  `bson:"rp_category"`
	Strategy       string  `bson:"strategy"`
	StrategyParams string  `bson:"strategy_params"`
	Weight         float64 `bson:"weight"`
	precision      int
}

type LCRCost struct {
	Entry         *LCREntry
	SupplierCosts []*LCRSupplierCost
}

type LCRSupplierCost struct {
	Supplier       string
	Cost           *dec.Dec
	Duration       time.Duration
	Error          string // Not error due to JSON automatic serialization into struct
	QOS            map[string]*dec.Dec
	qosSortParams  []string
	supplierQueues []*StatsQueue // used for load distribution
}

func (lsc *LCRSupplierCost) getCost() *dec.Dec {
	if lsc.Cost == nil {
		lsc.Cost = dec.New()
	}
	return lsc.Cost
}

func (lcr *LCR) FullID() string {
	return utils.ConcatKey(lcr.Direction, lcr.Category, lcr.Account, lcr.Subject)
}

func (lcr *LCR) Sort() {
	sort.Slice(lcr.Activations, func(i, j int) bool {
		return lcr.Activations[i].ActivationTime.Before(lcr.Activations[j].ActivationTime)
	})
}

func (lcr *LCR) setParentLCR() {
	for _, lcra := range lcr.Activations {
		lcra.parentLCR = lcr
	}
}

func (lcr *LCR) precision() int {
	precision := 0
	if lcr.Direction != utils.ANY {
		precision++
	}
	if lcr.Tenant != utils.ANY {
		precision++
	}
	if lcr.Category != utils.ANY {
		precision++
	}
	if lcr.Account != utils.ANY {
		precision++
	}
	if lcr.Subject != utils.ANY {
		precision++
	}
	return precision
}

func (le *LCREntry) GetQOSLimits() (minASR, maxASR float64, minPDD, maxPDD, minACD, maxACD, minTCD, maxTCD time.Duration, minACC, maxACC, minTCC, maxTCC, minDDC, maxDDC float64) {
	// MIN_ASR;MAX_ASR;MIN_PDD;MAX_PDD;MIN_ACD;MAX_ACD;MIN_TCD;MAX_TCD;MIN_ACC;MAX_ACC;MIN_TCC;MAX_TCC;MIN_DDC;MAX_DDC
	minASR, maxASR, minPDD, maxPDD, minACD, maxACD, minTCD, maxTCD, minACC, maxACC, minTCC, maxTCC, minDDC, maxDDC = -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1
	params := strings.Split(le.StrategyParams, utils.INFIELD_SEP)
	if len(params) == 14 {
		var err error
		if minASR, err = strconv.ParseFloat(params[0], 64); err != nil {
			minASR = -1
		}
		if maxASR, err = strconv.ParseFloat(params[1], 64); err != nil {
			maxASR = -1
		}
		if minPDD, err = utils.ParseDurationWithSecs(params[2]); err != nil {
			minPDD = -1
		}
		if maxPDD, err = utils.ParseDurationWithSecs(params[3]); err != nil {
			maxPDD = -1
		}
		if minACD, err = utils.ParseDurationWithSecs(params[4]); err != nil {
			minACD = -1
		}
		if maxACD, err = utils.ParseDurationWithSecs(params[5]); err != nil {
			maxACD = -1
		}
		if minTCD, err = utils.ParseDurationWithSecs(params[6]); err != nil {
			minTCD = -1
		}
		if maxTCD, err = utils.ParseDurationWithSecs(params[7]); err != nil {
			maxTCD = -1
		}
		if minACC, err = strconv.ParseFloat(params[8], 64); err != nil {
			minACC = -1
		}
		if maxACC, err = strconv.ParseFloat(params[9], 64); err != nil {
			maxACC = -1
		}
		if minTCC, err = strconv.ParseFloat(params[10], 64); err != nil {
			minTCC = -1
		}
		if maxTCC, err = strconv.ParseFloat(params[11], 64); err != nil {
			maxTCC = -1
		}
		if minDDC, err = strconv.ParseFloat(params[12], 64); err != nil {
			minDDC = -1
		}
		if maxDDC, err = strconv.ParseFloat(params[13], 64); err != nil {
			maxDDC = -1
		}
	}
	return
}

func (le *LCREntry) GetParams() []string {
	// ASR;ACD
	params := strings.Split(le.StrategyParams, utils.INFIELD_SEP)
	// eliminate empty strings
	var cleanParams []string
	for _, p := range params {
		p = strings.TrimSpace(p)
		if p != "" {
			cleanParams = append(cleanParams, p)
		}
	}
	if len(cleanParams) == 0 && le.Strategy == LCR_STRATEGY_QOS {
		return []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC} // Default QoS stats if none configured
	}
	return cleanParams
}

type LCREntriesSorter []*LCREntry

func (es LCREntriesSorter) Sort() {
	sort.Slice(es, func(j, i int) bool {
		// we need the best earlyer in the list
		return es[i].Weight < es[j].Weight ||
			(es[i].Weight == es[j].Weight && es[i].precision < es[j].precision)

	})
}

func (lcra *LCRActivation) GetLCREntryForPrefix(destination string) *LCREntry {
	var potentials LCREntriesSorter
	if dests, err := ratingStorage.GetDestinations(lcra.parentLCR.Tenant, destination, "", utils.DestMatching, utils.CACHED); err == nil {
		for _, dest := range dests {
			for _, entry := range lcra.Entries {
				if entry.DestinationID == dest.Name {
					entry.precision = len(dest.Code)
					potentials = append(potentials, entry)
				}
			}
		}
	}
	if len(potentials) > 0 {
		// sort by precision and weight
		potentials.Sort()
		return potentials[0]
	}
	// return the *any entry if it exists
	for _, entry := range lcra.Entries {
		if entry.DestinationID == utils.ANY {
			return entry
		}
	}
	return nil
}

func (lc *LCRCost) Sort() {
	switch lc.Entry.Strategy {
	case LCR_STRATEGY_LOWEST, LCR_STRATEGY_QOS_THRESHOLD:
		sort.Sort(LowestSupplierCostSorter(lc.SupplierCosts))
	case LCR_STRATEGY_HIGHEST:
		sort.Sort(HighestSupplierCostSorter(lc.SupplierCosts))
	case LCR_STRATEGY_QOS:
		sort.Sort(QOSSorter(lc.SupplierCosts))
	case LCR_STRATEGY_LOAD:
		lc.SortLoadDistribution()
		sort.Sort(HighestSupplierCostSorter(lc.SupplierCosts))
	}
}

func (lc *LCRCost) SortLoadDistribution() {
	// find the time window that is common to all qeues
	scoreBoard := make(map[time.Duration]int) // register TimeWindow across suppliers

	var winnerTimeWindow time.Duration
	maxScore := 0
	for _, supCost := range lc.SupplierCosts {
		timeWindowFlag := make(map[time.Duration]bool) // flags appearance in same supplier
		for _, sq := range supCost.supplierQueues {
			if !timeWindowFlag[sq.conf.TimeWindow] {
				timeWindowFlag[sq.conf.TimeWindow] = true
				scoreBoard[sq.conf.TimeWindow]++
			}
			if scoreBoard[sq.conf.TimeWindow] > maxScore {
				maxScore = scoreBoard[sq.conf.TimeWindow]
				winnerTimeWindow = sq.conf.TimeWindow
			}
		}
	}
	supplierQueues := make(map[*LCRSupplierCost]*StatsQueue)
	for _, supCost := range lc.SupplierCosts {
		for _, sq := range supCost.supplierQueues {
			if sq.conf.TimeWindow == winnerTimeWindow {
				supplierQueues[supCost] = sq
				break
			}
		}
	}
	/*for supplier, sq := range supplierQueues {
		log.Printf("Useful supplier qeues: %s %v", supplier, sq.conf.TimeWindow)
	}*/
	// if all have less than ratio return random order
	// if some have a cdr count not divisible by ratio return them first and all ordered by cdr times, oldest first
	// if all have a multiple of ratio return in the order of cdr times, oldest first

	// first put them in one of the above categories
	haveRatiolessSuppliers := false
	for supCost, sq := range supplierQueues {
		ratio := lc.GetSupplierRatio(supCost.Supplier)
		if ratio == -1 {
			supCost.Cost = dec.NewVal(-1, 0)
			haveRatiolessSuppliers = true
			continue
		}
		cdrCount := sq.length()
		if cdrCount < ratio {
			supCost.Cost = dec.NewFloat(float64(LOW_PRIORITY_LIMIT + rand.Intn(RAND_LIMIT)))
			continue
		}
		if cdrCount%ratio == 0 {
			supCost.Cost = dec.NewFloat(float64(MED_PRIORITY_LIMIT+rand.Intn(RAND_LIMIT)) + (time.Now().Sub(sq.getLastSetupTime()).Seconds() / RAND_LIMIT))
			continue
		} else {
			supCost.Cost = dec.NewFloat(float64(HIGH_PRIORITY_LIMIT+rand.Intn(RAND_LIMIT)) + (time.Now().Sub(sq.getLastSetupTime()).Seconds() / RAND_LIMIT))
			continue
		}
	}
	if haveRatiolessSuppliers {
		var filteredSupplierCost []*LCRSupplierCost
		for _, supCost := range lc.SupplierCosts {
			if supCost.Cost.Cmp(dec.NewVal(-1, 0)) != 0 {
				filteredSupplierCost = append(filteredSupplierCost, supCost)
			}
		}
		lc.SupplierCosts = filteredSupplierCost
	}
}

// used in load distribution strategy only
// receives a long supplier id and will return the ratio found in strategy params
func (lc *LCRCost) GetSupplierRatio(supplier string) int {
	// parse strategy params
	ratios := make(map[string]int)
	params := strings.Split(lc.Entry.StrategyParams, utils.INFIELD_SEP)
	for _, param := range params {
		ratioSlice := strings.Split(param, utils.CONCATENATED_KEY_SEP)
		if len(ratioSlice) != 2 {
			utils.Logger.Warn("bad format in load distribution strategy", zap.String("param", lc.Entry.StrategyParams))
			continue
		}
		p, err := strconv.Atoi(ratioSlice[1])
		if err != nil {
			utils.Logger.Warn("bad format in load distribution strategy", zap.String("param", lc.Entry.StrategyParams))
			continue
		}
		ratios[ratioSlice[0]] = p
	}
	parts := strings.Split(supplier, utils.CONCATENATED_KEY_SEP)
	if len(parts) > 0 {
		supplierSubject := parts[len(parts)-1]
		if ratio, found := ratios[supplierSubject]; found {
			return ratio
		}
		if ratio, found := ratios[utils.META_DEFAULT]; found {
			return ratio
		}
	}
	if len(ratios) == 0 {
		return 1 // use random/last cdr date sorting
	}
	return -1 // exclude missing suppliers
}

func (lc *LCRCost) HasErrors() bool {
	for _, supplCost := range lc.SupplierCosts {

		if len(supplCost.Error) != 0 {
			return true
		}
	}
	return false
}

func (lc *LCRCost) LogErrors() {
	for _, supplCost := range lc.SupplierCosts {
		if len(supplCost.Error) != 0 {
			utils.Logger.Error("LCR_ERROR: supplier", zap.String("supplier", supplCost.Supplier), zap.String("error", supplCost.Error))
		}
	}
}

func (lc *LCRCost) SuppliersSlice() ([]string, error) {
	if lc.Entry == nil {
		return nil, utils.ErrNotFound
	}
	supps := []string{}
	for _, supplCost := range lc.SupplierCosts {
		if supplCost.Error != "" {
			continue // Do not add the supplier with cost errors to list of suppliers available
		}
		if dtcs, err := utils.NewDTCSFromRPKey(supplCost.Supplier); err != nil {
			return nil, err
		} else if len(dtcs.Subject) != 0 {
			supps = append(supps, dtcs.Subject)
		}
	}
	if len(supps) == 0 {
		return nil, utils.ErrNotFound
	}
	return supps, nil
}

// Returns a list of suppliers separated via
func (lc *LCRCost) SuppliersString() (string, error) {
	supps, err := lc.SuppliersSlice()
	if err != nil {
		return "", err
	}
	supplStr := ""
	for idx, suppl := range supps {
		if idx != 0 {
			supplStr += utils.FIELDS_SEP
		}
		supplStr += suppl
	}
	return supplStr, nil
}

type LowestSupplierCostSorter []*LCRSupplierCost

func (lscs LowestSupplierCostSorter) Len() int {
	return len(lscs)
}

func (lscs LowestSupplierCostSorter) Swap(i, j int) {
	lscs[i], lscs[j] = lscs[j], lscs[i]
}

func (lscs LowestSupplierCostSorter) Less(i, j int) bool {
	return lscs[i].getCost().Cmp(lscs[j].getCost()) < 0
}

type HighestSupplierCostSorter []*LCRSupplierCost

func (hscs HighestSupplierCostSorter) Len() int {
	return len(hscs)
}

func (hscs HighestSupplierCostSorter) Swap(i, j int) {
	hscs[i], hscs[j] = hscs[j], hscs[i]
}

func (hscs HighestSupplierCostSorter) Less(i, j int) bool {
	return hscs[i].getCost().Cmp(hscs[j].getCost()) > 0
}

type QOSSorter []*LCRSupplierCost

func (qoss QOSSorter) Len() int {
	return len(qoss)
}

func (qoss QOSSorter) Swap(i, j int) {
	qoss[i], qoss[j] = qoss[j], qoss[i]
}

func (qoss QOSSorter) Less(i, j int) bool {
	for _, param := range qoss[i].qosSortParams {
		// if one of the supplier is missing the qos parram skip to next one
		if _, exists := qoss[i].QOS[param]; !exists {
			continue
		}
		if _, exists := qoss[j].QOS[param]; !exists {
			continue
		}
		// skip to next param
		if qoss[i].QOS[param] == qoss[j].QOS[param] {
			continue
		}
		// -1 is the best
		if qoss[j].QOS[param].Cmp(dec.MinusOne) == 0 {
			return false
		}
		// more is better
		if qoss[i].QOS[param].Cmp(dec.MinusOne) == 0 || qoss[i].QOS[param].Cmp(qoss[j].QOS[param]) > 0 {
			return true
		}
	}
	return false
}

type LCRList []*LCR // used in GetLCR

func (lcrl LCRList) Sort() {
	sort.Slice(lcrl, func(j, i int) bool {
		// we need higher precision earlyer in the list
		return lcrl[i].precision() < lcrl[j].precision()
	})
}
