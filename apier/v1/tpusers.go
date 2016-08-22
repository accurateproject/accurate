package v1

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Creates a new alias within a tariff plan
func (self *ApierV1) SetTPUser(attrs utils.TPUsers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject", "Group"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tm := engine.APItoModelUsers(&attrs)
	if err := self.StorDb.SetTpUsers(tm); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPUser struct {
	TPid   string // Tariff plan id
	UserId string
}

// Queries specific User on Tariff plan
func (self *ApierV1) GetTPUser(attr AttrGetTPUser, reply *utils.TPUsers) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	usr := &engine.TpUser{
		Tpid:   attr.TPid,
		Masked: true,
	}
	usr.SetId(attr.UserId)
	if tms, err := self.StorDb.GetTpUsers(usr); err != nil {
		return utils.NewErrServerError(err)
	} else if len(tms) == 0 {
		return utils.ErrNotFound
	} else {
		tmMap, err := engine.TpUsers(tms).GetUsers()
		if err != nil {
			return err
		}
		*reply = *tmMap[usr.GetId()]
	}
	return nil
}

type AttrGetTPUserIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries alias identities on specific tariff plan.
func (self *ApierV1) GetTPUserIds(attrs AttrGetTPUserIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_USERS, utils.TPDistinctIds{"tenant", "user_name"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific User on Tariff plan
func (self *ApierV1) RemTPUser(attrs AttrGetTPUser, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "UserId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_USERS, attrs.TPid, map[string]string{"tag": attrs.UserId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
