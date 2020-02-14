package engine

import (
	"sort"
	"time"

	"github.com/accurateproject/accurate/utils"
	"github.com/gorhill/cronexpr"
	"go.uber.org/zap"
)

const (
	FORMAT = "2006-1-2 15:04:05 MST"
)

type ActionTiming struct {
	UUID       string        `bson:"uuid"`
	Timing     *RateInterval `bson:"timing"`
	ActionsID  string        `bson:"actions_id"`
	Weight     float64       `bson:"weight"`
	actions    Actions
	actionPlan *ActionPlan
	accountIDs utils.StringMap
	stCache    time.Time // cached time of the next start
}

type Task struct {
	UUID      string `bson:"uuid"`
	Tenant    string `bson:"tenant"`
	AccountID string `bson:"account_id"`
	ActionsID string `bson:"actions_id"`
}

type ActionPlanBinding struct {
	Tenant     string `bson:"tenant"`
	Account    string `bson:"account"`
	ActionPlan string `bson:"action_plan"`
}

type ActionPlan struct {
	Tenant        string          `bson:"tenant"`
	Name          string          `bson:"name"`
	ActionTimings []*ActionTiming `bson:"action_timings"`
}

// SetParentActionPlan populates parent to all action timings
func (apl *ActionPlan) SetParentActionPlan() {
	for _, at := range apl.ActionTimings {
		at.actionPlan = apl
	}
}

func (t *Task) Execute() error {
	return (&ActionTiming{
		UUID:       t.UUID,
		ActionsID:  t.ActionsID,
		accountIDs: utils.StringMap{t.AccountID: true},
	}).Execute()
}

func (at *ActionTiming) GetNextStartTime(now time.Time) (t time.Time) {
	if !at.stCache.IsZero() {
		return at.stCache
	}
	i := at.Timing
	if i == nil || i.Timing == nil {
		return
	}
	// Normalize
	if i.Timing.StartTime == "" {
		i.Timing.StartTime = "00:00:00"
	}
	if len(i.Timing.Years) > 0 && len(i.Timing.Months) == 0 {
		i.Timing.Months = append(i.Timing.Months, 1)
	}
	if len(i.Timing.Months) > 0 && len(i.Timing.MonthDays) == 0 {
		i.Timing.MonthDays = append(i.Timing.MonthDays, 1)
	}
	at.stCache = cronexpr.MustParse(i.Timing.CronString()).Next(now)
	return at.stCache
}

func (at *ActionTiming) ResetStartTimeCache() {
	at.stCache = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
}

func (at *ActionTiming) SetActions(as Actions) {
	at.actions = as
}

func (at *ActionTiming) SetAccountIDs(accIDs utils.StringMap) {
	at.accountIDs = accIDs
}

func (at *ActionTiming) GetAccountIDs() utils.StringMap {
	return at.accountIDs
}

func (at *ActionTiming) SetActionPlan(apl *ActionPlan) {
	at.actionPlan = apl
}

func (at *ActionTiming) GetActionPlanID() string {
	return at.actionPlan.Name
}

func (at *ActionTiming) getActions() (as []*Action, err error) {
	if at.actions == nil {
		ag, err := ratingStorage.GetActionGroup(at.actionPlan.Tenant, at.ActionsID, utils.CACHED)
		if err != nil {
			return nil, err
		}
		ag.SetParentGroup() // populate parent to all actions
		at.actions = ag.Actions
	}
	at.actions.Sort()
	return at.actions, err
}

func (at *ActionTiming) getActionPlanBindings() Iterator {
	if at.accountIDs == nil {
		return ratingStorage.Iterator(ColApb, "", map[string]interface{}{"tenant": at.actionPlan.Tenant, "action_plan": at.actionPlan.Name})
	}
	return NewFakeAPBIterator(at.actionPlan.Tenant, at.actionPlan.Name, at.accountIDs.Slice())
}

