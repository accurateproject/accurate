
package v1

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new Actions profile within a tariff plan
func (self *ApierV1) SetTPActions(attrs utils.TPActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId", "Actions"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, action := range attrs.Actions {
		requiredFields := []string{"Identifier", "Weight"}
		if action.BalanceType != "" { // Add some inter-dependent parameters - if balanceType then we are not talking about simply calling actions
			requiredFields = append(requiredFields, "Direction", "Units")
		}
		if missing := utils.MissingStructFields(action, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), action.Identifier, missing)
		}
	}
	as := engine.APItoModelAction(&attrs)
	if err := self.StorDb.SetTpActions(as); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPActions struct {
	TPid      string // Tariff plan id
	ActionsId string // Actions id
}

// Queries specific Actions profile on tariff plan
func (self *ApierV1) GetTPActions(attrs AttrGetTPActions, reply *utils.TPActions) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if acts, err := self.StorDb.GetTpActions(attrs.TPid, attrs.ActionsId); err != nil {
		return utils.NewErrServerError(err)
	} else if len(acts) == 0 {
		return utils.ErrNotFound
	} else {
		as, err := engine.TpActions(acts).GetActions()
		if err != nil {

		}
		*reply = utils.TPActions{TPid: attrs.TPid, ActionsId: attrs.ActionsId, Actions: as[attrs.ActionsId]}
	}
	return nil
}

type AttrGetTPActionIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries Actions identities on specific tariff plan.
func (self *ApierV1) GetTPActionIds(attrs AttrGetTPActionIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ACTIONS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific Actions on Tariff plan
func (self *ApierV1) RemTPActions(attrs AttrGetTPActions, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionsId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_ACTIONS, attrs.TPid, map[string]string{"tag": attrs.ActionsId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
