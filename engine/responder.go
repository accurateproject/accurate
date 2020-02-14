package engine

import (
	"errors"
	"fmt"
	"net/rpc"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/accurateproject/accurate/balancer2go"
	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"go.uber.org/zap"
)

// Individual session run
type SessionRun struct {
	DerivedCharger *utils.DerivedCharger // Needed in reply
	CallDescriptor *CallDescriptor
	CallCosts      []*CallCost
}

type AttrGetLcr struct {
	*CallDescriptor
	*LCRFilter
	*utils.Paginator
}

type Responder struct {
	Bal           *balancer2go.Balancer
	ExitChan      chan bool
	Stats         rpcclient.RpcClientConnection
	Timeout       time.Duration
	Timezone      string
	cnt           int64
	responseCache *cache2go.ResponseCache
}

func (rs *Responder) SetTimeToLive(timeToLive time.Duration, out *int) error {
	rs.responseCache = cache2go.NewResponseCache(timeToLive)
	return nil
}

func (rs *Responder) getCache() *cache2go.ResponseCache {
	if rs.responseCache == nil {
		rs.responseCache = cache2go.NewResponseCache(0)
	}
	return rs.responseCache
}

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (rs *Responder) GetCost(arg *CallDescriptor, reply *CallCost) (err error) {
	rs.cnt += 1
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, false); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	if rs.Bal != nil {
		r, e := rs.getCallCost(arg, "Responder.GetCost")
		*reply, err = *r, e
	} else {
		r, e := Guardian.Guard(func() (interface{}, error) {
			return arg.GetCost()
		}, 0, utils.ConcatKey(arg.Tenant, arg.getAccountName()))
		if r != nil {
			*reply = *r.(*CallCost)
		}
		if e != nil {
			return e
		}
	}
	return
}

func (rs *Responder) Debit(arg *CallDescriptor, reply *CallCost) (err error) {
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, false); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}

	if rs.Bal != nil {
		r, e := rs.getCallCost(arg, "Responder.Debit")
		*reply, err = *r, e
	} else {
		r, e := arg.Debit()
		if e != nil {
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	return
}

func (rs *Responder) MaxDebit(arg *CallDescriptor, reply *CallCost) (err error) {
	cacheKey := utils.MAX_DEBIT_CACHE_PREFIX + arg.UniqueID + arg.RunID + arg.DurationIndex.String()
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*CallCost))
		}
		return item.Err
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, false); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}

	if rs.Bal != nil {
		r, e := rs.getCallCost(arg, "Responder.MaxDebit")
		*reply, err = *r, e
	} else {
		r, e := arg.MaxDebit()
		if e != nil {
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{
				Err: e,
			})
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{
		Value: reply,
		Err:   err,
	})
	return
}

func (rs *Responder) RefundIncrements(arg *CallDescriptor, reply *float64) (err error) {
	cacheKey := utils.REFUND_INCR_CACHE_PREFIX + arg.UniqueID + arg.RunID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*float64))
		}
		return item.Err
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, false); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{
			Err: err,
		})
		return err
	}

	if rs.Bal != nil {
		*reply, err = rs.callMethod(arg, "Responder.RefundIncrements")
	} else {
		err = arg.RefundIncrements()
	}
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{
		Value: reply,
		Err:   err,
	})
	return
}

func (rs *Responder) RefundRounding(arg *CallDescriptor, reply *float64) (err error) {
	cacheKey := utils.REFUND_ROUND_CACHE_PREFIX + arg.UniqueID + arg.RunID + arg.DurationIndex.String()
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*float64))
		}
		return item.Err
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, false); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{
			Err: err,
		})
		return err
	}

	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{
		Value: reply,
		Err:   err,
	})
	return
}

func (rs *Responder) GetMaxSessionTime(arg *CallDescriptor, reply *float64) (err error) {
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(arg, false); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrAlias{
			Destination: arg.Destination,
			Direction:   arg.Direction,
			Tenant:      arg.Tenant,
			Category:    arg.Category,
			Account:     arg.Account,
			Subject:     arg.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, arg, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}

	if rs.Bal != nil {
		*reply, err = rs.callMethod(arg, "Responder.GetMaxSessionTime")
	} else {
		r, e := arg.GetMaxSessionDuration()
		*reply, err = float64(r), e
	}
	return
}

