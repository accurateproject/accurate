package cdre

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

var cdreData = []byte(`
{
	"cdre": {
		"*default": {
			"cdr_format": "fwv",
			"header_fields": [{
				"tag": "TypeOfRecord",
				"type": "*constant",
				"value": "10",
				"width": 2
			}, {
				"tag": "Filler1",
				"type": "*filler",
				"width": 3
			}, {
				"tag": "DistributorCode",
				"type": "*constant",
				"value": "VOI",
				"width": 3
			}, {
				"tag": "FileSeqNr",
				"type": "*handler",
				"value": "*export_id",
				"width": 5,
				"strip": "right",
				"padding": "zeroleft"
			}, {
				"tag": "LastCdr",
				"type": "*handler",
				"value": "*last_cdr_atime",
				"width": 12,
				"layout": "020106150400"
			}, {
				"tag": "FileCreationfTime",
				"type": "*handler",
				"value": "*time_now",
				"width": 12,
				"layout": "020106150400"
			}, {
				"tag": "FileVersion",
				"type": "*constant",
				"value": "01",
				"width": 2
			}, {
				"tag": "Filler2",
				"type": "*filler",
				"width": 105
			}],
			"content_fields": [{
				"tag": "TypeOfRecord",
				"type": "*constant",
				"value": "20",
				"width": 2
			}, {
				"tag": "Account",
				"type": "*composed",
				"value": "Account",
				"width": 12,
				"strip": "left",
				"padding": "right"
			}, {
				"tag": "Subject",
				"type": "*composed",
				"value": "Subject",
				"width": 5,
				"strip": "right",
				"padding": "right"
			}, {
				"tag": "CLI",
				"type": "*composed",
				"value": "cli",
				"width": 15,
				"strip": "xright",
				"padding": "right"
			}, {
				"tag": "Destination",
				"type": "*composed",
				"value": "Destination",
				"width": 24,
				"strip": "xright",
				"padding": "right"
			}, {
				"tag": "TOR",
				"type": "*constant",
				"value": "02",
				"width": 2
			}, {
				"tag": "SubtypeTOR",
				"type": "*constant",
				"value": "11",
				"width": 4,
				"padding": "right"
			}, {
				"tag": "SetupTime",
				"type": "*composed",
				"value": "SetupTime",
				"width": 12,
				"strip": "right",
				"padding": "right",
				"layout": "020106150400"
			}, {
				"tag": "Duration",
				"type": "*composed",
				"value": "Usage",
				"width": 6,
				"strip": "right",
				"padding": "right",
				"layout": "seconds"
			}, {
				"tag": "DataVolume",
				"type": "*filler",
				"width": 6
			}, {
				"tag": "TaxCode",
				"type": "*constant",
				"value": "1",
				"width": 1
			}, {
				"tag": "OperatorCode",
				"type": "*composed",
				"value": "opercode",
				"width": 2,
				"strip": "right",
				"padding": "right"
			}, {
				"tag": "ProductId",
				"type": "*composed",
				"value": "productid",
				"width": 5,
				"strip": "right",
				"padding": "right"
			}, {
				"tag": "NetworkId",
				"type": "*constant",
				"value": "3",
				"width": 1
			}, {
				"tag": "CallId",
				"type": "*composed",
				"value": "OriginID",
				"width": 16,
				"padding": "right"
			}, {
				"tag": "Filler",
				"type": "*filler",
				"width": 8
			}, {
				"tag": "Filler",
				"type": "*filler",
				"width": 8
			}, {
				"tag": "TerminationCode",
				"type": "*composed",
				"value": "operator;product",
				"width": 5,
				"strip": "right",
				"padding": "right"
			}, {
				"tag": "Cost",
				"type": "*composed",
				"value": "Cost",
				"width": 9,
				"padding": "zeroleft"
			}, {
				"tag": "DestinationPrivacy",
				"type": "*masked_destination",
				"width": 1
			}],
			"trailer_fields": [{
				"tag": "TypeOfRecord",
				"type": "*constant",
				"value": "90",
				"width": 2
			}, {
				"tag": "Filler1",
				"type": "*filler",
				"width": 3
			}, {
				"tag": "DistributorCode",
				"type": "*constant",
				"value": "VOI",
				"width": 3
			}, {
				"tag": "FileSeqNr",
				"type": "*handler",
				"value": "*export_id",
				"width": 5,
				"strip": "right",
				"padding": "zeroleft"
			}, {
				"tag": "NumberOfRecords",
				"type": "*handler",
				"value": "*cdrs_number",
				"width": 6,
				"padding": "zeroleft"
			}, {
				"tag": "CdrsDuration",
				"type": "*handler",
				"value": "*cdrs_duration",
				"width": 8,
				"padding": "zeroleft",
				"layout": "seconds"
			}, {
				"tag": "FirstCdrTime",
				"type": "*handler",
				"value": "*first_cdr_atime",
				"width": 12,
				"layout": "020106150400"
			}, {
				"tag": "LastCdrTime",
				"type": "*handler",
				"value": "*last_cdr_atime",
				"width": 12,
				"layout": "020106150400"
			}, {
				"tag": "Filler2",
				"type": "*filler",
				"width": 93
			}]
		}
	}
}
`)

