package engine

import (
	"errors"

	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

func CheckVersion(acntDB AccountingStorage) error {
	// get current db version
	if acntDB == nil {
		acntDB = accountingStorage
	}
	dbVersion, err := acntDB.GetStructVersion()
	if err != nil {
		lhIter := acntDB.Iterator(ColLht, "-$natural", nil)
		loadHistory := make([]*utils.LoadInstance, 0)
		lhIter.All(&loadHistory)
		if err := lhIter.Close(); err != nil {
			return err
		}
		utils.Logger.Info("Load history: ", zap.Any("lhist", loadHistory))
		if len(loadHistory) == 0 {
			// no data, write version
			if err := acntDB.SetStructVersion(CurrentVersion); err != nil {
				utils.Logger.Warn("could not write current version to db", zap.Error(err))
			}
		} else if loadHistory[0].LoadID == "" {
			// has data but no version => run migration
			msg := "Could not detect data structures version: run appropriate migration"
			utils.Logger.Panic(msg)
			return errors.New(msg)
		}
	} else {
		// comparing versions
		if len(CurrentVersion.CompareAndMigrate(dbVersion)) > 0 {
			// write the new values
			msg := "Migration needed: please backup cgr data and run cgr-cloader -migrate"
			utils.Logger.Panic(msg)
			return errors.New(msg)
		}
	}
	return nil
}

var (
	CurrentVersion = &StructVersion{
		Destinations:    "1",
		RatingPlans:     "1",
		RatingProfiles:  "1",
		Lcrs:            "1",
		DerivedChargers: "1",
		Actions:         "1",
		ActionPlans:     "1",
		ActionTriggers:  "1",
		SharedGroups:    "1",
		Accounts:        "1",
		CdrStats:        "1",
		Users:           "1",
		Alias:           "1",
		PubSubs:         "1",
		LoadHistory:     "1",
		Cdrs:            "1",
		SMCosts:         "1",
		ResourceLimits:  "1",
	}
)

type StructVersion struct {
	//  rating
	Destinations    string `bson:"destinations"`
	RatingPlans     string `bson:"rating_plans"`
	RatingProfiles  string `bson:"rating_profiles"`
	Lcrs            string `bson:"lcrs"`
	DerivedChargers string `bson:"derived_chargers"`
	Actions         string `bson:"actions"`
	ActionPlans     string `bson:"action_plans"`
	ActionTriggers  string `bson:"action_triggers"`
	SharedGroups    string `bson:"shared_groups"`
	// accounting
	Accounts    string `bson:"accounts"`
	CdrStats    string `bson:"cdr_stats"`
	Users       string `bson:"users"`
	Alias       string `bson:"alias"`
	PubSubs     string `bson:"pub_subs"`
	LoadHistory string `bson:"load_history"`
	// cdr
	Cdrs           string `bson:"cdrs"`
	SMCosts        string `bson:"sm_costs"`
	ResourceLimits string `bson:"resource_limits"`
}

type MigrationInfo struct {
	Prefix         string
	DbVersion      string
	CurrentVersion string
}

func (sv *StructVersion) CompareAndMigrate(dbVer *StructVersion) []*MigrationInfo {
	var migrationInfoList []*MigrationInfo
	if sv.Destinations != dbVer.Destinations {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.DESTINATION_PREFIX,
			DbVersion:      dbVer.Destinations,
			CurrentVersion: CurrentVersion.Destinations,
		})

	}
	if sv.RatingPlans != dbVer.RatingPlans {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.RATING_PLAN_PREFIX,
			DbVersion:      dbVer.RatingPlans,
			CurrentVersion: CurrentVersion.RatingPlans,
		})
	}
	if sv.RatingProfiles != dbVer.RatingProfiles {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.RATING_PROFILE_PREFIX,
			DbVersion:      dbVer.RatingProfiles,
			CurrentVersion: CurrentVersion.RatingProfiles,
		})
	}
	if sv.Lcrs != dbVer.Lcrs {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.LCR_PREFIX,
			DbVersion:      dbVer.Lcrs,
			CurrentVersion: CurrentVersion.Lcrs,
		})
	}
	if sv.DerivedChargers != dbVer.DerivedChargers {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.DERIVEDCHARGERS_PREFIX,
			DbVersion:      dbVer.DerivedChargers,
			CurrentVersion: CurrentVersion.DerivedChargers,
		})
	}
	if sv.Actions != dbVer.Actions {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.ACTION_PREFIX,
			DbVersion:      dbVer.Actions,
			CurrentVersion: CurrentVersion.Actions,
		})
	}
	if sv.ActionPlans != dbVer.ActionPlans {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.ACTION_PLAN_PREFIX,
			DbVersion:      dbVer.ActionPlans,
			CurrentVersion: CurrentVersion.ActionPlans,
		})
	}
	if sv.ActionTriggers != dbVer.ActionTriggers {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.ACTION_TRIGGER_PREFIX,
			DbVersion:      dbVer.ActionTriggers,
			CurrentVersion: CurrentVersion.ActionTriggers,
		})
	}
	if sv.SharedGroups != dbVer.SharedGroups {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.SHARED_GROUP_PREFIX,
			DbVersion:      dbVer.SharedGroups,
			CurrentVersion: CurrentVersion.SharedGroups,
		})
	}
	if sv.Accounts != dbVer.Accounts {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.ACCOUNT_PREFIX,
			DbVersion:      dbVer.Accounts,
			CurrentVersion: CurrentVersion.Accounts,
		})
	}
	if sv.CdrStats != dbVer.CdrStats {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.CDR_STATS_PREFIX,
			DbVersion:      dbVer.CdrStats,
			CurrentVersion: CurrentVersion.CdrStats,
		})
	}
	if sv.Users != dbVer.Users {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.USERS_PREFIX,
			DbVersion:      dbVer.Users,
			CurrentVersion: CurrentVersion.Users,
		})
	}
	if sv.Alias != dbVer.Alias {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.ALIASES_PREFIX,
			DbVersion:      dbVer.Alias,
			CurrentVersion: CurrentVersion.Alias,
		})
	}
	if sv.PubSubs != dbVer.PubSubs {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.PUBSUB_SUBSCRIBERS_PREFIX,
			DbVersion:      dbVer.PubSubs,
			CurrentVersion: CurrentVersion.PubSubs,
		})
	}
	if sv.LoadHistory != dbVer.LoadHistory {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.LOADINST_KEY,
			DbVersion:      dbVer.LoadHistory,
			CurrentVersion: CurrentVersion.LoadHistory,
		})
	}
	if sv.Cdrs != dbVer.Cdrs {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.CDRS_SOURCE,
			DbVersion:      dbVer.RatingPlans,
			CurrentVersion: CurrentVersion.RatingPlans,
		})
	}
	if sv.SMCosts != dbVer.SMCosts {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.SMG,
			DbVersion:      dbVer.SMCosts,
			CurrentVersion: CurrentVersion.SMCosts,
		})
	}
	if sv.ResourceLimits != dbVer.ResourceLimits {
		migrationInfoList = append(migrationInfoList, &MigrationInfo{
			Prefix:         utils.ResourceLimitsPrefix,
			DbVersion:      dbVer.ResourceLimits,
			CurrentVersion: CurrentVersion.ResourceLimits,
		})
	}
	return migrationInfoList
}
