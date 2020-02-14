package engine

import (
	"fmt"
	"strings"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	ColTmg = "timings"
	ColDst = "destinations"
	ColRts = "rates"
	ColDrt = "destination_rates"
	ColAct = "actions"
	ColApl = "action_plans"
	ColTsk = "tasks"
	ColApb = "action_plan_bindings"
	ColAtr = "action_triggers"
	ColRpl = "rating_plans"
	ColRpf = "rating_profiles"
	ColAcc = "accounts"
	ColShg = "shared_groups"
	ColLcr = "lcr_rules"
	ColDcs = "derived_chargers"
	ColAls = "aliases"
	ColStq = "stat_queues"
	ColQcr = "stat_qcdrs"
	ColPbs = "pubsub"
	ColUsr = "users"
	ColCrs = "cdr_stats"
	ColLht = "load_history"
	ColVer = "versions"
	ColRL  = "resource_limits"
	ColCdr = "cdrs"
	ColSmc = "sm_costs"
	ColSac = "simple_accounts"
)

var (
	UniqueIDLow        = strings.ToLower(utils.UniqueID)
	RunIDLow           = strings.ToLower(utils.MEDI_RUNID)
	OrderIDLow         = strings.ToLower(utils.ORDERID)
	OriginHostLow      = strings.ToLower(utils.CDRHOST)
	OriginIDLow        = strings.ToLower(utils.ACCID)
	ToRLow             = strings.ToLower(utils.TOR)
	CDRHostLow         = strings.ToLower(utils.CDRHOST)
	CDRSourceLow       = strings.ToLower(utils.CDRSOURCE)
	RequestTypeLow     = strings.ToLower(utils.REQTYPE)
	DirectionLow       = strings.ToLower(utils.DIRECTION)
	TenantLow          = strings.ToLower(utils.TENANT)
	CategoryLow        = strings.ToLower(utils.CATEGORY)
	AccountLow         = strings.ToLower(utils.ACCOUNT)
	SubjectLow         = strings.ToLower(utils.SUBJECT)
	SupplierLow        = strings.ToLower(utils.SUPPLIER)
	DisconnectCauseLow = strings.ToLower(utils.DISCONNECT_CAUSE)
	SetupTimeLow       = strings.ToLower(utils.SETUP_TIME)
	AnswerTimeLow      = strings.ToLower(utils.ANSWER_TIME)
	CreatedAtLow       = strings.ToLower(utils.CreatedAt)
	UpdatedAtLow       = strings.ToLower(utils.UpdatedAt)
	UsageLow           = strings.ToLower(utils.USAGE)
	PDDLow             = strings.ToLower(utils.PDD)
	CostDetailsLow     = strings.ToLower(utils.COST_DETAILS)
	DestinationLow     = strings.ToLower(utils.DESTINATION)
	CostLow            = strings.ToLower(utils.COST)

	indexes = map[string]map[string][]mgo.Index{
		utils.TariffPlanDB: map[string][]mgo.Index{
			ColDst: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "code", "name"}, Unique: true},
				mgo.Index{Key: []string{"tenant", "code"}, Unique: false},
				mgo.Index{Key: []string{"tenant", "name"}, Unique: false},
			},
			ColTmg: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColApb: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "account", "action_plan"}, Unique: true, DropDups: true},
				mgo.Index{Key: []string{"tenant", "account"}, Unique: false},
				mgo.Index{Key: []string{"tenant", "action_plan"}, Unique: false},
			},
			ColApl: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColAtr: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColRpl: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColShg: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColAct: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColCrs: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
				mgo.Index{Key: []string{"tenant"}},
			},
			ColRpf: []mgo.Index{
				mgo.Index{Key: []string{"direction", "tenant", "category", "subject"}, Unique: true},
				mgo.Index{Key: []string{"direction", "tenant", "category"}, Unique: false}, // for lcr
			},
			ColDcs: []mgo.Index{
				mgo.Index{Key: []string{"direction", "tenant", "category", "account", "subject"}, Unique: true},
			},
			ColLcr: []mgo.Index{
				mgo.Index{Key: []string{"direction", "tenant", "category", "account", "subject"}, Unique: true},
			},
		},
		utils.DataDB: map[string][]mgo.Index{
			ColAcc: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColSac: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColStq: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColQcr: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: false},
				mgo.Index{Key: []string{"tenant", "name", "event_time"}, Unique: false},
			},
			ColUsr: []mgo.Index{
				mgo.Index{Key: []string{"tenant", "name"}, Unique: true},
			},
			ColAls: []mgo.Index{
				mgo.Index{Key: []string{"direction", "tenant", "category", "account", "subject", "context"}, Unique: true},
				mgo.Index{Key: []string{"tenant", "context", "index.target", "index.alias"}, Unique: false},
			},
			//colRls = "reverse_aliases"
			//ColPbs = "pubsub"
		},
		utils.CdrDB: map[string][]mgo.Index{
			ColCdr: []mgo.Index{
				mgo.Index{Key: []string{UniqueIDLow, RunIDLow}, Unique: true},
				mgo.Index{Key: []string{UniqueIDLow, RunIDLow, OriginIDLow}, Unique: true},
			},
			ColSmc: []mgo.Index{
				mgo.Index{Key: []string{UniqueIDLow, RunIDLow}, Unique: true},
				mgo.Index{Key: []string{OriginHostLow, OriginIDLow}, Unique: true},
			},
		},
	}
)

