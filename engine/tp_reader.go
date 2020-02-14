package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

type LoadStats struct {
	Tenants     utils.StringMap
	CDRStats    map[string]utils.StringMap // tenant [ids]
	UserTenants utils.StringMap
}

func NewLoadStats() *LoadStats {
	return &LoadStats{
		Tenants:     utils.StringMap{},
		CDRStats:    make(map[string]utils.StringMap),
		UserTenants: utils.StringMap{},
	}
}

type TpReader struct {
	ratingStorage     RatingStorage
	accountingStorage AccountingStorage
	timezone          string
	loadStats         *LoadStats
}

func NewTpReader(ratingStorage RatingStorage, accountingStorage AccountingStorage, timeZone string) *TpReader {
	return &TpReader{
		ratingStorage:     ratingStorage,
		accountingStorage: accountingStorage,
		timezone:          timeZone,
		loadStats:         NewLoadStats(),
	}
}

func (tpr *TpReader) LoadDestination(el interface{}) error {
	element := el.(*utils.TpDestination)
	tpr.loadStats.Tenants[element.Tenant] = true
	return tpr.ratingStorage.SetDestination(&Destination{
		Tenant: element.Tenant,
		Code:   element.Code,
		Name:   element.Tag,
	})
}

func (tpr *TpReader) LoadTiming(el interface{}) error {
	element := el.(*utils.TpTiming)
	tpr.loadStats.Tenants[element.Tenant] = true
	return tpr.ratingStorage.SetTiming(&Timing{
		Tenant:    element.Tenant,
		Name:      element.Tag,
		Years:     element.Years,
		Months:    element.Months,
		MonthDays: element.MonthDays,
		WeekDays:  element.WeekDays,
		Time:      element.Time,
	})
}

func (tpr *TpReader) LoadRate(el interface{}) error {
	element := el.(*utils.TpRate)
	tpr.loadStats.Tenants[element.Tenant] = true
	r := &Rate{
		Tenant: element.Tenant,
		Name:   element.Tag,
	}
	for _, slot := range element.Slots {
		var err error
		var ru time.Duration
		if slot.RateUnit != "" {
			ru, err = utils.ParseDurationWithSecs(slot.RateUnit)
			if err != nil {
				return fmt.Errorf("error parsing rate unit %s: %s", slot.RateUnit, err.Error())
			}
		} else {
			ru = time.Minute
		}
		var ri time.Duration
		if slot.RateIncrement != "" {
			ri, err = utils.ParseDurationWithSecs(slot.RateIncrement)
			if err != nil {
				return fmt.Errorf("error parsing rate increment %s: %s", slot.RateIncrement, err.Error())
			}
		} else {
			ri = 1 * time.Second
		}
		var gis time.Duration
		if slot.GroupIntervalStart != "" {
			gis, err = utils.ParseDurationWithSecs(slot.GroupIntervalStart)
			if err != nil {
				return fmt.Errorf("error parsing group interval start %s: %s", slot.GroupIntervalStart, err.Error())
			}
		} else {
			gis = 0 * time.Second
		}
		r.Slots = append(r.Slots, &RateSlot{
			ConnectFee:         slot.ConnectFee,
			Rate:               slot.Rate,
			RateUnit:           ru,
			RateIncrement:      ri,
			GroupIntervalStart: gis,
		})
	}

	return tpr.ratingStorage.SetRate(r)
}

