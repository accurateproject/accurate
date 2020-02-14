package v1

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

type AttrAcntAction struct {
	Tenant  string
	Account string
}

type AccountActionTiming struct {
	ActionPlanID string    // The id of the ActionPlanId profile attached to the account
	UUID         string    // The id to reference this particular ActionTiming
	ActionsID    string    // The id of actions which will be executed
	NextExecTime time.Time // Next execution time
}

func (api *ApiV1) GetAccountActionPlan(attr AttrAcntAction, reply *[]*AccountActionTiming) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(strings.Join(missing, ","), "")
	}
	accountATs := make([]*AccountActionTiming, 0) // needs to be initialized if remains empty
	iter := api.ratingDB.Iterator(engine.ColApb, "", map[string]interface{}{"tenant": attr.Tenant, "account": attr.Account})

	var binding engine.ActionPlanBinding
	for iter.Next(&binding) {
		ap, err := api.ratingDB.GetActionPlan(binding.Tenant, binding.ActionPlan, utils.CACHED)
		if err != nil {
			return utils.NewErrServerError(err)
		}
		for _, at := range ap.ActionTimings {
			accountATs = append(accountATs, &AccountActionTiming{
				ActionPlanID: ap.Name,
				UUID:         at.UUID,
				ActionsID:    at.ActionsID,
				NextExecTime: at.GetNextStartTime(time.Now()),
			})
		}

	}

	if err := iter.Close(); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = accountATs
	return nil
}

type AttrRemActionTiming struct {
	Tenant          string // Tenant it belongs to
	ActionPlanID    string // Id identifying the ActionPlan profile
	ActionTimingID  string // Internal id identifying particular ActionTiming, *all for all user related ActionTimings to be canceled
	Account         string // Account name
	ReloadScheduler bool   // If set it will reload the scheduler after adding
}

// Removes an ActionTimings or parts of it depending on filters being set
func (api *ApiV1) RemActionTiming(attr AttrRemActionTiming, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "ActionPlanID"}); len(missing) != 0 { // Only mandatory ActionPlanId
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if attr.Account != "" { // remove only account binding
			if err := api.ratingDB.RemoveActionPlanBindings(attr.Tenant, attr.Account, attr.ActionPlanID); err != nil {
				*reply = err.Error()
				return 0, err
			}
		}

		if attr.ActionTimingID != "" { // delete only a action timing from action plan
			ap, err := api.ratingDB.GetActionPlan(attr.Tenant, attr.ActionPlanID, utils.CACHED)
			if err != nil {
				return 0, utils.NewErrServerError(err)
			} else if ap == nil {
				return 0, utils.ErrNotFound
			}

			for i, at := range ap.ActionTimings {
				if at.UUID == attr.ActionTimingID {
					ap.ActionTimings[i] = ap.ActionTimings[len(ap.ActionTimings)-1]
					ap.ActionTimings = ap.ActionTimings[:len(ap.ActionTimings)-1]
					break
				}
			}
			if err := api.ratingDB.SetActionPlan(ap); err != nil {
				*reply = err.Error()
				return 0, err
			}
		}

		if attr.ActionPlanID != "" { // delete the entire action plan
			if err := api.ratingDB.SetActionPlan(&engine.ActionPlan{Tenant: attr.Tenant, Name: attr.ActionPlanID}); err != nil { // nil ActionTimings will delete the action plan
				*reply = err.Error()
				return 0, err
			}
			// delete all bindings
			if err := api.ratingDB.RemoveActionPlanBindings(attr.Tenant, "", attr.ActionPlanID); err != nil {
				*reply = err.Error()
				return 0, err
			}
		}
		return 0, nil
	}, 0, utils.ACTION_PLAN_PREFIX)
	if err != nil {
		*reply = err.Error()
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler && api.sched != nil {
		api.sched.Reload(true)
	}
	*reply = OK
	return nil
}

type AttrRemoveAccount struct {
	Tenant          string
	Account         string
	ReloadScheduler bool
}

