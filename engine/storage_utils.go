package engine

import (
	"os"
	"path"

	"go.uber.org/zap"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
)

// Various helpers to deal with database

func ConfigureRatingStorage(host, port, name, user, pass string, cacheCfg *config.Cache, loadHistorySize int) (db RatingStorage, err error) {
	utils.Logger.Info("Connecting to ratingDB...", zap.String("host", host), zap.String("port", port), zap.String("db", name), zap.String("user", user))
	db, err = NewMongoStorage(host, port, name, user, pass, utils.TariffPlanDB, nil, cacheCfg, loadHistorySize)
	if err != nil {
		return nil, err
	}
	return
}

func ConfigureAccountingStorage(host, port, name, user, pass string, cacheCfg *config.Cache, loadHistorySize int) (db AccountingStorage, err error) {
	utils.Logger.Info("Connecting to dataDB...", zap.String("host", host), zap.String("port", port), zap.String("db", name), zap.String("user", user))
	db, err = NewMongoStorage(host, port, name, user, pass, utils.DataDB, nil, cacheCfg, loadHistorySize)
	if err != nil {
		return nil, err
	}
	return
}

func ConfigureCdrStorage(host, port, name, user, pass string, maxConn, maxIdleConn int, cdrsIndexes []string) (db CdrStorage, err error) {
	utils.Logger.Info("Connecting to cdrDB...", zap.String("host", host), zap.String("port", port), zap.String("db", name), zap.String("user", user))
	db, err = NewMongoStorage(host, port, name, user, pass, utils.CdrDB, cdrsIndexes, nil, 1)
	if err != nil {
		return nil, err
	}
	return
}

func LoadTariffPlanFromFolder(tpPath, timezone string, ratingDb RatingStorage, accountingDb AccountingStorage) (*TpReader, error) {
	tpr := NewTpReader(ratingDb, accountingDb, timezone)

	if reader, err := os.Open(path.Join(tpPath, utils.DESTINATIONS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpDestination{} }, tpr.LoadDestination); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.DESTINATIONS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.TIMINGS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpTiming{} }, tpr.LoadTiming); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.TIMINGS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.RATES_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpRate{} }, tpr.LoadRate); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.RATES_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.DESTINATION_RATES_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpDestinationRate{} }, tpr.LoadDestinationRate); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.DESTINATION_RATES_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.RATING_PLANS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpRatingPlan{} }, tpr.LoadRatingPlan); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.RATING_PLANS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.RATING_PROFILES_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpRatingProfile{} }, tpr.LoadRatingProfile); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.RATING_PROFILES_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.SHARED_GROUPS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpSharedGroup{} }, tpr.LoadSharedGroup); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.SHARED_GROUPS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.LCRS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpLcrRule{} }, tpr.LoadLCR); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.LCRS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.ACTIONS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpActionGroup{} }, tpr.LoadActionGroup); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.ACTIONS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.ACTION_PLANS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpActionPlan{} }, tpr.LoadActionPlan); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.ACTION_PLANS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.ACTION_TRIGGERS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpActionTrigger{} }, tpr.LoadActionTrigger); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.ACTION_TRIGGERS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.ACCOUNT_ACTIONS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpAccountAction{} }, tpr.LoadAccountAction); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.ACCOUNT_ACTIONS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.DERIVED_CHARGERS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpDerivedCharger{} }, tpr.LoadDerivedCharger); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.DERIVED_CHARGERS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.CDR_STATS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpCdrStats{} }, tpr.LoadCdrStats); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.CDR_STATS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.USERS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpUser{} }, tpr.LoadUser); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.USERS_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.ALIASES_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpAlias{} }, tpr.LoadAlias); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.ALIASES_JSON, zap.Error(err))
	}
	if reader, err := os.Open(path.Join(tpPath, utils.RESOURCE_LIMITS_JSON)); err == nil {
		if err := utils.LoadJSON(reader, func() interface{} { return &utils.TpResourceLimit{} }, tpr.LoadResourceLimit); err != nil {
			return nil, err
		}
		if err := reader.Close(); err != nil {
			return nil, err
		}
	} else {
		utils.Logger.Warn(utils.RESOURCE_LIMITS_JSON, zap.Error(err))
	}
	return tpr, nil
}

// Stores one Cost coming from SM
type SMCost struct {
	UniqueID    string
	RunID       string
	OriginHost  string
	OriginID    string
	CostSource  string
	Usage       float64
	CostDetails *CallCost
}

type AttrCDRSStoreSMCost struct {
	Cost           *SMCost
	CheckDuplicate bool
}

type FakeAPBIterator struct {
	tenant     string
	actionPlan string
	data       []string
	index      int
}

func NewFakeAPBIterator(tenant, actionPlan string, data []string) *FakeAPBIterator {
	return &FakeAPBIterator{tenant: tenant, actionPlan: actionPlan, data: data, index: 0}
}

func (si *FakeAPBIterator) All(result interface{}) error {
	apbs := result.(*[]*ActionPlanBinding)
	for _, acc := range si.data {
		*apbs = append(*apbs, &ActionPlanBinding{
			Tenant:     si.tenant,
			Account:    acc,
			ActionPlan: si.actionPlan,
		})
	}
	return nil
}

func (si *FakeAPBIterator) Close() error {
	si.index = 0
	return nil
}

func (si *FakeAPBIterator) Done() bool {
	return si.index == len(si.data)
}

func (si *FakeAPBIterator) Err() error {
	return nil
}

func (si *FakeAPBIterator) Next(result interface{}) bool {
	if si.index >= len(si.data) {
		return false
	}
	apb := result.(*ActionPlanBinding)
	*apb = ActionPlanBinding{
		Tenant:     si.tenant,
		Account:    si.data[si.index],
		ActionPlan: si.actionPlan,
	}
	si.index++
	return true
}

func (si *FakeAPBIterator) Timeout() bool {
	return false
}
