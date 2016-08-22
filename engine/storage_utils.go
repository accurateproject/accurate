package engine

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
)

// Various helpers to deal with database

func ConfigureRatingStorage(db_type, host, port, name, user, pass, marshaler string, cacheCfg *config.CacheConfig, loadHistorySize int) (db RatingStorage, err error) {
	var d Storage
	switch db_type {
	case utils.REDIS:
		var db_nb int
		db_nb, err = strconv.Atoi(name)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return nil, err
		}
		if port != "" {
			host += ":" + port
		}
		d, err = NewRedisStorage(host, db_nb, pass, marshaler, utils.REDIS_MAX_CONNS, cacheCfg, loadHistorySize)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, nil, cacheCfg, loadHistorySize)
		db = d.(RatingStorage)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s' or '%s'",
			db_type, utils.REDIS, utils.MONGO))
	}
	if err != nil {
		return nil, err
	}
	return d.(RatingStorage), nil
}

func ConfigureAccountingStorage(db_type, host, port, name, user, pass, marshaler string, cacheCfg *config.CacheConfig, loadHistorySize int) (db AccountingStorage, err error) {
	var d AccountingStorage
	switch db_type {
	case utils.REDIS:
		var db_nb int
		db_nb, err = strconv.Atoi(name)
		if err != nil {
			utils.Logger.Crit("Redis db name must be an integer!")
			return nil, err
		}
		if port != "" {
			host += ":" + port
		}
		d, err = NewRedisStorage(host, db_nb, pass, marshaler, utils.REDIS_MAX_CONNS, cacheCfg, loadHistorySize)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, nil, cacheCfg, loadHistorySize)
		db = d.(AccountingStorage)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are '%s' or '%s'",
			db_type, utils.REDIS, utils.MONGO))
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ConfigureStorStorage(db_type, host, port, name, user, pass, marshaler string, maxConn, maxIdleConn int, cdrsIndexes []string) (db Storage, err error) {
	var d Storage
	switch db_type {
	/*
		case utils.REDIS:
			var db_nb int
			db_nb, err = strconv.Atoi(name)
			if err != nil {
				utils.Logger.Crit("Redis db name must be an integer!")
				return nil, err
			}
			if port != "" {
				host += ":" + port
			}
			d, err = NewRedisStorage(host, db_nb, pass, marshaler)
	*/
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, nil, nil, 1)
	case utils.POSTGRES:
		d, err = NewPostgresStorage(host, port, name, user, pass, maxConn, maxIdleConn)
	case utils.MYSQL:
		d, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are [%s, %s, %s]",
			db_type, utils.MYSQL, utils.MONGO, utils.POSTGRES))
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ConfigureLoadStorage(db_type, host, port, name, user, pass, marshaler string, maxConn, maxIdleConn int, cdrsIndexes []string) (db LoadStorage, err error) {
	var d LoadStorage
	switch db_type {
	case utils.POSTGRES:
		d, err = NewPostgresStorage(host, port, name, user, pass, maxConn, maxIdleConn)
	case utils.MYSQL:
		d, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, cdrsIndexes, nil, 1)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are [%s, %s, %s]",
			db_type, utils.MYSQL, utils.MONGO, utils.POSTGRES))
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ConfigureCdrStorage(db_type, host, port, name, user, pass string, maxConn, maxIdleConn int, cdrsIndexes []string) (db CdrStorage, err error) {
	var d CdrStorage
	switch db_type {
	case utils.POSTGRES:
		d, err = NewPostgresStorage(host, port, name, user, pass, maxConn, maxIdleConn)
	case utils.MYSQL:
		d, err = NewMySQLStorage(host, port, name, user, pass, maxConn, maxIdleConn)
	case utils.MONGO:
		d, err = NewMongoStorage(host, port, name, user, pass, cdrsIndexes, nil, 1)
	default:
		err = errors.New(fmt.Sprintf("Unknown db '%s' valid options are [%s, %s, %s]",
			db_type, utils.MYSQL, utils.MONGO, utils.POSTGRES))
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Stores one Cost coming from SM
type SMCost struct {
	CGRID       string
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