func NewMongoStorage(host, port, db, user, pass, storageType string, cdrsIndexes []string, cacheCfg *config.Cache, loadHistorySize int) (ms *MongoStorage, err error) {
	address := fmt.Sprintf("%s:%s", host, port)
	if user != "" && pass != "" {
		address = fmt.Sprintf("%s:%s@%s", user, pass, address)
	}
	session, err := mgo.Dial(address)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Strong, true)
	ms = &MongoStorage{db: db, session: session, cacheCfg: cacheCfg, loadHistorySize: loadHistorySize, cdrsIndexes: cdrsIndexes, storageType: storageType}
	if cNames, err := session.DB(ms.db).CollectionNames(); err != nil {
		return nil, err
	} else if len(cNames) == 0 { // create indexes only if database is empty
		if err = ms.EnsureIndexes(); err != nil {
			return nil, err
		}
	}
	return
}

type MongoStorage struct {
	session         *mgo.Session
	db              string
	cacheCfg        *config.Cache
	loadHistorySize int
	cdrsIndexes     []string
	storageType     string
}

func (ms *MongoStorage) conn(col string) (*mgo.Session, *mgo.Collection) {
	sessionCopy := ms.session.Copy()
	return sessionCopy, sessionCopy.DB(ms.db).C(col)
}

func (ms *MongoStorage) RemoveTenant(tenant string, collections ...string) error {
	tpCollections := []string{ColTmg, ColDst, ColRts, ColDrt, ColAct, ColApl, ColTsk, ColApb, ColAtr, ColRpl, ColRpf, ColShg, ColLcr, ColDcs, ColCrs}
	dataCollections := []string{ColAcc, ColSac, ColAls, ColStq, ColQcr, ColPbs, ColUsr, ColRL}
	cdrCollections := []string{ColCdr, ColSmc}
	var colls []string
	for _, col := range collections {
		if col == utils.TariffPlanDB {
			colls = append(colls, tpCollections...)
			continue
		}
		if col == utils.DataDB {
			colls = append(colls, dataCollections...)
			continue
		}
		if col == utils.CdrDB {
			colls = append(colls, cdrCollections...)
			continue
		}
		colls = append(colls, col)
	}
	for _, collName := range colls {
		session, col := ms.conn(collName)
		defer session.Close()
		if _, err := col.RemoveAll(bson.M{"tenant": tenant}); err != nil && err != mgo.ErrNotFound {
			return err
		}
	}
	return nil
}

func filter(params bson.M) bson.M {
	filteredParams := bson.M{}
	for k, v := range params {
		sv, ok := v.(string)
		if !ok || sv != "" {
			filteredParams[k] = v
		}
	}
	return filteredParams
}

func (ms *MongoStorage) Count(collection string) (int, error) {
	session, col := ms.conn(collection)
	defer session.Close()
	return col.Find(nil).Count()
}

func (ms *MongoStorage) Iterator(collection, sort string, fltr map[string]interface{}) Iterator {
	session, col := ms.conn(collection)
	defer session.Close()
	q := col.Find(filter(bson.M(fltr)))
	if sort != "" {
		q.Sort(sort)
	}
	return q.Iter()
}

func (ms *MongoStorage) GetAllPaged(tenant string, out interface{}, collection string, limit, offset int) error {
	session, col := ms.conn(collection)
	defer session.Close()
	q := col.Find(filter(bson.M{"tenant": tenant}))
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Skip(offset)
	}
	err := q.All(out)
	if err == mgo.ErrNotFound {
		return utils.ErrNotFound
	}
	return err
}

func (ms *MongoStorage) GetByNames(tenant string, names []string, out interface{}, collection string) error {
	session, col := ms.conn(collection)
	defer session.Close()
	flt := filter(bson.M{"tenant": tenant})
	if len(names) > 0 {
		flt["name"] = bson.M{"$in": names}
	}
	err := col.Find(flt).All(out)
	if err == mgo.ErrNotFound {
		return utils.ErrNotFound
	}
	return err
}

