package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"

	"strings"
)

/*
Structure containing information about user's credit (minutes, cents, sms...).'
This can represent a user or a shared group.
*/
type Account struct {
	Tenant            string                          `bson:"tenant"`
	Name              string                          `bson:"name"`
	BalanceMap        map[string]Balances             `bson:"balance_map"`
	UnitCounters      UnitCounters                    `bson:"unit_counters"`
	TriggerIDs        utils.StringMap                 `bson:"trigger_ids"` // trigger groups ids
	TriggerRecords    map[string]*ActionTriggerRecord `bson:"trigger_records"`
	AllowNegative     bool                            `bson:"allow_negative"`
	Disabled          bool                            `bson:"disabled"`
	executingTriggers bool
	triggers          ActionTriggers
}

func (acc *Account) getTriggers() ActionTriggers {
	if acc.triggers != nil {
		return acc.triggers
	}
	for atgName := range acc.TriggerIDs {
		atg, err := ratingStorage.GetActionTriggers(acc.Tenant, atgName, utils.CACHED)
		if err != nil || atg == nil {
			utils.Logger.Error("error getting triggers for ID: ", zap.String("id", atgName), zap.Error(err))
			continue
		}
		atg.SetParentGroup()
		acc.triggers = append(acc.triggers, atg.ActionTriggers...)
	}
	acc.triggers.Sort()

	if len(acc.triggers) > 0 && acc.TriggerRecords == nil {
		acc.TriggerRecords = make(map[string]*ActionTriggerRecord)
	}
	acc.InitTriggerRecords()

	return acc.triggers
}

// User's available minutes for the specified destination
func (ub *Account) getCreditForPrefix(cd *CallDescriptor) (duration time.Duration, credit *dec.Dec, balances Balances) {
	creditBalances := ub.getBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, utils.MONETARY, "")

	unitBalances := ub.getBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, cd.TOR, "")
	//log.Printf("Credit: %v Unit: %v", creditBalances, unitBalances)
	// gather all balances from shared groups
	var extendedCreditBalances Balances
	for _, cb := range creditBalances {
		if len(cb.SharedGroups) > 0 {
			for sg := range cb.SharedGroups {
				if sharedGroup, _ := ratingStorage.GetSharedGroup(ub.Tenant, sg, utils.CACHED); sharedGroup != nil {
					sgb := sharedGroup.GetBalances(cd.Destination, cd.Category, cd.Direction, utils.MONETARY, ub)
					sgb = sharedGroup.SortBalancesByStrategy(cb, sgb)
					extendedCreditBalances = append(extendedCreditBalances, sgb...)
				}
			}
		} else {
			extendedCreditBalances = append(extendedCreditBalances, cb)
		}
	}
	var extendedMinuteBalances Balances
	for _, mb := range unitBalances {
		if len(mb.SharedGroups) > 0 {
			for sg := range mb.SharedGroups {
				if sharedGroup, _ := ratingStorage.GetSharedGroup(ub.Tenant, sg, utils.CACHED); sharedGroup != nil {
					sgb := sharedGroup.GetBalances(cd.Destination, cd.Category, cd.Direction, cd.TOR, ub)
					sgb = sharedGroup.SortBalancesByStrategy(mb, sgb)
					extendedMinuteBalances = append(extendedMinuteBalances, sgb...)
				}
			}
		} else {
			extendedMinuteBalances = append(extendedMinuteBalances, mb)
		}
	}
	credit = extendedCreditBalances.GetTotalValue()
	balances = extendedMinuteBalances
	for _, b := range balances {
		d, c := b.GetMinutesForCredit(cd, credit)
		credit = c
		duration += d
	}
	return
}

