package v1

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Receive CDRs via RPC methods
type CdrsV1 struct {
	CdrSrv *engine.CdrServer
}

// Designed for CGR internal usage
// Deprecated
func (self *CdrsV1) ProcessCdr(cdr *engine.CDR, reply *string) error {
	return self.ProcessCDR(cdr, reply)
}

// Designed for CGR internal usage
func (self *CdrsV1) ProcessCDR(cdr *engine.CDR, reply *string) error {
	return self.CdrSrv.V1ProcessCDR(cdr, reply)
}

// Designed for external programs feeding CDRs to AccuRate
// Deprecated
func (self *CdrsV1) ProcessExternalCdr(cdr *engine.ExternalCDR, reply *string) error {
	return self.ProcessExternalCDR(cdr, reply)
}

// Designed for external programs feeding CDRs to AccuRate
func (self *CdrsV1) ProcessExternalCDR(cdr *engine.ExternalCDR, reply *string) error {
	if err := self.CdrSrv.ProcessExternalCdr(cdr); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Remotely (re)rating
// Deprecated
func (self *CdrsV1) RateCdrs(attrs utils.AttrRateCdrs, reply *string) error {
	return self.RateCDRs(attrs, reply)
}

// Remotely (re)rating
func (self *CdrsV1) RateCDRs(attrs utils.AttrRateCdrs, reply *string) error {
	cdrsFltr, err := attrs.AsCDRsFilter(self.CdrSrv.Timezone())
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := self.CdrSrv.RateCDRs(cdrsFltr, attrs.SendToStats); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (self *CdrsV1) StoreSMCost(attr engine.AttrCDRSStoreSMCost, reply *string) error {
	return self.CdrSrv.V1StoreSMCost(attr, reply)
}
