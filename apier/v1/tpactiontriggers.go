
package v1

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Creates a new ActionTriggers profile within a tariff plan
func (self *ApierV1) SetTPActionTriggers(attrs utils.TPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs,
		[]string{"TPid", "ActionTriggersId"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, at := range attrs.ActionTriggers {
		at.Id = attrs.ActionTriggersId
	}
	ats := engine.APItoModelActionTrigger(&attrs)
	if err := self.StorDb.SetTpActionTriggers(ats); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPActionTriggers struct {
	TPid             string // Tariff plan id
	ActionTriggersId string // ActionTrigger id
}

// Queries specific ActionTriggers profile on tariff plan
func (self *ApierV1) GetTPActionTriggers(attrs AttrGetTPActionTriggers, reply *utils.TPActionTriggers) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTriggersId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ats, err := self.StorDb.GetTpActionTriggers(attrs.TPid, attrs.ActionTriggersId); err != nil {
		return utils.NewErrServerError(err)
	} else if len(ats) == 0 {
		return utils.ErrNotFound
	} else {
		atsMap, err := engine.TpActionTriggers(ats).GetActionTriggers()
		if err != nil {
			return err
		}
		atRply := &utils.TPActionTriggers{
			TPid:             attrs.TPid,
			ActionTriggersId: attrs.ActionTriggersId,
			ActionTriggers:   atsMap[attrs.ActionTriggersId],
		}
		*reply = *atRply
	}
	return nil
}

type AttrGetTPActionTriggerIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries ActionTriggers identities on specific tariff plan.
func (self *ApierV1) GetTPActionTriggerIds(attrs AttrGetTPActionTriggerIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ACTION_TRIGGERS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific ActionTriggers on Tariff plan
func (self *ApierV1) RemTPActionTriggers(attrs AttrGetTPActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionTriggersId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_ACTION_TRIGGERS, attrs.TPid, map[string]string{"tag": attrs.ActionTriggersId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
