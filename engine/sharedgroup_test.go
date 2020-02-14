package engine

import (
	"reflect"
	"testing"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

func TestSharedSetGet(t *testing.T) {
	tenant := "tenant13"
	name := "TEST_SG100"
	sg := &SharedGroup{
		Tenant: tenant,
		Name:   name,
		AccountParameters: map[string]*SharingParam{
			"test": &SharingParam{Strategy: STRATEGY_HIGHEST},
		},
		MemberIDs: utils.NewStringMap("1", "2", "3"),
	}
	err := ratingStorage.SetSharedGroup(sg)
	if err != nil {
		t.Error("Error storing Shared groudp: ", err)
	}
	received, err := ratingStorage.GetSharedGroup(tenant, name, utils.CACHE_SKIP)
	if err != nil || received == nil || !reflect.DeepEqual(sg, received) {
		t.Error("Error getting shared group: ", err, received)
	}
	received, err = ratingStorage.GetSharedGroup(tenant, name, utils.CACHED)
	if err != nil || received == nil || !reflect.DeepEqual(sg, received) {
		t.Error("Error getting cached shared group: ", err, received)
	}

}

func TestSharedPopBalanceByStrategyLow(t *testing.T) {
	bc := Balances{
		&Balance{Value: dec.NewFloat(2.0)},
		&Balance{UUID: "uuuu", Value: dec.NewFloat(1.0), account: &Account{Tenant: "t1", Name: "test"}},
		&Balance{Value: dec.NewFloat(3.0)},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParam{
		"test": &SharingParam{Strategy: STRATEGY_LOWEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[1], bc)
	if len(sbc) != 3 ||
		sbc[0].Value.String() != "1" ||
		sbc[1].Value.String() != "2" {
		t.Error("Error sorting balance chain: ", sbc[0].GetValue())
	}
}

func TestSharedPopBalanceByStrategyHigh(t *testing.T) {
	bc := Balances{
		&Balance{UUID: "uuuu", Value: dec.NewFloat(2.0), account: &Account{Tenant: "t1", Name: "test"}},
		&Balance{Value: dec.NewFloat(1.0)},
		&Balance{Value: dec.NewFloat(3.0)},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParam{
		"test": &SharingParam{Strategy: STRATEGY_HIGHEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	if len(sbc) != 3 ||
		sbc[0].Value.String() != "3" ||
		sbc[1].Value.String() != "2" {
		t.Error("Error sorting balance chain: ", sbc)
	}
}

func TestSharedPopBalanceByStrategyMineHigh(t *testing.T) {
	bc := Balances{
		&Balance{UUID: "uuuu", Value: dec.NewFloat(2.0), account: &Account{Tenant: "t1", Name: "test"}},
		&Balance{Value: dec.NewFloat(1.0)},
		&Balance{Value: dec.NewFloat(3.0)},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParam{
		"test": &SharingParam{Strategy: STRATEGY_MINE_HIGHEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	if len(sbc) != 3 ||
		sbc[0].Value.String() != "2" ||
		sbc[1].Value.String() != "3" {
		t.Error("Error sorting balance chain: ", sbc)
	}
}

/*func TestSharedPopBalanceByStrategyRandomHigh(t *testing.T) {
	bc := Balances{
		&Balance{UUID: "uuuu", Value: 2.0, account: &Account{Id: "test"}},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParam{
		"test": &SharingParam{Strategy: STRATEGY_RANDOM}},
	}
	x := bc[0]
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	firstTest := (sbc[0].UUID == x.UUID)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	secondTest := (sbc[0].UUID == x.UUID)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	thirdTest := (sbc[0].UUID == x.UUID)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	fourthTest := (sbc[0].UUID == x.UUID)
	if firstTest && secondTest && thirdTest && fourthTest {
		t.Error("Something is wrong with balance randomizer")
	}
}*/
