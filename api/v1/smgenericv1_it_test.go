// +build integration

package v1

import (
	"encoding/json"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/sessionmanager"
	"github.com/accurateproject/accurate/utils"
)

var smgV1CfgPath string
var smgV1Cfg *config.Config
var smgV1Rpc *rpc.Client
var smgV1LoadInst utils.LoadInstance // Share load information between tests

func TestSMGV1InitCfg(t *testing.T) {
	smgV1CfgPath = path.Join(*dataDir, "conf", "samples", "smgeneric")
	// Init config first
	var err error
	config.Reset()
	if err = config.LoadPath(smgV1CfgPath); err != nil {
		t.Error(err)
	}
	smgV1Cfg = config.Get()
	smgV1Cfg.General.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

// Remove data in both rating and accounting db
func TestSMGV1ResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(smgV1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSMGV1ResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(smgV1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGV1StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(smgV1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGV1RpcConn(t *testing.T) {
	var err error
	smgV1Rpc, err = jsonrpc.Dial("tcp", *smgV1Cfg.Listen.RpcJson) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGV1LoadTariffPlanFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := smgV1Rpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &smgV1LoadInst); err != nil {
		t.Error(err)
	} else if smgV1LoadInst.RatingLoadID != "" && smgV1LoadInst.AccountingLoadID != "" {
		t.Error("Non Empty loadId received, loadInstance: ", smgV1LoadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestSMGV1CacheStats(t *testing.T) {
	var rcvStats *utils.CacheStats

	expectedStats := &utils.CacheStats{Destinations: 0, RatingPlans: 4, RatingProfiles: 0, Actions: 7, ActionPlans: 4, SharedGroups: 0, Aliases: 0, ResourceLimits: 0,
		DerivedChargers: 0, LcrProfiles: 0, CdrStats: 6, Users: 3}
	var args utils.AttrCacheStats
	if err := smgV1Rpc.Call("ApierV2.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV2.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV2.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
	}
}

// Make sure account was debited properly
func TestSMGV1AccountsBefore(t *testing.T) {
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := smgV1Rpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		jsn, _ := json.Marshal(reply)
		t.Errorf("Calling ApierV2.GetBalance received: %s", jsn)
	}
}

// Make sure account was debited properly
func TestSMGV1GetMaxUsage(t *testing.T) {
	setupReq := &sessionmanager.SMGenericEvent{utils.REQTYPE: utils.META_PREPAID, utils.TENANT: "cgrates.org",
		utils.ACCOUNT: "1003", utils.DESTINATION: "1002", utils.SETUP_TIME: "2015-11-10T15:20:00Z"}
	var maxTime float64
	if err := smgV1Rpc.Call("SMGenericV1.MaxUsage", setupReq, &maxTime); err != nil {
		t.Error(err)
	} else if maxTime != 2700 {
		t.Errorf("Calling ApierV2.MaxUsage got maxTime: %f", maxTime)
	}
}

func TestSMGV1StopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
