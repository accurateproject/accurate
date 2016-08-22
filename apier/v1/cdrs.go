package v1

import (
	"errors"
	"fmt"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Retrieves the callCost out of CGR logDb
func (apier *ApierV1) GetCallCostLog(attrs utils.AttrGetCallCost, reply *engine.SMCost) error {
	if attrs.CgrId == "" {
		return utils.NewErrMandatoryIeMissing("CgrId")
	}
	if attrs.RunId == "" {
		attrs.RunId = utils.META_DEFAULT
	}
	if smcs, err := apier.CdrDb.GetSMCosts(attrs.CgrId, attrs.RunId, "", ""); err != nil {
		return utils.NewErrServerError(err)
	} else if len(smcs) == 0 {
		return utils.ErrNotFound
	} else {
		*reply = *smcs[0]
	}
	return nil
}

// Retrieves CDRs based on the filters
func (apier *ApierV1) GetCdrs(attrs utils.AttrGetCdrs, reply *[]*engine.ExternalCDR) error {
	cdrsFltr, err := attrs.AsCDRsFilter(apier.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrs, _, err := apier.CdrDb.GetCDRs(cdrsFltr, false); err != nil {
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

// Remove Cdrs out of CDR storage
func (apier *ApierV1) RemCdrs(attrs utils.AttrRemCdrs, reply *string) error {
	if len(attrs.CgrIds) == 0 {
		return fmt.Errorf("%s:CgrIds", utils.ErrMandatoryIeMissing.Error())
	}
	if _, _, err := apier.CdrDb.GetCDRs(&utils.CDRsFilter{CGRIDs: attrs.CgrIds}, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

// New way of removing CDRs
func (apier *ApierV1) RemoveCDRs(attrs utils.RPCCDRsFilter, reply *string) error {
	cdrsFilter, err := attrs.AsCDRsFilter(apier.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if _, _, err := apier.CdrDb.GetCDRs(cdrsFilter, true); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

// New way of (re-)rating CDRs
func (apier *ApierV1) RateCDRs(attrs utils.AttrRateCDRs, reply *string) error {
	if apier.CDRs == nil {
		return errors.New("CDRS_NOT_ENABLED")
	}
	return apier.CDRs.Call("CDRsV1.RateCDRs", attrs, reply)
}
