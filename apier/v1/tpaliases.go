package v1

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Creates a new alias within a tariff plan
func (self *ApierV1) SetTPAlias(attrs utils.TPAliases, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Direction", "Tenant", "Category", "Account", "Subject", "Group"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tm := engine.APItoModelAliases(&attrs)
	if err := self.StorDb.SetTpAliases(tm); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPAlias struct {
	TPid    string // Tariff plan id
	AliasId string
}

// Queries specific Alias on Tariff plan
func (self *ApierV1) GetTPAlias(attr AttrGetTPAlias, reply *utils.TPAliases) error {
	if missing := utils.MissingStructFields(&attr, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	al := &engine.TpAlias{Tpid: attr.TPid}
	al.SetId(attr.AliasId)
	if tms, err := self.StorDb.GetTpAliases(al); err != nil {
		return utils.NewErrServerError(err)
	} else if len(tms) == 0 {
		return utils.ErrNotFound
	} else {
		tmMap, err := engine.TpAliases(tms).GetAliases()
		if err != nil {
			return err
		}
		*reply = *tmMap[al.GetId()]
	}
	return nil
}

type AttrGetTPAliasIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries alias identities on specific tariff plan.
func (self *ApierV1) GetTPAliasIds(attrs AttrGetTPAliasIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ALIASES, utils.TPDistinctIds{"direction", "tenant", "category", "account", "subject", "context"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Alias on Tariff plan
func (self *ApierV1) RemTPAlias(attrs AttrGetTPAlias, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "AliasId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_ALIASES, attrs.TPid, map[string]string{"tag": attrs.AliasId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