// sets all the fields of the balance
func (acc *Account) setBalanceAction(a *Action) error {
	if a == nil {
		return errors.New("nil action")
	}
	if a.TOR == "" {
		return errors.New("missing action tor")
	}
	balanceType := a.TOR
	if acc.BalanceMap == nil {
		acc.BalanceMap = make(map[string]Balances, 1)
	}
	var previousSharedGroups utils.StringMap // kept for comparison
	var balance *Balance
	var found bool

	for _, b := range acc.BalanceMap[balanceType] {
		if b.IsExpired() {
			continue
		}
		match, err := a.getFilter().Query(b, false)
		if err != nil {
			utils.Logger.Error(fmt.Sprintf("<set balance action> error matching balance: %s", err.Error()))
			return err
		}
		if match {
			previousSharedGroups = b.SharedGroups
			balance = b
			found = true
			break // only set one balance
		}
	}

	// if it is not found then we add it to the list
	if balance == nil {
		balance = &Balance{}
		balance.UUID = utils.GenUUID() // alway overwrite the uuid for consistency
		acc.BalanceMap[balanceType] = append(acc.BalanceMap[balanceType], balance)
	}
	newBalance, err := a.getBalance(nil)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("<set balance action> error unmarshalling new balance: %s", err.Error()))
		return err
	}
	if newBalance.ID == utils.META_DEFAULT {
		balance.ID = utils.META_DEFAULT
		balance.Value = newBalance.Value
	} else {
		if _, err := a.getBalance(balance); err != nil {
			return err
		}
	}

	if !found || !previousSharedGroups.Equal(balance.SharedGroups) {
		_, err := Guardian.Guard(func() (interface{}, error) {
			for sgID := range balance.SharedGroups {
				// add shared group member
				sg, err := ratingStorage.GetSharedGroup(acc.Tenant, sgID, utils.CACHED)
				if err != nil || sg == nil {
					//than is problem
					utils.Logger.Warn("Could not get shared group: ", zap.String("id", sgID))

				} else {
					if _, found := sg.MemberIDs[acc.Name]; !found {
						// add member and save
						if sg.MemberIDs == nil {
							sg.MemberIDs = make(utils.StringMap)
						}
						sg.MemberIDs[acc.Name] = true
						ratingStorage.SetSharedGroup(sg)
					}
				}
			}
			return 0, nil
		}, 0, balance.SharedGroups.Slice()...)
		if err != nil {
			return err
		}
	}
	acc.InitCounters()
	acc.ExecuteActionTriggers(nil, false)
	return nil
}

// Debits some amount of user's specified balance adding the balance if it does not exists.
// Returns the remaining credit in user's balance.
func (ub *Account) debitBalanceAction(a *Action, reset bool) error {
	//log.Print("Action: ", utils.ToIJSON(a))
	if a == nil {
		return errors.New("nil action")
	}
	bClone, err := a.getBalance(nil)
	//log.Print("bClone: ", utils.ToIJSON(bClone))
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("<debit balance action> error unmarshalling new balance: %s", err.Error()))
		return err
	}
	if bClone == nil {
		return errors.New("nil balance")
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]Balances, 1)
	}
	found := false
	for _, b := range ub.BalanceMap[a.TOR] {
		if b.IsExpired() {
			continue // just to be safe (cleaned expired balances above)
		}
		b.account = ub
		match, err := a.getFilter().Query(b, false)
		//log.Printf("match: %v b: %+v (%s)", match, b, a.ActionType)
		if err != nil {
			utils.Logger.Warn("<debitBalanceAction> action ", zap.String("filter", a.Filter1), zap.Error(err))
		}
		if match {
			if reset {
				b.SetValue(dec.Zero)
			}
			b.SubstractValue(bClone.GetValue())
			b.dirty = true
			found = true
			a.getBalanceValue().Set(b.GetValue())
		}
	}
	// if it is not found then we add it to the list
	if !found {
		// check if the Id is *default (user trying to create the default balance)
		// use only it's value value
		if bClone.ID == utils.META_DEFAULT {
			bClone = &Balance{
				ID:    utils.META_DEFAULT,
				Value: dec.New().Neg(bClone.GetValue()),
			}
		} else {
			if !bClone.GetValue().IsZero() {
				bClone.GetValue().Neg(bClone.GetValue())
			}
		}
		bClone.dirty = true // Mark the balance as dirty since we have modified and it should be checked by action triggers
		a.balanceValue = bClone.GetValue()
		bClone.UUID = utils.GenUUID() // alway overwrite the uuid for consistency
		// load ValueFactor if defined in extra parametrs
		if a.Params != "" && strings.Contains(a.Params, "ValueFactor") {
			x := struct {
				ValueFactor ValueFactor
			}{}
			err := json.Unmarshal([]byte(a.Params), &x)
			if err == nil {
				bClone.Factor = x.ValueFactor
			} else {
				utils.Logger.Warn("Could load value factor from actions: ", zap.String("params", a.Params))
			}
		}
		ub.BalanceMap[a.TOR] = append(ub.BalanceMap[a.TOR], bClone)
		if len(bClone.SharedGroups) > 0 {
			_, err := Guardian.Guard(func() (interface{}, error) {
				for sgID := range bClone.SharedGroups {
					// add shared group member
					sg, err := ratingStorage.GetSharedGroup(ub.Tenant, sgID, utils.CACHED)
					if err != nil || sg == nil {
						//than is problem
						utils.Logger.Warn("Could not get shared group: ", zap.String("id", sgID))
					} else {
						if _, found := sg.MemberIDs[ub.Name]; !found {
							// add member and save
							if sg.MemberIDs == nil {
								sg.MemberIDs = make(utils.StringMap)
							}
							sg.MemberIDs[ub.Name] = true
							ratingStorage.SetSharedGroup(sg)
						}
					}
				}
				return 0, nil
			}, 0, bClone.SharedGroups.Slice()...)
			if err != nil {
				return err
			}
		}
	}
	ub.InitCounters()
	ub.ExecuteActionTriggers(nil, false)
	return nil
}

