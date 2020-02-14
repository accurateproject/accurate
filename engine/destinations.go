package engine

import (
	"encoding/json"

	"github.com/accurateproject/accurate/history"
	"github.com/accurateproject/accurate/utils"
)

/*
Structure that gathers multiple destination prefixes under a common id.
*/
type Destination struct {
	Tenant string `bson:"tenant"`
	Code   string `bson:"code"`
	Name   string `bson:"name"`
}

// history record method
func (d *Destination) GetHistoryRecord(deleted bool) history.Record {
	js, _ := json.Marshal(d)
	return history.Record{
		Id:       d.Name,
		Filename: history.DESTINATIONS_FN,
		Payload:  js,
		Deleted:  deleted,
	}
}

func (d *Destination) CacheKey() string {
	return utils.ConcatKey(d.Tenant, d.Code)
}

// Reverse search in cache to see if prefix belongs to destination id
func CachedDestHasPrefix(tenant, name, code string) bool {
	if dests, err := ratingStorage.GetDestinations(tenant, code, name, utils.DestExact, utils.CACHED); err == nil {
		return len(dests) > 0
	}
	return false
}

type Destinations []*Destination

func (ds Destinations) getNames() utils.StringMap {
	result := utils.StringMap{}
	for _, d := range ds {
		result[d.Name] = true
	}
	return result
}

func (ds Destinations) getBest() *Destination {
	if len(ds) > 0 {
		return ds[0]
	}
	return nil
}
