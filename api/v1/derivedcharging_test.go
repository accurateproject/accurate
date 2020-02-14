package v1

import (
	"testing"

	"github.com/accurateproject/accurate/utils"
)

/*
func TestGetEmptyDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var dcs utils.DerivedChargers
	if err := apiObject.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, apiObject.Config.DerivedChargers) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}
*/
/*
func TestSetDC(t *testing.T) {
	dcs1 := []*utils.DerivedCharger{
		&utils.DerivedCharger{RunID: "extra1", Fields: `{"RequestType":"prepaid", "Account":"rif", "Subject":"rif"}`},
		&utils.DerivedCharger{RunID: "extra2", Fields: `{"Account":"ivo", "Subject":"ivo"}`},
	}
	attrs := AttrSetDerivedChargers{Direction: "*out", Tenant: "test", Category: "call", Account: "dan", Subject: "dan", DerivedChargers: dcs1}
	var reply string
	if err := apiObject.SetDerivedChargers(attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}*/

/*
func TestGetDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "test", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	eDcs := []*utils.DerivedCharger{
		&utils.DerivedCharger{RunID: "extra1", Fields: `{"RequestType":"prepaid", "Account":"rif", "Subject":"rif"}`},
		&utils.DerivedCharger{RunID: "extra2", Fields: `{"Account":"ivo", "Subject":"ivo"}`},
	}
	var dcs utils.DerivedChargerGroup
	if err := apiObject.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, eDcs) {
		t.Errorf("Expecting: %+v, received: %+v", eDcs, dcs)
	}
}*/

func TestRemDC(t *testing.T) {
	attrs := AttrRemDerivedChargers{Direction: "*out", Tenant: "test", Category: "call", Account: "dan", Subject: "dan"}
	var reply string
	if err := apiObject.RemDerivedChargers(attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

/*
func TestGetEmptyDC2(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "test", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var dcs utils.DerivedChargers
	if err := apiObject.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, apiObject.Config.DerivedChargers) {
		t.Error("Returned DerivedChargers not matching the configured ones")
	}
}
*/