func (tpr *TpReader) LoadDestinationRate(el interface{}) error {
	element := el.(*utils.TpDestinationRate)
	tpr.loadStats.Tenants[element.Tenant] = true
	dr := &DestinationRate{
		Tenant:   element.Tenant,
		Name:     element.Tag,
		Bindings: make(map[string]*DestinationRateBinding),
	}
	for _, binding := range element.Bindings {
		// check destination code
		destinationExists := binding.DestinationCode == utils.ANY
		var dests []*Destination
		if binding.DestinationCode != "" && !destinationExists {
			dCodes, err := tpr.ratingStorage.GetDestinations(element.Tenant, binding.DestinationCode, "", utils.DestExact, utils.CACHE_SKIP)
			if err != nil || len(dCodes) == 0 {
				return fmt.Errorf("could not get destination for code %s (%v)", binding.DestinationCode, err)
			}
			dests = append(dests, dCodes...)
		}
		// check destination tag
		destinationExists = binding.DestinationTag == utils.ANY
		if binding.DestinationTag != "" && !destinationExists {
			dTags, err := tpr.ratingStorage.GetDestinations(element.Tenant, "", binding.DestinationTag, utils.DestExact, utils.CACHE_SKIP)
			if err != nil || len(dTags) == 0 {
				return fmt.Errorf("could not get destination for tag %s (%v)", binding.DestinationTag, err)
			}
			dests = append(dests, dTags...)
		}

		// check rate
		rate, err := tpr.ratingStorage.GetRate(element.Tenant, binding.RatesTag)
		if err != nil || rate == nil {
			return fmt.Errorf("could not get rate for tag %s (%v)", binding.RatesTag, err)
		}
		// get unique codes
		uniqueCodes := make(map[string]string)
		if binding.DestinationTag == utils.ANY || binding.DestinationCode == utils.ANY {
			uniqueCodes[utils.ANY] = utils.ANY
		}
		for _, d := range dests {
			uniqueCodes[d.Code] = d.Name
		}
		for destinationCode, destinationName := range uniqueCodes {
			dr.Bindings[fmt.Sprintf("%s_%s", destinationCode, binding.RatesTag)] = &DestinationRateBinding{
				DestinationCode: destinationCode,
				DestinationName: destinationName,
				RateID:          binding.RatesTag,
				MaxCost:         binding.MaxCost,
				MaxCostStrategy: binding.MaxCostStrategy,
			}
		}
	}
	return tpr.ratingStorage.SetDestinationRate(dr)
}

func (tpr *TpReader) LoadRatingPlan(el interface{}) error {
	element := el.(*utils.TpRatingPlan)
	tpr.loadStats.Tenants[element.Tenant] = true
	rp := &RatingPlan{
		Tenant: element.Tenant,
		Name:   element.Tag,
	}

	for _, rpBinding := range element.Bindings {
		timing, err := tpr.ratingStorage.GetTiming(rp.Tenant, rpBinding.TimingTag)
		if err != nil || timing == nil {
			return fmt.Errorf("could not get timing for tag %s (%v)", rpBinding.TimingTag, err)
		}

		drate, err := tpr.ratingStorage.GetDestinationRate(rp.Tenant, rpBinding.DestinationRatesTag)
		if err != nil || drate == nil {
			return fmt.Errorf("could not find destination rate for tag %s (%s)", rpBinding.DestinationRatesTag, err.Error())
		}
		for _, drBinding := range drate.Bindings {
			rate, err := tpr.ratingStorage.GetRate(rp.Tenant, drBinding.RateID)
			if err != nil || rate == nil {
				return fmt.Errorf("could not find rate for tag %s (%v)", drBinding.RateID, err)
			}

			ri := &RateInterval{
				Timing: &RITiming{
					Years:     timing.Years,
					Months:    timing.Months,
					MonthDays: timing.MonthDays,
					WeekDays:  timing.WeekDays,
					StartTime: timing.Time,
				},
				Weight: rpBinding.Weight,
				Rating: &RIRate{
					ConnectFee:      dec.NewFloat(rate.Slots[0].ConnectFee),
					MaxCost:         dec.NewFloat(drBinding.MaxCost),
					MaxCostStrategy: drBinding.MaxCostStrategy,
				},
			}
			for _, rs := range rate.Slots {
				ri.Rating.Rates = append(ri.Rating.Rates, &RateInfo{
					GroupIntervalStart: rs.GroupIntervalStart,
					Value:              dec.NewFloat(rs.Rate),
					RateIncrement:      rs.RateIncrement,
					RateUnit:           rs.RateUnit,
				})
			}
			rp.AddRateInterval(drBinding.DestinationCode, drBinding.DestinationName, ri)

		}
	}
	if !rp.isContinous() {
		return fmt.Errorf("The rating plan %s is not covering all weekdays", rp.Name)
	}
	if crazyRate := rp.getFirstUnsaneRating(); crazyRate != "" {
		return fmt.Errorf("The rate %s is invalid", crazyRate)
	}
	if crazyTiming := rp.getFirstUnsaneTiming(); crazyTiming != "" {
		return fmt.Errorf("The timing %s is invalid", crazyTiming)
	}
	return tpr.ratingStorage.SetRatingPlan(rp)
}

