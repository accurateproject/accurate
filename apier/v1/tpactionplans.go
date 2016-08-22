package v1

import (
	"fmt"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Creates a new ActionTimings profile within a tariff plan
func (self *ApierV1) SetTPActionPlan(attrs utils.TPActionPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "ActionPlanId", "ActionPlan"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	for _, at := range attrs.ActionPlan {
		requiredFields := []string{"ActionsId", "TimingId", "Weight"}
		if missing := utils.MissingStructFields(at, requiredFields); len(missing) != 0 {
			return fmt.Errorf("%s:Action:%s:%v", utils.ErrMandatoryIeMissing.Error(), at.ActionsId, missing)
		}
	}
	ap := engine.APItoModelActionPlan(&attrs)
	if err := self.StorDb.SetTpActionPlans(ap); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPActionPlan struct {
	TPid string // Tariff plan id
	Id   string // ActionPlans id
}

// Queries specific ActionPlan profile on tariff plan
func (self *ApierV1) GetTPActionPlan(attrs AttrGetTPActionPlan, reply *utils.TPActionPlan) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Id"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ats, err := self.StorDb.GetTpActionPlans(attrs.TPid, attrs.Id); err != nil {
		return utils.NewErrServerError(err)
	} else if len(ats) == 0 {
		return utils.ErrNotFound
	} else { // Got the data we need, convert it
		aps, err := engine.TpActionPlans(ats).GetActionPlans()
		if err != nil {
			return err
		}
		atRply := &utils.TPActionPlan{
			TPid:         attrs.TPid,
			ActionPlanId: attrs.Id,
			ActionPlan:   aps[attrs.Id],
		}
		*reply = *atRply
	}
	return nil
}

type AttrGetTPActionPlanIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries ActionPlan identities on specific tariff plan.
func (self *ApierV1) GetTPActionPlanIds(attrs AttrGetTPActionPlanIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_ACTION_PLANS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific ActionPlan on Tariff plan
func (self *ApierV1) RemTPActionPlan(attrs AttrGetTPActionPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "Id"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_ACTION_PLANS, attrs.TPid, map[string]string{"tag": attrs.Id}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
