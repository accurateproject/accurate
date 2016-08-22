
package engine

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"reflect"

	"github.com/cgrates/cgrates/utils"
	"github.com/ugorji/go/codec"
	"gopkg.in/mgo.v2/bson"
)

type Storage interface {
	Close()
	Flush(string) error
	GetKeysForPrefix(string) ([]string, error)
	PreloadCacheForPrefix(string) error
	RebuildReverseForPrefix(string) error
}

// Interface for storage providers.
type RatingStorage interface {
	Storage
	HasData(string, string) (bool, error)
	PreloadRatingCache() error
	GetRatingPlan(string, bool) (*RatingPlan, error)
	SetRatingPlan(*RatingPlan) error
	GetRatingProfile(string, bool) (*RatingProfile, error)
	SetRatingProfile(*RatingProfile) error
	RemoveRatingProfile(string) error
	GetDestination(string, bool) (*Destination, error)
	SetDestination(*Destination) error
	RemoveDestination(string) error
	SetReverseDestination(*Destination) error
	GetReverseDestination(string, bool) ([]string, error)
	UpdateReverseDestination(*Destination, *Destination) error
	GetLCR(string, bool) (*LCR, error)
	SetLCR(*LCR) error
	SetCdrStats(*CdrStats) error
	GetCdrStats(string) (*CdrStats, error)
	GetAllCdrStats() ([]*CdrStats, error)
	GetDerivedChargers(string, bool) (*utils.DerivedChargers, error)
	SetDerivedChargers(string, *utils.DerivedChargers) error
	GetActions(string, bool) (Actions, error)
	SetActions(string, Actions) error
	RemoveActions(string) error
	GetSharedGroup(string, bool) (*SharedGroup, error)
	SetSharedGroup(*SharedGroup) error
	GetActionTriggers(string, bool) (ActionTriggers, error)
	SetActionTriggers(string, ActionTriggers) error
	RemoveActionTriggers(string) error
	GetActionPlan(string, bool) (*ActionPlan, error)
	SetActionPlan(string, *ActionPlan, bool) error
	GetAllActionPlans() (map[string]*ActionPlan, error)
	PushTask(*Task) error
	PopTask() (*Task, error)
}

type AccountingStorage interface {
	Storage
	PreloadAccountingCache() error
	GetAccount(string) (*Account, error)
	SetAccount(*Account) error
	RemoveAccount(string) error
	GetCdrStatsQueue(string) (*StatsQueue, error)
	SetCdrStatsQueue(*StatsQueue) error
	GetSubscribers() (map[string]*SubscriberData, error)
	SetSubscriber(string, *SubscriberData) error
	RemoveSubscriber(string) error
	SetUser(*UserProfile) error
	GetUser(string) (*UserProfile, error)
	GetUsers() ([]*UserProfile, error)
	RemoveUser(string) error
	SetAlias(*Alias) error
	GetAlias(string, bool) (*Alias, error)
	RemoveAlias(string) error
	SetReverseAlias(*Alias) error
	GetReverseAlias(string, bool) ([]string, error)
	UpdateReverseAlias(*Alias, *Alias) error
	GetResourceLimit(string, bool) (*ResourceLimit, error)
	SetResourceLimit(*ResourceLimit) error
	RemoveResourceLimit(string) error
	GetLoadHistory(int, bool) ([]*utils.LoadInstance, error)
	AddLoadHistory(*utils.LoadInstance, int) error
	GetStructVersion() (*StructVersion, error)
	SetStructVersion(*StructVersion) error
}

type CdrStorage interface {
	Storage
	SetCDR(*CDR, bool) error
	SetSMCost(smc *SMCost) error
	GetSMCosts(cgrid, runid, originHost, originIDPrfx string) ([]*SMCost, error)
	GetCDRs(*utils.CDRsFilter, bool) ([]*CDR, int64, error)
}

type LoadStorage interface {
	Storage
	LoadReader
	LoadWriter
}

