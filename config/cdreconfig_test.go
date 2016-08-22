package config

import (
	"reflect"
	"testing"

	"github.com/accurateproject/accurate/utils"
)

func TestCdreCfgClone(t *testing.T) {
	cgrIdRsrs, _ := utils.ParseRSRFields("cgrid", utils.INFIELD_SEP)
	runIdRsrs, _ := utils.ParseRSRFields("mediation_runid", utils.INFIELD_SEP)
	emptyFields := []*CfgCdrField{}
	initContentFlds := []*CfgCdrField{
		&CfgCdrField{Tag: "CgrId",
			Type:    "*composed",
			FieldId: "cgrid",
			Value:   cgrIdRsrs},
		&CfgCdrField{Tag: "RunId",
			Type:    "*composed",
			FieldId: "mediation_runid",
			Value:   runIdRsrs},
	}
	initCdreCfg := &CdreConfig{
		CdrFormat:               "csv",
		FieldSeparator:          rune(','),
		DataUsageMultiplyFactor: 1.0,
		CostMultiplyFactor:      1.0,
		CostRoundingDecimals:    -1,
		CostShiftDigits:         0,
		MaskDestinationID:       "MASKED_DESTINATIONS",
		MaskLength:              0,
		ExportDirectory:         "/var/spool/cgrates/cdre",
		ContentFields:           initContentFlds,
	}
	eClnContentFlds := []*CfgCdrField{
		&CfgCdrField{Tag: "CgrId",
			Type:    "*composed",
			FieldId: "cgrid",
			Value:   cgrIdRsrs},
		&CfgCdrField{Tag: "RunId",
			Type:    "*composed",
			FieldId: "mediation_runid",
			Value:   runIdRsrs},
	}
	eClnCdreCfg := &CdreConfig{
		CdrFormat:               "csv",
		FieldSeparator:          rune(','),
		DataUsageMultiplyFactor: 1.0,
		CostMultiplyFactor:      1.0,
		CostRoundingDecimals:    -1,
		CostShiftDigits:         0,
		MaskDestinationID:       "MASKED_DESTINATIONS",
		MaskLength:              0,
		ExportDirectory:         "/var/spool/cgrates/cdre",
		HeaderFields:            emptyFields,
		ContentFields:           eClnContentFlds,
		TrailerFields:           emptyFields,
	}
	clnCdreCfg := initCdreCfg.Clone()
	if !reflect.DeepEqual(eClnCdreCfg, clnCdreCfg) {
		t.Errorf("Cloned result: %+v", clnCdreCfg)
	}
	initCdreCfg.DataUsageMultiplyFactor = 1024.0
	if !reflect.DeepEqual(eClnCdreCfg, clnCdreCfg) { // MOdifying a field after clone should not affect cloned instance
		t.Errorf("Cloned result: %+v", clnCdreCfg)
	}
	initContentFlds[0].Tag = "Destination"
	if !reflect.DeepEqual(eClnCdreCfg, clnCdreCfg) { // MOdifying a field after clone should not affect cloned instance
		t.Errorf("Cloned result: %+v", clnCdreCfg)
	}
	clnCdreCfg.CostShiftDigits = 2
	if initCdreCfg.CostShiftDigits != 0 {
		t.Error("Unexpected CostShiftDigits: ", initCdreCfg.CostShiftDigits)
	}
	clnCdreCfg.ContentFields[0].FieldId = "destination"
	if initCdreCfg.ContentFields[0].FieldId != "cgrid" {
		t.Error("Unexpected change of FieldId: ", initCdreCfg.ContentFields[0].FieldId)
	}

}
