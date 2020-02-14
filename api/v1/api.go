package v1

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/scheduler"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
)

const (
	OK = utils.OK
)

type ApiV1 struct {
	ratingDB    engine.RatingStorage
	accountDB   engine.AccountingStorage
	cdrDB       engine.CdrStorage
	sched       *scheduler.Scheduler
	cfg         *config.Config
	responder   *engine.Responder
	cdrStatsSrv rpcclient.RpcClientConnection
	users       rpcclient.RpcClientConnection
	tpReader    *engine.TpReader
}

func NewAPIV1(ratingDB engine.RatingStorage, accountingDB engine.AccountingStorage, cdrDB engine.CdrStorage, sched *scheduler.Scheduler, cfg *config.Config, responder *engine.Responder, cdrStatsSrv rpcclient.RpcClientConnection, users rpcclient.RpcClientConnection) *ApiV1 {
	return &ApiV1{
		ratingDB:    ratingDB,
		accountDB:   accountingDB,
		cdrDB:       cdrDB,
		sched:       sched,
		cfg:         cfg,
		responder:   responder,
		cdrStatsSrv: cdrStatsSrv,
		users:       users,
	}
}

type AttrLoadTpFromFolder struct {
	FolderPath string // Take files from folder absolute path
	FlushDB    bool   // Flush previous data before loading new one
}

func (api *ApiV1) LoadTariffPlanFromFolder(attr AttrLoadTpFromFolder, reply *utils.LoadInstance) error {
	if len(attr.FolderPath) == 0 {
		return fmt.Errorf("%s:%s", utils.ErrMandatoryIeMissing.Error(), "FolderPath")
	}
	if fi, err := os.Stat(attr.FolderPath); err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			return utils.ErrInvalidPath
		}
		return utils.NewErrServerError(err)
	} else if !fi.IsDir() {
		return utils.ErrInvalidPath
	}

	if attr.FlushDB {
		api.ratingDB.Flush()
	}
	tpReader, err := engine.LoadTariffPlanFromFolder(attr.FolderPath, *api.cfg.General.DefaultTimezone, api.ratingDB, api.accountDB)

	loadStats := tpReader.LoadStats()
	// Reload scheduler and cache
	r := ""

	if err = api.ReloadCache(utils.AttrReloadCache{Tenants: loadStats.Tenants.Slice()}, &r); err != nil {
		log.Printf("WARNING: Got error on cache reload: %s\n", err.Error())
	}

	if err = api.ReloadScheduler("", &r); err != nil {
		log.Printf("WARNING: Got error on scheduler reload: %s\n", err.Error())
	}

	if api.cdrStatsSrv != nil {
		for tenant, slice := range loadStats.CDRStats {
			statsQueueIDs := slice.Slice()
			var reply string
			if err := api.cdrStatsSrv.Call("CDRStatsV1.ReloadQueues", utils.AttrStatsQueueIDs{Tenant: tenant, IDs: statsQueueIDs}, &reply); err != nil {
				log.Printf("WARNING: Failed reloading stat queues, error: %s\n", err.Error())
			}
		}
	}

	if api.users != nil {
		for tenant := range loadStats.UserTenants {
			var reply string
			if err := api.users.Call("UsersV1.ReloadUsers", engine.AttrReloadUsers{Tenant: tenant}, &reply); err != nil {
				log.Printf("WARNING: Failed reloading users data, error: %s\n", err.Error())
			}
		}
	}

	/*loadHistList, err := api.AccountDb.GetLoadHistory(1, utils.CACHE_SKIP)
	if err != nil {
		return err
	}
	if len(loadHistList) > 0 {
		*reply = *loadHistList[0]
	}*/

	return nil
}