func (ub *Account) getBalancesForPrefix(prefix, category, direction, tor string, sharedGroup string) Balances {
	var balances Balances
	balances = append(balances, ub.BalanceMap[tor]...)
	if tor != utils.MONETARY && tor != utils.GENERIC {
		balances = append(balances, ub.BalanceMap[utils.GENERIC]...)
	}
	var usefulBalances Balances
	for _, b := range balances {

		if b.Disabled {
			continue
		}
		if b.IsExpired() || (len(b.SharedGroups) == 0 && b.GetValue().LteZero() && !b.Blocker) {
			continue
		}
		if sharedGroup != "" && b.SharedGroups[sharedGroup] == false {
			continue
		}
		if !b.MatchCategory(category) {
			continue
		}
		if b.HasDirection() && b.Directions[direction] == false {
			continue
		}
		b.account = ub

		if len(b.DestinationIDs) > 0 && b.DestinationIDs[utils.ANY] == false {
			if dests, err := ratingStorage.GetDestinations(ub.Tenant, prefix, "", utils.DestMatching, utils.CACHED); err == nil {
				//log.Print("HERE: ", prefix, utils.ToJSON(dests), b.DestinationIDs)
				foundPrefix := ""
				allInclude := true // whether it is excluded or included
				for _, dest := range dests {
					dID := dest.Name
					inclDest, found := b.DestinationIDs[dID]
					if found {
						foundPrefix = dest.Code
						allInclude = allInclude && inclDest
					}
				}
				//log.Print("foundprefix: ", foundPrefix, allInclude)
				// check wheter all destination ids in the balance were exclusions
				allExclude := true
				for _, inclDest := range b.DestinationIDs {
					if inclDest {
						allExclude = false
						break
					}
				}
				if foundPrefix != "" || allExclude {
					if allInclude {
						b.precision = len(foundPrefix)
						usefulBalances = append(usefulBalances, b)
					} else {
						b.precision = 1 // fake to exit the outer loop
					}
				}
			}
		} else {
			usefulBalances = append(usefulBalances, b)
		}
	}
	// resort by precision
	usefulBalances.Sort()
	// clear precision
	for _, b := range usefulBalances {
		b.precision = 0
	}
	return usefulBalances
}

// like getBalancesForPrefix but expanding shared balances
func (account *Account) getAlldBalancesForPrefix(destination, category, direction, balanceType string) (bc Balances) {
	balances := account.getBalancesForPrefix(destination, category, direction, balanceType, "")
	for _, b := range balances {
		if len(b.SharedGroups) > 0 {
			for sgID := range b.SharedGroups {
				sharedGroup, err := ratingStorage.GetSharedGroup(account.Tenant, sgID, utils.CACHED)
				if err != nil || sharedGroup == nil {
					utils.Logger.Warn("Could not get shared group: ", zap.String("id", sgID))
					continue
				}
				sharedBalances := sharedGroup.GetBalances(destination, category, direction, balanceType, account)
				sharedBalances = sharedGroup.SortBalancesByStrategy(b, sharedBalances)
				bc = append(bc, sharedBalances...)
			}
		} else {
			bc = append(bc, b)
		}
	}
	return
}

