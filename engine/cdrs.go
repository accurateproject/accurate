package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"github.com/jeffail/tunny"
	"go.uber.org/zap"
)

var (
	cdrServer         *CdrServer                                                                    // Share the server so we can use it in http handlers
	debitRequestTypes = []string{utils.META_POSTPAID, utils.META_PREPAID, utils.META_PSEUDOPREPAID} // Prepaid - Cost can be recalculated in case of missing records from SM
	pool              *tunny.WorkPool
)

type CallCostLog struct {
	UniqueID       string
	Source         string
	RunID          string
	Usage          float64 // real usage (not increment rounded)
	CallCost       *CallCost
	CheckDuplicate bool
}

// Handler for generic cgr cdr http
func genericCdrHandler(w http.ResponseWriter, r *http.Request) {
	cdr, err := NewCgrCdrFromHttpReq(r, *cdrServer.cfg.General.DefaultTimezone)
	if err != nil {
		utils.Logger.Error("<CDRS> Could not create CDR entry: ", zap.Error(err))
		return
	}
	if _, err := pool.SendWork(cdr.AsStoredCdr(cdrServer.Timezone())); err != nil {
		utils.Logger.Error("<CDRS> error processing cdr: ", zap.Error(err))
	}
	w.WriteHeader(http.StatusOK)
}

// Handler for fs http
func fsCdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fsCdr, err := NewFSCdr(body, cdrServer.cfg)
	if err != nil {
		utils.Logger.Error("<CDRS> Could not create CDR entry: ", zap.Error(err))
		return
	}
	if _, err := pool.SendWork(fsCdr.AsStoredCdr(cdrServer.Timezone())); err != nil {
		utils.Logger.Error("<CDRS> error processing cdr: ", zap.Error(err))
	}
	w.WriteHeader(http.StatusOK)
}

func NewCdrServer(cfg *config.Config, cdrDB CdrStorage, dataDB AccountingStorage, rater, pubsub, users, aliases, stats rpcclient.RpcClientConnection) (*CdrServer, error) {
	if rater == nil || reflect.ValueOf(rater).IsNil() { // Work around so we store actual nil instead of nil interface value, faster to check here than in CdrServer code
		rater = nil
	}
	if pubsub == nil || reflect.ValueOf(pubsub).IsNil() {
		pubsub = nil
	}
	if users == nil || reflect.ValueOf(users).IsNil() {
		users = nil
	}
	if aliases == nil || reflect.ValueOf(aliases).IsNil() {
		aliases = nil
	}
	if stats == nil || reflect.ValueOf(stats).IsNil() {
		stats = nil
	}

	cdrServer := &CdrServer{cfg: cfg, cdrDB: cdrDB, dataDB: dataDB, rals: rater, pubsub: pubsub, users: users, aliases: aliases, stats: stats, guard: Guardian, sas: GetSimpleAccounts()}
	var err error
	pool, err = tunny.CreatePool(runtime.NumCPU(), func(object interface{}) interface{} {
		cdr := object.(*CDR)

		if err := cdrServer.processCdr(cdr); err != nil {
			utils.Logger.Error("<CDRS> error storing CDR entry: ", zap.Error(err))
		}
		return nil
	}).Open()
	if err != nil {
		utils.Logger.Error("<CDRS> error on worker pool init: ", zap.Error(err))
	}
	//defer pool.Close()

	return cdrServer, nil
}

type CdrServer struct {
	cfg           *config.Config
	cdrDB         CdrStorage
	dataDB        AccountingStorage
	rals          rpcclient.RpcClientConnection
	pubsub        rpcclient.RpcClientConnection
	users         rpcclient.RpcClientConnection
	aliases       rpcclient.RpcClientConnection
	stats         rpcclient.RpcClientConnection
	guard         *GuardianLock
	responseCache *cache2go.ResponseCache
	sas           *SimpleAccounts
	httpPoster    *utils.HTTPPoster // used for replication
}

func (cdrs *CdrServer) Timezone() string {
	return *cdrs.cfg.General.DefaultTimezone
}
func (cdrs *CdrServer) SetTimeToLive(timeToLive time.Duration, out *int) error {
	cdrs.responseCache = cache2go.NewResponseCache(timeToLive)
	return nil
}

