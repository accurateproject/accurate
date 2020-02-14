package general_tests

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

var tutSMGCfgPath string
var tutSMGCfg *config.Config
var tutSMGRpc *rpc.Client
var smgLoadInst utils.LoadInstance // Share load information between tests

func TestTutSMGInitCfg(t *testing.T) {
	if !*testLocal {
		return
	}
	tutSMGCfgPath = path.Join(*dataDir, "conf", "samples", "smgeneric")
	// Init config first
	var err error
	config.Reset()
	if err = config.LoadPath(tutSMGCfgPath); err != nil {
		t.Error(err)
	}
	tutSMGCfg := config.Get()
	tutSMGCfg.General.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

// Remove data in both rating and accounting db
func TestTutSMGResetDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTutSMGResetStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitStorDb(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTutSMGStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := engine.StopStartEngine(tutSMGCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTutSMGRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	tutSMGRpc, err = jsonrpc.Dial("tcp", *tutSMGCfg.Listen.RpcJson) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTutSMGLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutSMGRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &smgLoadInst); err != nil {
		t.Error(err)
	} else if smgLoadInst.RatingLoadID == "" || smgLoadInst.AccountingLoadID == "" {
		t.Error("Empty loadId received, loadInstance: ", smgLoadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestTutSMGCacheStats(t *testing.T) {
	if !*testLocal {
		return
	}
	var rcvStats *utils.CacheStats

	expectedStats := &utils.CacheStats{Destinations: 7, RatingPlans: 4, RatingProfiles: 9, Actions: 8, ActionPlans: 4, SharedGroups: 1, Aliases: 1, ResourceLimits: 0,
		DerivedChargers: 1, LcrProfiles: 5, CdrStats: 6, Users: 3}
	var args utils.AttrCacheStats
	if err := tutSMGRpc.Call("ApierV2.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV2.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV2.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
	}
}

/*
// Make sure account was debited properly
func TestTutSMGAccountsBefore(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		jsn, _ := json.Marshal(reply)
		t.Errorf("Calling ApierV2.GetBalance received: %s", jsn)
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1003"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1004"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1007"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 0.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1005"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err == nil || !strings.HasSuffix(err.Error(), "does not exist") {
		t.Error("Got error on ApierV2.GetAccount: %v", err)
	}
}

*/

func TestTutSMGStopCgrEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