func (ub *Account) debitCreditBalance(cd *CallDescriptor, count bool, dryRun bool, goNegative bool) (cc *CallCost, err error) {
	usefulUnitBalances := ub.getAlldBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, cd.TOR)
	usefulMoneyBalances := ub.getAlldBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, utils.MONETARY)
	//utils.Logger.Debug(fmt.Sprintf("%+v, %+v", usefulMoneyBalances, usefulUnitBalances))
	//log.Print("STARTCD: ", utils.ToIJSON(cd), dryRun)
	//log.Printf("%s, %s", utils.ToIJSON(usefulMoneyBalances), utils.ToIJSON(usefulUnitBalances))
	var leftCC *CallCost
	cc = cd.CreateCallCost()

	generalBalanceChecker := true
	for generalBalanceChecker {
		generalBalanceChecker = false

		// debit minutes
		unitBalanceChecker := true
		for unitBalanceChecker {
			// try every balance multiple times in case one becomes active or ratig changes
			unitBalanceChecker = false
			//log.Printf("InitialCD: %+v", cd)
			for _, balance := range usefulUnitBalances {
				//log.Printf("Unit balance: %+v", balance)
				//utils.Logger.Info(fmt.Sprintf("CD BEFORE UNIT: %+v", cd))

				partCC, debitErr := balance.debitUnits(cd, balance.account, usefulMoneyBalances, count, dryRun, len(cc.Timespans) == 0)
				//log.Print("here: ", utils.ToIJSON(partCC), err)
				if debitErr != nil {
					return nil, debitErr
				}
				//utils.Logger.Info(fmt.Sprintf("CD AFTER UNIT: %+v", cd))
				if partCC != nil {
					//log.Printf("partCC: %+v", utils.ToIJSON(partCC))
					cc.Timespans = append(cc.Timespans, partCC.Timespans...)
					cc.negativeConnectFee = partCC.negativeConnectFee
					// for i, ts := range cc.Timespans {
					//  log.Printf("cc.times[an[%d]: %+v\n", i, ts)
					// }
					cd.TimeStart = cc.GetEndTime()
					//log.Printf("CD: %+v", cd)
					//log.Printf("CD: %+v - %+v", cd.TimeStart, cd.TimeEnd)
					// check if the calldescriptor is covered
					if cd.GetDuration() <= 0 {
						goto COMMIT
					}
					unitBalanceChecker = true
					generalBalanceChecker = true
					// check for max cost disconnect
					if dryRun && partCC.maxCostDisconect {
						// only return if we are in dry run (max call duration)
						return
					}
				}
				// check for blocker
				if dryRun && balance.Blocker {
					//log.Print("BLOCKER!")
					return // don't go to next balances
				}
			}
		}
		// debit money
		moneyBalanceChecker := true
		for moneyBalanceChecker {
			// try every balance multiple times in case one becomes active or ratig changes
			moneyBalanceChecker = false
			for _, balance := range usefulMoneyBalances {
				//utils.Logger.Info(fmt.Sprintf("Money balance: %+v", balance))
				//log.Printf("CD BEFORE MONEY: %s", utils.ToIJSON(cd))
				partCC, debitErr := balance.debitMoney(cd, balance.account, usefulMoneyBalances, count, dryRun, len(cc.Timespans) == 0)
				//log.Print("here: ", utils.ToIJSON(partCC), err)
				if debitErr != nil {
					return nil, debitErr
				}
				//utils.Logger.Info(fmt.Sprintf("CD AFTER MONEY: %+v", cd))
				if partCC != nil {
					cc.Timespans = append(cc.Timespans, partCC.Timespans...)
					cc.negativeConnectFee = partCC.negativeConnectFee
					//log.Print("partCC: ", utils.ToIJSON(partCC))
					/*for i, ts := range cc.Timespans {
						log.Printf("cc.times[an[%d]: %+v\n", i, ts)
					}*/
					cd.TimeStart = cc.GetEndTime()
					//log.Printf("CD: %+v", cd)
					//log.Printf("CD: %+v - %+v", cd.TimeStart, cd.TimeEnd)
					// check if the calldescriptor is covered
					if cd.GetDuration() <= 0 {
						goto COMMIT
					}
					moneyBalanceChecker = true
					generalBalanceChecker = true
					if dryRun && partCC.maxCostDisconect {
						// only return if we are in dry run (max call duration)
						return
					}
				}
				// check for blocker
				if dryRun && balance.Blocker {
					//log.Print("BLOCKER!")
					return // don't go to next balances
				}
			}
		}
		//log.Printf("END CD: %+v", cd)
	}
	//log.Print("After balances CD: ", utils.ToIJSON(cd))
	leftCC, err = cd.getCost()
	//log.Printf("leftCC: %s (%v)", utils.ToIJSON(leftCC), err)
	if err != nil {
		utils.Logger.Error("Error getting cost for left CC: ", zap.String("tenant", cd.Tenant), zap.String("subject", cd.Subject), zap.Error(err))
	}
	if leftCC.GetCost().IsZero() && len(leftCC.Timespans) > 0 {
		// put AccountID ubformation in increments
		for _, ts := range leftCC.Timespans {
			if ts.Increments.CompIncrement.BalanceInfo == nil {
				ts.Increments.CompIncrement.BalanceInfo = &DebitInfo{}
			}
			ts.Increments.CompIncrement.BalanceInfo.AccountID = ub.Name

		}
		cc.Timespans = append(cc.Timespans, leftCC.Timespans...)
	}

	if leftCC.GetCost().GtZero() && goNegative {
		initialLength := len(cc.Timespans)
		cc.Timespans = append(cc.Timespans, leftCC.Timespans...)
		if initialLength == 0 {
			// this is the first add, debit the connect fee
			ub.DebitConnectionFee(cc, usefulMoneyBalances, count, true)
		}
		//log.Printf("Left CC: %+v ", leftCC)
		// get the default money balanance
		// and go negative on it with the amount still unpaid
		if len(leftCC.Timespans) > 0 && leftCC.GetCost().GtZero() && !ub.AllowNegative && !dryRun {
			utils.Logger.Warn("<Rater> Going negative on account with AllowNegative: false", zap.String("tenant", cd.Tenant), zap.String("accID", cd.getAccountName()))
		}
		leftCC.Timespans.Decompress()
		for _, ts := range leftCC.Timespans {
			if ts.Increments == nil {
				ts.createIncrementsSlice()
			}
			ts.Increments.Reset()
			for incrIndex, increment := ts.Increments.Next(); increment != nil; incrIndex, increment = ts.Increments.Next() {
				cost := increment.Cost
				defaultBalance := ub.GetDefaultMoneyBalance()
				defaultBalance.SubstractValue(cost)
				if increment.BalanceInfo.Monetary == nil {
					increment.BalanceInfo.Monetary = &MonetaryInfo{
						UUID: defaultBalance.UUID,
						ID:   defaultBalance.ID,
					}
				}
				increment.BalanceInfo.Monetary.getValue().Set(defaultBalance.Value)
				increment.BalanceInfo.AccountID = ub.Name
				increment.paid++
				if count {
					b := &Balance{
						Directions:     utils.StringMap{leftCC.Direction: true},
						DestinationIDs: utils.StringMap{leftCC.Destination: true},
					}
					b.SetValue(cost)
					pats := ub.countUnits(cost, utils.MONETARY, leftCC, b)
					increment.AddPostATIDs(incrIndex, pats)
				}
			}
		}
	}

