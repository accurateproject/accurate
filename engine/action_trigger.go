package engine

import (
	"fmt"
	"sort"
	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

type ActionTriggerGroup struct {
	Tenant         string         `bson:"tenant"`
	Name           string         `bson:"name"` // original csv tag
	ActionTriggers ActionTriggers `bson:"action_triggers"`
}

// SetParentActionPlan populates parent to all action timings
func (atrg *ActionTriggerGroup) SetParentGroup() {
	for _, atr := range atrg.ActionTriggers {
		atr.parentGroup = atrg
	}
}

type ActionTrigger struct {
	UniqueID      string `bson:"unique_id"`      // individual id
	ThresholdType string `bson:"threshold_type"` //*min_event_counter, *max_event_counter, *min_balance_counter, *max_balance_counter, *min_balance, *max_balance, *balance_expired
	// stats: `bson:""` *min_asr, *max_asr, *min_acd, *max_acd, *min_tcd, *max_tcd, *min_acc, *max_acc, *min_tcc, *max_tcc, *min_ddc, *max_ddc
	ThresholdValue *dec.Dec      `bson:"threshold_value"`
	Recurrent      bool          `bson:"recurrent"` // reset excuted flag each run
	MinSleep       time.Duration `bson:"min_sleep"` // Minimum duration between two executions in case of recurrent triggers
	ExpirationDate time.Time     `bson:"expiration_date"`
	ActivationDate time.Time     `bson:"activation_date"`
	TOR            string        `bson:"trigger_type"` // free string can be used to select balance type
	Filter         string        `bson:"filter"`
	Weight         float64       `bson:"weight"`
	ActionsID      string        `bson:"actions_id"`
	MinQueuedItems int           `bson:"min_queued_items"` // Trigger actions only if this number is hit (stats only)
	parentGroup    *ActionTriggerGroup
	filter         *utils.StructQ
}

type ActionTriggerRecord struct {
	UniqueID          string    `bson:"unique_id"` // for query matching
	Recurrent         bool      `bson:"recurrent"`
	Executed          bool      `bson:"executed"`
	ExpirationDate    time.Time `bson:"expiration_date"`
	ActivationDate    time.Time `bson:"activation_date"`
	LastExecutionTime time.Time `bson:"last_execution_time"`
}

func (at *ActionTrigger) getFilter() (*utils.StructQ, error) {
	if at.filter != nil {
		return at.filter, nil
	}
	var err error
	at.filter, err = utils.NewStructQ(at.Filter)
	return at.filter, err
}

func (at *ActionTrigger) Execute(acc *Account, sq *StatsQueueTriggered) (err error) {
	// check for min sleep time
	var trRec *ActionTriggerRecord
	if acc != nil {
		trRec = acc.TriggerRecords[at.UniqueID]
	}
	if sq != nil {
		trRec = sq.TriggerRecords[at.UniqueID]
	}
	lastExecutionTime := trRec.LastExecutionTime
	if at.Recurrent && !lastExecutionTime.IsZero() && time.Since(lastExecutionTime) < at.MinSleep {
		return
	}
	trRec.LastExecutionTime = time.Now()
	if acc != nil && acc.Disabled {
		return fmt.Errorf("User %s is disabled and there are triggers in action!", acc.FullID())
	}
	// does NOT need to Lock() because it is triggered from a method that took the Lock
	var aag *ActionGroup
	aag, err = ratingStorage.GetActionGroup(at.parentGroup.Tenant, at.ActionsID, utils.CACHED)
	if err != nil || aag == nil {
		utils.Logger.Error("Failed to get actions: ", zap.String("tenant", at.parentGroup.Tenant), zap.String("id", at.ActionsID), zap.Error(err))
		return
	}
	aag.Actions.Sort()
	trRec.Executed = true
	transactionFailed := false
	removeAccountActionFound := false
	for _, a := range aag.Actions {
		// check action filter
		if len(a.ExecFilter) > 0 {
			matched, err := acc.matchActionFilter(a.ExecFilter)
			if err != nil {
				return err
			}
			if !matched {
				continue
			}
		}

		actionFunction, exists := getActionFunc(a.ActionType)
		if !exists {
			utils.Logger.Error("Function type %v not available, aborting execution!", zap.String("action type", a.ActionType))
			transactionFailed = false
			break
		}
		//go utils.Logger.Info(fmt.Sprintf("Executing %v, %v: %v", acc, sq, a))
		if err := actionFunction(acc, sq, a, aag.Actions); err != nil {
			utils.Logger.Error("Error executing action ", zap.String("action type", a.ActionType), zap.Error(err))
			transactionFailed = false
			break
		}
		if a.ActionType == REMOVE_ACCOUNT {
			removeAccountActionFound = true
		}
	}
	if transactionFailed || at.Recurrent {
		trRec.Executed = false
	}
	if !transactionFailed && acc != nil && !removeAccountActionFound {
		Publish(CgrEvent{
			"EventName": utils.EVT_ACTION_TRIGGER_FIRED,
			"Uuid":      at.UniqueID,
			"Tenant":    at.parentGroup.Tenant,
			"Id":        at.parentGroup.Name,
			"ActionIds": at.ActionsID,
		})
		accountingStorage.SetAccount(acc)
	}
	return
}

// makes a shallow copy of the receiver
func (at *ActionTrigger) Clone() *ActionTrigger {
	clone := new(ActionTrigger)
	*clone = *at
	return clone
}

func (at *ActionTrigger) Equals(oat *ActionTrigger) bool {
	// ids only
	return at.parentGroup.Name == oat.parentGroup.Name && at.UniqueID == oat.UniqueID
}

func (at *ActionTrigger) IsActive(t time.Time) bool {
	return at.ActivationDate.IsZero() || t.After(at.ActivationDate)
}

func (at *ActionTrigger) IsExpired(t time.Time) bool {
	return !at.ExpirationDate.IsZero() && t.After(at.ExpirationDate)
}

// Structure to store actions according to weight
type ActionTriggers []*ActionTrigger

func (atpl ActionTriggers) Sort() {
	sort.Slice(atpl, func(j, i int) bool {
		//we need higher weights earlyer in the list
		return atpl[i].Weight < atpl[j].Weight
	})
}