// Returns MaxSessionTime for an event received in SessionManager, considering DerivedCharging for it
func (rs *Responder) GetDerivedMaxSessionTime(ev *CDR, reply *float64) error {
	if rs.Bal != nil {
		return errors.New("unsupported method on the balancer")
	}
	cacheKey := utils.GET_DERIV_MAX_SESS_TIME + ev.UniqueID + ev.RunID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*float64))
		}
		return item.Err
	}
	if ev.Subject == "" {
		ev.Subject = ev.Account
	}

	// replace user profile fields
	if err := LoadUserProfile(ev, false); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}

	// replace aliases
	if err := LoadAlias(
		&AttrAlias{
			Destination: ev.Destination,
			Direction:   ev.Direction,
			Tenant:      ev.Tenant,
			Category:    ev.Category,
			Account:     ev.Account,
			Subject:     ev.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, ev, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}

	maxCallDuration := -1.0
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.Tenant, Category: ev.Category, Direction: ev.Direction, Account: ev.Account, Subject: ev.Subject}
	dcs := &utils.DerivedChargerGroup{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	dcs, _ = dcs.AppendDefaultRun()

	for _, dc := range dcs.Chargers {
		if pass, err := dc.CheckRunFilter(ev); err != nil || !pass {
			if err != nil {
				rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
				return err
			}
			continue
		}

		forkedEv := ev.Clone()
		if err := dc.ChangeFields(forkedEv); err != nil {
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}

		if utils.IsSliceMember([]string{utils.META_RATED, utils.RATED}, forkedEv.RequestType) { // Only consider prepaid and pseudoprepaid for MaxSessionTime
			continue
		}
		loc, err := time.LoadLocation(rs.Timezone)
		if err != nil {
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}
		startTime := forkedEv.SetupTime.In(loc)
		usage := ev.Usage
		if usage == 0 {
			usage = config.Get().SmGeneric.MaxCallDuration.D()
		}
		cd := &CallDescriptor{
			UniqueID:    forkedEv.UniqueID,
			RunID:       dc.RunID,
			TOR:         forkedEv.ToR,
			Direction:   forkedEv.Direction,
			Tenant:      forkedEv.Tenant,
			Category:    forkedEv.Category,
			Subject:     forkedEv.Subject,
			Account:     forkedEv.Account,
			Destination: forkedEv.Destination,
			TimeStart:   startTime,
			TimeEnd:     startTime.Add(usage),
			ExtraFields: forkedEv.ExtraFields,
		}
		var remainingDuration float64
		err = rs.GetMaxSessionTime(cd, &remainingDuration)
		if err != nil {
			*reply = 0
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}
		if forkedEv.RequestType == utils.META_POSTPAID {
			// Only consider prepaid and pseudoprepaid for MaxSessionTime, do it here for unauthorized destination error check
			continue
		}
		// Set maxCallDuration, smallest out of all forked sessions
		if maxCallDuration == -1.0 { // first time we set it /not initialized yet
			maxCallDuration = remainingDuration
		} else if maxCallDuration > remainingDuration {
			maxCallDuration = remainingDuration
		}
	}
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: maxCallDuration})
	*reply = maxCallDuration
	return nil
}