// Retrieves actions attached to specific ActionsId within cache
func (api *ApiV1) GetActions(attr AttrGetMultiple, reply *map[string]engine.Actions) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenant")
	}

	retActions := make(map[string]engine.Actions)
	var offset, limit int
	if attr.Offset != nil {
		offset = *attr.Offset
		if offset < 0 {
			offset = 0
		}
	}
	if attr.Limit != nil {
		limit = *attr.Limit
		if limit <= 0 {
			limit = 1
		}
	}
	ag := &engine.ActionGroup{}
	if len(attr.IDs) != 0 {
		actionKeys := attr.IDs
		var limitedActions []string
		if limit != 0 {
			max := math.Min(float64(offset+limit), float64(len(actionKeys)))
			limitedActions = actionKeys[offset:int(max)]
		} else {
			limitedActions = actionKeys[offset:]
		}
		err := api.ratingDB.GetByNames(attr.Tenant, limitedActions, ag, engine.ColAct)
		if err != nil && err != utils.ErrNotFound { // Not found is not an error here
			return err
		}
		if ag != nil {
			retActions[ag.Name] = ag.Actions
		}
	} else {
		if err := api.ratingDB.GetAllPaged(attr.Tenant, ag, engine.ColAct, offset, limit); err != nil && err != utils.ErrNotFound {
			return err
		}
		if ag != nil {
			retActions[ag.Name] = ag.Actions
		}
	}
	*reply = retActions
	return nil
}

func (api *ApiV1) GetDestinations(attr AttrGetMultiple, reply *engine.Destinations) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenant")
	}

	retDestinations := make(engine.Destinations, 0)
	var offset, limit int
	if attr.Offset != nil {
		offset = *attr.Offset
		if offset < 0 {
			offset = 0
		}
	}
	if attr.Limit != nil {
		limit = *attr.Limit
		if limit <= 0 {
			limit = 1
		}
	}
	if len(attr.IDs) != 0 {
		actionKeys := attr.IDs
		var limitedDestinations []string
		if limit != 0 {
			max := math.Min(float64(offset+limit), float64(len(actionKeys)))
			limitedDestinations = actionKeys[offset:int(max)]
		} else {
			limitedDestinations = actionKeys[offset:]
		}
		err := api.ratingDB.GetByNames(attr.Tenant, limitedDestinations, &retDestinations, engine.ColDst)
		if err != nil && err != utils.ErrNotFound { // Not found is not an error here
			return err
		}

	} else {
		if err := api.ratingDB.GetAllPaged(attr.Tenant, &retDestinations, engine.ColDst, offset, limit); err != nil && err != utils.ErrNotFound {
			return err
		}
	}
	*reply = retDestinations
	return nil
}

type AttrGetSingle struct {
	Tenant string
	ID     string
}

func (api *ApiV1) GetDestination(attr AttrGetSingle, reply *engine.Destinations) error {
	if dst, err := api.ratingDB.GetDestinations(attr.Tenant, "", attr.ID, utils.DestExact, utils.CACHED); err != nil {
		return utils.ErrNotFound
	} else {
		*reply = dst
	}
	return nil
}

type AttrRemoveDestination struct {
	Tenant         string
	DestinationIDs []string
	Codes          []string
}

func (api *ApiV1) RemoveDestination(attr AttrRemoveDestination, reply *string) error {
	for _, name := range attr.DestinationIDs {
		if err := api.ratingDB.RemoveDestinations(attr.Tenant, "", name); err != nil && err != utils.ErrNotFound {
			*reply = err.Error()
			return err
		}
	}

	for _, code := range attr.Codes {
		if err := api.ratingDB.RemoveDestinations(attr.Tenant, code, ""); err != nil && err != utils.ErrNotFound {
			*reply = err.Error()
			return err
		}
	}

	return nil
}

func (apier *ApiV1) GetSharedGroup(attr AttrGetSingle, reply *engine.SharedGroup) error {
	sg, err := apier.ratingDB.GetSharedGroup(attr.Tenant, attr.ID, utils.CACHED)
	if err != nil && err != utils.ErrNotFound { // Not found is not an error here
		return err
	}
	if sg != nil {
		*reply = *sg
	}
	return nil
}

func (api *ApiV1) GetRatingPlan(attr AttrGetSingle, reply *engine.RatingPlan) error {
	if rpln, err := api.ratingDB.GetRatingPlan(attr.Tenant, attr.ID, utils.CACHED); err != nil {
		return utils.ErrNotFound
	} else {
		*reply = *rpln
	}
	return nil
}

