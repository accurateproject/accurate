// +build integration

package general_tests

import (
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/accurateproject/accurate/apier/v1"
	"github.com/accurateproject/accurate/apier/v2"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

var tpCfgPath string
var tpCfg *config.Config
var tpRPC *rpc.Client
var tpLoadInst utils.LoadInstance // Share load information between tests

func TestTpInitCfg(t *testing.T) {
	tpCfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	// Init config first
	var err error
	config.Reset()
	if err = config.LoadPath(tpCfgPath); err != nil {
		t.Error(err)
	}
	tpCfg = config.Get()
	tpCfg.General.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

// Remove data in both rating and accounting db
func TestTpResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTpResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTpStartEngine(t *testing.T) {
	fmt.Printf("Before starting: %v\n", time.Now())
	if _, err := engine.StopStartEngine(tpCfgPath, 1000); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("After starting: %v\n", time.Now())
}

// Connect rpc client to rater
func TestTpRpcConn(t *testing.T) {
	var err error
	tpRPC, err = jsonrpc.Dial("tcp", *tpCfg.Listen.RpcJson) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTpLoadTariffPlanFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	if err := tpRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &tpLoadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestTpBalanceCounter(t *testing.T) {
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Destination:   "+49",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(20) * time.Second),
	}
	var cc engine.CallCost
	if err := tpRPC.Call("Responder.Debit", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if acnt.UnitCounters[utils.MONETARY][1].Counters[0].Value != 20.0 {
		t.Errorf("Calling ApierV2.GetBalance received: %s", utils.ToIJSON(acnt))
	}
}

func TestTpActionTriggers(t *testing.T) {
	var atrs engine.ActionTriggers
	if err := tpRPC.Call("ApierV1.GetActionTriggers", v1.AttrGetActionTriggers{GroupIDs: []string{}}, &atrs); err != nil {
		t.Error("Got error on ApierV1.GetActionTriggers: ", err.Error())
	} else if len(atrs) != 9 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	}
	var reply string
	if err := tpRPC.Call("ApierV1.SetActionTrigger", v1.AttrSetActionTrigger{
		GroupID:   "TestATR",
		UniqueID:  "Unique atr id",
		BalanceID: utils.StringPointer("BID1"),
	}, &reply); err != nil {
		t.Error("Got error on ApierV1.SetActionTrigger: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling v1.SetActionTrigger got: %v", reply)
	}

	if err := tpRPC.Call("ApierV1.GetActionTriggers", v1.AttrGetActionTriggers{GroupIDs: []string{}}, &atrs); err != nil {
		t.Error("Got error on ApierV1.GetActionTriggers: ", err.Error())
	} else if len(atrs) != 10 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	}
	if err := tpRPC.Call("ApierV1.GetActionTriggers", v1.AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err != nil {
		t.Error("Got error on ApierV1.GetActionTriggers: ", err.Error())
	} else if len(atrs) != 1 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	}
	if atrs[0].ID != "TestATR" ||
		atrs[0].UniqueID != "Unique atr id" ||
		*atrs[0].Balance.ID != "BID1" {
		t.Error("Wrong action trigger set: ", utils.ToIJSON(atrs[0]))
	}
}

func TestTpZeroCost(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1012"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	}
	balanceValueBefore := acnt.BalanceMap[utils.MONETARY][0].Value
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "free",
		Account:       "1012",
		Destination:   "+49",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(20) * time.Second),
	}
	var cc engine.CallCost
	if err := tpRPC.Call("Responder.Debit", cd, &cc); err != nil {
		t.Error("Got error on Responder.Debit: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", utils.ToIJSON(cc))
	}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if acnt.BalanceMap[utils.MONETARY][0].Value != balanceValueBefore {
		t.Errorf("Calling ApierV2.GetAccount received: %s", utils.ToIJSON(acnt))
	}
}

func TestTpZeroNegativeCost(t *testing.T) {
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "free",
		Account:       "1013",
		Destination:   "+4915",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(20) * time.Second),
	}
	var cc engine.CallCost
	if err := tpRPC.Call("Responder.Debit", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", utils.ToIJSON(cc))
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1013"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if acnt.BalanceMap[utils.VOICE][0].Value != 100.0 {
		t.Errorf("Calling ApierV2.GetAccount received: %s", utils.ToIJSON(acnt))
	}
}

func TestTpExecuteActionCgrRpc(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{ActionsId: "RPC"}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "rpc"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	}
}

func TestTpExecuteActionCgrRpcAcc(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1016",
		ActionsId: "RPC_DEST",
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var dests []*engine.Destination
	attrs := &v2.AttrGetDestinations{DestinationIDs: []string{"1016"}}
	if err := tpRPC.Call("ApierV2.GetDestinations", attrs, &dests); err != nil {
		t.Error("Got error on ApierV2.GetDestinations: ", err.Error())
	}
}

