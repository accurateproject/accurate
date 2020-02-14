package engine

import (
	"testing"

	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/utils"
)

func TestAliasesGetAlias(t *testing.T) {
	alias := Alias{}
	err := aliasService.Call("AliasesV1.GetAlias", &Alias{
		Direction: "*out",
		Tenant:    "test",
		Category:  "call",
		Account:   "dan",
		Subject:   "dan",
		Context:   "*rating",
	}, &alias)
	if err != nil ||
		len(alias.Values) != 2 ||
		len(alias.Values[0].Fields) != 76 {
		t.Error("Error getting alias: ", err, alias, alias.Values[0].Fields)
	}
}

func TestAliasesSetters(t *testing.T) {
	var out string
	if err := aliasService.Call("AliasesV1.SetAlias", &AttrAddAlias{
		Alias: &Alias{
			Direction: "*out",
			Tenant:    "test",
			Category:  "call",
			Account:   "set",
			Subject:   "set",
			Context:   "*rating",
			Values: AliasValues{&AliasValue{
				DestinationID: utils.ANY,
				Fields:        `{"Account":{"$rpl":["1234", "1235"]}}`,
				Weight:        10,
			}},
		},
		Overwrite: true,
	}, &out); err != nil || out != utils.OK {
		t.Error("Error setting alias: ", err, out)
	}
	r := &Alias{}
	if err := aliasService.Call("AliasesV1.GetAlias", &Alias{
		Direction: "*out",
		Tenant:    "test",
		Category:  "call",
		Account:   "set",
		Subject:   "set",
		Context:   "*rating",
	}, r); err != nil || len(r.Values) != 1 || len(r.Values[0].Fields) != 37 {
		t.Errorf("Error getting alias: %+v", r.Values[0].Fields)
	}

	if err := aliasService.Call("AliasesV1.SetAlias", &AttrAddAlias{
		Alias: &Alias{
			Direction: "*out",
			Tenant:    "test",
			Category:  "call",
			Account:   "set",
			Subject:   "set",
			Context:   "*rating",
			Values: AliasValues{&AliasValue{
				DestinationID: "NAT",
				Fields:        `{"Subject":{"$rpl":["1234", "1235"]}}`,
				Weight:        10,
			}},
		},
		Overwrite: false,
	}, &out); err != nil || out != utils.OK {
		t.Error("Error updateing alias: ", err, out)
	}
	if err := aliasService.Call("AliasesV1.GetAlias", &Alias{
		Direction: "*out",
		Tenant:    "test",
		Category:  "call",
		Account:   "set",
		Subject:   "set",
		Context:   "*rating",
	}, r); err != nil ||
		len(r.Values) != 2 {
		t.Errorf("Error getting alias: %s", utils.ToIJSON(r.Values))
	}
	if err := aliasService.Call("AliasesV1.SetAlias", &AttrAddAlias{
		Alias: &Alias{
			Direction: "*out",
			Tenant:    "test",
			Category:  "call",
			Account:   "set",
			Subject:   "set",
			Context:   "*rating",
			Values: AliasValues{&AliasValue{
				DestinationID: "NAT",
				Fields:        `{"Subject":{"$rpl":["1111", "2222"]}}`,
				Weight:        10,
			}},
		},
		Overwrite: false,
	}, &out); err != nil || out != utils.OK {
		t.Error("Error updateing alias: ", err, out)
	}
	if err := aliasService.Call("AliasesV1.GetAlias", &Alias{
		Direction: "*out",
		Tenant:    "test",
		Category:  "call",
		Account:   "set",
		Subject:   "set",
		Context:   "*rating",
	}, r); err != nil || len(r.Values) != 2 {
		t.Errorf("Error getting alias: %+v", r.Values[0].Fields)
	}
	if err := aliasService.Call("AliasesV1.SetAlias", &AttrAddAlias{
		Alias: &Alias{
			Direction: "*out",
			Tenant:    "test",
			Category:  "call",
			Account:   "set",
			Subject:   "set",
			Context:   "*rating",
			Values: AliasValues{&AliasValue{
				DestinationID: "RET",
				Fields:        `{"Subject":{"$rpl":["3333", "4444"]}}`,
				Weight:        10,
			}},
		},
		Overwrite: false,
	}, &out); err != nil || out != utils.OK {
		t.Error("Error updateing alias: ", err, out)
	}
	if err := aliasService.Call("AliasesV1.GetAlias", &Alias{
		Direction: "*out",
		Tenant:    "test",
		Category:  "call",
		Account:   "set",
		Subject:   "set",
		Context:   "*rating",
	}, r); err != nil ||
		len(r.Values) != 3 {
		t.Errorf("Error getting alias: %s", utils.ToIJSON(r.Values))
	}
}

func TestAliasesLoadAlias(t *testing.T) {
	var response string
	cd := &CallDescriptor{
		Direction:   "*out",
		Tenant:      "test",
		Category:    "call",
		Account:     "rif",
		Subject:     "rif",
		Destination: "444",
		ExtraFields: map[string]string{
			"Cli":   "0723",
			"Other": "stuff",
		},
	}
	err := LoadAlias(
		&AttrAlias{
			Direction:   "*out",
			Tenant:      "test",
			Category:    "call",
			Account:     "dan",
			Subject:     "dan",
			Context:     "*rating",
			Destination: "444",
		}, cd, "ExtraFields")
	if err != nil || cd == nil {
		t.Error("Error getting alias: ", err, response)
	}
	if cd.Subject != "rif1" ||
		cd.ExtraFields["Cli"] != "0724" {
		t.Errorf("Aliases failed to change interface: %+v", cd)
	}
}

func TestAliasesCache(t *testing.T) {
	_, err := accountingStorage.GetAlias(utils.OUT, "test", "call", "remo", "remo", utils.ALIAS_CONTEXT_RATING, utils.CACHED)
	if err != nil {
		t.Error("Error getting alias: ", err)
	}
	key := "*out:call:remo:remo:*rating"
	a, found := cache2go.Get("test", utils.ALIASES_PREFIX+key)
	if !found || a == nil {
		t.Error("Error getting alias from cache: ", err, a)
	}
}