COMMIT:
	if !dryRun {
		// save darty shared balances
		usefulMoneyBalances.SaveDirtyBalances(ub)
		usefulUnitBalances.SaveDirtyBalances(ub)
	}
	//log.Printf("Final CC: %+v", cc)
	return
}

func (ub *Account) GetDefaultMoneyBalance() *Balance {
	for _, balance := range ub.BalanceMap[utils.MONETARY] {
		if balance.IsDefault() {
			return balance
		}
	}
	// create default balance
	defaultBalance := &Balance{
		UUID: utils.GenUUID(),
		ID:   utils.META_DEFAULT,
	} // minimum weight
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]Balances)
	}
	ub.BalanceMap[utils.MONETARY] = append(ub.BalanceMap[utils.MONETARY], defaultBalance)
	return defaultBalance
}

// Scans the action trigers and execute the actions for which trigger is met
func (acc *Account) ExecuteActionTriggers(a *Action, post bool) (postATIDs []string) {
	if acc.executingTriggers {
		return
	}
	acc.executingTriggers = true
	defer func() {
		acc.executingTriggers = false
	}()
	for _, at := range acc.getTriggers() {
		// check is effective
		if at.IsExpired(time.Now()) || !at.IsActive(time.Now()) {
			continue
		}

		if acc.TriggerRecords[at.UniqueID].Executed {
			// trigger is marked as executed, so skipp it until
			// the next reset (see RESET_TRIGGERS action type)
			continue
		}
		//log.Print("AT: ", utils.ToIJSON(at))
		if a != nil {
			match, err := a.getFilter().Query(at, false)
			if err != nil {
				utils.Logger.Error(fmt.Sprintf("<ExecuteActionTriggers> action filter %s errored: %s", a.Filter1, err.Error()))
			}
			if !match {
				continue
			}
		}
		if strings.Contains(at.ThresholdType, "counter") {
			/*if (at.Balance.ID == nil || *at.Balance.ID != "") && at.UniqueID != "" {
				at.Balance.ID = utils.StringPointer(at.UniqueID)
			}*/
			for _, counters := range acc.UnitCounters {
				for _, uc := range counters {
					if strings.Contains(at.ThresholdType, uc.CounterType[1:]) {
						for _, c := range uc.Counters {
							if strings.HasPrefix(at.ThresholdType, "*max") {
								if c.Filter == at.Filter && c.getValue().Cmp(at.ThresholdValue) >= 0 {
									if !post {
										if err := at.Execute(acc, nil); err != nil {
											utils.Logger.Error("error execute action triggers: ", zap.Error(err))
										}
									} else {
										acc.TriggerRecords[at.UniqueID].Executed = true
										postATIDs = append(postATIDs, at.UniqueID)
									}
								}
							} else { //MIN
								if c.Filter == at.Filter && c.getValue().Cmp(at.ThresholdValue) <= 0 {
									if !post {
										if err := at.Execute(acc, nil); err != nil {
											utils.Logger.Error("error execute action triggers: ", zap.Error(err))
										}
									} else {
										acc.TriggerRecords[at.UniqueID].Executed = true
										postATIDs = append(postATIDs, at.UniqueID)
									}
								}
							}
						}
					}
				}
			}
		} else { // BALANCE
			for _, b := range acc.BalanceMap[at.TOR] {
				if !b.dirty && at.ThresholdType != utils.TRIGGER_BALANCE_EXPIRED { // do not check clean balances
					continue
				}
				switch at.ThresholdType {
				case utils.TRIGGER_MAX_BALANCE:
					match, err := b.MatchActionTrigger(at)
					if err != nil {
						utils.Logger.Error(fmt.Sprintf("<ActionTrigger> action trigger filter error %s %s", at.Filter, err.Error()))
					}
					if match && b.GetValue().Cmp(at.ThresholdValue) >= 0 {
						if !post {
							at.Execute(acc, nil)
						} else {
							acc.TriggerRecords[at.UniqueID].Executed = true
							postATIDs = append(postATIDs, at.UniqueID)
						}
					}
				case utils.TRIGGER_MIN_BALANCE:
					match, err := b.MatchActionTrigger(at)
					if err != nil {
						utils.Logger.Error(fmt.Sprintf("<ActionTrigger> action trigger filter error %s %s", at.Filter, err.Error()))
					}
					if match && b.GetValue().Cmp(at.ThresholdValue) <= 0 {
						if !post {
							at.Execute(acc, nil)
						} else {
							acc.TriggerRecords[at.UniqueID].Executed = true
							postATIDs = append(postATIDs, at.UniqueID)
						}
					}
				case utils.TRIGGER_BALANCE_EXPIRED:
					match, err := b.MatchActionTrigger(at)
					if err != nil {
						utils.Logger.Error(fmt.Sprintf("<ActionTrigger> action trigger filter error %s %s", at.Filter, err.Error()))
					}
					if match && b.IsExpired() {
						if !post {
							at.Execute(acc, nil)
						} else {
							acc.TriggerRecords[at.UniqueID].Executed = true
							postATIDs = append(postATIDs, at.UniqueID)
						}
					}
				}
			}
		}
	}
	acc.CleanExpiredStuff()
	return
}

