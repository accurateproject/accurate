package engine

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
)

// Temporary export AliasService for the ApierV1 to be able to emulate old APIs
func GetAliasService() rpcclient.RpcClientConnection {
	return aliasService
}

type Alias struct {
	Direction string        `bson:"direction"`
	Tenant    string        `bson:"tenant"`
	Category  string        `bson:"category"`
	Account   string        `bson:"account"`
	Subject   string        `bson:"subject"`
	Context   string        `bson:"context"`
	Index     []*AliasIndex `bson:"index"`
	Values    AliasValues   `bson:"values"`
}

type AliasValue struct {
	DestinationID string  `bson:"destination_id"`
	Fields        string  `bson:"fields"`
	Weight        float64 `bson:"weight"`
	fields        *utils.StructQ
}

type AliasIndex struct {
	Target string `bson:"target"`
	Alias  string `bson:"alias"`
}

func (av *AliasValue) Equals(other *AliasValue) bool {
	return av.DestinationID == other.DestinationID &&
		av.Fields == other.Fields &&
		av.Weight == other.Weight
}

func (av *AliasValue) getFields() *utils.StructQ {
	if av.fields != nil {
		return av.fields
	}
	av.fields, _ = utils.NewStructQ(av.Fields) // error should be checked on load
	return av.fields
}

type AliasValues []*AliasValue

func (avs AliasValues) Sort() {
	sort.Slice(avs, func(j, i int) bool { // get higher weight in front
		return avs[i].Weight < avs[j].Weight
	})
}

func (avs AliasValues) GetValueByDestId(destID string) *AliasValue {
	for _, value := range avs {
		if value.DestinationID == destID {
			return value
		}
	}
	return nil
}

func (al *Alias) FullID() string {
	return utils.ConcatKey(al.Direction, al.Category, al.Account, al.Subject, al.Context)
}

func (al *Alias) precision() int {
	precision := 0
	if al.Direction != utils.ANY {
		precision++
	}
	if al.Category != utils.ANY {
		precision++
	}
	if al.Account != utils.ANY {
		precision++
	}
	if al.Subject != utils.ANY {
		precision++
	}
	if al.Context != utils.ANY {
		precision++
	}
	return precision
}

type AttrAlias struct {
	Direction   string
	Tenant      string
	Category    string
	Account     string
	Subject     string
	Context     string
	Destination string
}

type AttrReverseAlias struct {
	Tenant  string
	Context string
	Target  string
	Alias   string
}

type AttrMatchingAlias struct {
	Tenant      string
	Direction   string
	Category    string
	Account     string
	Subject     string
	Context     string
	Destination string
	Target      string
	Original    string
}

type AliasService interface {
	SetAlias(Alias, *string) error
	UpdateAlias(Alias, *string) error
	RemoveAlias(Alias, *string) error
	GetAlias(Alias, *Alias) error
	GetReverseAlias(AttrReverseAlias, *map[string][]*Alias) error
}

type AliasHandler struct {
	accountingDb AccountingStorage
	mu           sync.RWMutex
}

func NewAliasHandler(accountingDb AccountingStorage) *AliasHandler {
	return &AliasHandler{
		accountingDb: accountingDb,
	}
}

type AttrAddAlias struct {
	Alias     *Alias
	Overwrite bool
}

// SetAlias will set/overwrite specified alias
func (am *AliasHandler) SetAlias(attr *AttrAddAlias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	var oldAlias *Alias
	if !attr.Overwrite { // get previous value
		oldAlias, _ = am.accountingDb.GetAlias(attr.Alias.Direction, attr.Alias.Tenant, attr.Alias.Category, attr.Alias.Account, attr.Alias.Subject, attr.Alias.Context, utils.CACHED)
	}

	if attr.Overwrite || oldAlias == nil {
		if err := am.accountingDb.SetAlias(attr.Alias); err != nil {
			*reply = err.Error()
			return err
		}
	} else {
		for _, value := range attr.Alias.Values {
			found := false
			if value.DestinationID == "" {
				value.DestinationID = utils.ANY
			}
			for _, oldValue := range oldAlias.Values {
				if oldValue.DestinationID == value.DestinationID {
					if oldValue.Fields != value.Fields {
						oldValue.Fields = value.Fields
						oldValue.fields = nil
					}
					oldValue.Weight = value.Weight
					found = true
					break
				}
			}
			if !found {
				oldAlias.Values = append(oldAlias.Values, value)
			}
		}
		if err := am.accountingDb.SetAlias(oldAlias); err != nil {
			*reply = err.Error()
			return err
		}
	}

	*reply = utils.OK
	return nil
}

func (am *AliasHandler) RemoveAlias(al *Alias, reply *string) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	if err := am.accountingDb.RemoveAlias(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Context); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

func (am *AliasHandler) GetReverseAlias(attr *AttrReverseAlias, result *map[string][]*Alias) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	aliases := make(map[string][]*Alias)
	if als, err := am.accountingDb.GetReverseAlias(attr.Tenant, attr.Context, attr.Target, attr.Alias, utils.CACHED); err == nil {
		for _, al := range als {
			for _, av := range al.Values {
				// search for target: alias in fields
				if strings.Contains(av.Fields, fmt.Sprintf(`"%s":"%s"`, attr.Target, attr.Alias)) {
					aliases[av.DestinationID] = append(aliases[av.DestinationID], al)
					break
				}
			}

		}
	}
	*result = aliases
	return nil
}

func (am *AliasHandler) GetAlias(al *Alias, result *Alias) error {
	am.mu.RLock()
	defer am.mu.RUnlock()
	//log.Print("ALIAS: ", utils.ToIJSON(al))
	if alias, err := am.accountingDb.GetAlias(al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Context, utils.CACHED); err == nil && alias != nil {
		//log.Print("GOT: ", utils.ToIJSON(alias))
		*result = *(alias) // copy
		return nil
	}
	return utils.ErrNotFound
}

func (am *AliasHandler) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(am).MethodByName(parts[1])
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

func LoadAlias(attr *AttrAlias, in interface{}, extraFields string) error {
	if aliasService == nil { // no alias service => no fun
		return nil
	}
	response := Alias{}
	if err := aliasService.Call("AliasesV1.GetAlias", &Alias{
		Direction: attr.Direction,
		Tenant:    attr.Tenant,
		Category:  attr.Category,
		Account:   attr.Account,
		Subject:   attr.Subject,
		Context:   attr.Context,
	}, &response); err != nil {
		return err
	}

	// sort according to weight
	values := response.Values
	values.Sort()

	var rightFields *utils.StructQ
	// if destination does not metter get first alias
	if attr.Destination == "" || attr.Destination == utils.ANY {
		rightFields = values[0].getFields()
	}

	if rightFields == nil {
		// check destination ids
		if dests, err := ratingStorage.GetDestinations(attr.Tenant, attr.Destination, "", utils.DestMatching, utils.CACHED); err == nil {
			destNames := dests.getNames()
			for _, value := range values {
				if value.DestinationID == utils.ANY || destNames[value.DestinationID] {
					rightFields = value.getFields()
				}
				if rightFields != nil {
					break
				}
			}
		}
	}

	if rightFields != nil {
		if _, err := rightFields.Query(in, true); err != nil {
			return err
		}
	}
	return nil
}

type AliasList []*Alias // used in GetAlias

func (all AliasList) Sort() {
	// we need higher precision earlyer in the list
	sort.Slice(all, func(j, i int) bool {
		return all[i].precision() < all[j].precision()
	})
}
