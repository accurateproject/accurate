// +build integration

package v1

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

var cdrstCfgPath string
var cdrstCfg *config.Config
var cdrstRpc *rpc.Client

func TestCDRStatsLclLoadConfig(t *testing.T) {
	var err error
	cdrstCfgPath = path.Join(*dataDir, "conf", "samples", "cdrstats")
	config.Reset()
	if err = config.LoadPath(cfgPath); err != nil {
		t.Error(err)
	}
	cdrstCfg = config.Get()
}

func TestCDRStatsitInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdrstCfg); err != nil {
		t.Fatal(err)
	}
}

func TestCDRStatsitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrstCfgPath, 1000); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestCDRStatsitRpcConn(t *testing.T) {
	var err error
	cdrstRpc, err = jsonrpc.Dial("tcp", *cdrstCfg.Listen.RpcJson) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func TestCDRStatsitLoadTariffPlanFromFolder(t *testing.T) {
	reply := ""
	// Simple test that command is executed without errors
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "cdrstats")}
	if err := cdrstRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestCDRStatsitGetQueueIds2(t *testing.T) {
	var queueIds []string
	eQueueIds := []string{"CDRST3", "CDRST4"}
	if err := cdrstRpc.Call("CDRStatsV1.GetQueueIds", "", &queueIds); err != nil {
		t.Error("Calling CDRStatsV1.GetQueueIds, got error: ", err.Error())
	} else if len(eQueueIds) != len(queueIds) {
		t.Errorf("Expecting: %v, received: %v", eQueueIds, queueIds)
	}
	var rcvMetrics map[string]float64
	expectedMetrics := map[string]float64{"ASR": -1, "ACD": -1}
	if err := cdrstRpc.Call("CDRStatsV1.GetMetrics", AttrGetMetrics{StatsQueueId: "CDRST4"}, &rcvMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedMetrics, rcvMetrics) {
		t.Errorf("Expecting: %v, received: %v", expectedMetrics, rcvMetrics)
	}
}

func TestCDRStatsitPostCdrs(t *testing.T) {
	storedCdrs := []*engine.CDR{
		&engine.CDR{UniqueID: utils.Sha1("dsafdsafa", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsafa",
			OriginHost: "192.168.1.1", Source: "test",
			RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		&engine.CDR{UniqueID: utils.Sha1("dsafdsafb", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsafb",
			OriginHost: "192.168.1.1", Source: "test",
			RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(),
			AnswerTime: time.Now(), RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(5) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		&engine.CDR{UniqueID: utils.Sha1("dsafdsafc", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsafc",
			OriginHost: "192.168.1.1", Source: "test",
			RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(), AnswerTime: time.Now(),
			RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(30) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
		&engine.CDR{UniqueID: utils.Sha1("dsafdsafd", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsafd",
			OriginHost: "192.168.1.1", Source: "test",
			RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
			Category: "call", Account: "1001", Subject: "1001", Destination: "+4986517174963", SetupTime: time.Now(), AnswerTime: time.Time{},
			RunID: utils.DEFAULT_RUNID,
			Usage: time.Duration(0) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := cdrstRpc.Call("CdrsV1.ProcessCdr", cdr, &reply); err != nil {
			t.Error(err)
		}
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func TestCDRStatsitGetMetrics1(t *testing.T) {
	var rcvMetrics2 map[string]float64
	expectedMetrics2 := map[string]float64{"ASR": 75, "ACD": 15}
	if err := cdrstRpc.Call("CDRStatsV1.GetMetrics", AttrGetMetrics{StatsQueueId: "CDRST4"}, &rcvMetrics2); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedMetrics2, rcvMetrics2) {
		t.Errorf("Expecting: %v, received: %v", expectedMetrics2, rcvMetrics2)
	}
}

// Test stats persistence
func TestCDRStatsitStatsPersistence(t *testing.T) {
	time.Sleep(time.Duration(2) * time.Second) // Allow stats to be updated in dataDb
	if _, err := engine.StopStartEngine(cdrstCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	var err error
	cdrstRpc, err = jsonrpc.Dial("tcp", *cdrstCfg.Listen.RpcJson) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	var rcvMetrics map[string]float64
	expectedMetrics := map[string]float64{"ASR": 75, "ACD": 15}
	if err := cdrstRpc.Call("CDRStatsV1.GetMetrics", AttrGetMetrics{StatsQueueId: "CDRST4"}, &rcvMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedMetrics, rcvMetrics) {
		t.Errorf("Expecting: %v, received: %v", expectedMetrics, rcvMetrics)
	}
}

func TestCDRStatsitResetMetrics(t *testing.T) {
	var reply string
	if err := cdrstRpc.Call("CDRStatsV1.ResetQueues", utils.AttrCDRStatsReloadQueues{StatsQueueIds: []string{"CDRST4"}}, &reply); err != nil {
		t.Error("Calling CDRStatsV1.ResetQueues, got error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var rcvMetrics2 map[string]float64
	expectedMetrics2 := map[string]float64{"ASR": -1, "ACD": -1}
	if err := cdrstRpc.Call("CDRStatsV1.GetMetrics", AttrGetMetrics{StatsQueueId: "CDRST4"}, &rcvMetrics2); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedMetrics2, rcvMetrics2) {
		t.Errorf("Expecting: %v, received: %v", expectedMetrics2, rcvMetrics2)
	}
}

func TestCDRStatsitKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