var hdrCfgFlds, contentCfgFlds, trailerCfgFlds []*config.CdrField

// Write one CDR and test it's results only for content buffer
func TestWriteCdr(t *testing.T) {
	var err error
	wrBuf := &bytes.Buffer{}
	config.Reset()
	if err := config.LoadBytes(cdreData, true); err != nil {
		t.Fatal("error loading config: ", err)
	}

	cfg := config.Get()
	//log.Printf("CDRE: %s", utils.ToIJSON(cdreCfg.Cdre))

	hdrCfgFlds = (*cfg.Cdre)["*default"].HeaderFields
	contentCfgFlds = (*cfg.Cdre)["*default"].ContentFields
	trailerCfgFlds = (*cfg.Cdre)["*default"].TrailerFields

	cdreCfg := &config.Cdre{
		CdrFormat:     utils.StringPointer(utils.CDRE_FIXED_WIDTH),
		HeaderFields:  hdrCfgFlds,
		ContentFields: contentCfgFlds,
		TrailerFields: trailerCfgFlds,
	}

	cdr := &engine.CDR{UniqueID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		ToR: utils.VOICE, OrderID: 1, OriginID: "dsafdsaf", OriginHost: "192.168.1.1",
		RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second, RunID: utils.DEFAULT_RUNID, Cost: dec.NewFloat(2.34567),
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	cdre, err := NewCdrExporter([]*engine.CDR{cdr}, nil, cdreCfg, utils.CDRE_FIXED_WIDTH, ',', "fwv_1", 0.0, 0.0, 0.0, 0.0, 0.0,
		*cfg.General.RoundingDecimals, *cfg.General.HttpSkipTlsVerify)
	if err != nil {
		t.Error(err)
	}
	eHeader := "10   VOIfwv_107111308420018011511340001                                                                                                         \n"
	eContentOut := "201001        1001                1002                    0211  07111308420010          1       3dsafdsaf                             0002.34570\n"
	eTrailer := "90   VOIfwv_100000100000010071113084200071113084200                                                                                             \n"
	if err := cdre.writeOut(wrBuf); err != nil {
		t.Error(err)
	}
	allOut := wrBuf.String()
	eAllOut := eHeader + eContentOut + eTrailer
	if math.Mod(float64(len(allOut)), 145) != 0 {
		t.Error("Unexpected export content length", len(allOut))
	} else if len(allOut) != len(eAllOut) {
		t.Errorf("Output does not match expected length. Have output %q, expecting: %q", allOut, eAllOut)
	}
	// Test stats
	if !cdre.firstCdrATime.Equal(cdr.AnswerTime) {
		t.Error("Unexpected firstCdrATime in stats: ", cdre.firstCdrATime)
	} else if !cdre.lastCdrATime.Equal(cdr.AnswerTime) {
		t.Error("Unexpected lastCdrATime in stats: ", cdre.lastCdrATime)
	} else if cdre.numberOfRecords != 1 {
		t.Error("Unexpected number of records in the stats: ", cdre.numberOfRecords)
	} else if cdre.totalDuration != cdr.Usage {
		t.Error("Unexpected total duration in the stats: ", cdre.totalDuration)
	} else if cdre.GetTotalCost().Cmp(dec.New().Set(cdr.GetCost()).Round(int32(cdre.cgrPrecision))) != 0 {
		t.Error("Unexpected total cost in the stats: ", cdre.totalCost)
	}
	if cdre.FirstOrderId() != 1 {
		t.Error("Unexpected FirstOrderId", cdre.FirstOrderId())
	}
	if cdre.LastOrderId() != 1 {
		t.Error("Unexpected LastOrderId", cdre.LastOrderId())
	}
	/*if cdre.getTotalCost() != utils.Round(cdr.Cost, cdre.cgrPrecision, utils.ROUNDING_MIDDLE) {
		t.Error("Unexpected TotalCost: ", cdre.TotalCost())
	}*/
}

func TestWriteCdrs(t *testing.T) {
	wrBuf := &bytes.Buffer{}
	cdreCfg := &config.Cdre{
		CdrFormat:     utils.StringPointer(utils.CDRE_FIXED_WIDTH),
		HeaderFields:  hdrCfgFlds,
		ContentFields: contentCfgFlds,
		TrailerFields: trailerCfgFlds,
	}
	cdr1 := &engine.CDR{UniqueID: utils.Sha1("aaa1", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
		ToR: utils.VOICE, OrderID: 2, OriginID: "aaa1", OriginHost: "192.168.1.1", RequestType: utils.META_RATED, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1001", Subject: "1001", Destination: "1010",
		SetupTime:  time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage:      time.Duration(10) * time.Second, RunID: utils.DEFAULT_RUNID, Cost: dec.NewFloat(2.25),
		ExtraFields: map[string]string{"productnumber": "12341", "fieldextr2": "valextr2"},
	}
	cdr2 := &engine.CDR{UniqueID: utils.Sha1("aaa2", time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String()),
		ToR: utils.VOICE, OrderID: 4, OriginID: "aaa2", OriginHost: "192.168.1.2", RequestType: utils.META_PREPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1002", Subject: "1002", Destination: "1011",
		SetupTime:  time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 7, 42, 26, 0, time.UTC),
		Usage:      time.Duration(5) * time.Minute, RunID: utils.DEFAULT_RUNID, Cost: dec.NewFloat(1.40001),
		ExtraFields: map[string]string{"productnumber": "12342", "fieldextr2": "valextr2"},
	}
	cdr3 := &engine.CDR{}
	cdr4 := &engine.CDR{UniqueID: utils.Sha1("aaa3", time.Date(2013, 11, 7, 9, 42, 18, 0, time.UTC).String()),
		ToR: utils.VOICE, OrderID: 3, OriginID: "aaa4", OriginHost: "192.168.1.4", RequestType: utils.META_POSTPAID, Direction: "*out", Tenant: "cgrates.org",
		Category: "call", Account: "1004", Subject: "1004", Destination: "1013",
		SetupTime:  time.Date(2013, 11, 7, 9, 42, 18, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 9, 42, 26, 0, time.UTC),
		Usage:      time.Duration(20) * time.Second, RunID: utils.DEFAULT_RUNID, Cost: dec.NewFloat(2.34567),
		ExtraFields: map[string]string{"productnumber": "12344", "fieldextr2": "valextr2"},
	}
	config.Reset()
	cfg := config.Get()
	cdre, err := NewCdrExporter([]*engine.CDR{cdr1, cdr2, cdr3, cdr4}, nil, cdreCfg, utils.CDRE_FIXED_WIDTH, ',',
		"fwv_1", 0.0, 0.0, 0.0, 0.0, 0.0, *cfg.General.RoundingDecimals, *cfg.General.HttpSkipTlsVerify)
	if err != nil {
		t.Error(err)
	}
	if err := cdre.writeOut(wrBuf); err != nil {
		t.Error(err)
	}
	if len(wrBuf.String()) != 725 {
		t.Error("Output buffer does not contain expected info. Expecting len: 725, got: ", len(wrBuf.String()))
	}
	// Test stats
	if !cdre.firstCdrATime.Equal(cdr2.AnswerTime) {
		t.Error("Unexpected firstCdrATime in stats: ", cdre.firstCdrATime)
	}
	if !cdre.lastCdrATime.Equal(cdr4.AnswerTime) {
		t.Error("Unexpected lastCdrATime in stats: ", cdre.lastCdrATime)
	}
	if cdre.numberOfRecords != 3 {
		t.Error("Unexpected number of records in the stats: ", cdre.numberOfRecords)
	}
	if cdre.totalDuration != time.Duration(330)*time.Second {
		t.Error("Unexpected total duration in the stats: ", cdre.totalDuration)
	}
	if cdre.totalCost.Cmp(dec.NewFloat(5.9957)) != 0 {
		t.Error("Unexpected total cost in the stats: ", cdre.totalCost)
	}
	if cdre.FirstOrderId() != 2 {
		t.Error("Unexpected FirstOrderId", cdre.FirstOrderId())
	}
	if cdre.LastOrderId() != 4 {
		t.Error("Unexpected LastOrderId", cdre.LastOrderId())
	}
	if cdre.GetTotalCost().Cmp(dec.NewFloat(5.9957)) != 0 {
		t.Error("Unexpected TotalCost: ", cdre.GetTotalCost())
	}
}
