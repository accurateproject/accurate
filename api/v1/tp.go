package v1

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func (api *ApiV1) getTpReader() *engine.TpReader {
	if api.tpReader == nil {
		api.tpReader = engine.NewTpReader(api.ratingDB, api.accountDB, *api.cfg.General.DefaultTimezone)
	}
	return api.tpReader
}

func (api *ApiV1) SetTpDestination(tp utils.TpDestination, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Code", "Tag"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadDestination(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpTiming(tp utils.TpTiming, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadTiming(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpRate(tp utils.TpRate, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag", "Slots"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadRate(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpDestinationRate(tp utils.TpDestinationRate, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag", "Bindings"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadDestinationRate(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpRatingPlan(tp utils.TpRatingPlan, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag", "Bindings"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadRatingPlan(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpRatingProfile(tp utils.TpRatingProfile, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Direction", "Category", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadRatingProfile(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpSharedGroup(tp utils.TpSharedGroup, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag", "MemberIDs"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadSharedGroup(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpLCRule(tp utils.TpLcrRule, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Direction", "Category", "Account", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadLCR(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpActionGroup(tp utils.TpActionGroup, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadActionGroup(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpActionPlan(tp utils.TpActionPlan, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag", "ActionTimings"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadActionPlan(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpActionTrigger(tp utils.TpActionTrigger, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag", "Triggers"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadActionTrigger(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpAccountAction(tp utils.TpAccountAction, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadAccountAction(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpDerivedCharger(tp utils.TpDerivedCharger, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Direction", "Category", "Account", "Subject"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadDerivedCharger(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpCdrStats(tp utils.TpCdrStats, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadCdrStats(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpUser(tp utils.TpUser, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadUser(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpAlias(tp utils.TpAlias, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Direction", "Category", "Account", "Subject", "Context"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadAlias(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

func (api *ApiV1) SetTpResourceLimit(tp utils.TpResourceLimit, reply *string) (err error) {
	if missing := utils.MissingStructFields(&tp, []string{"Tenant", "Tag"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.getTpReader().LoadResourceLimit(&tp); err != nil {
		*reply = err.Error()
	}
	return err
}

type AttrRemoveTenant struct {
	Tenant string
}

func (api *ApiV1) RemoveTpAllTenant(attr AttrRemoveTenant, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.ratingDB.RemoveTenant(attr.Tenant, utils.TariffPlanDB); err != nil {
		*reply = err.Error()
		return err
	}
	if err = api.accountDB.RemoveTenant(attr.Tenant, utils.DataDB); err != nil {
		*reply = err.Error()
		return err
	}
	return nil
}

func (api *ApiV1) RemoveTpTariffPlanTenant(attr AttrRemoveTenant, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.ratingDB.RemoveTenant(attr.Tenant, utils.TariffPlanDB); err != nil {
		*reply = err.Error()
		return err
	}
	if err = api.accountDB.RemoveTenant(attr.Tenant, utils.DataDB); err != nil {
		*reply = err.Error()
		return err
	}
	return nil
}
func (api *ApiV1) RemoveTpDataTenant(attr AttrRemoveTenant, reply *string) (err error) {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	*reply = OK
	if err = api.ratingDB.RemoveTenant(attr.Tenant, utils.TariffPlanDB); err != nil {
		*reply = err.Error()
		return err
	}
	if err = api.accountDB.RemoveTenant(attr.Tenant, utils.DataDB); err != nil {
		*reply = err.Error()
		return err
	}
	return nil
}