func (tpr *TpReader) LoadRatingProfile(el interface{}) error {
	element := el.(*utils.TpRatingProfile)
	tpr.loadStats.Tenants[element.Tenant] = true
	rpf := &RatingProfile{
		Direction: element.Direction,
		Tenant:    element.Tenant,
		Category:  element.Category,
		Subject:   element.Subject,
	}

	for _, rpfa := range element.Activations {
		at, err := utils.ParseDate(rpfa.ActivationTime)
		if err != nil {
			return fmt.Errorf("cannot parse activation time from %s", rpfa.ActivationTime)
		}
		ratingPlan, err := tpr.ratingStorage.GetRatingPlan(rpf.Tenant, rpfa.RatingPlanTag, utils.CACHE_SKIP)
		if err != nil || ratingPlan == nil {
			return fmt.Errorf("could not load rating plans for tag: %s (%v)", rpfa.RatingPlanTag, err)
		}
		rpf.RatingPlanActivations = append(rpf.RatingPlanActivations, &RatingPlanActivation{
			ActivationTime:  at,
			RatingPlanID:    rpfa.RatingPlanTag,
			FallbackKeys:    rpfa.FallbackSubjects,
			CdrStatQueueIDs: rpfa.CdrStatQueueIDs,
		})
	}
	return tpr.ratingStorage.SetRatingProfile(rpf)
}

func (tpr *TpReader) LoadSharedGroup(el interface{}) error {
	element := el.(*utils.TpSharedGroup)
	tpr.loadStats.Tenants[element.Tenant] = true
	sg := &SharedGroup{
		Tenant:            element.Tenant,
		Name:              element.Tag,
		AccountParameters: make(map[string]*SharingParam),
		MemberIDs:         utils.StringMap{},
	}
	for acc, sgParam := range element.AccountParameters {
		sg.AccountParameters[acc] = &SharingParam{
			Strategy:      sgParam.Strategy,
			RatingSubject: sgParam.RatingSubject,
		}
	}
	for _, acc := range element.MemberIDs {
		sg.MemberIDs[acc] = true
	}
	return tpr.ratingStorage.SetSharedGroup(sg)
}

func (tpr *TpReader) LoadLCR(el interface{}) error {
	element := el.(*utils.TpLcrRule)
	tpr.loadStats.Tenants[element.Tenant] = true
	lcr := &LCR{
		Direction: element.Direction,
		Tenant:    element.Tenant,
		Category:  element.Category,
		Account:   element.Account,
		Subject:   element.Subject,
	}

	for _, activation := range element.Activations {
		at, err := utils.ParseDate(activation.ActivationTime)
		if err != nil {
			return fmt.Errorf("cannot parse activation time from %s", activation.ActivationTime)
		}
		a := &LCRActivation{
			ActivationTime: at,
		}
		lcr.Activations = append(lcr.Activations, a)
		for _, entry := range activation.Entries {
			// check rating profile prefix
			rpf, err := tpr.ratingStorage.GetRatingProfile(lcr.Direction, lcr.Tenant, entry.RPCategory, "", false, utils.CACHE_SKIP)
			if err != nil || rpf == nil {
				return fmt.Errorf("[LCR] could not find ratingProfiles with %s %s %s (%v)", lcr.Direction, lcr.Tenant, lcr.Category, err)
			}
			// check destination
			d, err := tpr.ratingStorage.GetDestinations(lcr.Tenant, "", entry.DestinationTag, utils.DestExact, utils.CACHE_SKIP)
			if err != nil || d == nil {
				return fmt.Errorf("[LCR] could not find rating destination with tag %s (%v)", entry.DestinationTag, err)
			}
			a.Entries = append(a.Entries, &LCREntry{
				DestinationID:  entry.DestinationTag,
				RPCategory:     entry.RPCategory,
				Strategy:       entry.Strategy,
				StrategyParams: entry.StrategyParams,
				Weight:         entry.Weight,
			})
		}
	}

	return tpr.ratingStorage.SetLCR(lcr)
}

