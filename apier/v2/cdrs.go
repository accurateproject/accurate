package v2

import (
	"github.com/accurateproject/accurate/apier/v1"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Retrieves CDRs based on the filters
func (apier *ApierV2) GetCdrs(attrs utils.RPCCDRsFilter, reply *[]*engine.ExternalCDR) error {
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

func (apier *ApierV2) CountCdrs(attrs utils.RPCCDRsFilter, reply *int64) error {
	cdrsFltr, err := attrs.AsCDRsFilter(apier.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrsFltr.Count = true
	if _, count, err := apier.CdrDb.GetCDRs(cdrsFltr, false); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = count
	}
	return nil
}

// Receive CDRs via RPC methods, not included with APIer because it has way less dependencies and can be standalone
type CdrsV2 struct {
	v1.CdrsV1
}