func (api *ApiV1) RemoveAccount(attr AttrRemoveAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		// remove it from all action plans
		_, err := engine.Guardian.Guard(func() (interface{}, error) {
			if err := api.ratingDB.RemoveActionPlanBindings(attr.Tenant, attr.Account, ""); err != nil {
				return 0, err
			}
			return 0, nil
		}, 0, utils.ACTION_PLAN_PREFIX)
		if err != nil {
			return 0, err
		}

		if err := api.accountDB.RemoveAccount(attr.Tenant, attr.Account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, attr.Account)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler {
		// reload scheduler
		if api.sched != nil {
			api.sched.Reload(true)
		}
	}
	*reply = OK
	return nil
}

type AttrSetBalance struct {
	Tenant     string
	Account    string
	TOR        string // *monetary, *voice, *sms (BalanceType)
	ExecFilter string
	Filter     string
	Params     string
	Overwrite  bool // When true it will reset if the balance is already there
}

func (api *ApiV1) AddBalance(attr AttrSetBalance, reply *string) error {
	return api.modifyBalance(engine.TOPUP, attr, reply)
}
func (api *ApiV1) DebitBalance(attr AttrSetBalance, reply *string) error {
	return api.modifyBalance(engine.DEBIT, attr, reply)
}

func (api *ApiV1) SetBalance(attr AttrSetBalance, reply *string) error {
	return api.modifyBalance(engine.SET_BALANCE, attr, reply)
}

func (api *ApiV1) RemoveBalances(attr AttrSetBalance, reply *string) error {
	return api.modifyBalance(engine.REMOVE_BALANCE, attr, reply)
}

func (api *ApiV1) modifyBalance(aType string, attr AttrSetBalance, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"Tenant", "Account", "Filter", "Params"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	if _, err := api.accountDB.GetAccount(attr.Tenant, attr.Account); err != nil {
		// create account if not exists
		if aType != engine.REMOVE_BALANCE {
			if err := api.accountDB.SetAccount(&engine.Account{
				Tenant: attr.Tenant,
				Name:   attr.Account,
			}); err != nil {
				*reply = err.Error()
				return err
			}
		} else {
			*reply = "account not found"
			return utils.ErrNotFound
		}
	}

	at := &engine.ActionTiming{}
	// set parent action plan for Tenant
	apl := &engine.ActionPlan{
		Tenant:        attr.Tenant,
		ActionTimings: []*engine.ActionTiming{at},
	}
	apl.SetParentActionPlan()
	at.SetAccountIDs(utils.StringMap{attr.Account: true})

	if attr.Overwrite {
		aType += "_reset" // => *topup_reset/*debit_reset
	}
	at.SetActions(engine.Actions{&engine.Action{
		ActionType: aType,
		TOR:        attr.TOR,
		Params:     attr.Params,
		ExecFilter: attr.ExecFilter,
		Filter1:    attr.Filter,
	}})
	if err := at.Execute(); err != nil {
		*reply = err.Error()
		return err
	}
	*reply = OK
	return nil
}

type AttrGetMultiple struct {
	Tenant string
	IDs    []string
	utils.Paginator
}

func (api *ApiV1) GetAccounts(attr AttrGetMultiple, reply *[]*engine.Account) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenant")
	}

	retAccounts := make([]*engine.Account, 0)
	var offset, limit int
	if attr.Offset != nil {
		offset = *attr.Offset
		if offset < 0 {
			offset = 0
		}
	}
	if attr.Limit != nil {
		limit = *attr.Limit
		if limit <= 0 {
			limit = 1
		}
	}
	if len(attr.IDs) != 0 {
		accountKeys := attr.IDs
		var limitedAccounts []string
		if limit != 0 {
			max := math.Min(float64(offset+limit), float64(len(accountKeys)))
			limitedAccounts = accountKeys[offset:int(max)]
		} else {
			limitedAccounts = accountKeys[offset:]
		}
		err := api.accountDB.GetByNames(attr.Tenant, limitedAccounts, &retAccounts, engine.ColAcc)
		if err != nil && err != utils.ErrNotFound { // Not found is not an error here
			return err
		}
	} else {
		if err := api.accountDB.GetAllPaged(attr.Tenant, &retAccounts, engine.ColAcc, offset, limit); err != nil && err != utils.ErrNotFound {
			return err
		}
	}
	*reply = retAccounts
	return nil
}