// Mark all action trigers as ready for execution
// If the action is not nil it acts like a filter
func (acc *Account) ResetActionTriggers(a *Action) {
	for _, trRec := range acc.TriggerRecords {
		if a != nil {
			match, err := a.getFilter().Query(trRec, false)
			if err != nil {
				utils.Logger.Error(fmt.Sprintf("<ResetActionTriggers> action filter %s errored: %s", a.Filter1, err.Error()))
			}
			if !match {
				continue
			}
		}
		trRec.Executed = false
	}
	acc.ExecuteActionTriggers(a, false)
}

// Sets/Unsets recurrent flag for action triggers
func (acc *Account) SetRecurrent(a *Action, recurrent bool) {
	for _, trRec := range acc.TriggerRecords {
		if a != nil {
			match, err := a.getFilter().Query(trRec, false)
			if err != nil {
				utils.Logger.Error(fmt.Sprintf("<SetRecurrent> action filter %s errored: %s", a.Filter1, err.Error()))
			}
			if !match {
				continue
			}
		}

		trRec.Recurrent = recurrent
	}
}

// Increments the counter for the type
func (acc *Account) countUnits(amount *dec.Dec, kind string, cc *CallCost, b *Balance) (postATIDs []string) {
	acc.UnitCounters.addUnits(amount, kind, cc, b)
	postActionTrigger := false
	if cc != nil {
		postActionTrigger = cc.postActionTrigger
	}
	postATIDs = acc.ExecuteActionTriggers(nil, postActionTrigger)
	return
}

