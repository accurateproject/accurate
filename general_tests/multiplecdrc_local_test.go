package general_tests

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

var cfgPath string
var cfg *config.Config
var rater *rpc.Client

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var testCalls = flag.Bool("calls", false, "Run test calls simulation, not by default.")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var storDbType = flag.String("stordb_type", "mysql", "The type of the storDb database <mysql>")
var waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")

func startEngine() error {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return errors.New("Cannot find cgr-engine executable")
	}
	stopEngine()
	engine := exec.Command(enginePath, "-config_dir", cfgPath)
	if err := engine.Start(); err != nil {
		return fmt.Errorf("Cannot start cgr-engine: %s", err.Error())
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time to rater to fire up
	return nil
}

func stopEngine() error {
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	return nil
}

func TestMCDRCLoadConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cfgPath = path.Join(*dataDir, "conf", "samples", "multiplecdrc")
	config.Reset()
	if err = config.LoadPath(cfgPath); err != nil {
		t.Error(err)
	}
	cfg = config.Get()
}

// Remove data in both rating and accounting db
func TestMCDRCResetDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestMCDRCEmptyTables(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestMCDRCCreateCdrDirs(t *testing.T) {
	if !*testLocal {
		return
	}
	for _, cdrcProfiles := range cfg.CdrcProfiles() {
		for _, cdrcInst := range cdrcProfiles {
			for _, dir := range []string{*cdrcInst.CdrInDir, *cdrcInst.CdrOutDir} {
				if err := os.RemoveAll(dir); err != nil {
					t.Fatal("Error removing folder: ", dir, err)
				}
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Error creating folder: ", dir, err)
				}
			}
		}
	}
}

// Connect rpc client to rater
func TestMCDRCRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	startEngine()
	var err error
	rater, err = jsonrpc.Dial("tcp", *cfg.Listen.RpcJson) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Test here LoadTariffPlanFromFolder
func TestMCDRCApierLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Simple test that command is executed without errors
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	if err := rater.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// The default scenario, out of cdrc defined in .cfg file
func TestMCDRCHandleCdr1File(t *testing.T) {
	if !*testLocal {
		return
	}
	var fileContent1 = `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,default,*voice,dsafdsaf,rated,*out,cgrates.org,call,1001,1001,+4986517174963,2013-11-07 08:42:25 +0000 UTC,2013-11-07 08:42:26 +0000 UTC,10000000000,1.0100,val_extra3,"",val_extra1
dbafe9c8614c785a65aabd116dd3959c3c56f7f7,default,*voice,dsafdsag,rated,*out,cgrates.org,call,1001,1001,+4986517174964,2013-11-07 09:42:25 +0000 UTC,2013-11-07 09:42:26 +0000 UTC,20000000000,1.0100,val_extra3,"",val_extra1
`
	fileName := "file1.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cgrates/cdrc1/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of first .xml config
func TestMCDRCHandleCdr2File(t *testing.T) {
	if !*testLocal {
		return
	}
	var fileContent = `616350843,20131022145011,20131022172857,3656,1001,,,data,mo,640113,0.000000,1.222656,1.222660
616199016,20131022154924,20131022154955,3656,1001,086517174963,,voice,mo,31,0.000000,0.000000,0.000000
800873243,20140516063739,20140516063739,9774,1001,+49621621391,,sms,mo,1,0.00000,0.00000,0.00000`
	fileName := "file2.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cgrates/cdrc2/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

// Scenario out of second .xml config
func TestMCDRCHandleCdr3File(t *testing.T) {
	if !*testLocal {
		return
	}
	var fileContent = `4986517174960;4986517174963;Sample Mobile;08.04.2014  22:14:29;08.04.2014  22:14:29;1;193;Offeak;0,072728833;31619
4986517174960;4986517174964;National;08.04.2014  20:34:55;08.04.2014  20:34:55;1;21;Offeak;0,0079135;311`
	fileName := "file3.csv"
	tmpFilePath := path.Join("/tmp", fileName)
	if err := ioutil.WriteFile(tmpFilePath, []byte(fileContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpFilePath, path.Join("/tmp/cgrates/cdrc3/in", fileName)); err != nil {
		t.Fatal("Error moving file to processing directory: ", err)
	}
}

func TestMCDRCStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	stopEngine()
}
