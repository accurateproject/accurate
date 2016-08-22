package v1

// This file deals with tp_rates management over APIs

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Creates a new rate within a tariff plan
func (self *ApierV1) SetTPRate(attrs utils.TPRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateId", "RateSlots"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	r := engine.APItoModelRate(&attrs)
	if err := self.StorDb.SetTpRates(r); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPRate struct {
	TPid   string // Tariff plan id
	RateId string // Rate id
}

// Queries specific Rate on tariff plan
func (self *ApierV1) GetTPRate(attrs AttrGetTPRate, reply *utils.TPRate) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rts, err := self.StorDb.GetTpRates(attrs.TPid, attrs.RateId); err != nil {
		return utils.NewErrServerError(err)
	} else if len(rts) == 0 {
		return utils.ErrNotFound
	} else {
		rtsMap, err := engine.TpRates(rts).GetRates()
		if err != nil {
			return err
		}
		*reply = *rtsMap[attrs.RateId]
	}
	return nil
}

type AttrGetTPRateIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries rate identities on specific tariff plan.
func (self *ApierV1) GetTPRateIds(attrs AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_RATES, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Rate on Tariff plan
func (self *ApierV1) RemTPRate(attrs AttrGetTPRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RateId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_RATES, attrs.TPid, map[string]string{"tag": attrs.RateId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