func (cdrs *CdrServer) getCache() *cache2go.ResponseCache {
	if cdrs.responseCache == nil {
		cdrs.responseCache = cache2go.NewResponseCache(0)
	}
	return cdrs.responseCache
}

func (cdrs *CdrServer) RegisterHandlersToServer(server *utils.Server) {
	cdrServer = cdrs // Share the server object for handlers
	server.RegisterHTTPFunc("/cdr_http", genericCdrHandler)
	server.RegisterHTTPFunc("/freeswitch_json", fsCdrHandler)
}

// Used to process external CDRs
func (cdrs *CdrServer) ProcessExternalCdr(eCDR *ExternalCDR) error {
	cdr, err := NewCDRFromExternalCDR(eCDR, *cdrs.cfg.General.DefaultTimezone)
	if err != nil {
		return err
	}
	return cdrs.processCdr(cdr)
}

func (cdrs *CdrServer) storeSMCost(smCost *SMCost, checkDuplicate bool) error {
	smCost.CostDetails.UpdateCost()                                                 // make sure the total cost reflect the increments
	smCost.CostDetails.UpdateRatedUsage()                                           // make sure rated usage is updated
	lockKey := utils.CDRS_SOURCE + smCost.UniqueID + smCost.RunID + smCost.OriginID // Will lock on this ID
	if checkDuplicate {
		_, err := cdrs.guard.Guard(func() (interface{}, error) {
			smCosts, err := cdrs.cdrDB.GetSMCosts(smCost.UniqueID, smCost.RunID, "", "")
			if err != nil {
				return nil, err
			}
			if len(smCosts) != 0 {
				return nil, utils.ErrExists
			}
			return nil, cdrs.cdrDB.SetSMCost(smCost)
		}, time.Duration(2*time.Second), lockKey) // FixMe: Possible deadlock with Guard from SMG session close()
		return err
	}
	return cdrs.cdrDB.SetSMCost(smCost)
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (cdrs *CdrServer) processCdr(cdr *CDR) (err error) {
	if cdr.Direction == "" {
		cdr.Direction = utils.OUT
	}
	if cdr.RequestType == "" {
		cdr.RequestType = *cdrs.cfg.General.DefaultRequestType
	}
	if cdr.Tenant == "" {
		cdr.Tenant = *cdrs.cfg.General.DefaultTenant
	}
	if cdr.Category == "" {
		cdr.Category = *cdrs.cfg.General.DefaultCategory
	}
	if cdr.Subject == "" { // Use account information as rating subject if missing
		cdr.Subject = cdr.Account
	}
	if !cdr.Rated { // Enforce the RunID if CDR is not rated
		cdr.RunID = utils.MetaRaw
	}
	if cdr.RunID == utils.MetaRaw {
		cdr.Cost = dec.NewVal(-1, 0)
	}
	if *cdrs.cfg.Cdrs.StoreCdrs { // Store RawCDRs, this we do sync so we can reply with the status
		if cdr.CostDetails != nil {
			cdr.CostDetails.UpdateCost()
			cdr.CostDetails.UpdateRatedUsage()
		}
		if err := cdrs.cdrDB.SetCDR(cdr, false); err != nil {
			utils.Logger.Error("<CDRS> Storing primary ", zap.Any("CDR", cdr), zap.Error(err))
			return err // Error is propagated back and we don't continue processing the CDR if we cannot store it
		}
	}
	// Attach raw CDR to stats
	if cdrs.stats != nil { // Send raw CDR to stats
		var out int
		go cdrs.stats.Call("CDRStatsV1.AppendCDR", cdr, &out)
	}
	if len(cdrs.cfg.Cdrs.CdrReplication) != 0 { // Replicate raw CDR
		go cdrs.replicateCdr(cdr)
	}

	if cdrs.rals != nil && !cdr.Rated { // CDRs not rated will be processed by Rating
		cdrs.deriveRateStoreStatsReplicate(cdr, *cdrs.cfg.Cdrs.StoreCdrs, cdrs.stats != nil, len(cdrs.cfg.Cdrs.CdrReplication) != 0)
	}
	return nil
}

// Returns error if not able to properly store the CDR, mediation is async since we can always recover offline
func (cdrs *CdrServer) deriveRateStoreStatsReplicate(cdr *CDR, store, stats, replicate bool) error {
	cdrRuns, err := cdrs.deriveCdrs(cdr)
	if err != nil {
		utils.Logger.Error("<CDRS> error getting derived chargers for ", zap.Any("CDR", cdr), zap.Error(err))
		return err
	}
	var ratedCDRs []*CDR // Gather all CDRs received from rating subsystem
	for _, cdrRun := range cdrRuns {
		if err := LoadUserProfile(cdrRun, true); err != nil {
			utils.Logger.Error("<CDRS> UserS handling for ", zap.Any("CDR", cdrRun), zap.Error(err))
			continue
		}
		if err := LoadAlias(&AttrAlias{
			Destination: cdrRun.Destination,
			Direction:   cdrRun.Direction,
			Tenant:      cdrRun.Tenant,
			Category:    cdrRun.Category,
			Account:     cdrRun.Account,
			Subject:     cdrRun.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, cdrRun, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
			utils.Logger.Error("<CDRS> Aliasing ", zap.Any("CDR", cdrRun), zap.Error(err))
			continue
		}
		rcvRatedCDRs, err := cdrs.rateCDR(cdrRun)
		if err != nil {
			cdrRun.Cost = dec.NewVal(-1, 0) // If there was an error, mark the CDR
			cdrRun.ExtraInfo = err.Error()
			rcvRatedCDRs = []*CDR{cdrRun}
		}
		ratedCDRs = append(ratedCDRs, rcvRatedCDRs...)
	}
	// Request should be processed by SureTax
	for _, ratedCDR := range ratedCDRs {
		if ratedCDR.RunID == utils.META_SURETAX {
			if err := SureTaxProcessCdr(ratedCDR); err != nil {
				ratedCDR.Cost = dec.NewVal(-1, 0)
				ratedCDR.ExtraInfo = err.Error() // Something failed, write the error in the ExtraInfo
			}
		}
	}
	// Store AccountSummary if requested
	if *cdrs.cfg.Cdrs.AccountSummary {
		for _, ratedCDR := range ratedCDRs {
			if utils.IsSliceMember(debitRequestTypes, ratedCDR.RequestType) {
				acnt, err := cdrs.dataDB.GetAccount(ratedCDR.Tenant, ratedCDR.Account)
				if err != nil {
					utils.Logger.Error("<CDRS> Querying AccountDigest for account: ", zap.String("tenant", ratedCDR.Tenant), zap.String("name", ratedCDR.Account), zap.Error(err))
				} else if acnt.Name != "" {
					ratedCDR.AccountSummary = acnt.AsAccountSummary()
				}
			}
		}
	}
	// Store rated CDRs
	if store {
		for _, ratedCDR := range ratedCDRs {
			if ratedCDR.CostDetails != nil {
				ratedCDR.CostDetails.UpdateCost()
				ratedCDR.CostDetails.UpdateRatedUsage()
			}
			if err := cdrs.cdrDB.SetCDR(ratedCDR, true); err != nil {
				utils.Logger.Error("<CDRS> Storing rated ", zap.Any("CDR", ratedCDR), zap.Error(err))
			}
		}
	}
	// Attach CDR to stats
	if stats { // Send CDR to stats
		for _, ratedCDR := range ratedCDRs {
			var out int
			if err := cdrs.stats.Call("CDRStatsV1.AppendCDR", ratedCDR, &out); err != nil {
				utils.Logger.Error("<CDRS> Could not send CDR to stats: ", zap.Error(err))
			}
		}
	}
	if replicate {
		for _, ratedCDR := range ratedCDRs {
			if err := cdrs.replicateCdr(ratedCDR); err != nil {
				utils.Logger.Error("error replicating CDR: ", zap.Error(err))
			}
		}
	}
	return nil
}

func (cdrs *CdrServer) deriveCdrs(cdr *CDR) ([]*CDR, error) {
	dfltCDRRun := cdr.Clone()
	cdrRuns := []*CDR{dfltCDRRun}
	if cdr.RunID != utils.MetaRaw { // Only derive *raw CDRs
		return cdrRuns, nil
	}
	dfltCDRRun.RunID = utils.META_DEFAULT // Rewrite *raw with *default since we have it as first run
	if err := LoadUserProfile(cdr, true); err != nil {
		utils.Logger.Error("No user for ", zap.Any("CDR", cdr), zap.Error(err))
		return nil, err
	}
	if err := LoadAlias(&AttrAlias{
		Destination: cdr.Destination,
		Direction:   cdr.Direction,
		Tenant:      cdr.Tenant,
		Category:    cdr.Category,
		Account:     cdr.Account,
		Subject:     cdr.Subject,
		Context:     utils.ALIAS_CONTEXT_RATING,
	}, cdr, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return nil, err
	}
	attrsDC := &utils.AttrDerivedChargers{Tenant: cdr.Tenant, Category: cdr.Category, Direction: cdr.Direction,
		Account: cdr.Account, Subject: cdr.Subject, Destination: cdr.Destination}
	var dcs utils.DerivedChargerGroup
	if err := cdrs.rals.Call("Responder.GetDerivedChargers", attrsDC, &dcs); err != nil {
		utils.Logger.Error("Could not get derived charging ", zap.String("uniqueid", cdr.UniqueID), zap.Error(err))
		return nil, err
	}
	for _, dc := range dcs.Chargers {
		if check, err := dc.CheckRunFilter(cdr); err != nil {
			return nil, err
		} else {
			if !check {
				continue
			}
		}

		forkedCdr := cdr.Clone()
		if err := dc.ChangeFields(forkedCdr); err != nil {
			utils.Logger.Error("Could not fork CGR with ", zap.String("uniqueid", cdr.UniqueID), zap.String("runid", dc.RunID), zap.Error(err))
			continue // do not add it to the forked CDR list
		}
		if !forkedCdr.Rated {
			forkedCdr.Cost = dec.NewVal(-1, 0) // Make sure that un-rated CDRs start with Cost -1
		}
		cdrRuns = append(cdrRuns, forkedCdr)
	}
	return cdrRuns, nil
}

// rateCDR will populate cost field
// Returns more than one rated CDR in case of SMCost retrieved based on prefix
func (cdrs *CdrServer) rateCDR(cdr *CDR) ([]*CDR, error) {
	var qryCC *CallCost
	var err error
	if cdr.RequestType == utils.META_NONE {
		return nil, nil
	}
	cdr.ExtraInfo = "" // for re-rating
	var cdrsRated []*CDR
	_, hasLastUsed := cdr.ExtraFields[utils.LastUsed]
	if utils.META_PREPAID == cdr.RequestType && (cdr.Usage != 0 || hasLastUsed) {
		// Should be previously calculated and stored in DB
		delay := utils.Fib()
		var smCosts []*SMCost
		for i := 0; i < *cdrs.cfg.Cdrs.SmCostRetries; i++ {
			smCosts, err = cdrs.cdrDB.GetSMCosts(cdr.UniqueID, cdr.RunID, cdr.OriginHost, cdr.ExtraFields[utils.OriginIDPrefix])
			if err == nil && len(smCosts) != 0 {
				break
			}
			if i != 3 {
				time.Sleep(delay())
			}
		}
		if len(smCosts) != 0 { // Cost retrieved from SMCost table
			for _, smCost := range smCosts {
				cdrClone := cdr.Clone()
				cdrClone.OriginID = smCost.OriginID
				if cdr.Usage == 0 {
					cdrClone.Usage = time.Duration(smCost.Usage * utils.NANO_MULTIPLIER) // Usage is float as seconds, convert back to duration
				}
				cdrClone.Cost = smCost.CostDetails.Cost
				cdrClone.CostDetails = smCost.CostDetails
				cdrClone.CostSource = smCost.CostSource
				cdrsRated = append(cdrsRated, cdrClone)
			}
			return cdrsRated, nil
		} else { //calculate CDR as for pseudoprepaid
			utils.Logger.Warn("<Cdrs> WARNING: Could not find CallCostLog will recalculate", zap.String("uniqueid", cdr.UniqueID), zap.String("source", utils.SESSION_MANAGER_SOURCE), zap.String("runid", cdr.RunID))
			qryCC, err = cdrs.getCostFromRater(cdr)
		}
	} else {
		qryCC, err = cdrs.getCostFromRater(cdr)
	}
	if err != nil {
		return nil, err
	} else if qryCC != nil {
		cdr.Cost = qryCC.Cost
		cdr.CostDetails = qryCC
	}
	return []*CDR{cdr}, nil
}

// Retrive the cost from engine
func (cdrs *CdrServer) getCostFromRater(cdr *CDR) (*CallCost, error) {
	cc := new(CallCost)
	var err error
	timeStart := cdr.AnswerTime
	if timeStart.IsZero() { // Fix for FreeSWITCH unanswered calls
		timeStart = cdr.SetupTime
	}
	cd := &CallDescriptor{
		TOR:             cdr.ToR,
		Direction:       cdr.Direction,
		Tenant:          cdr.Tenant,
		Category:        cdr.Category,
		Subject:         cdr.Subject,
		Account:         cdr.Account,
		Destination:     cdr.Destination,
		TimeStart:       timeStart,
		TimeEnd:         timeStart.Add(cdr.Usage),
		DurationIndex:   cdr.Usage,
		PerformRounding: true,
	}

	if utils.IsSliceMember(debitRequestTypes, cdr.RequestType) {
		err = cdrs.rals.Call("Responder.Debit", cd, cc)
	} else {
		err = cdrs.rals.Call("Responder.GetCost", cd, cc)
		if err == nil && cdrs.sas != nil {
			err = cdrs.sas.Debit(cd.Tenant, cd.getAccountName(), cd.Category, cc.GetCost())
		}
	}
	if err != nil {
		return cc, err
	}
	cdr.CostSource = utils.CDRS_SOURCE
	return cc, nil
}

// ToDo: Add websocket support
func (cdrs *CdrServer) replicateCdr(cdr *CDR) error {
	for _, rplCfg := range cdrs.cfg.Cdrs.CdrReplication {
		passesFilters := true
		for _, cdfFltr := range rplCfg.CdrFilter {
			if !cdfFltr.FilterPasses(cdr.FieldAsString(cdfFltr)) {
				passesFilters = false
				break
			}
		}
		if !passesFilters { // Not passes filters, ignore this replication
			continue
		}
		var body interface{}
		var content = ""
		switch rplCfg.Transport {
		case utils.MetaHTTPjsonCDR:
			content = utils.CONTENT_JSON
			jsn, err := json.Marshal(cdr)
			if err != nil {
				return err
			}
			body = jsn
		case utils.MetaHTTPjsonMap:
			content = utils.CONTENT_JSON
			expMp, err := cdr.AsExportMap(rplCfg.ContentFields, *cdrs.cfg.General.HttpSkipTlsVerify, nil)
			if err != nil {
				return err
			}
			jsn, err := json.Marshal(expMp)
			if err != nil {
				return err
			}
			body = jsn
		case utils.META_HTTP_POST:
			content = utils.CONTENT_FORM
			expMp, err := cdr.AsExportMap(rplCfg.ContentFields, *cdrs.cfg.General.HttpSkipTlsVerify, nil)
			if err != nil {
				return err
			}
			vals := url.Values{}
			for fld, val := range expMp {
				vals.Set(fld, val)
			}
			body = vals
		}
		var errChan chan error
		if rplCfg.Synchronous {
			errChan = make(chan error)
		}
		go func(body interface{}, rplCfg *config.CdrReplication, content string, errChan chan error) {
			fallbackPath := path.Join(
				*cdrs.cfg.General.HttpFailedDir,
				rplCfg.FallbackFileName())
			if _, err := cdrs.httpPoster.Post(rplCfg.Address, content, body, rplCfg.Attempts, fallbackPath); err != nil {
				utils.Logger.Error("<CDRReplicator> Replicating ", zap.Any("CDR", cdr), zap.Error(err))
				if rplCfg.Synchronous {
					errChan <- err
				}
			}
			if rplCfg.Synchronous {
				errChan <- nil
			}
		}(body, rplCfg, content, errChan)
		if rplCfg.Synchronous { // Synchronize here
			<-errChan
		}
	}
	return nil
}

// Called by rate/re-rate API, FixMe: deprecate it once new APIer structure is operational
func (cdrs *CdrServer) RateCDRs(cdrFltr *utils.CDRsFilter, sendToStats bool) error {
	cdrList, _, err := cdrs.cdrDB.GetCDRs(cdrFltr, false)
	if err != nil {
		return err
	}
	for _, cdr := range cdrList {
		if err := cdrs.deriveRateStoreStatsReplicate(cdr, *cdrs.cfg.Cdrs.StoreCdrs, sendToStats, len(cdrs.cfg.Cdrs.CdrReplication) != 0); err != nil {
			utils.Logger.Error("<CDRS> Processing ", zap.Any("CDR", cdr), zap.Error(err))
		}
	}
	return nil
}

// Internally used and called from CDRSv1
// Cached requests for HA setups
func (cdrs *CdrServer) V1ProcessCDR(cdr *CDR, reply *string) error {
	if len(cdr.UniqueID) == 0 { // Populate UniqueID if not present
		cdr.ComputeUniqueID()
	}
	cacheKey := "V1ProcessCDR" + cdr.UniqueID + cdr.RunID
	if item, err := cdrs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = item.Value.(string)
		}
		return item.Err
	}
	if err := cdrs.processCdr(cdr); err != nil {
		cdrs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return utils.NewErrServerError(err)
	}
	cdrs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: utils.OK})
	*reply = utils.OK
	return nil
}

