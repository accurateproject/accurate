package utils

import (
	"reflect"
	"testing"
)

func TestAppendDerivedChargerGroup(t *testing.T) {
	var err error

	dcs := &DerivedChargerGroup{Chargers: make([]*DerivedCharger, 0)}
	if _, err := dcs.Append(&DerivedCharger{RunID: DEFAULT_RUNID}); err == nil {
		t.Error("Failed to detect using of the default RunID")
	}
	if dcs, err = dcs.Append(&DerivedCharger{RunID: "FIRST_RunID"}); err != nil {
		t.Error("Failed to add RunID")
	} else if len(dcs.Chargers) != 1 {
		t.Error("Unexpected number of items inside DerivedChargerGroup configuration", len(dcs.Chargers))
	}
	if dcs, err = dcs.Append(&DerivedCharger{RunID: "SECOND_RunID"}); err != nil {
		t.Error("Failed to add RunID")
	} else if len(dcs.Chargers) != 2 {
		t.Error("Unexpected number of items inside DerivedChargerGroup configuration", len(dcs.Chargers))
	}
	if _, err := dcs.Append(&DerivedCharger{RunID: "SECOND_RunID"}); err == nil {
		t.Error("Failed to detect duplicate RunID")
	}
}

func TestDerivedChargerGroupKey(t *testing.T) {
	if dcKey := DerivedChargerGroupKey("*out", "cgrates.org", "call", "dan", "dan"); dcKey != "*out:cgrates.org:call:dan:dan" {
		t.Error("Unexpected derived chargers key: ", dcKey)
	}
}

func TestAppendDefaultRun(t *testing.T) {
	dc1 := &DerivedChargerGroup{}
	dcDf := &DerivedCharger{RunID: DEFAULT_RUNID, RunFilter: "", Fields: ""}
	eDc1 := &DerivedChargerGroup{Chargers: []*DerivedCharger{dcDf}}
	if dc1, _ = dc1.AppendDefaultRun(); !reflect.DeepEqual(dc1, eDc1) {
		t.Errorf("Expecting: %+v, received: %+v", eDc1.Chargers[0], dc1.Chargers[0])
	}
	dc2 := &DerivedChargerGroup{Chargers: []*DerivedCharger{
		&DerivedCharger{RunID: "extra1", RunFilter: "", Fields: `{"RequestType":reqtype2", "Account":"rif"}, "Subject":{"$set":"rif"}}`},
		&DerivedCharger{RunID: "extra2", RunFilter: "", Fields: `{"Account":ivo", "Subject":ivo"}`},
	}}
	eDc2 := &DerivedChargerGroup{}
	eDc2.Chargers = append(dc2.Chargers, dcDf)
	if dc2, _ = dc2.AppendDefaultRun(); !reflect.DeepEqual(dc2, eDc2) {
		t.Errorf("Expecting: %+v, received: %+v", eDc2.Chargers, dc2.Chargers)
	}
}
