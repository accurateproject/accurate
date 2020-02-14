package engine

import "github.com/accurateproject/accurate/utils"

type Storage interface {
	Close()
	Ping() error
	Flush() error
	EnsureIndexes() error
	Count(col string) (int, error)
	Iterator(col, sort string, filter map[string]interface{}) Iterator
	GetAllPaged(tenant string, out interface{}, collection string, limit, offset int) error
	GetByNames(tenant string, names []string, out interface{}, collection string) error
	PreloadCacheForPrefix(string) error
	RemoveTenant(tenant string, collections ...string) error
}

// Interface for storage providers.
type RatingStorage interface {
	Storage
	PreloadRatingCache() error
	GetTiming(tenant, name string) (*Timing, error)
	SetTiming(*Timing) error
	GetRate(tenant, name string) (*Rate, error)
	SetRate(*Rate) error
	GetDestinationRate(tenant, name string) (*DestinationRate, error)
	SetDestinationRate(*DestinationRate) error
	GetRatingPlan(tenant, name, cacheParam string) (*RatingPlan, error)
	SetRatingPlan(*RatingPlan) error
	GetRatingProfile(direction, tenant, category, subject string, prefixMatching bool, cacheParam string) (*RatingProfile, error)
	GetRatingProfiles(direction, tenant, category, subject, cacheParam string) ([]*RatingProfile, error)
	SetRatingProfile(*RatingProfile) error
	RemoveRatingProfile(direction, tenant, category, subject string) error
	GetDestinations(tenant, code, name, strategy, cacheParam string) (Destinations, error)
	SetDestination(*Destination) error
	RemoveDestination(*Destination) error
	RemoveDestinations(tenant, code, name string) error
	GetLCR(direction, tenant, category, account, subject string, prefixMatching bool, cacheParam string) (*LCR, error)
	SetLCR(*LCR) error
	SetCdrStats(*CdrStats) error
	GetCdrStats(tenant, name string) (*CdrStats, error)
	RemoveCdrStats(tenant, name string) error
	GetDerivedChargers(direction, tenant, category, account, subject, cacheParam string) (utils.DerivedChargers, error)
	SetDerivedChargers(*utils.DerivedChargerGroup) error
	GetActionGroup(tenant, name, cacheParam string) (*ActionGroup, error)
	SetActionGroup(*ActionGroup) error
	RemoveActionGroup(tenant, name string) error
	GetSharedGroup(tenant, name, cacheParam string) (*SharedGroup, error)
	SetSharedGroup(*SharedGroup) error
	GetActionTriggers(tenant, name, cacheParam string) (*ActionTriggerGroup, error)
	SetActionTriggers(*ActionTriggerGroup) error
	RemoveActionTriggers(tenant, name string) error
	GetActionPlanBinding(tenant, accountName, actionPlanName string) (*ActionPlanBinding, error)
	SetActionPlanBinding(apb *ActionPlanBinding) error
	RemoveActionPlanBindings(tenant, accountName, actionPlanName string) error
	GetActionPlan(tenant, name, cacheParam string) (*ActionPlan, error)
	SetActionPlan(*ActionPlan) error
	PushTask(*Task) error
	PopTask() (*Task, error)
}

type AccountingStorage interface {
	Storage
	PreloadAccountingCache() error
	GetAccount(tenant, name string) (*Account, error)
	SetAccount(*Account) error
	RemoveAccount(tenant, name string) error
	SetSimpleAccount(*SimpleAccount) error
	RemoveSimpleAccount(tenant, name string) error
	GetCdrStatsQueue(tenant, name string) (*StatsQueue, error)
	SetCdrStatsQueue(*StatsQueue) error
	RemoveCdrStatsQueue(tenant, name string) error
	PushQCDR(qcdr *QCDR) error
	PopQCDR(tenant, name string, filter map[string]interface{}, limit int) ([]*QCDR, error)
	RemoveQCDRs(tenant, name string) error
	GetSubscribers() (map[string]*SubscriberData, error)
	SetSubscriber(string, *SubscriberData) error
	RemoveSubscriber(string) error
	SetUser(*UserProfile) error
	GetUser(tenant, name string) (*UserProfile, error)
	RemoveUser(tenant, user string) error
	SetAlias(*Alias) error
	GetAlias(direction, tenant, category, account, subject, context, cacheParam string) (*Alias, error)
	RemoveAlias(direction, tenant, category, account, subject, context string) error
	GetReverseAlias(tenant, context, target, alias, cacheParam string) ([]*Alias, error)
	GetResourceLimit(string, bool, string) (*ResourceLimit, error)
	SetResourceLimit(*ResourceLimit, string) error
	RemoveResourceLimit(string, string) error
	AddLoadHistory(*utils.LoadInstance) error
	GetStructVersion() (*StructVersion, error)
	SetStructVersion(*StructVersion) error
}

type CdrStorage interface {
	Storage
	SetCDR(cdr *CDR, update bool) error
	SetSMCost(smc *SMCost) error
	GetSMCosts(uniqueid, runid, originHost, originIDPrfx string) ([]*SMCost, error)
	GetCDRs(*utils.CDRsFilter, bool) ([]*CDR, int64, error)
}

type Iterator interface {
	All(result interface{}) error
	Close() error
	Done() bool
	Err() error
	Next(result interface{}) bool
	Timeout() bool
}
