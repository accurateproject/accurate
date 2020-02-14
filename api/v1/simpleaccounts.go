package v1

import (
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

type AttrSetSimpleAccount struct {
	Tenant     string
	Account    string
	MaxBalance *float64
	Disabled   *bool
}

func (api *ApiV1) SimpleAccountSet(attr AttrSetSimpleAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sas := engine.GetSimpleAccounts()
	if sas == nil {
		*reply = "simple accounts not enabled"
		return utils.ErrNotImplemented
	}
	acc, err := sas.GetAccount(attr.Tenant, attr.Account)
	if err != nil && err != utils.ErrNotFound {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	if acc == nil {
		disabled := false
		if attr.Disabled != nil {
			disabled = *attr.Disabled
		}
		var maxBalance *dec.Dec = nil
		if attr.MaxBalance != nil {
			maxBalance = dec.NewFloat(*attr.MaxBalance)
		}

		err = sas.NewAccount(attr.Tenant, attr.Account, disabled, maxBalance)
		if err != nil {
			*reply = err.Error()
			return utils.NewErrServerError(err)
		}
	} else {
		if attr.Disabled != nil {
			sas.SetDisabled(attr.Tenant, attr.Account, *attr.Disabled)
		}
		if attr.MaxBalance != nil {
			sas.SetMaxBalance(attr.Tenant, attr.Account, dec.NewFloat(*attr.MaxBalance))
		}
	}
	*reply = utils.OK
	return nil
}

func (api *ApiV1) SimpleAccountGet(attr AttrGetAccount, reply *engine.SimpleAccount) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sas := engine.GetSimpleAccounts()
	if sas == nil {
		return utils.ErrNotImplemented
	}
	account, err := sas.GetAccount(attr.Tenant, attr.Account)
	if err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = *account
	return nil
}

func (api *ApiV1) SimpleAccountRemove(attr AttrGetAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sas := engine.GetSimpleAccounts()
	if sas == nil {
		return utils.ErrNotImplemented
	}
	err := sas.RemoveAccount(attr.Tenant, attr.Account)
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return nil
}

type AttrDebitSimpleAccount struct {
	Tenant   string
	Account  string
	Category string
	Value    float64
}

func (api *ApiV1) SimpleAccountDebit(attr AttrDebitSimpleAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account", "Category", "Value"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sas := engine.GetSimpleAccounts()
	if sas == nil {
		return utils.ErrNotImplemented
	}
	err := sas.Debit(attr.Tenant, attr.Account, attr.Category, dec.NewFloat(attr.Value))
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return nil
}

func (api *ApiV1) SimpleAccountSetValue(attr AttrDebitSimpleAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account", "Value"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sas := engine.GetSimpleAccounts()
	if sas == nil {
		return utils.ErrNotImplemented
	}
	err := sas.SetValue(attr.Tenant, attr.Account, attr.Category, dec.NewFloat(attr.Value))
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return nil
}

type AttrReloadSimpleAccounts struct {
	Tenant string
}

func (api *ApiV1) SimpleAccountsReload(attr AttrReloadSimpleAccounts, reply *string) error {
	sas := engine.GetSimpleAccounts()
	if sas == nil {
		return utils.ErrNotImplemented
	}
	sas.LoadFromStorage(attr.Tenant)
	*reply = utils.OK
	return nil
}
