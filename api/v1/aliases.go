package v1

import (
	"errors"
	"fmt"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

type AttrAddAccountAliases struct {
	Tenant, Category, Account string
	Aliases                   []string
}

func (api *ApiV1) AddAccountAliases(attrs AttrAddAccountAliases, reply *string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account", "Aliases"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if attrs.Category == "" {
		attrs.Category = utils.CALL
	}
	aliases := engine.GetAliasService()
	if aliases == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}
	var ignr string
	for _, alias := range attrs.Aliases {
		als := engine.Alias{Direction: utils.META_OUT, Tenant: attrs.Tenant, Category: attrs.Category, Account: alias, Subject: alias, Context: utils.ALIAS_CONTEXT_RATING,
			Values: engine.AliasValues{&engine.AliasValue{
				DestinationID: utils.META_ANY,
				Fields:        fmt.Sprintf(`{"Account":"%s", "Subject":"%s"}`, attrs.Account, attrs.Account),
				Weight:        10.0}}}
		if err := aliases.Call("AliasesV1.SetAlias", &engine.AttrAddAlias{Alias: &als}, &ignr); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	*reply = utils.OK
	return nil
}

// Remove aliases configured for an account
func (api *ApiV1) RemAccountAliases(attr utils.TenantAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	aliases := engine.GetAliasService()
	if aliases == nil {
		return errors.New("ALIASES_NOT_ENABLED")
	}

	var r string
	if err := aliases.Call("AliasesV1.RemoveAliasByFieldValue", engine.AttrReverseAlias{Tenant: attr.Tenant, Context: utils.ALIAS_CONTEXT_RATING, Target: "Account", Alias: attr.Account}, &r); err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = utils.OK
	return nil
}