func (acc *Account) processPostActionTriggers(exe, unexe []string) {
	for _, uniqueID := range exe {
		if at := acc.getActionTrigger(uniqueID); at != nil {
			if err := at.Execute(acc, nil); err != nil {
				utils.Logger.Error("error executing post action triggers: ", zap.Error(err))
			}
		}
	}
	for _, uniqueID := range unexe {
		if at := acc.getActionTrigger(uniqueID); at != nil {
			acc.TriggerRecords[at.UniqueID].Executed = false
		}
	}
}

func (acc *Account) getActionTrigger(uniqueID string) *ActionTrigger {
	for _, at := range acc.getTriggers() {
		if at.UniqueID == uniqueID {
			return at
		}
	}
	return nil
}

// Create counters for all triggered actions
func (acc *Account) InitCounters() {
	oldUcs := acc.UnitCounters
	acc.UnitCounters = make(UnitCounters)
	ucTempMap := make(map[string]*UnitCounter)
	for _, at := range acc.getTriggers() {
		//log.Print("AT: ", utils.ToJSON(at))
		if !strings.Contains(at.ThresholdType, "counter") {
			continue
		}
		ct := utils.COUNTER_EVENT //default
		if strings.Contains(at.ThresholdType, "balance") {
			ct = utils.COUNTER_BALANCE
		}
		uc, exists := ucTempMap[at.TOR+ct]
		//log.Print("CT: ", at.Balance.GetType()+ct)
		if !exists {
			uc = &UnitCounter{
				CounterType: ct,
			}
			ucTempMap[at.TOR+ct] = uc
			uc.Counters = make(CounterFilters, 0)
			acc.UnitCounters[at.TOR] = append(acc.UnitCounters[at.TOR], uc)
		}

		c := &CounterFilter{UniqueID: at.UniqueID, Filter: at.Filter}
		/*if (c.Filter.ID == nil || *c.Filter.ID == "") && at.UniqueID != "" {
			c.Filter.ID = utils.StringPointer(at.UniqueID)
		}*/
		//log.Print("C: ", utils.ToJSON(c))
		if !uc.Counters.HasCounter(c) {
			uc.Counters = append(uc.Counters, c)
		}
	}
	// copy old counter values
	for key, counters := range acc.UnitCounters {
		oldCounters, found := oldUcs[key]
		if !found {
			continue
		}
		for _, uc := range counters {
			for _, oldUc := range oldCounters {
				if uc.CopyCounterValues(oldUc) {
					break
				}
			}
		}
	}
	if len(acc.UnitCounters) == 0 {
		acc.UnitCounters = nil // leave it nil if empty
	}
}

func (acc *Account) InitTriggerRecords() { // no need to call this explicitly unless you know what you are doing
	newTriggerRecords := make(map[string]*ActionTriggerRecord)
	for _, atr := range acc.triggers { // not getTriggers => infinite loop
		if _, found := acc.TriggerRecords[atr.UniqueID]; !found {
			newTriggerRecords[atr.UniqueID] = &ActionTriggerRecord{
				UniqueID:       atr.UniqueID,
				Recurrent:      atr.Recurrent,
				ExpirationDate: atr.ExpirationDate,
				ActivationDate: atr.ActivationDate,
			}
		} else {
			newTriggerRecords[atr.UniqueID] = acc.TriggerRecords[atr.UniqueID]
		}
	}
	acc.TriggerRecords = newTriggerRecords
}

func (acc *Account) CleanExpiredStuff() {
	for key, bm := range acc.BalanceMap {
		for i := 0; i < len(bm); i++ {
			if bm[i].IsExpired() {
				// delete it
				bm = append(bm[:i], bm[i+1:]...)
			}
		}
		acc.BalanceMap[key] = bm
	}
	// TODO: clean expired triggers (clean enitire groups?)
	/*for i := 0; i < len(acc.ActionTriggers); i++ {
		if acc.ActionTriggers[i].IsExpired(time.Now()) {
			acc.ActionTriggers = append(acc.ActionTriggers[:i], acc.ActionTriggers[i+1:]...)
		}
	}*/
}

func (acc *Account) allBalancesExpired() bool {
	for _, bm := range acc.BalanceMap {
		for i := 0; i < len(bm); i++ {
			if !bm[i].IsExpired() {
				return false
			}
		}
	}
	return true
}

// returns the shared groups that this user balance belnongs to
func (acc *Account) GetSharedGroups() (groups []string) {
	for _, balanceChain := range acc.BalanceMap {
		for _, b := range balanceChain {
			for sg := range b.SharedGroups {
				groups = append(groups, sg)
			}
		}
	}
	return
}

