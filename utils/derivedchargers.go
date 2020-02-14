package utils

import (
	"errors"
	"sort"
)

// Wraps regexp compiling in case of rsr fields
func NewDerivedCharger(runID, runFilter, fields string) (dc *DerivedCharger, err error) {
	if len(runID) == 0 {
		return nil, errors.New("Empty run id field")
	}
	dc = &DerivedCharger{RunID: runID}
	if runFilter != "" {
		dc.RunFilter = runFilter
		if dc.runFilter, err = NewStructQ(dc.RunFilter); err != nil {
			return nil, err
		}
	}
	if fields != "" {
		dc.Fields = fields
		if dc.fields, err = NewStructQ(dc.Fields); err != nil {
			return nil, err
		}
	}
	return dc, nil
}

type DerivedCharger struct {
	RunID     string `bson:"run_id"`      // Unique runId in the chain
	RunFilter string `bson:"run_filters"` // Only run the charger if all the filters match
	Fields    string `bson:"fields"`      // fields changer
	runFilter *StructQ
	fields    *StructQ
}

func (dc *DerivedCharger) CheckRunFilter(o interface{}) (bool, error) {
	if dc.runFilter == nil {
		var err error
		if dc.runFilter, err = NewStructQ(dc.RunFilter); err != nil {
			return false, err
		}
	}
	if dc.RunFilter == "" || dc.runFilter == nil {
		return true, nil
	}
	return dc.runFilter.Query(o, false)
}

func (dc *DerivedCharger) ChangeFields(o interface{}) error {
	if dc.fields == nil {
		var err error
		if dc.fields, err = NewStructQ(dc.Fields); err != nil {
			return err
		}
	}
	if dc.Fields == "" || dc.fields == nil {
		return nil
	}
	_, err := dc.fields.Query(o, true)
	return err
}

func (dc *DerivedCharger) Equal(other *DerivedCharger) bool {
	return dc.RunID == other.RunID &&
		dc.RunFilter == other.RunFilter &&
		dc.Fields == other.Fields
}

func DerivedChargerGroupKey(direction, tenant, category, account, subject string) string {
	return ConcatKey(direction, tenant, category, account, subject)
}

type DerivedChargerGroup struct {
	Direction      string            `bson:"direction"`
	Tenant         string            `bson:"tenant"`
	Category       string            `bson:"category"`
	Account        string            `bson:"account"`
	Subject        string            `bson:"subject"`
	DestinationIDs StringMap         `bson:"destination_ids"`
	Chargers       []*DerivedCharger `bson:"chargers"`
}

// Precheck that RunId is unique
func (dcs *DerivedChargerGroup) Append(dc *DerivedCharger) (*DerivedChargerGroup, error) {
	if dc.RunID == DEFAULT_RUNID {
		return nil, errors.New("Reserved RunId")
	}
	for _, dcLocal := range dcs.Chargers {
		if dcLocal.RunID == dc.RunID {
			return nil, errors.New("Duplicated RunId")
		}
	}
	dcs.Chargers = append(dcs.Chargers, dc)
	return dcs, nil
}

func (dcs *DerivedChargerGroup) AppendDefaultRun() (*DerivedChargerGroup, error) {
	dcDf, _ := NewDerivedCharger(DEFAULT_RUNID, "", "")
	dcs.Chargers = append(dcs.Chargers, dcDf)
	return dcs, nil
}

func (dcs *DerivedChargerGroup) Equal(other *DerivedChargerGroup) bool {
	dcs.DestinationIDs.Equal(other.DestinationIDs)
	for i, dc := range dcs.Chargers {
		if !dc.Equal(other.Chargers[i]) {
			return false
		}
	}
	return true
}

func (dcs *DerivedChargerGroup) precision() int {
	precision := 0
	if dcs.Direction != ANY {
		precision++
	}
	if dcs.Tenant != ANY {
		precision++
	}
	if dcs.Category != ANY {
		precision++
	}
	if dcs.Account != ANY {
		precision++
	}
	if dcs.Subject != ANY {
		precision++
	}
	return precision
}

type DerivedChargers []*DerivedChargerGroup // used in GetDerivedChargers

func (dcs DerivedChargers) Sort() {
	sort.Slice(dcs, func(j, i int) bool { return dcs[i].precision() < dcs[j].precision() })
}
