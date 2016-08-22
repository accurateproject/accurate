package v1

// This file deals with tp_destination_rates management over APIs

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Creates a new DestinationRate profile within a tariff plan
func (self *ApierV1) SetTPDestinationRate(attrs utils.TPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationRateId", "DestinationRates"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	drs := engine.APItoModelDestinationRate(&attrs)
	if err := self.StorDb.SetTpDestinationRates(drs); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPDestinationRate struct {
	TPid              string // Tariff plan id
	DestinationRateId string // Rate id
	utils.Paginator
}

// Queries specific DestinationRate profile on tariff plan
func (self *ApierV1) GetTPDestinationRate(attrs AttrGetTPDestinationRate, reply *utils.TPDestinationRate) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationRateId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if drs, err := self.StorDb.GetTpDestinationRates(attrs.TPid, attrs.DestinationRateId, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if len(drs) == 0 {
		return utils.ErrNotFound
	} else {
		drsMap, err := engine.TpDestinationRates(drs).GetDestinationRates()
		if err != nil {
			return err
		}
		*reply = *drsMap[attrs.DestinationRateId]
	}
	return nil
}

type AttrTPDestinationRateIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries DestinationRate identities on specific tariff plan.
func (self *ApierV1) GetTPDestinationRateIds(attrs AttrGetTPRateIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_DESTINATION_RATES, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific DestinationRate on Tariff plan
func (self *ApierV1) RemTPDestinationRate(attrs AttrGetTPDestinationRate, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "DestinationRateId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_DESTINATION_RATES, attrs.TPid, map[string]string{"tag": attrs.DestinationRateId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
