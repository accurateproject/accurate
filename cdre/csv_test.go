package cdre

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func TestCsvCdrWriter(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	storedCdr1 := &engine.CDR{
		CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage: time.Duration(10) * time.Second, RunID: utils.DEFAULT_RUNID,
		ExtraFields: map[string]string{"extra1": "val_extra1", "extra2": "val_extra2", "extra3": "val_extra3"}, Cost: 1.01,
	}
	cdre, err := NewCdrExporter([]*engine.CDR{storedCdr1}, nil, cfg.CdreProfiles["*default"], utils.CSV, ',', "firstexport", 0.0, 0.0, 0.0, 0.0, 0.0, 0, 4,
		cfg.RoundingDecimals, "", 0, cfg.HttpSkipTlsVerify, "")
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6,*default,*voice,dsafdsaf,*rated,*out,cgrates.org,call,1001,1001,1002,2013-11-07T08:42:25Z,2013-11-07T08:42:26Z,10,1.0100`
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.01 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}

func TestAlternativeFieldSeparator(t *testing.T) {
	writer := &bytes.Buffer{}
	cfg, _ := config.NewDefaultCGRConfig()
	storedCdr1 := &engine.CDR{CGRID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()), ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002", SetupTime: time.Unix(1383813745, 0).UTC(), AnswerTime: time.Unix(1383813746, 0).UTC(),
		Usage: time.Duration(10) * time.Second, RunID: utils.DEFAULT_RUNID,
		ExtraFields: map[string]string{"extra1": "val_extra1", "extra2": "val_extra2", "extra3": "val_extra3"}, Cost: 1.01,
	}
	cdre, err := NewCdrExporter([]*engine.CDR{storedCdr1}, nil, cfg.CdreProfiles["*default"], utils.CSV, '|',
		"firstexport", 0.0, 0.0, 0.0, 0.0, 0.0, 0, 4, cfg.RoundingDecimals, "", 0, cfg.HttpSkipTlsVerify, "")
	if err != nil {
		t.Error("Unexpected error received: ", err)
	}
	csvWriter := csv.NewWriter(writer)
	if err := cdre.writeCsv(csvWriter); err != nil {
		t.Error("Unexpected error: ", err)
	}
	expected := `dbafe9c8614c785a65aabd116dd3959c3c56f7f6|*default|*voice|dsafdsaf|*rated|*out|cgrates.org|call|1001|1001|1002|2013-11-07T08:42:25Z|2013-11-07T08:42:26Z|10|1.0100`
	result := strings.TrimSpace(writer.String())
	if result != expected {
		t.Errorf("Expected: \n%s received: \n%s.", expected, result)
	}
	if cdre.TotalCost() != 1.01 {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}
}
