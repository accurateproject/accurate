
package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSharedSetGet(t *testing.T) {
	id := "TEST_SG100"
	sg := &SharedGroup{
		Id: id,
		AccountParameters: map[string]*SharingParameters{
			"test": &SharingParameters{Strategy: STRATEGY_HIGHEST},
		},
		MemberIds: utils.NewStringMap("1", "2", "3"),
	}
	err := ratingStorage.SetSharedGroup(sg)
	if err != nil {
		t.Error("Error storing Shared groudp: ", err)
	}
	received, err := ratingStorage.GetSharedGroup(id, true)
	if err != nil || received == nil || !reflect.DeepEqual(sg, received) {
		t.Error("Error getting shared group: ", err, received)
	}
	received, err = ratingStorage.GetSharedGroup(id, false)
	if err != nil || received == nil || !reflect.DeepEqual(sg, received) {
		t.Error("Error getting cached shared group: ", err, received)
	}

}

func TestSharedPopBalanceByStrategyLow(t *testing.T) {
	bc := Balances{
		&Balance{Value: 2.0},
		&Balance{Uuid: "uuuu", Value: 1.0, account: &Account{ID: "test"}},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParameters{
		"test": &SharingParameters{Strategy: STRATEGY_LOWEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[1], bc)
	if len(sbc) != 3 ||
		sbc[0].Value != 1.0 ||
		sbc[1].Value != 2.0 {
		t.Error("Error sorting balance chain: ", sbc[0].GetValue())
	}
}

func TestSharedPopBalanceByStrategyHigh(t *testing.T) {
	bc := Balances{
		&Balance{Uuid: "uuuu", Value: 2.0, account: &Account{ID: "test"}},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParameters{
		"test": &SharingParameters{Strategy: STRATEGY_HIGHEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	if len(sbc) != 3 ||
		sbc[0].Value != 3.0 ||
		sbc[1].Value != 2.0 {
		t.Error("Error sorting balance chain: ", sbc)
	}
}

func TestSharedPopBalanceByStrategyMineHigh(t *testing.T) {
	bc := Balances{
		&Balance{Uuid: "uuuu", Value: 2.0, account: &Account{ID: "test"}},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParameters{
		"test": &SharingParameters{Strategy: STRATEGY_MINE_HIGHEST}},
	}
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	if len(sbc) != 3 ||
		sbc[0].Value != 2.0 ||
		sbc[1].Value != 3.0 {
		t.Error("Error sorting balance chain: ", sbc)
	}
}

/*func TestSharedPopBalanceByStrategyRandomHigh(t *testing.T) {
	bc := Balances{
		&Balance{Uuid: "uuuu", Value: 2.0, account: &Account{Id: "test"}},
		&Balance{Value: 1.0},
		&Balance{Value: 3.0},
	}
	sg := &SharedGroup{AccountParameters: map[string]*SharingParameters{
		"test": &SharingParameters{Strategy: STRATEGY_RANDOM}},
	}
	x := bc[0]
	sbc := sg.SortBalancesByStrategy(bc[0], bc)
	firstTest := (sbc[0].Uuid == x.Uuid)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	secondTest := (sbc[0].Uuid == x.Uuid)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	thirdTest := (sbc[0].Uuid == x.Uuid)
	sbc = sg.SortBalancesByStrategy(bc[0], bc)
	fourthTest := (sbc[0].Uuid == x.Uuid)
	if firstTest && secondTest && thirdTest && fourthTest {
		t.Error("Something is wrong with balance randomizer")
	}
}*/