func TestTpExecuteActionCgrRpcCdrStats(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{
		ActionsId: "RPC_CDRSTATS",
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var queue engine.StatsQueue
	if err := tpRPC.Call("CDRStatsV1.GetQueue", "qtest", &queue); err != nil {
		t.Error("Got error on CDRStatsV1.GetQueue: ", err.Error())
	}
}

func TestTpCreateExecuteActionMatch(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.SetActions", utils.AttrSetActions{
		ActionsId: "PAYMENT_2056bd2fe137082970f97102b64e42fd",
		Actions: []*utils.TPAction{
			&utils.TPAction{
				BalanceType:   "*monetary",
				Directions:    "*out",
				Identifier:    "*topup",
				RatingSubject: "",
				Units:         "10.500000",
				Weight:        10,
			},
		},
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions got reply: %s", reply)
	}
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1015",
		ActionsId: "PAYMENT_2056bd2fe137082970f97102b64e42fd",
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1015",
		ActionsId: "PAYMENT_2056bd2fe137082970f97102b64e42fd",
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1015"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	}
	if len(acnt.BalanceMap) != 1 ||
		len(acnt.BalanceMap[utils.MONETARY]) != 1 ||
		acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 21 {
		t.Error("error matching previous created balance: ", utils.ToIJSON(acnt.BalanceMap))
	}
}

func TestTpSetRemoveActions(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.SetActions", utils.AttrSetActions{
		ActionsId: "TO_BE_DELETED",
		Actions: []*utils.TPAction{
			&utils.TPAction{
				BalanceType:   "*monetary",
				Directions:    "*out",
				Identifier:    "*topup",
				RatingSubject: "",
				Units:         "10.500000",
				Weight:        10,
			},
		},
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions got reply: %s", reply)
	}
	actionsMap := make(map[string]engine.Actions)
	if err := tpRPC.Call("ApierV2.GetActions", v2.AttrGetActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &actionsMap); err != nil {
		t.Error("Got error on ApierV2.GetActions: ", err.Error())
	} else if len(actionsMap) != 1 {
		t.Errorf("Calling ApierV2.GetActions got reply: %s", utils.ToIJSON(actionsMap))
	}
	if err := tpRPC.Call("ApierV2.RemoveActions", v1.AttrRemoveActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.RemoveActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.RemoveActions got reply: %s", reply)
	}
	if err := tpRPC.Call("ApierV2.GetActions", v2.AttrGetActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &actionsMap); err == nil {
		t.Error("no error on ApierV2.GetActions: ", err)
	}
}

func TestTpRemoveActionsRefenced(t *testing.T) {

	// no more reference check for sake of speed!

	actionsMap := make(map[string]engine.Actions)
	if err := tpRPC.Call("ApierV2.GetActions", v2.AttrGetActions{
		ActionIDs: []string{"TOPUP_VOICE"},
	}, &actionsMap); err != nil {
		t.Error("Got error on ApierV2.GetActions: ", err.Error())
	} else if len(actionsMap) != 1 {
		t.Errorf("Calling ApierV2.GetActions got reply: %s", utils.ToIJSON(actionsMap))
	}
	var reply string
	if err := tpRPC.Call("ApierV2.RemoveActions", v1.AttrRemoveActions{
		ActionIDs: []string{"TOPUP_VOICE"},
	}, &reply); err != nil {
		t.Error("Error on ApierV2.RemoveActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.RemoveActions got reply: %s", reply)
	}
	/*
		if err := tpRPC.Call("ApierV2.GetActions", v2.AttrGetActions{
			ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
		}, &actionsMap); err == nil {
			t.Error("no error on ApierV2.GetActions: ", err)
		}
	*/
}

func TestTpApierResetAccountActionTriggers(t *testing.T) {
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1005"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.ActionTriggers[0].Executed != true {
		t.Skip("Skipping test since Executed is not yet true")
	}
	var reply string
	if err := tpRPC.Call("ApierV2.ResetAccountActionTriggers", v1.AttrResetAccountActionTriggers{
		Tenant:   "cgrates.org",
		Account:  "1005",
		GroupID:  "STANDARD_TRIGGERS",
		Executed: true,
	}, &reply); err != nil {
		t.Error("Error on ApierV2.ResetAccountActionTriggers: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.ResetAccountActionTriggers got reply: %s", reply)
	}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.ActionTriggers[0].Executed == false {
		t.Errorf("wrong action trigger executed flag: %s", utils.ToIJSON(acnt.ActionTriggers))
	}
}

func TestTpStopCgrEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