func (api *ApiV1) ExecuteAction(attr utils.AttrExecuteAction, reply *string) error {
	at := &engine.ActionTiming{
		ActionsID: attr.ActionsID,
	}
	// set parent action plan for Tenant
	apl := &engine.ActionPlan{
		Tenant:        attr.Tenant,
		ActionTimings: []*engine.ActionTiming{at},
	}
	apl.SetParentActionPlan()
	if attr.Tenant != "" && attr.Account != "" {
		at.SetAccountIDs(utils.StringMap{attr.Account: true})
	}
	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

/*
type AttrSetRatingProfile struct {
	Tenant                string                      // Tenant's Id
	Category              string                      // TypeOfRecord
	Direction             string                      // Traffic direction, OUT is the only one supported for now
	Subject               string                      // Rating subject, usually the same as account
	Overwrite             bool                        // Overwrite if exists
	RatingPlanActivations []*utils.TPRatingActivation // Activate rating plans at specific time
}

// Sets a specific rating profile working with data directly in the ratingDB without involving storDb
func (api *ApiV1) SetRatingProfile(attrs AttrSetRatingProfile, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "TOR", "Direction", "Subject", "RatingPlanActivations"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, rpa := range attrs.RatingPlanActivations {
		if missing := utils.MissingStructFields(rpa, []string{"ActivationTime", "RatingPlanId"}); len(missing) != 0 {
			return fmt.Errorf("%s:RatingPlanActivation:%v", utils.ErrMandatoryIeMissing.Error(), missing)
		}
	}
	tpRpf := utils.TPRatingProfile{Tenant: attrs.Tenant, Category: attrs.Category, Direction: attrs.Direction, Subject: attrs.Subject}
	keyId := tpRpf.KeyId()
	var rpfl *engine.RatingProfile
	if !attrs.Overwrite {
		if exists, err := api.ratingDB.HasData(utils.RATING_PROFILE_PREFIX, keyId); err != nil {
			return utils.NewErrServerError(err)
		} else if exists {
			var err error
			if rpfl, err = api.ratingDB.GetRatingProfile(keyId, utils.CACHED); err != nil {
				return utils.NewErrServerError(err)
			}
		}
	}
	if rpfl == nil {
		rpfl = &engine.RatingProfile{Id: keyId, RatingPlanActivations: make(engine.RatingPlanActivations, 0)}
	}
	for _, ra := range attrs.RatingPlanActivations {
		at, err := utils.ParseTimeDetectLayout(ra.ActivationTime, *api.Config.General.DefaultTimezone)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("%s:Cannot parse activation time from %v", utils.ErrServerError.Error(), ra.ActivationTime))
		}
		if exists, err := api.ratingDB.HasData(utils.RATING_PLAN_PREFIX, ra.RatingPlanId); err != nil {
			return utils.NewErrServerError(err)
		} else if !exists {
			return fmt.Errorf(fmt.Sprintf("%s:RatingPlanId:%s", utils.ErrNotFound.Error(), ra.RatingPlanId))
		}
		rpfl.RatingPlanActivations = append(rpfl.RatingPlanActivations, &engine.RatingPlanActivation{ActivationTime: at, RatingPlanId: ra.RatingPlanId,
			FallbackKeys: utils.FallbackSubjKeys(tpRpf.Direction, tpRpf.Tenant, tpRpf.Category, ra.FallbackSubjects)})
	}
	if err := api.ratingDB.SetRatingProfile(rpfl); err != nil {
		return utils.NewErrServerError(err)
	}
	cache2go.RemPrefixKey(utils.RATING_PLAN_PREFIX, utils.CACHE_SKIP)
	api.ratingDB.PreloadCacheForPrefix(utils.RATING_PLAN_PREFIX)
	*reply = OK
	return nil
}

func (api *ApiV1) SetActions(attrs utils.AttrSetActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"ActionsId", "Actions"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), action.Identifier, missing)
		}
	}
	if !attrs.Overwrite {
		if exists, err := api.ratingDB.HasData(utils.ACTION_PREFIX, attrs.ActionsId); err != nil {
			return utils.NewErrServerError(err)
		} else if exists {
			return utils.ErrExists
		}
	}
	storeActions := make(engine.Actions, len(attrs.Actions))
	for idx, apiAct := range attrs.Actions {
		var vf *utils.ValueFormula
		if apiAct.Units != "" {
			if x, err := utils.ParseBalanceFilterValue(apiAct.Units); err == nil {
				vf = x
			} else {
				return err
			}
		}

		var weight *float64
		if apiAct.BalanceWeight != "" {
			if x, err := strconv.ParseFloat(apiAct.BalanceWeight, 64); err == nil {
				weight = &x
			} else {
				return err
			}
		}

		a := &engine.Action{
			Id:               attrs.ActionsId,
			ActionType:       apiAct.Identifier,
			Weight:           apiAct.Weight,
			ExpirationString: apiAct.ExpiryTime,
			ExtraParameters:  apiAct.ExtraParameters,
			Filter:           apiAct.Filter,
			Balance: &engine.BalanceFilter{ // TODO: update this part
				Uuid:           utils.StringPointer(apiAct.BalanceUuid),
				ID:             utils.StringPointer(apiAct.BalanceId),
				Type:           utils.StringPointer(apiAct.BalanceType),
				Value:          vf,
				Weight:         weight,
				Directions:     utils.StringMapPointer(utils.ParseStringMap(apiAct.Directions)),
				DestinationIDs: utils.StringMapPointer(utils.ParseStringMap(apiAct.DestinationIds)),
				RatingSubject:  utils.StringPointer(apiAct.RatingSubject),
				SharedGroups:   utils.StringMapPointer(utils.ParseStringMap(apiAct.SharedGroups)),
			},
		}
		storeActions[idx] = a
	}
	if err := api.ratingDB.SetActions(attrs.ActionsId, storeActions); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}

// Retrieves actions attached to specific ActionsId within cache
func (api *ApiV1) GetActions(actsId string, reply *[]*utils.TPAction) error {
	if len(actsId) == 0 {
		return fmt.Errorf("%s ActionsId: %s", utils.ErrMandatoryIeMissing.Error(), actsId)
	}
	acts := make([]*utils.TPAction, 0)
	engActs, err := api.ratingDB.GetActions(actsId, utils.CACHED)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	for _, engAct := range engActs {
		act := &utils.TPAction{
			Identifier:      engAct.ActionType,
			ExpiryTime:      engAct.ExpirationString,
			ExtraParameters: engAct.ExtraParameters,
			Filter:          engAct.Filter,
			Weight:          engAct.Weight,
		}
		bf := engAct.Balance
		if bf != nil {
			act.BalanceType = bf.GetType()
			act.Units = strconv.FormatFloat(bf.GetValue(), 'f', -1, 64)
			act.Directions = bf.GetDirections().String()
			act.DestinationIds = bf.GetDestinationIDs().String()
			act.RatingSubject = bf.GetRatingSubject()
			act.SharedGroups = bf.GetSharedGroups().String()
			act.BalanceWeight = strconv.FormatFloat(bf.GetWeight(), 'f', -1, 64)
			act.TimingTags = bf.GetTimingIDs().String()
			act.BalanceId = bf.GetID()
			act.Categories = bf.GetCategories().String()
			act.BalanceBlocker = strconv.FormatBool(bf.GetBlocker())
			act.BalanceDisabled = strconv.FormatBool(bf.GetDisabled())
		}
		acts = append(acts, act)
	}
	*reply = acts
	return nil
}

type AttrSetActionPlan struct {
	Id              string            // Profile id
	ActionPlan      []*AttrActionPlan // Set of actions this Actions profile will perform
	Overwrite       bool              // If previously defined, will be overwritten
	ReloadScheduler bool              // Enables automatic reload of the scheduler (eg: useful when adding a single action timing)
}

type AttrActionPlan struct {
	ActionsId string  // Actions id
	Years     string  // semicolon separated list of years this timing is valid on, *any or empty supported
	Months    string  // semicolon separated list of months this timing is valid on, *any or empty supported
	MonthDays string  // semicolon separated list of month's days this timing is valid on, *any or empty supported
	WeekDays  string  // semicolon separated list of week day names this timing is valid on *any or empty supported
	Time      string  // String representing the time this timing starts on, *asap supported
	Weight    float64 // Binding's weight
}

func (api *ApiV1) SetActionPlan(attrs AttrSetActionPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Id", "ActionPlan"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, at := range attrs.ActionPlan {
		requiredFields := []string{"ActionsId", "Time", "Weight"}
		if missing := utils.MissingStructFields(at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), at.ActionsId, missing)
		}
	}
	if !attrs.Overwrite {
		if exists, err := api.ratingDB.HasData(utils.ACTION_PLAN_PREFIX, attrs.Id); err != nil {
			return utils.NewErrServerError(err)
		} else if exists {
			return utils.ErrExists
		}
	}
	ap := &engine.ActionPlan{
		Id: attrs.Id,
	}
	for _, apiAtm := range attrs.ActionPlan {
		if exists, err := api.ratingDB.HasData(utils.ACTION_PREFIX, apiAtm.ActionsId); err != nil {
			return utils.NewErrServerError(err)
		} else if !exists {
			return fmt.Errorf("%s:%s", utils.ErrBrokenReference.Error(), apiAtm.ActionsId)
		}
		timing := new(engine.RITiming)
		timing.Years.Parse(apiAtm.Years, ";")
		timing.Months.Parse(apiAtm.Months, ";")
		timing.MonthDays.Parse(apiAtm.MonthDays, ";")
		timing.WeekDays.Parse(apiAtm.WeekDays, ";")
		timing.StartTime = apiAtm.Time
		ap.ActionTimings = append(ap.ActionTimings, &engine.ActionTiming{
			Uuid:      utils.GenUUID(),
			Weight:    apiAtm.Weight,
			Timing:    &engine.RateInterval{Timing: timing},
			ActionsID: apiAtm.ActionsId,
		})
	}
	if err := api.ratingDB.SetActionPlan(ap.Id, ap, true); err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.ReloadScheduler {
		if api.Sched == nil {
			return errors.New("SCHEDULER_NOT_ENABLED")
		}
		api.Sched.Reload(true)
	}
	*reply = OK
	return nil
}

func (api *ApiV1) GetActionPlan(attr AttrGetSingle, reply *[]*engine.ActionPlan) error {
	var result []*engine.ActionPlan
	if attr.Id == "" || attr.Id == "*" {
		aplsMap, err := api.ratingDB.GetAllActionPlans()
		if err != nil {
			return err
		}
		for _, apls := range aplsMap {
			result = append(result, apls)
		}
	} else {
		apls, err := api.ratingDB.GetActionPlan(attr.Id, utils.CACHED)
		if err != nil {
			return err
		}
		result = append(result, apls)
	}
	*reply = result
	return nil
}

func (api *ApiV1) GetCacheStats(attrs utils.AttrCacheStats, reply *utils.CacheStats) error {
	cs := new(utils.CacheStats)
	cs.Destinations = cache2go.CountEntries(utils.DESTINATION_PREFIX)
	cs.ReverseDestinations = cache2go.CountEntries(utils.REVERSE_DESTINATION_PREFIX)
	cs.RatingPlans = cache2go.CountEntries(utils.RATING_PLAN_PREFIX)
	cs.RatingProfiles = cache2go.CountEntries(utils.RATING_PROFILE_PREFIX)
	cs.Actions = cache2go.CountEntries(utils.ACTION_PREFIX)
	cs.ActionPlans = cache2go.CountEntries(utils.ACTION_PLAN_PREFIX)
	cs.SharedGroups = cache2go.CountEntries(utils.SHARED_GROUP_PREFIX)
	cs.DerivedChargers = cache2go.CountEntries(utils.DERIVEDCHARGERS_PREFIX)
	cs.LcrProfiles = cache2go.CountEntries(utils.LCR_PREFIX)
	cs.Aliases = cache2go.CountEntries(utils.ALIASES_PREFIX)
	cs.ReverseAliases = cache2go.CountEntries(utils.REVERSE_ALIASES_PREFIX)
	cs.ResourceLimits = cache2go.CountEntries(utils.ResourceLimitsPrefix)
	if api.CdrStatsSrv != nil {
		var queueIds []string
		if err := api.CdrStatsSrv.Call("CDRStatsV1.GetQueueIds", 0, &queueIds); err != nil {
			return utils.NewErrServerError(err)
		}
		cs.CdrStats = len(queueIds)
	}
	if api.Users != nil {
		var ups engine.UserProfiles
		if err := api.Users.Call("UsersV1.GetUsers", &engine.UserProfile{}, &ups); err != nil {
			return utils.NewErrServerError(err)
		}
		cs.Users = len(ups)
	}
	*reply = *cs
	return nil
}


type AttrRemoveRatingProfile struct {
	Direction string
	Tenant    string
	Category  string
	Subject   string
}

func (arrp *AttrRemoveRatingProfile) GetId() (result string) {
	if arrp.Direction != "" && arrp.Direction != utils.ANY {
		result += arrp.Direction
		result += utils.CONCATENATED_KEY_SEP
	} else {
		return
	}
	if arrp.Tenant != "" && arrp.Tenant != utils.ANY {
		result += arrp.Tenant
		result += utils.CONCATENATED_KEY_SEP
	} else {
		return
	}

	if arrp.Category != "" && arrp.Category != utils.ANY {
		result += arrp.Category
		result += utils.CONCATENATED_KEY_SEP
	} else {
		return
	}
	if arrp.Subject != "" && arrp.Subject != utils.ANY {
		result += arrp.Subject
	}
	return
}

func (api *ApiV1) RemoveRatingProfile(attr AttrRemoveRatingProfile, reply *string) error {
	if attr.Direction == "" {
		attr.Direction = utils.OUT
	}
	if (attr.Subject != "" && utils.IsSliceMember([]string{attr.Direction, attr.Tenant, attr.Category}, "")) ||
		(attr.Category != "" && utils.IsSliceMember([]string{attr.Direction, attr.Tenant}, "")) ||
		attr.Tenant != "" && attr.Direction == "" {
		return utils.ErrMandatoryIeMissing
	}
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		err := api.ratingDB.RemoveRatingProfile(attr.GetId())
		if err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, "RemoveRatingProfile")
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (api *ApiV1) GetLoadHistory(attrs utils.Paginator, reply *[]*utils.LoadInstance) error {
	nrItems := -1
	offset := 0
	if attrs.Offset != nil { // For offset we need full data
		offset = *attrs.Offset
	} else if attrs.Limit != nil {
		nrItems = *attrs.Limit
	}
	loadHist, err := api.AccountDb.GetLoadHistory(nrItems, utils.CACHE_SKIP)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attrs.Offset != nil && attrs.Limit != nil { // Limit back to original
		nrItems = *attrs.Limit
	}
	if len(loadHist) == 0 || len(loadHist) <= offset || nrItems == 0 {
		return utils.ErrNotFound
	}
	if offset != 0 {
		nrItems = offset + nrItems
	}
	if nrItems == -1 || nrItems > len(loadHist) { // So we can use it in indexing bellow
		nrItems = len(loadHist)
	}
	*reply = loadHist[offset:nrItems]
	return nil
}
*/
type AttrRemoveActionGroups struct {
	Tenant   string
	GroupIDs []string
}

func (api *ApiV1) RemoveActionGroups(attr AttrRemoveActionGroups, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if len(attr.GroupIDs) == 0 {
		err := utils.ErrNotFound
		*reply = err.Error()
		return err
	}
	// The check could lead to very long execution time. So we decided to leave it at the user's risck.'
	/*
		stringMap := utils.NewStringMap(attr.ActionIDs...)
		keys, err := api.ratingDB.GetKeysForPrefix(utils.ACTION_TRIGGER_PREFIX, true)
		if err != nil {
			*reply = err.Error()
			return err
		}
		for _, key := range keys {
			getAttrs, err := api.ratingDB.GetActionTriggers(key[len(utils.ACTION_TRIGGER_PREFIX):])
			if err != nil {
				*reply = err.Error()
				return err
			}
			for _, atr := range getAttrs {
				if _, found := stringMap[atr.ActionsID]; found {
					// found action trigger referencing action; abort
					err := fmt.Errorf("action %s refenced by action trigger %s", atr.ActionsID, atr.ID)
					*reply = err.Error()
					return err
				}
			}
		}
		allAplsMap, err := api.ratingDB.GetAllActionPlans()
		if err != nil && err != utils.ErrNotFound {
			*reply = err.Error()
			return err
		}
		for _, apl := range allAplsMap {
			for _, atm := range apl.ActionTimings {
				if _, found := stringMap[atm.ActionsID]; found {
					err := fmt.Errorf("action %s refenced by action plan %s", atm.ActionsID, apl.Id)
					*reply = err.Error()
					return err
				}
			}

		}
	*/
	for _, agID := range attr.GroupIDs {
		if err := api.ratingDB.RemoveActionGroup(attr.Tenant, agID); err != nil {
			*reply = err.Error()
			return err
		}
	}
	*reply = utils.OK
	return nil
}

func (api *ApiV1) ReloadScheduler(input string, reply *string) error {
	if api.sched == nil {
		return utils.ErrNotFound
	}
	api.sched.Reload(true)
	*reply = OK
	return nil
}

func (api *ApiV1) ReloadCache(attr utils.AttrReloadCache, reply *string) error {
	for _, tenant := range attr.Tenants {
		cache2go.Flush(tenant)
	}
	*reply = utils.OK
	return nil
}

type AttrEmpty struct{}

func (api *ApiV1) EnsureIndexes(attr AttrEmpty, reply *string) error {
	if err := engine.EnsureIndexes(); err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