func (tpr *TpReader) LoadActionGroup(el interface{}) error {
	element := el.(*utils.TpActionGroup)
	tpr.loadStats.Tenants[element.Tenant] = true
	ag := &ActionGroup{
		Tenant: element.Tenant,
		Name:   element.Tag,
	}

	ag.Actions = make([]*Action, len(element.Actions))

	for idx, tpact := range element.Actions {
		a := &Action{
			ActionType: tpact.Action,
			Weight:     tpact.Weight,
			Params:     tpact.Params,
			ExecFilter: tpact.ExecFilter,
			Filter1:    tpact.Filter,
			TOR:        tpact.TOR,
		}
		// make filters and params standard json
		a.ExecFilter = strings.Replace(a.ExecFilter, `'`, `"`, -1)
		a.Filter1 = strings.Replace(a.Filter1, `'`, `"`, -1)
		a.Params = strings.Replace(a.Params, `'`, `"`, -1)

		// check filter field
		if len(tpact.ExecFilter) > 0 {
			if _, err := utils.NewStructQ(a.ExecFilter); err != nil {
				return fmt.Errorf("error parsing action %s exec filter %s (%v)", ag.Name, a.ExecFilter, err)
			}
		}
		ag.Actions[idx] = a
	}
	return tpr.ratingStorage.SetActionGroup(ag)
}

func (tpr *TpReader) LoadActionPlan(el interface{}) (err error) {
	element := el.(*utils.TpActionPlan)
	tpr.loadStats.Tenants[element.Tenant] = true
	apl := &ActionPlan{
		Tenant: element.Tenant,
		Name:   element.Tag,
	}
	for _, at := range element.ActionTimings {
		action, err := tpr.ratingStorage.GetActionGroup(apl.Tenant, at.ActionsTag, utils.CACHE_SKIP)
		if err != nil || action == nil {
			return fmt.Errorf("[ActionPlans] error geting actions: %s (%v)", at.ActionsTag, err)
		}

		timing, err := tpr.ratingStorage.GetTiming(apl.Tenant, at.TimingTag)
		if err != nil || timing == nil {
			return fmt.Errorf("[ActionPlans] error geting timing: %s (%v)", at.TimingTag, err)
		}

		apl.ActionTimings = append(apl.ActionTimings, &ActionTiming{
			UUID:   utils.GenUUID(),
			Weight: at.Weight,
			Timing: &RateInterval{
				Timing: &RITiming{
					Years:     timing.Years,
					Months:    timing.Months,
					MonthDays: timing.MonthDays,
					WeekDays:  timing.WeekDays,
					StartTime: timing.Time,
				},
			},
			ActionsID: at.ActionsTag,
		})
	}
	return tpr.ratingStorage.SetActionPlan(apl)
}

func (tpr *TpReader) LoadActionTrigger(el interface{}) error {
	element := el.(*utils.TpActionTrigger)
	tpr.loadStats.Tenants[element.Tenant] = true
	atrg := &ActionTriggerGroup{
		Tenant:         element.Tenant,
		Name:           element.Tag,
		ActionTriggers: make([]*ActionTrigger, len(element.Triggers)),
	}

	for idx, atr := range element.Triggers {
		expirationDate, err := utils.ParseTimeDetectLayout(atr.ExpiryTime, tpr.timezone)
		if err != nil {
			return err
		}
		activationDate, err := utils.ParseTimeDetectLayout(atr.ActivationTime, tpr.timezone)
		if err != nil {
			return err
		}
		minSleep, err := utils.ParseDurationWithSecs(atr.MinSleep)
		if err != nil {
			return err
		}
		if atr.UniqueID == "" {
			atr.UniqueID = utils.GenUUID()
		}
		atr.Filter = strings.Replace(atr.Filter, `'`, `"`, -1)
		atrg.ActionTriggers[idx] = &ActionTrigger{
			UniqueID:       atr.UniqueID,
			ThresholdType:  atr.ThresholdType,
			ThresholdValue: dec.NewFloat(atr.ThresholdValue),
			Recurrent:      atr.Recurrent,
			MinSleep:       minSleep,
			ExpirationDate: expirationDate,
			ActivationDate: activationDate,
			TOR:            atr.TOR,
			Filter:         atr.Filter,
			Weight:         atr.Weight,
			ActionsID:      atr.ActionsTag,
			MinQueuedItems: atr.MinQueuedItems,
		}

	}
	return tpr.ratingStorage.SetActionTriggers(atrg)
}