// EnsureIndexes creates db indexes
func (ms *MongoStorage) EnsureIndexes() error {
	dbSession := ms.session.Copy()
	defer dbSession.Close()
	db := dbSession.DB(ms.db)
	collectionIndexes := indexes[ms.storageType]

	for col, indexes := range collectionIndexes {
		for _, index := range indexes {
			if err := db.C(col).EnsureIndex(index); err != nil {
				return err
			}
		}
	}
	if ms.storageType == utils.CdrDB {
		// extra cdrs indexes
		for _, index := range ms.cdrsIndexes {
			if err := db.C(ColCdr).EnsureIndex(mgo.Index{Key: []string{index}, Unique: true}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ms *MongoStorage) Close() {
	ms.session.Close()
}

func (ms *MongoStorage) Ping() error {
	return ms.session.Ping()
}

func (ms *MongoStorage) Flush() (err error) {
	dbSession := ms.session.Copy()
	defer dbSession.Close()
	return dbSession.DB(ms.db).DropDatabase()
}

func (ms *MongoStorage) PreloadRatingCache() error {
	if ms.cacheCfg == nil {
		return nil
	}
	if ms.cacheCfg.Destinations != nil && ms.cacheCfg.Destinations.Precache {
		if err := ms.PreloadCacheForPrefix(utils.DESTINATION_PREFIX); err != nil {
			return err
		}
	}

	if ms.cacheCfg.RatingPlans != nil && ms.cacheCfg.RatingPlans.Precache {
		if err := ms.PreloadCacheForPrefix(utils.RATING_PLAN_PREFIX); err != nil {
			return err
		}
	}

	if ms.cacheCfg.RatingProfiles != nil && ms.cacheCfg.RatingProfiles.Precache {
		if err := ms.PreloadCacheForPrefix(utils.RATING_PROFILE_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.Lcr != nil && ms.cacheCfg.Lcr.Precache {
		if err := ms.PreloadCacheForPrefix(utils.LCR_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.CdrStats != nil && ms.cacheCfg.CdrStats.Precache {
		if err := ms.PreloadCacheForPrefix(utils.CDR_STATS_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.Actions != nil && ms.cacheCfg.Actions.Precache {
		if err := ms.PreloadCacheForPrefix(utils.ACTION_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.ActionPlans != nil && ms.cacheCfg.ActionPlans.Precache {
		if err := ms.PreloadCacheForPrefix(utils.ACTION_PLAN_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.ActionTriggers != nil && ms.cacheCfg.ActionTriggers.Precache {
		if err := ms.PreloadCacheForPrefix(utils.ACTION_TRIGGER_PREFIX); err != nil {
			return err
		}
	}
	if ms.cacheCfg.SharedGroups != nil && ms.cacheCfg.SharedGroups.Precache {
		if err := ms.PreloadCacheForPrefix(utils.SHARED_GROUP_PREFIX); err != nil {
			return err
		}
	}
	// add more prefixes if needed
	return nil
}

func (ms *MongoStorage) PreloadAccountingCache() error {
	if ms.cacheCfg == nil {
		return nil
	}
	if ms.cacheCfg.Aliases != nil && ms.cacheCfg.Aliases.Precache {
		if err := ms.PreloadCacheForPrefix(utils.ALIASES_PREFIX); err != nil {
			return err
		}
	}

	return nil
}

func (ms *MongoStorage) PreloadCacheForPrefix(prefix string) error {
	transID := cache2go.BeginTransaction()
	//FIXME: cache2go.RemPrefixKey(prefix, transID)
	switch prefix {
	case utils.RATING_PLAN_PREFIX:
		session, col := ms.conn(ColRpl)
		defer session.Close()
		iter := col.Find(bson.M{}).Iter()
		var rpl RatingPlan
		for iter.Next(&rpl) {
			_, err := ms.GetRatingPlan(rpl.Tenant, rpl.Name, transID)
			if err != nil {
				cache2go.RollbackTransaction(transID)
				return err
			}
		}
	default:
		return utils.ErrInvalidKey
	}
	cache2go.CommitTransaction(transID)
	return nil
}

func (ms *MongoStorage) GetRatingPlan(tenant, name, cacheParam string) (rp *RatingPlan, err error) {
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.RATING_PLAN_PREFIX+name); ok {
			if x != nil {
				return x.(*RatingPlan), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}
	rp = new(RatingPlan)
	session, col := ms.conn(ColRpl)
	defer session.Close()
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(rp)
	if err != nil {
		return nil, err
	}
	cache2go.Set(tenant, utils.RATING_PLAN_PREFIX+name, rp, cacheParam)
	return
}

func (ms *MongoStorage) SetRatingPlan(rp *RatingPlan) (err error) {
	session, col := ms.conn(ColRpl)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": rp.Tenant, "name": rp.Name}, rp)
	if err == nil && historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(), &response)
	}
	cache2go.Set(rp.Tenant, utils.RATING_PLAN_PREFIX+rp.Name, rp, utils.CACHE_SKIP)
	return err
}

func (ms *MongoStorage) GetRatingProfiles(direction, tenant, category, subject, cacheParam string) (rps []*RatingProfile, err error) {
	key := utils.ConcatKey(direction, category, subject)
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.RATING_PROFILE_PREFIX+key); ok {
			if x != nil {
				return x.([]*RatingProfile), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}
	session, col := ms.conn(ColRpf)
	defer session.Close()
	rps = make([]*RatingProfile, 0)
	// need filter for lcr load
	err = col.Find(filter(bson.M{
		"direction": direction,
		"tenant":    tenant,
		"category":  category,
		"subject":   subject,
	})).All(&rps)
	if err != nil {
		rps = nil
	}
	cache2go.Set(tenant, utils.RATING_PROFILE_PREFIX+key, rps, cacheParam)
	return
}

func (ms *MongoStorage) GetRatingProfile(direction, tenant, category, subject string, prefixMatching bool, cacheParam string) (rp *RatingProfile, err error) {
	prefix := ""
	if prefixMatching {
		prefix = "prefix"
	}
	key := utils.ConcatKey(direction, category, subject, prefix)
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.RATING_PROFILE_PREFIX+key); ok {
			if x != nil {
				return x.(*RatingProfile), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}
	session, col := ms.conn(ColRpf)
	defer session.Close()
	m := filter(bson.M{
		"direction": direction,
		"tenant":    tenant,
		"category":  category,
		"subject":   subject,
	}) // need filter for lcr load (tp_reader)
	if prefixMatching && subject != "" && subject != utils.ANY {
		x := make([]*RatingProfile, 0)
		m["subject"] = bson.M{"$in": utils.SplitPrefix(subject, MIN_PREFIX_MATCH)}
		err = col.Find(m).Sort("-subject").Limit(1).All(&x)
		if len(x) > 0 {
			rp = x[0]
		}
	} else {
		rp = new(RatingProfile)
		err = col.Find(m).One(rp)
	}

	if err != nil {
		rp = nil
	}
	cache2go.Set(tenant, utils.RATING_PROFILE_PREFIX+key, rp, cacheParam)
	return
}

func (ms *MongoStorage) SetRatingProfile(rp *RatingProfile) error {
	session, col := ms.conn(ColRpf)
	defer session.Close()
	_, err := col.Upsert(bson.M{"direction": rp.Direction, "tenant": rp.Tenant, "category": rp.Category, "subject": rp.Subject}, rp)
	if err == nil && historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", rp.GetHistoryRecord(false), &response)
	}
	cache2go.RemPrefixKey(rp.Tenant, utils.RATING_PROFILE_PREFIX, "")
	return err
}

func (ms *MongoStorage) RemoveRatingProfile(direction, tenant, category, subject string) error {
	session, col := ms.conn(ColRpf)
	defer session.Close()

	err := col.Remove(filter(bson.M{"direction": direction, "tenant": tenant, "category": category, "subject": subject}))
	if err == mgo.ErrNotFound {
		err = nil
	}
	if err != nil {
		return err
	}
	rpf := &RatingProfile{
		Direction: direction,
		Tenant:    tenant,
		Category:  category,
		Subject:   subject,
	}
	cache2go.RemPrefixKey(tenant, utils.RATING_PROFILE_PREFIX, "")
	if historyScribe != nil {
		var response int
		go historyScribe.Call("HistoryV1.Record", rpf.GetHistoryRecord(true), &response)
	}
	return nil
}

func (ms *MongoStorage) GetLCR(direction, tenant, category, account, subject string, prefixMatching bool, cacheParam string) (lcr *LCR, err error) {
	prefix := ""
	if prefixMatching {
		prefix = "prefix"
	}
	key := utils.ConcatKey(direction, category, account, subject, prefix)
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.LCR_PREFIX+key); ok {
			if x != nil {
				return x.(*LCR), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}
	session, col := ms.conn(ColLcr)
	defer session.Close()
	m := bson.M{
		"direction": bson.M{"$in": []string{direction, utils.ANY}},
		"tenant":    bson.M{"$in": []string{tenant, utils.ANY}},
		"category":  bson.M{"$in": []string{category, utils.ANY}},
		"account":   bson.M{"$in": []string{account, utils.ANY}},
		"subject":   bson.M{"$in": []string{subject, utils.ANY}},
	}
	if prefixMatching {
		lcrl := LCRList{}
		subjectList := utils.SplitPrefix(subject, MIN_PREFIX_MATCH)
		subjectList = append(subjectList, utils.ANY)
		m["subject"] = bson.M{"$in": subjectList}
		err = col.Find(m).Sort("-subject").Limit(1).All(&lcrl)
		if len(lcrl) > 0 {
			lcrl.Sort()
			lcr = lcrl[0] // best match
		}
	} else {
		lcr = &LCR{}
		err = col.Find(m).One(lcr)
	}

	if err != nil {
		lcr = nil
	}
	cache2go.Set(tenant, utils.LCR_PREFIX+key, lcr, cacheParam)
	return
}

func (ms *MongoStorage) SetLCR(lcr *LCR) error {
	session, col := ms.conn(ColLcr)
	defer session.Close()
	_, err := col.Upsert(bson.M{"direction": lcr.Direction, "tenant": lcr.Tenant, "category": lcr.Category, "account": lcr.Account, "subject": lcr.Subject}, lcr)
	cache2go.RemKey(lcr.Tenant, utils.LCR_PREFIX+lcr.FullID(), "")
	return err
}

func (ms *MongoStorage) GetDestinations(tenant, code, name, strategy string, cacheParam string) (result Destinations, err error) {
	key := utils.ConcatKey(code, name, strategy)
	if name == "" { // if search by name do not use cache (should be rare)
		if cacheParam == utils.CACHED {
			if x, ok := cache2go.Get(tenant, utils.DESTINATION_PREFIX+key); ok {
				if x != nil {
					return x.(Destinations), nil
				}
				return nil, utils.ErrNotFound
			}
			cacheParam = utils.CACHE_SKIP
		}
	}
	result = make([]*Destination, 0)

	session, col := ms.conn(ColDst)
	defer session.Close()

	switch strategy {
	case utils.DestExact:
		err = col.Find(filter(bson.M{"tenant": tenant, "code": code, "name": name})).All(&result)
	case utils.DestMatching:
		err = col.Find(filter(bson.M{"tenant": tenant, "code": bson.M{"$in": utils.SplitPrefix(code, MIN_PREFIX_MATCH)}})).Sort("-code").All(&result)
	default:
		err = utils.ErrInvalidKey
	}
	if err != nil {
		result = nil
	}
	if name == "" { // if search by name do not use cache (should be rare)
		cache2go.Set(tenant, utils.DESTINATION_PREFIX+key, result, cacheParam)
	}
	return
}

func (ms *MongoStorage) SetDestination(dest *Destination) (err error) {
	session, col := ms.conn(ColDst)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": dest.Tenant, "code": dest.Code, "name": dest.Name}, dest)
	cache2go.RemPrefixKey(dest.Tenant, utils.DESTINATION_PREFIX, "")
	if err == nil && historyScribe != nil {
		var response int
		historyScribe.Call("HistoryV1.Record", dest.GetHistoryRecord(false), &response)
	}
	return
}

func (ms *MongoStorage) RemoveDestination(dest *Destination) (err error) {
	session, col := ms.conn(ColDst)
	defer session.Close()
	col.Remove(bson.M{"tenant": dest.Tenant, "code": dest.Code, "name": dest.Name})
	if err == mgo.ErrNotFound {
		err = nil
	}
	if err != nil {
		return err
	}
	cache2go.RemPrefixKey(dest.Tenant, utils.DESTINATION_PREFIX, "") // remove all destinations because we don't know all the combinations'
	return
}

func (ms *MongoStorage) RemoveDestinations(tenant, code, name string) (err error) {
	session, col := ms.conn(ColDst)
	defer session.Close()
	_, err = col.RemoveAll(filter(bson.M{"tenant": tenant, "code": code, "name": name}))
	if err == mgo.ErrNotFound {
		err = nil
	}
	if err != nil {
		return err
	}
	cache2go.RemPrefixKey(tenant, utils.DESTINATION_PREFIX, "") // remove all destinations because we don't know all the combinations'
	return
}

func (ms *MongoStorage) GetTiming(tenant, name string) (result *Timing, err error) {
	if name == utils.ANY {
		return &Timing{
			Tenant:    tenant,
			Name:      utils.ANY,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
			Time:      "00:00:00",
		}, nil
	}
	if name == utils.ASAP {
		return &Timing{
			Tenant:    tenant,
			Name:      utils.ASAP,
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
			Time:      utils.ASAP,
		}, nil
	}

	session, col := ms.conn(ColTmg)
	defer session.Close()
	result = &Timing{}
	err = col.Find(filter(bson.M{"tenant": tenant, "name": name})).One(result)
	if err != nil {
		result = nil
	}
	return
}

func (ms *MongoStorage) SetTiming(tmg *Timing) (err error) {
	session, col := ms.conn(ColTmg)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": tmg.Tenant, "Name": tmg.Name}, tmg)
	return
}

func (ms *MongoStorage) GetRate(tenant, name string) (result *Rate, err error) {
	session, col := ms.conn(ColRts)
	defer session.Close()
	result = &Rate{}
	err = col.Find(filter(bson.M{"tenant": tenant, "name": name})).One(result)
	if err != nil {
		result = nil
	}
	return
}

func (ms *MongoStorage) SetRate(rt *Rate) (err error) {
	session, col := ms.conn(ColRts)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": rt.Tenant, "Name": rt.Name}, rt)
	return
}

func (ms *MongoStorage) GetDestinationRate(tenant, name string) (result *DestinationRate, err error) {
	session, col := ms.conn(ColDrt)
	defer session.Close()
	result = &DestinationRate{}
	err = col.Find(filter(bson.M{"tenant": tenant, "name": name})).One(result)
	if err != nil {
		result = nil
	}
	return
}

func (ms *MongoStorage) SetDestinationRate(drt *DestinationRate) (err error) {
	session, col := ms.conn(ColDrt)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": drt.Tenant, "name": drt.Name}, drt)
	return
}

func (ms *MongoStorage) GetActionGroup(tenant, name, cacheParam string) (ag *ActionGroup, err error) {
	key := utils.ConcatKey(tenant, name)
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.ACTION_PREFIX+key); ok {
			if x != nil {
				return x.(*ActionGroup), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}
	session, col := ms.conn(ColAct)
	defer session.Close()
	ag = &ActionGroup{}
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(ag)
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
		ag = nil
	}
	cache2go.Set(tenant, utils.ACTION_PREFIX+key, ag, cacheParam)
	return
}

func (ms *MongoStorage) SetActionGroup(ag *ActionGroup) error {
	session, col := ms.conn(ColAct)
	defer session.Close()
	_, err := col.Upsert(bson.M{"tenant": ag.Tenant, "name": ag.Name}, ag)
	cache2go.RemKey(ag.Tenant, utils.ACTION_PREFIX+ag.Name, "")
	return err
}

func (ms *MongoStorage) RemoveActionGroup(tenant, name string) error {
	session, col := ms.conn(ColAct)
	defer session.Close()
	cache2go.RemKey(tenant, utils.ACTION_PREFIX+name, "")
	err := col.Remove(bson.M{"tenant": tenant, "name": name})
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (ms *MongoStorage) GetSharedGroup(tenant, name, cacheParam string) (sg *SharedGroup, err error) {
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.SHARED_GROUP_PREFIX+name); ok {
			if x != nil {
				return x.(*SharedGroup), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}
	session, col := ms.conn(ColShg)
	defer session.Close()
	sg = &SharedGroup{}
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(sg)
	if err != nil {
		sg = nil
	}
	cache2go.Set(tenant, utils.SHARED_GROUP_PREFIX+name, sg, cacheParam)
	return
}

func (ms *MongoStorage) SetSharedGroup(sg *SharedGroup) (err error) {
	session, col := ms.conn(ColShg)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": sg.Tenant, "name": sg.Name}, sg)
	cache2go.RemKey(sg.Tenant, utils.SHARED_GROUP_PREFIX+sg.Name, "")
	return err
}

func (ms *MongoStorage) GetAccount(tenant, name string) (result *Account, err error) {
	result = new(Account)
	session, col := ms.conn(ColAcc)
	defer session.Close()
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(result)
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
		result = nil
	}
	return
}

func (ms *MongoStorage) SetAccount(acc *Account) error {
	//utils.Logger.Info("Saving acount: ", zap.String("account", utils.ToJSON(acc)))
	// never override existing account with an empty one
	// UPDATE: if all balances expired and were cleaned it makes
	// sense to write empty balance map
	if len(acc.BalanceMap) == 0 {
		if ac, err := ms.GetAccount(acc.Tenant, acc.Name); err == nil && !ac.allBalancesExpired() {
			ac.TriggerIDs = acc.TriggerIDs
			ac.TriggerRecords = acc.TriggerRecords
			ac.UnitCounters = acc.UnitCounters
			ac.AllowNegative = acc.AllowNegative
			ac.Disabled = acc.Disabled
			acc = ac
		}
	}
	session, col := ms.conn(ColAcc)
	defer session.Close()
	_, err := col.Upsert(bson.M{"tenant": acc.Tenant, "name": acc.Name}, acc)
	return err
}

func (ms *MongoStorage) RemoveAccount(tenant, name string) error {
	session, col := ms.conn(ColAcc)
	defer session.Close()
	err := col.Remove(bson.M{"tenant": tenant, "name": name})
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err

}

func (ms *MongoStorage) SetSimpleAccount(sa *SimpleAccount) error {
	session, col := ms.conn(ColSac)
	defer session.Close()
	_, err := col.Upsert(bson.M{"tenant": sa.Tenant, "name": sa.Name}, sa)
	return err
}

func (ms *MongoStorage) RemoveSimpleAccount(tenant, name string) error {
	session, col := ms.conn(ColSac)
	defer session.Close()
	err := col.Remove(bson.M{"tenant": tenant, "name": name})
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (ms *MongoStorage) GetCdrStatsQueue(tenant, name string) (sq *StatsQueue, err error) {
	sq = &StatsQueue{}
	session, col := ms.conn(ColStq)
	defer session.Close()
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(sq)
	if err != nil {
		sq = nil
	}
	return
}

func (ms *MongoStorage) SetCdrStatsQueue(sq *StatsQueue) (err error) {
	session, col := ms.conn(ColStq)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": sq.Tenant, "name": sq.Name}, sq)
	return
}

func (ms *MongoStorage) RemoveCdrStatsQueue(tenant, name string) (err error) {
	session, col := ms.conn(ColStq)
	defer session.Close()
	err = col.Remove(bson.M{"tenant": tenant, "name": name})
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (ms *MongoStorage) PushQCDR(qcdr *QCDR) error {
	session, col := ms.conn(ColQcr)
	defer session.Close()
	qcdr.ID = bson.NewObjectId()
	return col.Insert(qcdr)
}

func (ms *MongoStorage) PopQCDR(tenant, name string, fltr map[string]interface{}, limit int) ([]*QCDR, error) {
	session, col := ms.conn(ColQcr)
	defer session.Close()

	if fltr == nil {
		fltr = make(map[string]interface{})
	}
	fltr["tenant"] = tenant
	fltr["name"] = name
	q := col.Find(filter(bson.M(fltr))).Sort("event_time")
	if limit > 0 {
		q = q.Limit(limit)
	}
	qcdrs := make([]*QCDR, 0)
	err := q.All(&qcdrs)
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	objectIDs := make([]bson.ObjectId, len(qcdrs))
	for i, q := range qcdrs {
		objectIDs[i] = q.ID
	}

	// remove documents
	if _, err = col.RemoveAll(bson.M{"_id": bson.M{"$in": objectIDs}}); err != nil {
		return nil, err
	}
	return qcdrs, nil
}

func (ms *MongoStorage) RemoveQCDRs(tenant, name string) error {
	session, col := ms.conn(ColQcr)
	defer session.Close()
	_, err := col.RemoveAll(filter(bson.M{"tenant": tenant, "name": name}))
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (ms *MongoStorage) GetSubscribers() (result map[string]*SubscriberData, err error) {
	session, col := ms.conn(ColPbs)
	defer session.Close()
	iter := col.Find(nil).Iter()
	result = make(map[string]*SubscriberData)
	var kv struct {
		Key   string
		Value *SubscriberData
	}
	for iter.Next(&kv) {
		result[kv.Key] = kv.Value
	}
	err = iter.Close()
	return
}

func (ms *MongoStorage) SetSubscriber(key string, sub *SubscriberData) (err error) {
	session, col := ms.conn(ColPbs)
	defer session.Close()
	_, err = col.Upsert(bson.M{"key": key}, &struct {
		Key   string
		Value *SubscriberData
	}{Key: key, Value: sub})
	return err
}

func (ms *MongoStorage) RemoveSubscriber(key string) (err error) {
	session, col := ms.conn(ColPbs)
	defer session.Close()
	err = col.Remove(bson.M{"key": key})
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (ms *MongoStorage) SetUser(up *UserProfile) (err error) {
	session, col := ms.conn(ColUsr)
	defer session.Close()
	_, err = col.Upsert(bson.M{"tenant": up.Tenant, "name": up.Name}, up)
	return err
}

func (ms *MongoStorage) GetUser(tenant, name string) (up *UserProfile, err error) {
	session, col := ms.conn(ColUsr)
	defer session.Close()
	up = &UserProfile{}
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(up)
	if err != nil {
		up = nil
	}
	return
}

func (ms *MongoStorage) RemoveUser(tenant, name string) (err error) {
	session, col := ms.conn(ColUsr)
	defer session.Close()
	err = col.Remove(bson.M{"tenant": tenant, "name": name})
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (ms *MongoStorage) GetAlias(direction, tenant, category, account, subject, context, cacheParam string) (al *Alias, err error) {
	key := utils.ConcatKey(direction, category, account, subject, context)
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.ALIASES_PREFIX+key); ok {
			if x != nil {
				return x.(*Alias), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}

	session, col := ms.conn(ColAls)
	defer session.Close()
	all := AliasList{}
	err = col.Find(bson.M{
		"direction": bson.M{"$in": []string{direction, utils.ANY}},
		"tenant":    bson.M{"$in": []string{tenant, utils.ANY}},
		"category":  bson.M{"$in": []string{category, utils.ANY}},
		"account":   bson.M{"$in": []string{account, utils.ANY}},
		"subject":   bson.M{"$in": []string{subject, utils.ANY}},
		"context":   context,
	}).All(&all)
	if err == nil && len(all) > 0 {
		all.Sort() // sort by precision
		al = all[0]
	}
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
	}
	cache2go.Set(tenant, utils.ALIASES_PREFIX+key, al, cacheParam)

	return
}

func (ms *MongoStorage) SetAlias(al *Alias) (err error) {
	session, col := ms.conn(ColAls)
	defer session.Close()
	_, err = col.Upsert(bson.M{"direction": al.Direction, "tenant": al.Tenant, "category": al.Category, "account": al.Account, "subject": al.Subject, "context": al.Context}, al)
	cache2go.RemKey(al.Tenant, utils.ALIASES_PREFIX+al.FullID(), "")
	return err
}

func (ms *MongoStorage) RemoveAlias(direction, tenant, category, account, subject, context string) (err error) {
	key := utils.ALIASES_PREFIX + utils.ConcatKey(direction, category, account, subject, context)
	session, col := ms.conn(ColAls)
	defer session.Close()
	err = col.Remove(bson.M{"direction": direction, "tenant": tenant, "category": category, "account": account, "subject": subject,
		"context": context})
	if err == mgo.ErrNotFound {
		err = nil
	}
	if err != nil {
		return err
	}
	cache2go.RemKey(tenant, key, "")
	return
}

func (ms *MongoStorage) GetReverseAlias(tenant, context, target, alias, cacheParam string) (als []*Alias, err error) {
	key := utils.ConcatKey(context, target, alias)
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.ALIASES_PREFIX+key); ok {
			if x != nil {
				return x.([]*Alias), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}

	session, col := ms.conn(ColAls)
	defer session.Close()
	als = make([]*Alias, 0)
	err = col.Find(bson.M{"tenant": tenant, "context": context, "index.target": target, "index.alias": alias}).All(&als)
	if err == mgo.ErrNotFound {
		err = nil
		als = nil
	}
	if err != nil {
		return nil, err
	}
	cache2go.Set(tenant, utils.ALIASES_PREFIX+key, als, cacheParam)
	return
}

// Adds a single load instance to load history
func (ms *MongoStorage) AddLoadHistory(ldInst *utils.LoadInstance) error {
	session, col := ms.conn(ColLht)
	defer session.Close()
	return col.Insert(ldInst)
}

func (ms *MongoStorage) GetActionTriggers(tenant, name, cacheParam string) (atrg *ActionTriggerGroup, err error) {
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.ACTION_TRIGGER_PREFIX+name); ok {
			if x != nil {
				return x.(*ActionTriggerGroup), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}

	session, col := ms.conn(ColAtr)
	defer session.Close()
	atrg = &ActionTriggerGroup{}
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(atrg)
	if err != nil {
		atrg = nil
	}
	cache2go.Set(tenant, utils.ACTION_TRIGGER_PREFIX+name, atrg, cacheParam)
	return
}

func (ms *MongoStorage) SetActionTriggers(atrg *ActionTriggerGroup) (err error) {
	session, col := ms.conn(ColAtr)
	defer session.Close()
	if len(atrg.ActionTriggers) == 0 {
		err = col.Remove(bson.M{"tenant": atrg.Tenant, "name": atrg.Name}) // delete the atrg
		if err != mgo.ErrNotFound {
			return err
		}
		return nil
	}
	_, err = col.Upsert(bson.M{"tenant": atrg.Tenant, "name": atrg.Name}, atrg)
	cache2go.RemKey(atrg.Tenant, utils.ACTION_TRIGGER_PREFIX+atrg.Name, "")
	return err
}

func (ms *MongoStorage) RemoveActionTriggers(tenant, name string) error {
	session, col := ms.conn(ColAtr)
	defer session.Close()
	cache2go.RemKey(tenant, utils.ACTION_TRIGGER_PREFIX+name, "")
	err := col.Remove(bson.M{"tenant": tenant, "name": name})
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (ms *MongoStorage) GetActionPlan(tenant, name, cacheParam string) (apl *ActionPlan, err error) {
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.ACTION_PLAN_PREFIX+name); ok {
			if x != nil {
				return x.(*ActionPlan), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}
	session, col := ms.conn(ColApl)
	defer session.Close()
	apl = &ActionPlan{}
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(apl)
	if err != nil {
		return nil, err
	}
	cache2go.Set(tenant, utils.ACTION_PLAN_PREFIX+name, apl, cacheParam)
	return
}

func (ms *MongoStorage) SetActionPlan(apl *ActionPlan) (err error) {
	session, col := ms.conn(ColApl)
	defer session.Close()
	// clean dots from account ids map
	if len(apl.ActionTimings) == 0 {
		cache2go.RemKey(apl.Tenant, utils.ACTION_PLAN_PREFIX+apl.Name, "")
		err := col.Remove(bson.M{"tenant": apl.Tenant, "name": apl.Name})
		if err != mgo.ErrNotFound {
			return err
		}
		return nil
	}

	_, err = col.Upsert(bson.M{"tenant": apl.Tenant, "name": apl.Name}, apl)
	cache2go.RemKey(apl.Tenant, utils.ACTION_PLAN_PREFIX+apl.Name, "")
	return err
}

func (ms *MongoStorage) GetActionPlanBinding(tenant, account, actionPlan string) (*ActionPlanBinding, error) {
	session, col := ms.conn(ColApb)
	defer session.Close()
	apb := &ActionPlanBinding{}
	err := col.Find(bson.M{"tenant": tenant, "account": account, "action_plan": actionPlan}).One(apb)
	if err != nil {
		apb = nil
	}
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
	}
	return apb, err
}

func (ms *MongoStorage) SetActionPlanBinding(apb *ActionPlanBinding) error {
	session, col := ms.conn(ColApb)
	defer session.Close()
	_, err := col.Upsert(bson.M{"tenant": apb.Tenant, "account": apb.Account, "action_plan": apb.ActionPlan}, apb)
	return err
}

func (ms *MongoStorage) RemoveActionPlanBindings(tenant, account, actionPlan string) error {
	session, col := ms.conn(ColApb)
	defer session.Close()
	err := col.Remove(filter(bson.M{"tenant": tenant, "account": account, "action_plan": actionPlan}))
	if err == mgo.ErrNotFound {
		return nil
	}
	return err
}

func (ms *MongoStorage) PushTask(t *Task) error {
	session, col := ms.conn(ColTsk)
	defer session.Close()
	return col.Insert(t)
}

func (ms *MongoStorage) PopTask() (t *Task, err error) {
	session, col := ms.conn(ColTsk)
	defer session.Close()

	t = &Task{}
	_, err = col.Find(nil).Apply(mgo.Change{Remove: true}, &t)
	if err != nil {
		return nil, err
	}

	return
}

func (ms *MongoStorage) GetDerivedChargers(direction, tenant, category, account, subject, cacheParam string) (dcs utils.DerivedChargers, err error) {
	key := utils.ConcatKey(direction, category, account, subject)
	if cacheParam == utils.CACHED {
		if x, ok := cache2go.Get(tenant, utils.DERIVEDCHARGERS_PREFIX+key); ok {
			if x != nil {
				return x.(utils.DerivedChargers), nil
			}
			return nil, utils.ErrNotFound
		}
		cacheParam = utils.CACHE_SKIP
	}
	session, col := ms.conn(ColDcs)
	defer session.Close()
	dcs = make(utils.DerivedChargers, 0)
	err = col.Find(bson.M{
		"direction": direction,
		"tenant":    bson.M{"$in": []string{tenant, utils.ANY}},
		"category":  bson.M{"$in": []string{category, utils.ANY}},
		"account":   bson.M{"$in": []string{account, utils.ANY}},
		"subject":   bson.M{"$in": []string{subject, utils.ANY}},
	}).All(&dcs)
	if err != nil {
		dcs = nil
	} else {
		dcs.Sort() // sort by precision
	}
	if err == mgo.ErrNotFound {
		err = utils.ErrNotFound
	}
	cache2go.Set(tenant, utils.DERIVEDCHARGERS_PREFIX+key, dcs, cacheParam)
	return
}

func (ms *MongoStorage) SetDerivedChargers(dcs *utils.DerivedChargerGroup) (err error) {
	key := utils.ConcatKey(dcs.Direction, dcs.Category, dcs.Account, dcs.Subject)
	if dcs == nil || len(dcs.Chargers) == 0 {
		cache2go.RemKey(dcs.Tenant, utils.DERIVEDCHARGERS_PREFIX+key, "")
		session, col := ms.conn(ColDcs)
		defer session.Close()
		err = col.Remove(bson.M{"direction": dcs.Direction, "tenant": dcs.Tenant, "category": dcs.Category, "account": dcs.Account, "subject": dcs.Subject})
		if err != mgo.ErrNotFound {
			return err
		}
		return nil
	}
	session, col := ms.conn(ColDcs)
	defer session.Close()
	_, err = col.Upsert(bson.M{"direction": dcs.Direction, "tenant": dcs.Tenant, "category": dcs.Category, "account": dcs.Account, "subject": dcs.Subject}, dcs)
	cache2go.RemKey(dcs.Tenant, utils.DERIVEDCHARGERS_PREFIX+key, "")
	return err
}

func (ms *MongoStorage) SetCdrStats(cs *CdrStats) error {
	session, col := ms.conn(ColCrs)
	defer session.Close()
	_, err := col.Upsert(bson.M{"tenant": cs.Tenant, "name": cs.Name}, cs)
	return err
}

func (ms *MongoStorage) GetCdrStats(tenant, name string) (cs *CdrStats, err error) {
	cs = &CdrStats{}
	session, col := ms.conn(ColCrs)
	defer session.Close()
	err = col.Find(bson.M{"tenant": tenant, "name": name}).One(cs)
	return
}

func (ms *MongoStorage) RemoveCdrStats(tenant, name string) (err error) {
	session, col := ms.conn(ColCrs)
	defer session.Close()
	err = col.Remove(bson.M{"tenant": tenant, "name": name})
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (ms *MongoStorage) SetStructVersion(v *StructVersion) (err error) {
	session, col := ms.conn(ColVer)
	defer session.Close()
	return col.Insert(v)
}

func (ms *MongoStorage) GetStructVersion() (rsv *StructVersion, err error) {
	session, col := ms.conn(ColVer)
	defer session.Close()
	rsv = &StructVersion{}
	err = col.Find(nil).Sort("-$natural").Limit(1).One(rsv) // latest document
	if err == mgo.ErrNotFound {
		rsv = nil
	}
	return
}

func (ms *MongoStorage) GetResourceLimit(id string, skipCache bool, transactionID string) (rl *ResourceLimit, err error) {
	/*key := utils.ResourceLimitsPrefix + id
	if !skipCache {
		if x, ok := cache.Get(key); ok {
			if x != nil {
				return x.(*ResourceLimit), nil
			}
			return nil, utils.ErrNotFound
		}
	}
	session, col := ms.conn(ColRL)
	defer session.Close()
	if err = col.Find(bson.M{"id": id}).One(rl); err != nil {
		if err == mgo.ErrNotFound {
			err = utils.ErrNotFound
			cache.Set(key, nil, cacheCommit(transactionID), transactionID)
		}
		return nil, err
	}
	cache.Set(key, rl, cacheCommit(transactionID), transactionID)*/
	return
}

func (ms *MongoStorage) SetResourceLimit(rl *ResourceLimit, transactionID string) (err error) {
	/*session, col := ms.conn(ColRL)
	defer session.Close()
	_, err = col.Upsert(bson.M{"id": rl.ID}, rl)*/
	return err
}

func (ms *MongoStorage) RemoveResourceLimit(id string, transactionID string) error {
	/*session, col := ms.conn(ColRL)
	defer session.Close()
	if err := col.Remove(bson.M{"id": id}); err != nil {
		return err
	}
	cache.RemKey(utils.ResourceLimitsPrefix+id, cacheCommit(transactionID), transactionID)*/
	return nil
}
