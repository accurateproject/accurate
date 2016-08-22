package v1

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Creates a new SharedGroups profile within a tariff plan
func (self *ApierV1) SetTPSharedGroups(attrs utils.TPSharedGroups, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "SharedGroupsId", "SharedGroups"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	/*for _, action := range attrs.SharedGroups {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:SharedGroup:%s:%v", utils.ERR_MANDATORY_IE_MISSING, action.Identifier, missing)
		}
	}*/
	sg := engine.APItoModelSharedGroup(&attrs)
	if err := self.StorDb.SetTpSharedGroups(sg); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPSharedGroups struct {
	TPid           string // Tariff plan id
	SharedGroupsId string // SharedGroup id
}

// Queries specific SharedGroup on tariff plan
func (self *ApierV1) GetTPSharedGroups(attrs AttrGetTPSharedGroups, reply *utils.TPSharedGroups) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "SharedGroupsId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if sgs, err := self.StorDb.GetTpSharedGroups(attrs.TPid, attrs.SharedGroupsId); err != nil {
		return utils.NewErrServerError(err)
	} else if len(sgs) == 0 {
		return utils.ErrNotFound
	} else {
		sgMap, err := engine.TpSharedGroups(sgs).GetSharedGroups()
		if err != nil {
			return err
		}
		*reply = utils.TPSharedGroups{TPid: attrs.TPid, SharedGroupsId: attrs.SharedGroupsId, SharedGroups: sgMap[attrs.SharedGroupsId]}
	}
	return nil
}

type AttrGetTPSharedGroupIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries SharedGroups identities on specific tariff plan.
func (self *ApierV1) GetTPSharedGroupIds(attrs AttrGetTPSharedGroupIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_SHARED_GROUPS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific SharedGroups on Tariff plan
func (self *ApierV1) RemTPSharedGroups(attrs AttrGetTPSharedGroups, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "SharedGroupsId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_SHARED_GROUPS, attrs.TPid, map[string]string{"tag": attrs.SharedGroupsId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
