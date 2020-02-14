package utils

import "time"

// Structs here are one to one mapping of the tables and fields
// to be used by gorm orm

type TpTiming struct {
	Tenant    string
	Tag       string
	Years     Years
	Months    Months
	MonthDays MonthDays
	WeekDays  WeekDays
	Time      string
}

type TpDestination struct {
	Tenant string
	Code   string
	Tag    string
}

type TpRate struct {
	Tenant string
	Tag    string
	Slots  []*rateSlot
}

type rateSlot struct {
	ConnectFee         float64
	Rate               float64
	RateUnit           string
	RateIncrement      string
	GroupIntervalStart string
}

type TpDestinationRate struct {
	Tenant   string
	Tag      string
	Bindings []*destinationRateBinding
}

type destinationRateBinding struct {
	DestinationCode string
	DestinationTag  string
	RatesTag        string
	MaxCost         float64
	MaxCostStrategy string
}

type TpRatingPlan struct {
	Tenant   string
	Tag      string
	Bindings []*ratingPlanBinding
}

type ratingPlanBinding struct {
	DestinationRatesTag string
	TimingTag           string
	Weight              float64
}

type TpRatingProfile struct {
	Tenant      string
	Direction   string
	Category    string
	Subject     string
	Activations []*ratingProfileActivation
}

type ratingProfileActivation struct {
	ActivationTime   string
	RatingPlanTag    string
	FallbackSubjects []string
	CdrStatQueueIDs  []string
}

type TpLcrRule struct {
	Tenant      string
	Direction   string
	Category    string
	Account     string
	Subject     string
	Activations []*lcrActivation
}
type lcrActivation struct {
	ActivationTime string
	Entries        []*lcrEntry
}
type lcrEntry struct {
	DestinationTag string
	RPCategory     string
	Strategy       string
	StrategyParams string
	Weight         float64
}

type TpActionGroup struct {
	Tenant  string
	Tag     string
	Actions []*actionBody
}

type actionBody struct {
	Action     string
	TOR        string
	Params     string
	ExecFilter string
	Filter     string
	Weight     float64
}

type TpActionPlan struct {
	Tenant        string
	Tag           string
	ActionTimings []*actionTiming
}

type actionTiming struct {
	TimingTag  string  `bson:"timing"`
	ActionsTag string  `bson:"actions_id"`
	Weight     float64 `bson:"weight"`
}

type TpActionTrigger struct {
	Tenant   string
	Tag      string
	Triggers []*triggerBody
}

type triggerBody struct {
	UniqueID       string
	ThresholdType  string
	ThresholdValue float64
	Recurrent      bool
	MinSleep       string
	ExpiryTime     string
	ActivationTime string
	TOR            string
	Filter         string
	MinQueuedItems int
	ActionsTag     string
	Weight         float64
}

type TpAccountAction struct {
	Tenant            string
	Account           string
	ActionPlanTags    []string
	ActionTriggerTags []string
	AllowNegative     bool
	Disabled          bool
}

type TpSharedGroup struct {
	Tenant            string
	Tag               string
	AccountParameters map[string]*sharingParameters
	MemberIDs         []string
}

type sharingParameters struct {
	Strategy      string
	RatingSubject string
}

type TpDerivedCharger struct {
	Tenant         string
	Direction      string
	Category       string
	Account        string
	Subject        string
	DestinationIDs []string
	Chargers       []*derivedCharger
}

type derivedCharger struct {
	RunID  string
	Filter string
	Fields string
}

type TpCdrStats struct {
	Tenant            string
	Tag               string
	QueueLength       int
	TimeWindow        string
	Metrics           []string
	Filter            string
	ActionTriggerTags []string `index:"24" re:""`
	Disabled          bool
}

type TpUser struct {
	Tenant string
	Name   string
	Masked bool
	Index  map[string]string
	Query  string
	Weight float64
}

type TpAlias struct {
	Tenant    string
	Direction string
	Category  string
	Account   string
	Subject   string
	Context   string
	Indexes   []*aliasIndex
	Values    []*aliasValue
}

type aliasValue struct {
	DestinationTag string
	Fields         string
	Weight         float64
}

type aliasIndex struct {
	Target string
	Alias  string
}

type TpResourceLimit struct {
	Tenant  string
	Tag     string
	Filters []*rsFilter
}

type rsFilter struct {
	Type              string
	FieldName         string
	Values            []string
	ActivationTime    string
	Weight            float64
	Limit             float64
	ActionTriggersTag string
}

type TBLCDRs struct {
	ID              int64
	Cgrid           string
	RunID           string
	OriginHost      string
	Source          string
	OriginID        string
	Tor             string
	RequestType     string
	Direction       string
	Tenant          string
	Category        string
	Account         string
	Subject         string
	Destination     string
	SetupTime       time.Time
	Pdd             float64
	AnswerTime      time.Time
	Usage           float64
	Supplier        string
	DisconnectCause string
	ExtraFields     string
	Cost            float64
	CostDetails     string
	CostSource      string
	AccountSummary  string
	ExtraInfo       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

func (t TBLCDRs) TableName() string {
	return TBLCDRS
}

type SMCosts struct {
	ID          int64
	Cgrid       string
	RunID       string
	OriginHost  string
	OriginID    string
	CostSource  string
	Usage       float64
	CostDetails string
	CreatedAt   time.Time
	DeletedAt   *time.Time
}

func (t SMCosts) TableName() string {
	return TBLSMCosts
}