// RPC method, differs from storeSMCost through it's signature
func (cdrs *CdrServer) V1StoreSMCost(attr AttrCDRSStoreSMCost, reply *string) error {
	if attr.Cost.UniqueID == "" {
		return utils.NewCGRError(utils.CDRSCtx,
			utils.MandatoryIEMissingCaps, fmt.Sprintf("%s: UniqueID", utils.MandatoryInfoMissing),
			"SMCost: %+v with empty UniqueID")
	}
	cacheKey := "V1StoreSMCost" + attr.Cost.UniqueID + attr.Cost.RunID + attr.Cost.OriginID
	if item, err := cdrs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = item.Value.(string)
			utils.Logger.Error("<CDRS> Serving cached error for", zap.String("key", cacheKey), zap.String("reply", *reply))
		}
		return item.Err
	}
	if err := cdrs.storeSMCost(attr.Cost, attr.CheckDuplicate); err != nil {
		cdrs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return utils.NewErrServerError(err)
	}
	cdrs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: utils.OK})
	*reply = utils.OK
	return nil
}

// Called by rate/re-rate API, RPC method
func (cdrs *CdrServer) V1RateCDRs(attrs utils.AttrRateCDRs, reply *string) error {
	cdrFltr, err := attrs.RPCCDRsFilter.AsCDRsFilter(*cdrs.cfg.General.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrList, _, err := cdrs.cdrDB.GetCDRs(cdrFltr, false)
	if err != nil {
		return err
	}
	storeCDRs := *cdrs.cfg.Cdrs.StoreCdrs
	if attrs.StoreCDRs != nil {
		storeCDRs = *attrs.StoreCDRs
	}
	sendToStats := cdrs.stats != nil
	if attrs.SendToStatS != nil {
		sendToStats = *attrs.SendToStatS
	}
	replicate := len(cdrs.cfg.Cdrs.CdrReplication) != 0
	if attrs.ReplicateCDRs != nil {
		replicate = *attrs.ReplicateCDRs
	}
	for _, cdr := range cdrList {
		if err := cdrs.deriveRateStoreStatsReplicate(cdr, storeCDRs, sendToStats, replicate); err != nil {
			utils.Logger.Error("<CDRS> Processing ", zap.Any("CDR", cdr), zap.Error(err))
		}
	}
	return nil
}

func (cdrsrv *CdrServer) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// get method
	method := reflect.ValueOf(cdrsrv).MethodByName(parts[0][len(parts[0])-2:] + parts[1]) // Inherit the version in the method
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
