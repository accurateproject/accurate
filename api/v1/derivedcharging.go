package v1

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Get DerivedChargers applying to our call, appends general configured to account specific ones if that is configured
func (api *ApiV1) GetDerivedChargers(attrs utils.AttrDerivedChargers, reply *utils.DerivedChargerGroup) (err error) {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Direction", "Account", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if hDc, err := engine.HandleGetDerivedChargers(api.ratingDB, &attrs); err != nil {
		return utils.NewErrServerError(err)
	} else if hDc != nil {
		*reply = *hDc
	}
	return nil
}

/*
type AttrSetDerivedChargers struct {
	Direction, Tenant, Category, Account, Subject, DestinationIds string
	DerivedChargers                                               []*utils.DerivedCharger
	Overwrite                                                     bool // Do not overwrite if present in redis
}

func (api *ApiV1) SetDerivedChargers(attrs AttrSetDerivedChargers, reply *string) (err error) {
	if len(attrs.DerivedChargers) == 0 {
		return utils.NewErrMandatoryIeMissing("DerivedChargers")
	}
	if len(attrs.Direction) == 0 {
		attrs.Direction = utils.OUT
	}
	if len(attrs.Tenant) == 0 {
		attrs.Tenant = utils.ANY
	}
	if len(attrs.Category) == 0 {
		attrs.Category = utils.ANY
	}
	if len(attrs.Account) == 0 {
		attrs.Account = utils.ANY
	}
	if len(attrs.Subject) == 0 {
		attrs.Subject = utils.ANY
	}
	for _, dc := range attrs.DerivedChargers {
		if _, err = utils.ParseRSRFields(dc.RunFilters, utils.INFIELD_SEP); err != nil { // Make sure rules are OK before loading in db
			return fmt.Errorf("%s:%s", utils.ErrParserError.Error(), err.Error())
		}
	}
	dcKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject)
	if !attrs.Overwrite {
		if exists, err := api.ratingDB.HasData(utils.DERIVEDCHARGERS_PREFIX, dcKey); err != nil {
			return utils.NewErrServerError(err)
		} else if exists {
			return utils.ErrExists
		}
	}
	dstIds := strings.Split(attrs.DestinationIds, utils.INFIELD_SEP)
	dcs := &utils.DerivedChargers{DestinationIDs: utils.NewStringMap(dstIds...), Chargers: attrs.DerivedChargers}
	if err := api.ratingDB.SetDerivedChargers(dcKey, dcs); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
*/

type AttrRemDerivedChargers struct {
	Direction, Tenant, Category, Account, Subject string
}

func (api *ApiV1) RemDerivedChargers(attr AttrRemDerivedChargers, reply *string) error {
	if len(attr.Direction) == 0 {
		attr.Direction = utils.OUT
	}
	if len(attr.Tenant) == 0 {
		attr.Tenant = utils.ANY
	}
	if len(attr.Category) == 0 {
		attr.Category = utils.ANY
	}
	if len(attr.Account) == 0 {
		attr.Account = utils.ANY
	}
	if len(attr.Subject) == 0 {
		attr.Subject = utils.ANY
	}
	if err := api.ratingDB.SetDerivedChargers(&utils.DerivedChargerGroup{Tenant: attr.Tenant, Direction: attr.Direction, Category: attr.Category, Account: attr.Account, Subject: attr.Subject}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
