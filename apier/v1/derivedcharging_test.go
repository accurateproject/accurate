
package v1

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var apierDcT *ApierV1

func init() {
	dataStorage, _ := engine.NewMapStorage()
	cfg, _ := config.NewDefaultCGRConfig()
	apierDcT = &ApierV1{RatingDb: engine.RatingStorage(dataStorage), Config: cfg}
}

/*
func TestGetEmptyDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var dcs utils.DerivedChargers
	if err := apierDcT.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, apierDcT.Config.DerivedChargers) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}
*/

func TestSetDC(t *testing.T) {
	dcs1 := []*utils.DerivedCharger{
		&utils.DerivedCharger{RunID: "extra1", RequestTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunID: "extra2", RequestTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	attrs := AttrSetDerivedChargers{Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dan", DerivedChargers: dcs1}
	var reply string
	if err := apierDcT.SetDerivedChargers(attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func TestGetDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	eDcs := utils.DerivedChargers{DestinationIDs: utils.NewStringMap(),
		Chargers: []*utils.DerivedCharger{
			&utils.DerivedCharger{RunID: "extra1", RequestTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
				AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
			&utils.DerivedCharger{RunID: "extra2", RequestTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
				AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		}}
	var dcs utils.DerivedChargers
	if err := apierDcT.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, eDcs) {
		t.Errorf("Expecting: %v, received: %v", eDcs.DestinationIDs, dcs.DestinationIDs)
	}
}

func TestRemDC(t *testing.T) {
	attrs := AttrRemDerivedChargers{Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dan"}
	var reply string
	if err := apierDcT.RemDerivedChargers(attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

/*
func TestGetEmptyDC2(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var dcs utils.DerivedChargers
	if err := apierDcT.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, apierDcT.Config.DerivedChargers) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}
*/