// Used by SM to get all the prepaid CallDescriptors attached to a session
func (rs *Responder) GetSessionRuns(ev *CDR, sRuns *[]*SessionRun) error {
	if rs.Bal != nil {
		return errors.New("Unsupported method on the balancer")
	}
	cacheKey := utils.GET_SESS_RUNS_CACHE_PREFIX + ev.UniqueID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*sRuns = *(item.Value.(*[]*SessionRun))
		}
		return item.Err
	}
	if ev.Subject == "" {
		ev.Subject = ev.Account
	}
	//utils.Logger.Info(fmt.Sprintf("DC before: %+v", ev))
	// replace user profile fields
	if err := LoadUserProfile(ev, false); err != nil {
		return err
	}
	// replace aliases
	if err := LoadAlias(
		&AttrAlias{
			Destination: ev.Destination,
			Direction:   ev.Direction,
			Tenant:      ev.Tenant,
			Category:    ev.Category,
			Account:     ev.Account,
			Subject:     ev.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, ev, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		return err
	}
	//utils.Logger.Info(fmt.Sprintf("DC after: %+v", ev))
	attrsDC := &utils.AttrDerivedChargers{Tenant: ev.Tenant, Category: ev.Category, Direction: ev.Direction, Account: ev.Account, Subject: ev.Subject}
	//utils.Logger.Info(fmt.Sprintf("Derived chargers for: %+v", attrsDC))
	dcs := &utils.DerivedChargerGroup{}
	if err := rs.GetDerivedChargers(attrsDC, dcs); err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{
			Err: err,
		})
		return err
	}
	dcs, _ = dcs.AppendDefaultRun()
	//utils.Logger.Info(fmt.Sprintf("DCS: %v", len(dcs.Chargers)))
	sesRuns := make([]*SessionRun, 0)
	for _, dc := range dcs.Chargers {
		forkedEv := ev.Clone()
		if err := dc.ChangeFields(forkedEv); err != nil {
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}
		if forkedEv.RequestType != utils.META_PREPAID {
			continue // We only consider prepaid sessions
		}
		loc, err := time.LoadLocation(rs.Timezone)
		if err != nil {
			rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
			return err
		}
		startTime := forkedEv.AnswerTime.In(loc)
		if startTime.IsZero() { // AnswerTime not parsable, try SetupTime
			startTime = forkedEv.SetupTime.In(loc)
			if startTime.IsZero() {
				err := errors.New("Error parsing answer event start time")
				rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
				return err
			}
		}
		extraFields := ev.GetExtraFields()
		cd := &CallDescriptor{
			UniqueID:    forkedEv.UniqueID,
			RunID:       dc.RunID,
			TOR:         forkedEv.ToR,
			Direction:   forkedEv.Direction,
			Tenant:      forkedEv.Tenant,
			Category:    forkedEv.Category,
			Subject:     forkedEv.Subject,
			Account:     forkedEv.Account,
			Destination: forkedEv.Destination,
			TimeStart:   startTime,
			TimeEnd:     forkedEv.AnswerTime.Add(forkedEv.Usage),
			ExtraFields: extraFields}
		if flagsStr, hasFlags := extraFields[utils.CGRFlags]; hasFlags { // Force duration from extra fields
			flags := utils.StringMapFromSlice(strings.Split(flagsStr, utils.INFIELD_SEP))
			if _, hasFD := flags[utils.FlagForceDuration]; hasFD {
				cd.ForceDuration = true
			}
		}
		sesRuns = append(sesRuns, &SessionRun{DerivedCharger: dc, CallDescriptor: cd})
	}
	//utils.Logger.Info(fmt.Sprintf("RUNS: %v", len(sesRuns)))
	*sRuns = sesRuns
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: sRuns})
	return nil
}

func (rs *Responder) GetDerivedChargers(attrs *utils.AttrDerivedChargers, dcs *utils.DerivedChargerGroup) error {
	if rs.Bal != nil {
		return errors.New("BALANCER_UNSUPPORTED_METHOD")
	}
	if dcsH, err := HandleGetDerivedChargers(ratingStorage, attrs); err != nil {
		return err
	} else if dcsH != nil {
		*dcs = *dcsH
	}
	return nil
}

func (rs *Responder) GetLCR(attrs *AttrGetLcr, reply *LCRCost) error {
	cacheKey := utils.LCRCachePrefix + attrs.UniqueID + attrs.RunID
	if item, err := rs.getCache().Get(cacheKey); err == nil && item != nil {
		if item.Value != nil {
			*reply = *(item.Value.(*LCRCost))
		}
		return item.Err
	}
	if attrs.CallDescriptor.Subject == "" {
		attrs.CallDescriptor.Subject = attrs.CallDescriptor.Account
	}
	// replace user profile fields
	if err := LoadUserProfile(attrs.CallDescriptor, false); err != nil {
		return err
	}
	// replace aliases
	cd := attrs.CallDescriptor
	if err := LoadAlias(
		&AttrAlias{
			Destination: cd.Destination,
			Direction:   cd.Direction,
			Tenant:      cd.Tenant,
			Category:    cd.Category,
			Account:     cd.Account,
			Subject:     cd.Subject,
			Context:     utils.ALIAS_CONTEXT_RATING,
		}, cd, utils.EXTRA_FIELDS); err != nil && err != utils.ErrNotFound {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	lcrCost, err := attrs.CallDescriptor.GetLCR(rs.Stats, attrs.LCRFilter, attrs.Paginator)
	if err != nil {
		rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Err: err})
		return err
	}
	if lcrCost.Entry != nil && lcrCost.Entry.Strategy == LCR_STRATEGY_LOAD {
		for _, suppl := range lcrCost.SupplierCosts {
			suppl.Cost = dec.NewVal(-1, 0) // In case of load distribution we don't calculate costs
		}
	}
	rs.getCache().Cache(cacheKey, &cache2go.CacheItem{Value: lcrCost})
	*reply = *lcrCost
	return nil
}

