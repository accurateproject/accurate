package engine

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
)

var rds *RedisStorage
var err error

func TestConnectRedis(t *testing.T) {
	if !*testLocal {
		return
	}
	cfg, _ := config.NewDefaultCGRConfig()
	rds, err = NewRedisStorage(fmt.Sprintf("%s:%s", cfg.TpDbHost, cfg.TpDbPort), 4, cfg.TpDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
}

func TestFlush(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := rds.Flush(""); err != nil {
		t.Error("Failed to Flush redis database", err.Error())
	}
	rds.PreloadRatingCache()
}

func TestSetGetDerivedCharges(t *testing.T) {
	if !*testLocal {
		return
	}
	keyCharger1 := utils.ConcatenatedKey("*out", "cgrates.org", "call", "dan", "dan")
	charger1 := &utils.DerivedChargers{Chargers: []*utils.DerivedCharger{
		&utils.DerivedCharger{RunID: "extra1", RequestTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunID: "extra2", RequestTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}}
	if err := rds.SetDerivedChargers(keyCharger1, charger1); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	// Retrieve from db
	if rcvCharger, err := rds.GetDerivedChargers(keyCharger1, utils.CACHE_SKIP); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
	// Retrieve from cache
	if rcvCharger, err := rds.GetDerivedChargers(keyCharger1, utils.CACHED); err != nil {
		t.Error("Error when retrieving DerivedCHarger", err.Error())
	} else if !reflect.DeepEqual(rcvCharger, charger1) {
		t.Errorf("Expecting %v, received: %v", charger1, rcvCharger)
	}
}
