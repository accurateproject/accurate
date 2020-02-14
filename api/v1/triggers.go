package v1

import (
	"math"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Returns a list of ActionTriggers on an account
func (api *ApiV1) GetAccountActionTriggers(attrs AttrAcntAction, reply *[]string) error {
	if missing := utils.MissingStructFields(&attrs, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if account, err := api.accountDB.GetAccount(attrs.Tenant, attrs.Account); err != nil {
		return utils.NewErrServerError(err)
	} else {
		ats := account.TriggerIDs
		if ats == nil {
			ats = utils.StringMap{}
		}
		*reply = ats.Slice()
	}
	return nil
}

type AttrAddAccountActionTriggers struct {
	Tenant                 string
	Account                string
	ActionTriggerIDs       *[]string
	ActionTriggerOverwrite bool
	ActivationDate         string
	Executed               bool
}

func (api *ApiV1) AddAccountActionTriggers(attr AttrAddAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	actTime, err := utils.ParseTimeDetectLayout(attr.ActivationDate, *api.cfg.General.DefaultTimezone)
	if err != nil {
		*reply = err.Error()
		return err
	}
	var account *engine.Account
	_, err = engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := api.accountDB.GetAccount(attr.Tenant, attr.Account); err == nil {
			account = acc
		} else {
			return 0, err
		}
		if attr.ActionTriggerIDs != nil {
			if attr.ActionTriggerOverwrite || account.TriggerIDs == nil {
				account.TriggerIDs = utils.StringMap{}
			}
			for _, actionTriggerID := range *attr.ActionTriggerIDs {
				atrg, err := api.ratingDB.GetActionTriggers(attr.Tenant, actionTriggerID, utils.CACHED)
				if err != nil {
					return 0, err
				}

				account.TriggerIDs.Add(actionTriggerID)
				account.InitTriggerRecords()
				for _, atr := range atrg.ActionTriggers {
					triggerRecord := account.TriggerRecords[atr.UniqueID]
					triggerRecord.ActivationDate = actTime
					triggerRecord.Executed = attr.Executed
				}
			}
		}
		account.InitCounters()
		if err := api.accountDB.SetAccount(account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, attr.Account)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrRemoveAccountActionTriggers struct {
	Tenant  string
	Account string
	GroupID string
}

func (api *ApiV1) RemoveAccountActionTriggers(attr AttrRemoveAccountActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		var account *engine.Account
		if acc, err := api.accountDB.GetAccount(attr.Tenant, attr.Account); err == nil {
			account = acc
		} else {
			return 0, err
		}
		newActionTriggers := utils.StringMap{}

		for atID := range account.TriggerIDs {
			if attr.GroupID == "" || atID == attr.GroupID {
				// remove action trigger
				continue
			}
			newActionTriggers.Add(atID)
		}
		account.TriggerIDs = newActionTriggers
		account.InitCounters()
		if err := api.accountDB.SetAccount(account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, attr.Account)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

type AttrResetAccountActionTriggers struct {
	Tenant   string
	Account  string
	GroupID  string
	UniqueID string
	Executed bool
}

func (api *ApiV1) ResetAccountActionTriggers(attr AttrResetAccountActionTriggers, reply *string) error {

	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var account *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := api.accountDB.GetAccount(attr.Tenant, attr.Account); err == nil {
			account = acc
		} else {
			return 0, err
		}
		for key, atRecord := range account.TriggerRecords {
			if (attr.UniqueID == "" || atRecord.UniqueID == attr.UniqueID) &&
				(attr.GroupID == "" || key == attr.GroupID) {
				// reset action trigger
				atRecord.Executed = attr.Executed
			}

		}
		if attr.Executed == false {
			account.ExecuteActionTriggers(nil, false)
		}
		if err := api.accountDB.SetAccount(account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, attr.Account)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}

/*
type AttrSetAccountActionTriggers struct {
	Tenant                string
	Account               string
	GroupID               string
	UniqueID              string
	ThresholdType         *string
	ThresholdValue        *float64
	Recurrent             *bool
	Executed              *bool
	MinSleep              *string
	ExpirationDate        *string
	ActivationDate        *string
	BalanceID             *string
	BalanceType           *string
	BalanceDirections     *[]string
	BalanceDestinationIds *[]string
	BalanceWeight         *float64
	BalanceExpirationDate *string
	BalanceTimingTags     *[]string
	BalanceRatingSubject  *string
	BalanceCategories     *[]string
	BalanceSharedGroups   *[]string
	BalanceBlocker        *bool
	BalanceDisabled       *bool
	MinQueuedItems        *int
	ActionsID             *string
}

func (api *ApiV1) SetAccountActionTriggers(attr AttrSetAccountActionTriggers, reply *string) error {

	if missing := utils.MissingStructFields(&attr, []string{"Tenant", "Account"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	accID := utils.AccountKey(attr.Tenant, attr.Account)
	var account *engine.Account
	_, err := engine.Guardian.Guard(func() (interface{}, error) {
		if acc, err := api.accountDB.GetAccount(accID); err == nil {
			account = acc
		} else {
			return 0, err
		}
		for _, at := range account.ActionTriggers {
			if (attr.UniqueID == "" || at.UniqueID == attr.UniqueID) &&
				(attr.GroupID == "" || at.ID == attr.GroupID) {
				// we have a winner
				if attr.ThresholdType != nil {
					at.ThresholdType = *attr.ThresholdType
				}
				if attr.ThresholdValue != nil {
					at.ThresholdValue = *attr.ThresholdValue
				}
				if attr.Recurrent != nil {
					at.Recurrent = *attr.Recurrent
				}
				if attr.Executed != nil {
					at.Executed = *attr.Executed
				}
				if attr.MinSleep != nil {
					minSleep, err := utils.ParseDurationWithSecs(*attr.MinSleep)
					if err != nil {
						return 0, err
					}
					at.MinSleep = minSleep
				}
				if attr.ExpirationDate != nil {
					expTime, err := utils.ParseTimeDetectLayout(*attr.ExpirationDate, *api.cfg.General.DefaultTimezone)
					if err != nil {
						return 0, err
					}
					at.ExpirationDate = expTime
				}
				if attr.ActivationDate != nil {
					actTime, err := utils.ParseTimeDetectLayout(*attr.ActivationDate, *api.cfg.General.DefaultTimezone)
					if err != nil {
						return 0, err
					}
					at.ActivationDate = actTime
				}
				at.Balance = &engine.BalanceFilter{}
				if attr.BalanceID != nil {
					at.Balance.ID = attr.BalanceID
				}
				if attr.BalanceType != nil {
					at.Balance.Type = attr.BalanceType
				}
				if attr.BalanceDirections != nil {
					at.Balance.Directions = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDirections...))
				}
				if attr.BalanceDestinationIds != nil {
					at.Balance.DestinationIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDestinationIds...))
				}
				if attr.BalanceWeight != nil {
					at.Balance.Weight = attr.BalanceWeight
				}
				if attr.BalanceExpirationDate != nil {
					balanceExpTime, err := utils.ParseDate(*attr.BalanceExpirationDate)
					if err != nil {
						return 0, err
					}
					at.Balance.ExpirationDate = &balanceExpTime
				}
				if attr.BalanceTimingTags != nil {
					at.Balance.TimingIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceTimingTags...))
				}
				if attr.BalanceRatingSubject != nil {
					at.Balance.RatingSubject = attr.BalanceRatingSubject
				}
				if attr.BalanceCategories != nil {
					at.Balance.Categories = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceCategories...))
				}
				if attr.BalanceSharedGroups != nil {
					at.Balance.SharedGroups = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceSharedGroups...))
				}
				if attr.BalanceBlocker != nil {
					at.Balance.Blocker = attr.BalanceBlocker
				}
				if attr.BalanceDisabled != nil {
					at.Balance.Disabled = attr.BalanceDisabled
				}
				if attr.MinQueuedItems != nil {
					at.MinQueuedItems = *attr.MinQueuedItems
				}
				if attr.ActionsID != nil {
					at.ActionsID = *attr.ActionsID
				}
			}

		}
		account.ExecuteActionTriggers(nil, false)
		if err := api.accountDB.SetAccount(account); err != nil {
			return 0, err
		}
		return 0, nil
	}, 0, accID)
	if err != nil {
		*reply = err.Error()
		return err
	}
	*reply = utils.OK
	return nil
}
*/
type AttrRemoveActionTriggers struct {
	Tenant    string
	GroupIDs  []string
	UniqueIDs []string
}

func (api *ApiV1) RemoveActionTriggers(attr AttrRemoveActionTriggers, reply *string) error {
	if missing := utils.MissingStructFields(&attr, []string{"Tenant"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if len(attr.UniqueIDs) == 0 {
		for _, groupID := range attr.GroupIDs {
			err := api.ratingDB.RemoveActionTriggers(attr.Tenant, groupID)
			if err != nil {
				*reply = err.Error()
				return err
			}
		}
	} else {
		for _, groupID := range attr.GroupIDs {
			atrg, err := api.ratingDB.GetActionTriggers(attr.Tenant, groupID, utils.CACHED)
			if err != nil {
				*reply = err.Error()
				return err
			}
			var remainingAtrs engine.ActionTriggers
			for _, atr := range atrg.ActionTriggers {
				if utils.IsSliceMember(attr.UniqueIDs, atr.UniqueID) {
					continue
				}
				remainingAtrs = append(remainingAtrs, atr)
			}
			// set the cleared list back
			atrg.ActionTriggers = remainingAtrs
			err = api.ratingDB.SetActionTriggers(atrg)
			if err != nil {
				*reply = err.Error()
				return err
			}
		}

	}
	*reply = utils.OK
	return nil
}

/*
type AttrSetActionTrigger struct {
	GroupID               string
	UniqueID              string
	ThresholdType         *string
	ThresholdValue        *float64
	Recurrent             *bool
	MinSleep              *string
	ExpirationDate        *string
	ActivationDate        *string
	BalanceID             *string
	BalanceType           *string
	BalanceDirections     *[]string
	BalanceDestinationIds *[]string
	BalanceWeight         *float64
	BalanceExpirationDate *string
	BalanceTimingTags     *[]string
	BalanceRatingSubject  *string
	BalanceCategories     *[]string
	BalanceSharedGroups   *[]string
	BalanceBlocker        *bool
	BalanceDisabled       *bool
	MinQueuedItems        *int
	ActionsID             *string
}

func (api *ApiV1) SetActionTrigger(attr AttrSetActionTrigger, reply *string) error {

	if missing := utils.MissingStructFields(&attr, []string{"GroupID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	atrs, _ := api.ratingDB.GetActionTriggers(attr.GroupID, utils.CACHED)
	var newAtr *engine.ActionTrigger
	if attr.UniqueID != "" {
		//search for exiting one
		for _, atr := range atrs {
			if atr.UniqueID == attr.UniqueID {
				newAtr = atr
				break
			}
		}
	}

	if newAtr == nil {
		newAtr = &engine.ActionTrigger{}
		atrs = append(atrs, newAtr)
	}
	newAtr.ID = attr.GroupID
	if attr.UniqueID != "" {
		newAtr.UniqueID = attr.UniqueID
	} else {
		newAtr.UniqueID = utils.GenUUID()
	}

	if attr.ThresholdType != nil {
		newAtr.ThresholdType = *attr.ThresholdType
	}
	if attr.ThresholdValue != nil {
		newAtr.ThresholdValue = *attr.ThresholdValue
	}
	if attr.Recurrent != nil {
		newAtr.Recurrent = *attr.Recurrent
	}
	if attr.MinSleep != nil {
		minSleep, err := utils.ParseDurationWithSecs(*attr.MinSleep)
		if err != nil {
			*reply = err.Error()
			return err
		}
		newAtr.MinSleep = minSleep
	}
	if attr.ExpirationDate != nil {
		expTime, err := utils.ParseTimeDetectLayout(*attr.ExpirationDate, *api.cfg.General.DefaultTimezone)
		if err != nil {
			*reply = err.Error()
			return err
		}
		newAtr.ExpirationDate = expTime
	}
	if attr.ActivationDate != nil {
		actTime, err := utils.ParseTimeDetectLayout(*attr.ActivationDate, *api.cfg.General.DefaultTimezone)
		if err != nil {
			*reply = err.Error()
			return err
		}
		newAtr.ActivationDate = actTime
	}
	newAtr.Balance = &engine.BalanceFilter{}
	if attr.BalanceID != nil {
		newAtr.Balance.ID = attr.BalanceID
	}
	if attr.BalanceType != nil {
		newAtr.Balance.Type = attr.BalanceType
	}
	if attr.BalanceDirections != nil {
		newAtr.Balance.Directions = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDirections...))
	}
	if attr.BalanceDestinationIds != nil {
		newAtr.Balance.DestinationIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceDestinationIds...))
	}
	if attr.BalanceWeight != nil {
		newAtr.Balance.Weight = attr.BalanceWeight
	}
	if attr.BalanceExpirationDate != nil {
		balanceExpTime, err := utils.ParseDate(*attr.BalanceExpirationDate)
		if err != nil {
			*reply = err.Error()
			return err
		}
		newAtr.Balance.ExpirationDate = &balanceExpTime
	}
	if attr.BalanceTimingTags != nil {
		newAtr.Balance.TimingIDs = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceTimingTags...))
	}
	if attr.BalanceRatingSubject != nil {
		newAtr.Balance.RatingSubject = attr.BalanceRatingSubject
	}
	if attr.BalanceCategories != nil {
		newAtr.Balance.Categories = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceCategories...))
	}
	if attr.BalanceSharedGroups != nil {
		newAtr.Balance.SharedGroups = utils.StringMapPointer(utils.NewStringMap(*attr.BalanceSharedGroups...))
	}
	if attr.BalanceBlocker != nil {
		newAtr.Balance.Blocker = attr.BalanceBlocker
	}
	if attr.BalanceDisabled != nil {
		newAtr.Balance.Disabled = attr.BalanceDisabled
	}
	if attr.MinQueuedItems != nil {
		newAtr.MinQueuedItems = *attr.MinQueuedItems
	}
	if attr.ActionsID != nil {
		newAtr.ActionsID = *attr.ActionsID
	}

	if err := api.ratingDB.SetActionTriggers(attr.GroupID, atrs); err != nil {
		*reply = err.Error()
		return err
	}
	//no cache for action triggers
	*reply = utils.OK
	return nil
}
*/

func (api *ApiV1) GetActionTriggers(attr AttrGetMultiple, reply *[]*engine.ActionTriggerGroup) error {
	if len(attr.Tenant) == 0 {
		return utils.NewErrMandatoryIeMissing("Tenant")
	}

	var retActionTriggers []*engine.ActionTriggerGroup
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
		atrKeys := attr.IDs
		var limitedActionTriggers []string
		if limit != 0 {
			max := math.Min(float64(offset+limit), float64(len(atrKeys)))
			limitedActionTriggers = atrKeys[offset:int(max)]
		} else {
			limitedActionTriggers = atrKeys[offset:]
		}
		for _, atrName := range limitedActionTriggers {
			atrg, err := api.ratingDB.GetActionTriggers(attr.Tenant, atrName, utils.CACHED)
			if err != nil && err != utils.ErrNotFound { // Not found is not an error here
				return err
			}
			if atrg != nil {
				retActionTriggers = append(retActionTriggers, atrg)
			}
		}
	} else {
		retActionTriggers = make([]*engine.ActionTriggerGroup, 0)
		if err := api.accountDB.GetAllPaged(attr.Tenant, &retActionTriggers, engine.ColAtr, offset, limit); err != nil && err != utils.ErrNotFound {
			return err
		}
	}
	*reply = retActionTriggers
	return nil
}