func (account *Account) GetUniqueSharedGroupMembers(cd *CallDescriptor) (utils.StringMap, error) {
	var balances []*Balance
	balances = append(balances, account.getBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, utils.MONETARY, "")...)
	balances = append(balances, account.getBalancesForPrefix(cd.Destination, cd.Category, cd.Direction, cd.TOR, "")...)
	// gather all shared group ids
	var sharedGroupIds []string
	for _, b := range balances {
		for sg := range b.SharedGroups {
			sharedGroupIds = append(sharedGroupIds, sg)
		}
	}
	memberIds := make(utils.StringMap)
	for _, sgID := range sharedGroupIds {
		sharedGroup, err := ratingStorage.GetSharedGroup(account.Tenant, sgID, utils.CACHED)
		if err != nil {
			utils.Logger.Warn("Could not get shared group: ", zap.String("id", sgID))
			return nil, err
		}
		for memberID := range sharedGroup.MemberIDs {
			if memberID != account.Name { // account name is allready used for locking
				memberIds[memberID] = true
			}
		}
	}
	return memberIds, nil
}

func (acc *Account) Clone() *Account {
	newAcc := &Account{
		Tenant:         acc.Tenant,
		Name:           acc.Name,
		BalanceMap:     make(map[string]Balances, len(acc.BalanceMap)),
		UnitCounters:   nil, // not used when cloned (dryRun)
		TriggerIDs:     nil, // not used when cloned (dryRun)
		TriggerRecords: nil, // not used when cloned (dryRun)
		AllowNegative:  acc.AllowNegative,
		Disabled:       acc.Disabled,
	}
	for key, balanceChain := range acc.BalanceMap {
		newAcc.BalanceMap[key] = balanceChain.Clone()
	}
	return newAcc
}

func (acc *Account) DebitConnectionFee(cc *CallCost, usefulMoneyBalances Balances, count bool, block bool) bool {
	if cc.deductConnectFee {
		connectFee := cc.GetConnectFee()
		if connectFee.IsZero() {
			return true
		}
		//log.Print("CONNECT FEE: %f", connectFee)
		connectFeePaid := false
		for _, b := range usefulMoneyBalances {
			if b.GetValue().Cmp(connectFee) >= 0 {
				b.SubstractValue(connectFee)
				// the conect fee is not refundable!
				if count {
					inc := cc.GetFirstIncrement()
					pats := acc.countUnits(connectFee, utils.MONETARY, cc, b)
					inc.AddPostATIDs(1, pats)
				}
				connectFeePaid = true
				break
			}
			if b.Blocker && block { // stop here
				return false
			}
		}
		// debit connect fee
		if connectFee.GtZero() && !connectFeePaid {
			cc.negativeConnectFee = true
			// there are no money for the connect fee; go negative
			b := acc.GetDefaultMoneyBalance()
			b.SubstractValue(connectFee)
			// the conect fee is not refundable!
			if count {
				inc := cc.GetFirstIncrement()
				pats := acc.countUnits(connectFee, utils.MONETARY, cc, b)
				inc.AddPostATIDs(1, pats)
			}
		}
	}
	return true
}

func (acc *Account) matchActionFilter(condition string) (bool, error) {
	sm, err := utils.NewStructQ(condition)
	if err != nil {
		return false, err
	}
	for balanceType, balanceChain := range acc.BalanceMap {
		for _, b := range balanceChain {
			check, err := sm.Query(&struct {
				Type string
				*Balance
			}{
				Type:    balanceType,
				Balance: b,
			}, false)
			if err != nil {
				return false, err
			}
			if check {
				return true, nil
			}
		}
	}
	return false, nil
}

func (acc *Account) FullID() string {
	return utils.ConcatKey(acc.Tenant, acc.Name)
}

func (acc *Account) AsAccountSummary() *AccountSummary {
	ad := &AccountSummary{
		Tenant:        acc.Tenant,
		ID:            acc.Name,
		AllowNegative: acc.AllowNegative,
		Disabled:      acc.Disabled}
	for balanceType, balances := range acc.BalanceMap {
		for _, balance := range balances {
			ad.BalanceSummaries = append(ad.BalanceSummaries, balance.AsBalanceSummary(balanceType))
		}
	}
	return ad
}

func NewAccountSummaryFromJSON(jsn string) (acntSummary *AccountSummary, err error) {
	if !utils.IsSliceMember([]string{"", "null"}, jsn) { // Unmarshal only when content
		json.Unmarshal([]byte(jsn), &acntSummary)
	}
	return
}

// AccountDigest contains compressed information about an Account
type AccountSummary struct {
	Tenant           string
	ID               string
	BalanceSummaries []*BalanceSummary
	AllowNegative    bool
	Disabled         bool
}
