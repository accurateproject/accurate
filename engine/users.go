package engine

import (
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

type UserProfile struct {
	Tenant string            `bson:"tenant"`
	Name   string            `bson:"name"`
	Masked bool              `bson:"masked"`
	Index  map[string]string `bson:"index"`
	Query  string            `bson:"query"`
	Weight float64           `bson:"weight"`
	query  *utils.StructQ
}

func (up *UserProfile) Ponder() int {
	if up.query == nil {
		return 0
	}
	return up.query.Complexity()
}

func (up *UserProfile) Equals(o *UserProfile) bool {
	return up.Tenant == o.Tenant && up.Name == o.Name
}

func (up *UserProfile) getQuery() *utils.StructQ {
	if up.query != nil {
		return up.query
	}
	up.query, _ = utils.NewStructQ(up.Query) // the error should be checked during load
	return up.query
}

type UserProfiles []*UserProfile

func (ups UserProfiles) sort() {
	sort.Slice(ups, func(j, i int) bool { // get higher Weight and ponder in front
		return ups[i].Weight < ups[j].Weight ||
			(ups[i].Weight == ups[j].Weight && ups[i].Ponder() < ups[j].Ponder())
	})
}

func (ups *UserProfiles) remove(up *UserProfile) {
	index := -1
	for i, itUP := range *ups {
		if itUP.Equals(up) {
			index = i
			break
		}
	}
	if index != -1 {
		*ups = append((*ups)[:index], (*ups)[index+1:]...)
	}
}

func (ud *UserProfile) FullID() string {
	return utils.ConcatKey(ud.Tenant, ud.Name)
}

func (ud *UserProfile) SetID(id string) error {
	vals := strings.SplitN(id, utils.CONCATENATED_KEY_SEP, 2)
	ud.Tenant = vals[0]
	if len(vals) == 2 {
		ud.Name = vals[1]
	}
	return nil
}

type UserService interface {
	SetUser(UserProfile, *string) error
	RemoveUser(UserProfile, *string) error
	UpdateUser(UserProfile, *string) error
	GetUsers(AttrGetUsers, *UserProfiles) error
	AddIndex([]string, *string) error
	GetIndexes(string, *map[string][]string) error
	ReloadUsers(attr AttrReloadUsers, reply *string) error
}

type UserMap struct {
	table        map[string]*UserProfile
	index        map[string]UserProfiles
	indexKeys    utils.StringMap
	accountingDb AccountingStorage
	mu           sync.RWMutex
}

func NewUserMap(accountingDb AccountingStorage, indexes []string) (*UserMap, error) {
	um := newUserMap(accountingDb, utils.NewStringMap(indexes...))
	var reply string
	if err := um.ReloadUsers(AttrReloadUsers{}, &reply); err != nil {
		return nil, err
	}
	return um, nil
}

func newUserMap(accountingDb AccountingStorage, indexes utils.StringMap) *UserMap {
	return &UserMap{
		table:        make(map[string]*UserProfile),
		index:        make(map[string]UserProfiles),
		indexKeys:    indexes,
		accountingDb: accountingDb,
	}
}

func (um *UserMap) ReloadUsers(attr AttrReloadUsers, reply *string) error {
	um.mu.Lock()

	// backup old data
	oldTable := um.table
	oldIndex := um.index
	um.table = make(map[string]*UserProfile)
	um.index = make(map[string]UserProfiles)

	// load from db
	up := &UserProfile{}
	upIter := um.accountingDb.Iterator(ColUsr, "", map[string]interface{}{"tenant": attr.Tenant})
	for upIter.Next(up) {
		um.table[up.FullID()] = up
		up = &UserProfile{}
	}
	if err := upIter.Close(); err != nil {
		// restore old data before return
		um.table = oldTable
		um.index = oldIndex

		*reply = err.Error()
		return err
	}
	um.mu.Unlock()

	if len(um.indexKeys) != 0 {
		var s string
		if err := um.AddIndex(um.indexKeys.Slice(), &s); err != nil {
			utils.Logger.Error("error adding indexes to user profile service: ", zap.Stringer("keys", um.indexKeys), zap.Error(err))
			um.mu.Lock()
			um.table = oldTable
			um.index = oldIndex
			um.mu.Unlock()
			*reply = err.Error()
			return err
		}
	}
	*reply = utils.OK
	return nil
}

func (um *UserMap) SetUser(up *UserProfile, reply *string) error {
	um.mu.Lock()
	defer um.mu.Unlock()
	up.Query = strings.Replace(up.Query, `'`, `"`, -1)
	if err := um.accountingDb.SetUser(up); err != nil {
		*reply = err.Error()
		return err
	}
	um.table[up.FullID()] = up
	um.addIndex(up, um.indexKeys)
	*reply = utils.OK
	return nil
}

func (um *UserMap) RemoveUser(up *UserProfile, reply *string) error {
	um.mu.Lock()
	defer um.mu.Unlock()
	if err := um.accountingDb.RemoveUser(up.Tenant, up.Name); err != nil {
		*reply = err.Error()
		return err
	}
	delete(um.table, up.FullID())
	um.deleteIndex(up)
	*reply = utils.OK
	return nil
}

func (um *UserMap) UpdateUser(up *UserProfile, reply *string) error {
	um.mu.Lock()
	defer um.mu.Unlock()
	up.Query = strings.Replace(up.Query, `'`, `"`, -1)
	oldUp, found := um.table[up.FullID()]

	if err := um.accountingDb.SetUser(up); err != nil {
		*reply = err.Error()
		return err
	}
	um.table[up.FullID()] = up
	if found {
		um.deleteIndex(oldUp)
	}
	um.addIndex(up, um.indexKeys)
	*reply = utils.OK
	return nil
}

type AttrGetUsers struct {
	Object interface{}
	Masked bool
}

type AttrReloadUsers struct {
	Tenant string
}

func (um *UserMap) GetUsers(attr AttrGetUsers, results *UserProfiles) error {
	um.mu.RLock()
	defer um.mu.RUnlock()
	table := um.table // no index

	// search index
	indexedTable := make(map[string]*UserProfile)
	valuesMap := utils.FieldByName(attr.Object, um.indexKeys)
	for fieldName, valueInterface := range valuesMap {
		value, ok := valueInterface.(string)
		if !ok {
			continue
		}

		if indexedUPs, found := um.index[utils.ConcatKey(fieldName, value)]; found {
			for _, indexedUP := range indexedUPs {
				indexedTable[indexedUP.FullID()] = indexedUP
			}
		}
	}

	if len(indexedTable) > 0 {
		table = indexedTable
	}
	candidates := make(UserProfiles, 0) // It should [], not nil
	for _, tableUP := range table {
		// skip masked if not asked for
		if attr.Masked == false && tableUP.Masked == true {
			continue
		}
		//utils.Logger.Info("checking user", zap.String("query", tableUP.Query))
		if passed, err := tableUP.getQuery().Query(attr.Object, false); !passed || err != nil {
			if err != nil {
				utils.Logger.Error("<GetUsers> error checking query: ", zap.Error(err), zap.String("filter", tableUP.Query))
			}
			continue
		}
		// all filters passed, add to candidates
		candidates = append(candidates, tableUP)
		if *config.Get().Users.ComplexityMatch == false { // only return first match
			break
		}
	}
	if len(candidates) > 1 {
		candidates.sort()
	}
	*results = candidates
	return nil
}

func (um *UserMap) AddIndex(indexes []string, reply *string) error {
	um.mu.Lock()
	defer um.mu.Unlock()
	if um.indexKeys == nil {
		um.indexKeys = utils.StringMap{}
	}
	um.indexKeys.CopySlice(indexes)
	indexMap := utils.NewStringMap(indexes...)
	for _, up := range um.table {
		um.addIndex(up, indexMap)
	}
	*reply = utils.OK
	return nil
}

func (um *UserMap) addIndex(up *UserProfile, indexes utils.StringMap) {
	for index := range indexes {
		if index == "Tenant" {
			if up.Tenant != "" {
				indexKey := utils.ConcatKey(index, up.Tenant)
				um.index[indexKey] = append(um.index[indexKey], up)
			}
			continue
		}
		if index == "Name" {
			if up.Name != "" {
				indexKey := utils.ConcatKey(index, up.Name)
				um.index[indexKey] = append(um.index[indexKey], up)
			}
			continue
		}

		for k, v := range up.Index {
			if k == index && v != "" {
				indexKey := utils.ConcatKey(k, v)
				um.index[indexKey] = append(um.index[indexKey], up)
			}
		}
	}
}

func (um *UserMap) deleteIndex(up *UserProfile) {
	for index := range um.indexKeys {
		if index == "Tenant" {
			if up.Tenant != "" {
				indexKey := utils.ConcatKey(index, up.Tenant)
				x := um.index[indexKey]
				(&x).remove(up)
				um.index[indexKey] = x
				if len(um.index[indexKey]) == 0 {
					delete(um.index, indexKey)
				}
			}
			continue
		}
		if index == "Name" {
			if up.Name != "" {
				indexKey := utils.ConcatKey(index, up.Name)
				x := um.index[indexKey]
				(&x).remove(up)
				um.index[indexKey] = x
				if len(um.index[indexKey]) == 0 {
					delete(um.index, indexKey)
				}
			}
			continue
		}
		for k, v := range up.Index {
			if k == index && v != "" {
				indexKey := utils.ConcatKey(k, v)
				x := um.index[indexKey]
				(&x).remove(up)
				um.index[indexKey] = x
				if len(um.index[indexKey]) == 0 {
					delete(um.index, indexKey)
				}
			}
		}
	}
}

func (um *UserMap) GetIndexes(in string, reply *map[string][]string) error {
	um.mu.RLock()
	defer um.mu.RUnlock()
	indexes := make(map[string][]string)
	for key, values := range um.index {
		var vs []string
		for _, val := range values {
			vs = append(vs, val.FullID())
		}
		indexes[key] = vs
	}
	*reply = indexes
	return nil
}

func (um *UserMap) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(um).MethodByName(parts[1])
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

// extraFields - Field name in the interface containing extraFields information
func LoadUserProfile(in UsersNeeder, masked bool) error {
	if userService == nil || !in.NeedsUsers() {
		return nil
	}
	ups := UserProfiles{}
	if err := userService.Call("UsersV1.GetUsers", AttrGetUsers{Object: in, Masked: masked}, &ups); err != nil {
		return err
	}
	if len(ups) > 0 {
		up := ups[0]
		_, err := up.getQuery().Query(in, true)
		return err
	}
	return utils.ErrUserNotFound
}

type UsersNeeder interface {
	NeedsUsers() bool
}