func (tpr *TpReader) LoadAccountAction(el interface{}) error {
	element := el.(*utils.TpAccountAction)
	tpr.loadStats.Tenants[element.Tenant] = true
	acc := &Account{
		Tenant:        element.Tenant,
		Name:          element.Account,
		AllowNegative: element.AllowNegative,
		Disabled:      element.Disabled,
	}

	if len(element.ActionTriggerTags) > 0 {
		for _, atrTag := range element.ActionTriggerTags {
			if strings.TrimSpace(atrTag) == "" {
				continue
			}
			atrg, err := tpr.ratingStorage.GetActionTriggers(element.Tenant, atrTag, utils.CACHE_SKIP)
			if err != nil || atrg == nil {
				return fmt.Errorf("<LoadAccountActions> could not get action triggers for tag %s (%v)", atrTag, err)
			}
			if acc.TriggerIDs == nil {
				acc.TriggerIDs = utils.StringMap{}
			}
			acc.TriggerIDs[atrTag] = true
		}
		acc.InitCounters()
	}

	if len(element.ActionPlanTags) > 0 {
		for _, aplTag := range element.ActionPlanTags {
			if strings.TrimSpace(aplTag) == "" {
				continue
			}
			apl, err := tpr.ratingStorage.GetActionPlan(element.Tenant, aplTag, utils.CACHE_SKIP)
			if err != nil || apl == nil {
				return fmt.Errorf("<LoadAccountActions> could not get action plan for tag %s (%v)", aplTag, err)
			}
			// add action plan binding and push tasks
			for _, at := range apl.ActionTimings {
				if at.IsASAP() {
					if err = tpr.ratingStorage.PushTask(&Task{
						UUID:      utils.GenUUID(),
						Tenant:    acc.Tenant,
						AccountID: acc.Name,
						ActionsID: at.ActionsID,
					}); err != nil {
						return err
					}
				}
			}
			// remove previous binings
			if err := tpr.ratingStorage.RemoveActionPlanBindings(acc.Tenant, acc.Name, ""); err != nil {
				return err
			}
			// add new binding
			if err := tpr.ratingStorage.SetActionPlanBinding(&ActionPlanBinding{
				Tenant:     acc.Tenant,
				Account:    acc.Name,
				ActionPlan: apl.Name,
			}); err != nil {
				return err
			}
		}
	}
	return tpr.accountingStorage.SetAccount(acc)
}

func (tpr *TpReader) LoadDerivedCharger(el interface{}) error {
	element := el.(*utils.TpDerivedCharger)
	tpr.loadStats.Tenants[element.Tenant] = true
	dcs := &utils.DerivedChargerGroup{
		Direction:      element.Direction,
		Tenant:         element.Tenant,
		Category:       element.Category,
		Account:        element.Account,
		Subject:        element.Subject,
		DestinationIDs: utils.NewStringMap(element.DestinationIDs...),
	}

	for _, tpDc := range element.Chargers {
		tpDc.Filter = strings.Replace(tpDc.Filter, `'`, `"`, -1)
		tpDc.Fields = strings.Replace(tpDc.Fields, `'`, `"`, -1)
		dc, err := utils.NewDerivedCharger(tpDc.RunID, tpDc.Filter, tpDc.Fields)
		if err != nil {
			return err
		}

		if dc.RunID == "" {
			dc.RunID = utils.META_DEFAULT
		}

		dcs.Chargers = append(dcs.Chargers, dc)
	}
	return tpr.ratingStorage.SetDerivedChargers(dcs)
}

