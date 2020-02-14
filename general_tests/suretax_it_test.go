package general_tests

import (
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"testing"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

/*
Integration tests with SureTax platform.
Configuration file is kept outside of AccuRate repository since it contains sensitive customer information
*/

var testSureTax = flag.Bool("suretax", false, "Pefrom SureTax integration tests when this flag is activated")
var configDir = flag.String("config_dir", "", "CGR config dir path here")
var tpDir = flag.String("tp_dir", "", "CGR config dir path here")

var stiCfg *config.Config
var stiRpc *rpc.Client
var stiLoadInst utils.LoadInstance

func TestSTIInitCfg(t *testing.T) {
	if !*testSureTax {
		return
	}
	// Init config first
	var err error
	config.Reset()
	if err = config.LoadPath(*configDir); err != nil {
		t.Error(err)
	}
	//stiCfg := config.Get()
}

// Remove data in both rating and accounting db
func TestSTIResetDataDb(t *testing.T) {
	if !*testSureTax {
		return
	}
	if err := engine.InitDataDb(stiCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSTIResetStorDb(t *testing.T) {
	if !*testSureTax {
		return
	}
	if err := engine.InitStorDb(stiCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSTIStartEngine(t *testing.T) {
	if !*testSureTax {
		return
	}
	if _, err := engine.StopStartEngine(*configDir, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSTIRpcConn(t *testing.T) {
	if !*testSureTax {
		return
	}
	var err error
	stiRpc, err = jsonrpc.Dial("tcp", *stiCfg.Listen.RpcJson) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSTILoadTariffPlanFromFolder(t *testing.T) {
	if !*testSureTax {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: *tpDir}
	if err := stiRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &stiLoadInst); err != nil {
		t.Error(err)
	} else if stiLoadInst.RatingLoadID == "" || stiLoadInst.AccountingLoadID == "" {
		t.Error("Empty loadId received, loadInstance: ", stiLoadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestSTICacheStats(t *testing.T) {
	if !*testSureTax {
		return
	}
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 1, RatingPlans: 1, RatingProfiles: 1, DerivedChargers: 1}
	var args utils.AttrCacheStats
	if err := stiRpc.Call("ApierV2.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV2.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV2.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
	}
}

// Test CDR from external sources
func TestSTIProcessExternalCdr(t *testing.T) {
	if !*testSureTax {
		return
	}
	cdr := &engine.ExternalCDR{ToR: utils.VOICE,
		OriginID: "teststicdr1", OriginHost: "192.168.1.1", Source: "STI_TEST", RequestType: utils.META_RATED, Direction: utils.OUT,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "+14082342500", Destination: "+16268412300", Supplier: "SUPPL1",
		SetupTime: "2015-10-18T13:00:00Z", AnswerTime: "2015-10-18T13:00:00Z",
		Usage: "15s", PDD: "7.0", ExtraFields: map[string]string{"CustomerNumber": "000000534", "ZipCode": ""},
	}
	var reply string
	if err := stiRpc.Call("CdrsV2.ProcessExternalCdr", cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(2) * time.Second)
}

func TestSTIGetCdrs(t *testing.T) {
	if !*testSureTax {
		return
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, Accounts: []string{"1001"}}
	if err := stiRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.012 {
			t.Errorf("Unexpected Cost for CDR: %+v", cdrs[0])
		}
	}
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_SURETAX}, Accounts: []string{"1001"}}
	if err := stiRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0027 {
			t.Errorf("Unexpected Cost for CDR: %+v", cdrs[0])
		}
	}
}

func TestSTIStopCgrEngine(t *testing.T) {
	if !*testSureTax {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