type AttrGetAccount struct {
	Tenant  string
	Account string
}

func (api *ApiV1) GetAccount(attr AttrGetAccount, reply *engine.Account) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	account, err := api.accountDB.GetAccount(attr.Tenant, attr.Account)
	if err != nil {
		return utils.NewErrServerError(err)
	}

	*reply = *account
	return nil
}

type AttrSetAccount struct {
	Tenant                 string
	Account                string
	ActionPlanIDs          *[]string
	ActionPlansOverwrite   bool
	ActionTriggerIDs       *[]string
	ActionTriggerOverwrite bool
	AllowNegative          *bool
	Disabled               *bool
	ReloadScheduler        bool
}

func (api *ApiV1) SetAccount(attr AttrSetAccount, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		// get account if exists for disabled and allow negative
		acc, err := api.accountDB.GetAccount(attr.Tenant, attr.Account)
		if err != nil && err != utils.ErrNotFound {
			return 0, err
		}
		if acc == nil {
			acc = &engine.Account{
				Tenant: attr.Tenant,
				Name:   attr.Account,
			}
		}
		if attr.AllowNegative != nil {
			acc.AllowNegative = *attr.AllowNegative
		}
		if attr.Disabled != nil {
			acc.Disabled = *attr.Disabled
		}

		if attr.ActionPlanIDs != nil {
			_, err := engine.Guardian.Guard(func() (interface{}, error) {
				for _, actionPlanID := range *attr.ActionPlanIDs {
					ap, err := api.ratingDB.GetActionPlan(attr.Tenant, actionPlanID, utils.CACHED)
					if err != nil || ap == nil {
						return 0, fmt.Errorf("error getting action plan: %s (%s)", actionPlanID, err)
					}
					// check exitsing account binding

					binding, err := api.ratingDB.GetActionPlanBinding(attr.Tenant, attr.Account, actionPlanID)
					if err != nil && err != utils.ErrNotFound {
						return 0, err
					}
					if binding == nil { // no previous biniding to this action plan
						// create tasks
						for _, at := range ap.ActionTimings {
							if at.IsASAP() {
								t := &engine.Task{
									UUID:      utils.GenUUID(),
									AccountID: attr.Account,
									ActionsID: at.ActionsID,
								}
								if err = api.ratingDB.PushTask(t); err != nil {
									return 0, err
								}
							}
						}
					}
				}
				if attr.ActionPlansOverwrite {
					// clean previous action plan bindings
					if err := api.ratingDB.RemoveActionPlanBindings(attr.Tenant, attr.Account, ""); err != nil {
						return err, err
					}
				}
				// add new binings
				for _, actionPlanID := range *attr.ActionPlanIDs {
					if err := api.ratingDB.SetActionPlanBinding(&engine.ActionPlanBinding{
						Tenant:     attr.Tenant,
						Account:    attr.Account,
						ActionPlan: actionPlanID,
					}); err != nil {
						return err, err
					}
				}
				return 0, nil
			}, 0, utils.ACTION_PLAN_PREFIX)
			if err != nil {
				return 0, err
			}
		}

		if attr.ActionTriggerIDs != nil {
			if attr.ActionTriggerOverwrite || acc.TriggerIDs == nil {
				acc.TriggerIDs = utils.StringMap{}
			}
			for _, actionTriggerID := range *attr.ActionTriggerIDs {
				_, err := api.ratingDB.GetActionTriggers(attr.Tenant, actionTriggerID, utils.CACHED)
				if err != nil {
					return 0, fmt.Errorf("error getting action trigger: %s (%s)", actionTriggerID, err)
				}
				acc.TriggerIDs[actionTriggerID] = true
			}
		}
		acc.InitCounters()

		// All prepared, save account
		if err := api.accountDB.SetAccount(acc); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, attr.Account)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if attr.ReloadScheduler {
		// reload scheduler
		if api.sched != nil {
			api.sched.Reload(true)
		}
	}
	*reply = utils.OK // This will mark saving of the account, error still can show up in actionTimingsId
	return nil
}