func (tpr *TpReader) LoadCdrStats(el interface{}) (err error) {
	element := el.(*utils.TpCdrStats)
	if tpr.loadStats.CDRStats[element.Tenant] == nil {
		tpr.loadStats.CDRStats[element.Tenant] = utils.StringMap{}
	}
	tpr.loadStats.CDRStats[element.Tenant][element.Tag] = true
	tw, err := utils.ParseDurationWithSecs(element.TimeWindow)
	if err != nil {
		return fmt.Errorf("<LoadCdrStats> could parse time window %s for cdr stat %s (%v)", element.TimeWindow, element.Tag, err)
	}
	cs := &CdrStats{
		Tenant:      element.Tenant,
		Name:        element.Tag,
		QueueLength: element.QueueLength,
		TimeWindow:  tw,
		Metrics:     element.Metrics,
		Filter:      element.Filter,
		Disabled:    element.Disabled,
	}
	// make Filter standard json
	cs.Filter = strings.Replace(cs.Filter, `'`, `"`, -1)

	if len(element.ActionTriggerTags) > 0 {
		for _, atrTag := range element.ActionTriggerTags {
			atrg, err := tpr.ratingStorage.GetActionTriggers(element.Tenant, atrTag, utils.CACHE_SKIP)
			if err != nil || atrg == nil {
				return fmt.Errorf("<LoadCdrStats> could not get action triggers for tag %s (%v)", atrTag, err)
			}
			if cs.TriggerIDs == nil {
				cs.TriggerIDs = utils.StringMap{}
			}
			cs.TriggerIDs[atrTag] = true
		}
	}
	return tpr.ratingStorage.SetCdrStats(cs)
}

func (tpr *TpReader) LoadUser(el interface{}) error {
	element := el.(*utils.TpUser)
	tpr.loadStats.UserTenants[element.Tenant] = true
	up := &UserProfile{
		Tenant: element.Tenant,
		Name:   element.Name,
		Query:  element.Query,
		Index:  element.Index,
		Weight: element.Weight,
	}
	up.Query = strings.Replace(up.Query, `'`, `"`, -1)
	if _, err := utils.NewStructQ(up.Query); err != nil {
		return fmt.Errorf("<LoadUser> error parsing Query field: %v", err)
	}

	return tpr.accountingStorage.SetUser(up)
}

func (tpr *TpReader) LoadAlias(el interface{}) error {
	element := el.(*utils.TpAlias)
	tpr.loadStats.Tenants[element.Tenant] = true
	alias := &Alias{
		Direction: element.Direction,
		Tenant:    element.Tenant,
		Category:  element.Category,
		Account:   element.Account,
		Subject:   element.Subject,
		Context:   element.Context,
	}
	for _, tpv := range element.Values {
		av := &AliasValue{
			DestinationID: tpv.DestinationTag,
			Fields:        tpv.Fields,
			Weight:        tpv.Weight,
		}
		// check fields parse error
		av.Fields = strings.Replace(av.Fields, `'`, `"`, -1)
		if _, err := utils.NewStructQ(av.Fields); err != nil {
			return fmt.Errorf("<LoadAlias> error parsing Fields: %v", err)
		}
		alias.Values = append(alias.Values, av)

	}
	for _, tpi := range element.Indexes {
		alias.Index = append(alias.Index, &AliasIndex{Target: tpi.Target, Alias: tpi.Alias})
	}
	return tpr.accountingStorage.SetAlias(alias)
}

func (tpr *TpReader) LoadResourceLimit(el interface{}) error {
	//element := el.(*utils.TpResourceLimit)
	//tpr.loadStats.Tenants[element.Tenant] = true
	return nil
}

func (tpr *TpReader) LoadStats() *LoadStats {
	return tpr.loadStats
}
