package v1

import (
	"fmt"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Retrieves CDRs based on the filters
func (api *ApiV1) GetCdrs(attrs utils.RPCCDRsFilter, reply *[]*engine.ExternalCDR) error {
	cdrsFltr, err := attrs.AsCDRsFilter(*api.cfg.General.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrs, _, err := api.cdrDB.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else if len(cdrs) == 0 {
		*reply = make([]*engine.ExternalCDR, 0)
	} else {
		for _, cdr := range cdrs {
			*reply = append(*reply, cdr.AsExternalCDR())
		}
	}
	return nil
}

func (api *ApiV1) CountCdrs(attrs utils.RPCCDRsFilter, reply *int64) error {
	cdrsFltr, err := attrs.AsCDRsFilter(*api.cfg.General.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrsFltr.Count = true
	if _, count, err := api.cdrDB.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = count
	}
	return nil
}

// Retrieves the callCost out of CGR logDb
func (api *ApiV1) GetCallCostLog(attrs utils.AttrGetCallCost, reply *engine.SMCost) error {
	if attrs.UniqueID == "" {
		return utils.NewErrMandatoryIeMissing("UniqueID")
	}
	if attrs.RunID == "" {
		attrs.RunID = utils.META_DEFAULT
	}
	if smcs, err := api.cdrDB.GetSMCosts(attrs.UniqueID, attrs.RunID, "", ""); err != nil {
		return utils.NewErrServerError(err)
	} else if len(smcs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = *smcs[0]
	}
	return nil
}

// Remove Cdrs out of CDR storage
func (api *ApiV1) RemCdrs(attrs utils.AttrRemCdrs, reply *string) error {
	if len(attrs.UniqueIDs) == 0 {
		return fmt.Errorf("%s:UniqueIDs", utils.ErrMandatoryIeMissing.Error())
	}
	if _, _, err := api.cdrDB.GetCDRs(&utils.CDRsFilter{UniqueIDs: attrs.UniqueIDs}, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

// New way of removing CDRs
func (api *ApiV1) RemoveCDRs(attrs utils.RPCCDRsFilter, reply *string) error {
	cdrsFilter, err := attrs.AsCDRsFilter(*api.cfg.General.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if _, _, err := api.cdrDB.GetCDRs(cdrsFilter, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

// New way of (re-)rating CDRs
/*func (api *ApiV1) RateCDRs(attrs utils.AttrRateCDRs, reply *string) error {
	if api.cdrs == nil {
		return errors.New("CDRS_NOT_ENABLED")
	}
	return api.cdrs.Call("CDRsV1.RateCDRs", attrs, reply)
}*/