func (rs *Responder) FlushCache(arg *CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(arg, "Responder.FlushCache")
	} else {
		r, e := Guardian.Guard(func() (interface{}, error) {
			return 0, arg.FlushCache(arg.Tenant)
		}, 0, utils.ConcatKey(arg.Tenant, arg.getAccountName()))
		*reply, err = r.(float64), e
	}
	return
}

type AttrStatus struct {
	ShowBytes bool   // show mem in bytes
	Delay     string // delay in answer, used in some automated tests
}

func (rs *Responder) Status(attr AttrStatus, reply *map[string]interface{}) (err error) {
	if attr.Delay != "" {
		if delay, err := utils.ParseDurationWithSecs(attr.Delay); err == nil {
			time.Sleep(delay)
		}
	}
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]interface{})
	response[utils.InstanceID] = config.Get().General.InstanceID
	if rs.Bal != nil {
		response["Raters"] = rs.Bal.GetClientAddresses()
	}
	response["memstat"] = memstats.HeapAlloc
	response["footprint"] = memstats.Sys
	if !attr.ShowBytes {
		response["memstat"] = utils.SizeFmt(float64(response["memstat"].(uint64)), "")
		response["footprint"] = utils.SizeFmt(float64(response["footprint"].(uint64)), "")
	}
	response["goroutines"] = runtime.NumGoroutine()
	*reply = response
	return
}

func (rs *Responder) Shutdown(arg string, reply *string) (err error) {
	if rs.Bal != nil {
		rs.Bal.Shutdown("Responder.Shutdown")
	}
	ratingStorage.Close()
	accountingStorage.Close()
	cdrStorage.Close()
	defer func() { rs.ExitChan <- true }()
	*reply = "Done!"
	return
}

/*
The function that gets the information from the raters using balancer.
*/
func (rs *Responder) getCallCost(key *CallDescriptor, method string) (reply *CallCost, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := rs.Bal.Balance()
		if client == nil {
			utils.Logger.Info("<Balancer> Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			_, err = Guardian.Guard(func() (interface{}, error) {
				err = client.Call(method, *key, reply)
				return reply, err
			}, 0, utils.ConcatKey(key.Tenant, key.getAccountName()))
			if err != nil {
				utils.Logger.Error("<Balancer> Got en error from rater", zap.Error(err))
			}
		}
	}
	return
}

/*
The function that gets the information from the raters using balancer.
*/
func (rs *Responder) callMethod(key *CallDescriptor, method string) (reply float64, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := rs.Bal.Balance()
		if client == nil {
			utils.Logger.Info("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			_, err = Guardian.Guard(func() (interface{}, error) {
				err = client.Call(method, *key, &reply)
				return reply, err
			}, 0, utils.ConcatKey(key.Tenant, key.getAccountName()))
			if err != nil {
				utils.Logger.Error("error from rater", zap.Error(err))
			}
		}
	}
	return
}

/*
RPC method that receives a rater address, connects to it and ads the pair to the rater list for balancing
*/
func (rs *Responder) RegisterRater(clientAddress string, replay *int) error {
	utils.Logger.Info(fmt.Sprintf("Started rater %v registration...", clientAddress))
	time.Sleep(2 * time.Second) // wait a second for Rater to start serving
	client, err := rpc.Dial("tcp", clientAddress)
	if err != nil {
		utils.Logger.Error("Could not connect to client!")
		return err
	}
	rs.Bal.AddClient(clientAddress, client)
	utils.Logger.Info("rater registered succesfully.", zap.String("address", clientAddress))
	return nil
}

/*
RPC method that recives a rater addres gets the connections and closes it and removes the pair from rater list.
*/
func (rs *Responder) UnRegisterRater(clientAddress string, replay *int) error {
	client, ok := rs.Bal.GetClient(clientAddress)
	if ok {
		client.Close()
		rs.Bal.RemoveClient(clientAddress)
		utils.Logger.Info("rater %v unregistered succesfully", zap.String("address", clientAddress))
	} else {
		utils.Logger.Info("rerver %v was not on my watch", zap.String("address", clientAddress))
	}
	return nil
}

func (rs *Responder) GetTimeout(i int, d *time.Duration) error {
	*d = rs.Timeout
	return nil
}

func (rs *Responder) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(rs).MethodByName(parts[1])
	if !method.IsValid() {
		return utils.ErrNotImplemented
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
