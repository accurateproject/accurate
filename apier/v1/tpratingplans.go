package v1

// This file deals with tp_destrates_timing management over APIs

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Creates a new DestinationRateTiming profile within a tariff plan
func (self *ApierV1) SetTPRatingPlan(attrs utils.TPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId", "RatingPlanBindings"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rp := engine.APItoModelRatingPlan(&attrs)
	if err := self.StorDb.SetTpRatingPlans(rp); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = "OK"
	return nil
}

type AttrGetTPRatingPlan struct {
	TPid         string // Tariff plan id
	RatingPlanId string // Rate id
	utils.Paginator
}

// Queries specific RatingPlan profile on tariff plan
func (self *ApierV1) GetTPRatingPlan(attrs AttrGetTPRatingPlan, reply *utils.TPRatingPlan) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rps, err := self.StorDb.GetTpRatingPlans(attrs.TPid, attrs.RatingPlanId, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if len(rps) == 0 {
		return utils.ErrNotFound
	} else {
		rpsMap, err := engine.TpRatingPlans(rps).GetRatingPlans()
		if err != nil {
			return err
		}
		*reply = utils.TPRatingPlan{TPid: attrs.TPid, RatingPlanId: attrs.RatingPlanId, RatingPlanBindings: rpsMap[attrs.RatingPlanId]}
	}
	return nil
}

type AttrGetTPRatingPlanIds struct {
	TPid string // Tariff plan id
	utils.Paginator
}

// Queries RatingPlan identities on specific tariff plan.
func (self *ApierV1) GetTPRatingPlanIds(attrs AttrGetTPRatingPlanIds, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ids, err := self.StorDb.GetTpTableIds(attrs.TPid, utils.TBL_TP_RATING_PLANS, utils.TPDistinctIds{"tag"}, nil, &attrs.Paginator); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

// Removes specific RatingPlan on Tariff plan
func (self *ApierV1) RemTPRatingPlan(attrs AttrGetTPRatingPlan, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"TPid", "RatingPlanId"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := self.StorDb.RemTpData(utils.TBL_TP_RATING_PLANS, attrs.TPid, map[string]string{"tag": attrs.RatingPlanId}); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*reply = "OK"
	}
	return nil
}
