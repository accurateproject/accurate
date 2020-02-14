//+build integration

package cdrc

import (
	"io/ioutil"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path"
	"testing"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

var xmlCfgPath string
var xmlCfg *config.Config
var cdrcXmlCfgs []*config.Cdrc
var cdrcXmlCfg *config.Cdrc
var cdrcXmlRPC *rpc.Client
var xmlPathIn1, xmlPathOut1 string

func TestXmlITInitConfig(t *testing.T) {
	var err error
	xmlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrcxml")
	config.Reset()
	if err = config.LoadPath(xmlCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
	xmlCfg = config.Get()
}

// InitDb so we can rely on count
func TestXmlITInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(xmlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestXmlITCreateCdrDirs(t *testing.T) {
	for _, cdrcProfiles := range xmlCfg.CdrcProfiles() {
		for i, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{*cdrcInst.CdrInDir, *cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
			if i == 0 { // Initialize the folders to check later
				xmlPathIn1 = *cdrcInst.CdrInDir
				xmlPathOut1 = *cdrcInst.CdrOutDir
			}
		}
	}
}

func TestXmlITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(xmlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestXmlITRpcConn(t *testing.T) {
	var err error
	cdrcXmlRPC, err = jsonrpc.Dial("tcp", *xmlCfg.Listen.RpcJson) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// The default scenario, out of cdrc defined in .cfg file
func TestXmlITHandleCdr1File(t *testing.T) {
	fileName := "file1.xml"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(cdrXmlBroadsoft), 0644); err != nil {
		t.Fatal(err.Error)
	}
	if err := os.Rename(tmpFilePath, path.Join(xmlPathIn1, fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestXmlITProcessedFiles(t *testing.T) {
	time.Sleep(time.Duration(2**waitRater) * time.Millisecond)
	if outContent1, err := ioutil.ReadFile(path.Join(xmlPathOut1, "file1.xml")); err != nil {
		t.Error(err)
	} else if cdrXmlBroadsoft != string(outContent1) {
		t.Errorf("Expecting: %q, received: %q", cdrXmlBroadsoft, string(outContent1))
	}
}

func TestXmlITAnalyseCDRs(t *testing.T) {
	var reply []*engine.ExternalCDR
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	if err := cdrcXmlRPC.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{DestinationPrefixes: []string{"+4986517174963"}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}

}

func TestXmlITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
