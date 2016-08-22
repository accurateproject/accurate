package config

import (
	"github.com/accurateproject/accurate/utils"
)

func NewCfgCdrFieldFromCdrFieldJsonCfg(jsnCfgFld *CdrFieldJsonCfg) (*CfgCdrField, error) {
	var err error
	cfgFld := new(CfgCdrField)
	if jsnCfgFld.Tag != nil {
		cfgFld.Tag = *jsnCfgFld.Tag
	}
	if jsnCfgFld.Type != nil {
		cfgFld.Type = *jsnCfgFld.Type
	}
	if jsnCfgFld.Field_id != nil {
		cfgFld.FieldId = *jsnCfgFld.Field_id
	}
	if jsnCfgFld.Handler_id != nil {
		cfgFld.HandlerId = *jsnCfgFld.Handler_id
	}
	if jsnCfgFld.Value != nil {
		if cfgFld.Value, err = utils.ParseRSRFields(*jsnCfgFld.Value, utils.INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	if jsnCfgFld.Append != nil {
		cfgFld.Append = *jsnCfgFld.Append
	}
	if jsnCfgFld.Field_filter != nil {
		if cfgFld.FieldFilter, err = utils.ParseRSRFields(*jsnCfgFld.Field_filter, utils.INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	if jsnCfgFld.Width != nil {
		cfgFld.Width = *jsnCfgFld.Width
	}
	if jsnCfgFld.Strip != nil {
		cfgFld.Strip = *jsnCfgFld.Strip
	}
	if jsnCfgFld.Padding != nil {
		cfgFld.Padding = *jsnCfgFld.Padding
	}
	if jsnCfgFld.Layout != nil {
		cfgFld.Layout = *jsnCfgFld.Layout
	}
	if jsnCfgFld.Mandatory != nil {
		cfgFld.Mandatory = *jsnCfgFld.Mandatory
	}
	return cfgFld, nil
}

type CfgCdrField struct {
	Tag         string // Identifier for the administrator
	Type        string // Type of field
	FieldId     string // Field identifier
	HandlerId   string
	Value       utils.RSRFields
	Append      bool
	FieldFilter utils.RSRFields
	Width       int
	Strip       string
	Padding     string
	Layout      string
	Mandatory   bool
}

func CfgCdrFieldsFromCdrFieldsJsonCfg(jsnCfgFldss []*CdrFieldJsonCfg) ([]*CfgCdrField, error) {
	retFields := make([]*CfgCdrField, len(jsnCfgFldss))
	for idx, jsnFld := range jsnCfgFldss {
		if cfgFld, err := NewCfgCdrFieldFromCdrFieldJsonCfg(jsnFld); err != nil {
			return nil, err
		} else {
			retFields[idx] = cfgFld
		}
	}
	return retFields, nil
}