// LoadReader reads from .csv or TP tables and provides the data ready for the tp_db or data_db.
type LoadReader interface {
	GetTpIds() ([]string, error)
	GetTpTableIds(string, string, utils.TPDistinctIds, map[string]string, *utils.Paginator) ([]string, error)
	GetTpTimings(string, string) ([]TpTiming, error)
	GetTpDestinations(string, string) ([]TpDestination, error)
	GetTpRates(string, string) ([]TpRate, error)
	GetTpDestinationRates(string, string, *utils.Paginator) ([]TpDestinationRate, error)
	GetTpRatingPlans(string, string, *utils.Paginator) ([]TpRatingPlan, error)
	GetTpRatingProfiles(*TpRatingProfile) ([]TpRatingProfile, error)
	GetTpSharedGroups(string, string) ([]TpSharedGroup, error)
	GetTpCdrStats(string, string) ([]TpCdrstat, error)
	GetTpLCRs(*TpLcrRule) ([]TpLcrRule, error)
	GetTpUsers(*TpUser) ([]TpUser, error)
	GetTpAliases(*TpAlias) ([]TpAlias, error)
	GetTpDerivedChargers(*TpDerivedCharger) ([]TpDerivedCharger, error)
	GetTpActions(string, string) ([]TpAction, error)
	GetTpActionPlans(string, string) ([]TpActionPlan, error)
	GetTpActionTriggers(string, string) ([]TpActionTrigger, error)
	GetTpAccountActions(*TpAccountAction) ([]TpAccountAction, error)
	GetTpResourceLimits(string, string) (TpResourceLimits, error)
}

type LoadWriter interface {
	RemTpData(string, string, map[string]string) error
	SetTpTimings([]TpTiming) error
	SetTpDestinations([]TpDestination) error
	SetTpRates([]TpRate) error
	SetTpDestinationRates([]TpDestinationRate) error
	SetTpRatingPlans([]TpRatingPlan) error
	SetTpRatingProfiles([]TpRatingProfile) error
	SetTpSharedGroups([]TpSharedGroup) error
	SetTpCdrStats([]TpCdrstat) error
	SetTpUsers([]TpUser) error
	SetTpAliases([]TpAlias) error
	SetTpDerivedChargers([]TpDerivedCharger) error
	SetTpLCRs([]TpLcrRule) error
	SetTpActions([]TpAction) error
	SetTpActionPlans([]TpActionPlan) error
	SetTpActionTriggers([]TpActionTrigger) error
	SetTpAccountActions([]TpAccountAction) error
}

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type JSONMarshaler struct{}

func (jm *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jm *JSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type BSONMarshaler struct{}

func (jm *BSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return bson.Marshal(v)
}

func (jm *BSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return bson.Unmarshal(data, v)
}

type JSONBufMarshaler struct{}

func (jbm *JSONBufMarshaler) Marshal(v interface{}) (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(v)
	data = buf.Bytes()
	return
}

func (jbm *JSONBufMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}

type CodecMsgpackMarshaler struct {
	mh *codec.MsgpackHandle
}

func NewCodecMsgpackMarshaler() *CodecMsgpackMarshaler {
	cmm := &CodecMsgpackMarshaler{new(codec.MsgpackHandle)}
	mh := cmm.mh
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mh.RawToString = true
	return cmm
}

func (cmm *CodecMsgpackMarshaler) Marshal(v interface{}) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, cmm.mh)
	err = enc.Encode(v)
	return
}

func (cmm *CodecMsgpackMarshaler) Unmarshal(data []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(data, cmm.mh)
	return dec.Decode(&v)
}

type BincMarshaler struct {
	bh *codec.BincHandle
}

func NewBincMarshaler() *BincMarshaler {
	return &BincMarshaler{new(codec.BincHandle)}
}

func (bm *BincMarshaler) Marshal(v interface{}) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, bm.bh)
	err = enc.Encode(v)
	return
}

func (bm *BincMarshaler) Unmarshal(data []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(data, bm.bh)
	return dec.Decode(&v)
}

type GOBMarshaler struct{}

func (gm *GOBMarshaler) Marshal(v interface{}) (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = gob.NewEncoder(buf).Encode(v)
	data = buf.Bytes()
	return
}

func (gm *GOBMarshaler) Unmarshal(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(v)
}