func (at *ActionTiming) Execute() (err error) {
	at.ResetStartTimeCache()
	aac, err := at.getActions()
	if err != nil {
		utils.Logger.Error("Failed to get actions for ", zap.String("ID", at.ActionsID), zap.Error(err))
		return
	}
	apbIter := at.getActionPlanBindings()
	var apb ActionPlanBinding
	actionPlanHadBindings := false
	for apbIter.Next(&apb) {
		actionPlanHadBindings = true
		_, err = Guardian.Guard(func() (interface{}, error) {
			acc, err := accountingStorage.GetAccount(apb.Tenant, apb.Account)
			if err != nil {
				utils.Logger.Warn("Could not get account. Skipping!", zap.String("tenant", apb.Tenant), zap.String("id", apb.Account))
				return 0, err
			}
			transactionFailed := false
			removeAccountActionFound := false
			for _, a := range aac {
				// check action filter
				if len(a.ExecFilter) > 0 {
					matched, err := acc.matchActionFilter(a.ExecFilter)
					//log.Print("Checkng: ", a.ExecFilter, matched)
					if err != nil {
						return 0, err
					}
					if !matched {
						continue
					}
				}
				b, _ := a.getBalance(nil)
				if b == nil {
					b = &Balance{}
				}
				actionFunction, exists := getActionFunc(a.ActionType)
				if !exists {
					// do not allow the action plan to be rescheduled
					at.Timing = nil
					utils.Logger.Warn("Function type not available, aborting execution!", zap.String("action type", a.ActionType))
					transactionFailed = true
					break
				}
				if err := actionFunction(acc, nil, a, aac); err != nil {
					//log.Print("err: ", err)
					utils.Logger.Error("Error executing action ", zap.String("action type", a.ActionType), zap.Error(err))
					transactionFailed = true
					break
				}
				if a.ActionType == REMOVE_ACCOUNT {
					removeAccountActionFound = true
				}
			}
			//log.Print("transaction failed: ", transactionFailed)
			if !transactionFailed && !removeAccountActionFound {
				//log.Print("Write: ", utils.ToIJSON(acc))
				accountingStorage.SetAccount(acc)
			}
			return 0, nil
		}, 0, utils.ConcatKey(apb.Tenant, apb.Account))
	}
	if err := apbIter.Close(); err != nil {
		return err
	}
	if !actionPlanHadBindings { // action timing executing without accounts
		for _, a := range aac {
			actionFunction, exists := getActionFunc(a.ActionType)
			if !exists {
				// do not allow the action plan to be rescheduled
				at.Timing = nil
				utils.Logger.Error("Function type %v not available, aborting execution!", zap.String("action type", a.ActionType))
				break
			}
			if err := actionFunction(nil, nil, a, aac); err != nil {
				utils.Logger.Error("Error executing accountless action", zap.String("action type", a.ActionType), zap.Error(err))
				break
			}
		}
	}
	if err != nil {
		utils.Logger.Warn("Error executing action plan: %v", zap.Error(err))
		return err
	}
	Publish(CgrEvent{
		"EventName": utils.EVT_ACTION_TIMING_FIRED,
		"Uuid":      at.UUID,
		"Tenant":    at.actionPlan.Tenant,
		"Id":        at.actionPlan.Name,
		"ActionIds": at.ActionsID,
	})
	return
}

func (at *ActionTiming) IsASAP() bool {
	if at.Timing == nil {
		return false
	}
	return at.Timing.Timing.StartTime == utils.ASAP
}

// Structure to store actions according to execution time and weight
type ActionTimingPriorityList []*ActionTiming

func (atpl ActionTimingPriorityList) Len() int {
	return len(atpl)
}

func (atpl ActionTimingPriorityList) Swap(i, j int) {
	atpl[i], atpl[j] = atpl[j], atpl[i]
}

func (atpl ActionTimingPriorityList) Less(i, j int) bool {
	if atpl[i].GetNextStartTime(time.Now()).Equal(atpl[j].GetNextStartTime(time.Now())) {
		// higher weights earlyer in the list
		return atpl[i].Weight > atpl[j].Weight
	}
	return atpl[i].GetNextStartTime(time.Now()).Before(atpl[j].GetNextStartTime(time.Now()))
}

// using the len/swap/less to implement sort.Interface
func (atpl ActionTimingPriorityList) Sort() {
	sort.Sort(atpl)
}

// Structure to store actions according to weight
type ActionTimingWeightOnlyPriorityList []*ActionTiming

func (atpl ActionTimingWeightOnlyPriorityList) Sort() {
	sort.Slice(atpl, func(i, j int) bool {
		return atpl[i].Weight > atpl[j].Weight
	})
}
