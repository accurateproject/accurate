package v1

import (
	"fmt"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Computes the LCR for a specific request emulating a call
func (api *ApiV1) GetLcr(lcrReq engine.LcrRequest, lcrReply *engine.LcrReply) error {
	cd, err := lcrReq.AsCallDescriptor(*api.cfg.General.DefaultTimezone)
	if err != nil {
		return err
	}
	var lcrQried engine.LCRCost
	if err := api.responder.GetLCR(&engine.AttrGetLcr{CallDescriptor: cd, LCRFilter: lcrReq.LCRFilter, Paginator: lcrReq.Paginator}, &lcrQried); err != nil {
		return utils.NewErrServerError(err)
	}
	if lcrQried.Entry == nil {
		return utils.ErrNotFound
	}
	lcrReply.DestinationID = lcrQried.Entry.DestinationID
	lcrReply.RPCategory = lcrQried.Entry.RPCategory
	lcrReply.Strategy = lcrQried.Entry.Strategy
	for _, qriedSuppl := range lcrQried.SupplierCosts {
		if qriedSuppl.Error != "" {
			if !lcrReq.IgnoreErrors {
				return fmt.Errorf("%s:%s", utils.ErrServerError.Error(), "LCR_COMPUTE_ERRORS")
			}
			continue
		}
		if dtcs, err := utils.NewDTCSFromRPKey(qriedSuppl.Supplier); err != nil {
			return utils.NewErrServerError(err)
		} else {
			lcrReply.Suppliers = append(lcrReply.Suppliers, &engine.LcrSupplier{Supplier: dtcs.Subject, Cost: qriedSuppl.Cost, QOS: qriedSuppl.QOS})
		}
	}
	return nil
}

// Computes the LCR for a specific request emulating a call, returns a comma separated list of suppliers
func (api *ApiV1) GetLcrSuppliers(lcrReq engine.LcrRequest, suppliers *string) (err error) {
	cd, err := lcrReq.AsCallDescriptor(*api.cfg.General.DefaultTimezone)
	if err != nil {
		return err
	}
	var lcrQried engine.LCRCost
	if err := api.responder.GetLCR(&engine.AttrGetLcr{CallDescriptor: cd, LCRFilter: lcrReq.LCRFilter, Paginator: lcrReq.Paginator}, &lcrQried); err != nil {
		return utils.NewErrServerError(err)
	}
	if lcrQried.HasErrors() {
		lcrQried.LogErrors()
		if !lcrReq.IgnoreErrors {
			return fmt.Errorf("%s:%s", utils.ErrServerError.Error(), "LCR_COMPUTE_ERRORS")
		}
	}
	if suppliersStr, err := lcrQried.SuppliersString(); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = suppliersStr
	}
	return nil
}
