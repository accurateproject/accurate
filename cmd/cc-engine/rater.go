package main

import (
	"time"

	"github.com/accurateproject/accurate/api/v1"
	"github.com/accurateproject/accurate/balancer2go"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/history"
	"github.com/accurateproject/accurate/scheduler"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"go.uber.org/zap"
)

/*func init() {
	gob.Register(map[interface{}]struct{}{})
	gob.Register(engine.Actions{})
	gob.RegisterName("github.com/accurateproject/accurate/engine.ActionPlan", &engine.ActionPlan{})
	gob.Register([]*utils.LoadInstance{})
	gob.RegisterName("github.com/accurateproject/accurate/engine.RatingPlan", &engine.RatingPlan{})
	gob.RegisterName("github.com/accurateproject/accurate/engine.RatingProfile", &engine.RatingProfile{})
	gob.RegisterName("github.com/accurateproject/accurate/utils.DerivedChargers", &utils.DerivedChargers{})
	gob.Register(engine.AliasValues{})
}*/

func startBalancer(internalBalancerChan chan *balancer2go.Balancer, stopHandled *bool, exitChan chan bool) {
	bal := balancer2go.NewBalancer()
	go stopBalancerSignalHandler(bal, exitChan)
	*stopHandled = true
	internalBalancerChan <- bal
}

// Starts rater and reports on chan
func startRater(internalRaterChan chan rpcclient.RpcClientConnection, cacheDoneChan chan struct{}, internalBalancerChan chan *balancer2go.Balancer, internalSchedulerChan chan *scheduler.Scheduler,
	internalCdrStatSChan, internalHistorySChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan chan rpcclient.RpcClientConnection,
	server *utils.Server,
	ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, cdrDb engine.CdrStorage, stopHandled *bool, exitChan chan bool) {
	var waitTasks []chan struct{}

	//Cache load
	cacheTaskChan := make(chan struct{})
	waitTasks = append(waitTasks, cacheTaskChan)
	go func() {
		defer close(cacheTaskChan)

		histIterator := accountDb.Iterator(engine.ColLht, "-$natural", nil)
		var loadHist utils.LoadInstance
		histIterator.Next(&loadHist)
		if err := histIterator.Close(); err != nil {
			utils.Logger.Info("could not get load history", zap.Any("hist", loadHist), zap.Error(err))
			cacheDoneChan <- struct{}{}
			return
		}

		if err := ratingDb.PreloadRatingCache(); err != nil {
			utils.Logger.Panic("Cache rating error:", zap.Error(err))
			exitChan <- true
			return
		}

		cacheDoneChan <- struct{}{}
	}()

	// Retrieve scheduler for it's API methods
	var sched *scheduler.Scheduler // Need the scheduler in APIer
	if *cfg.Scheduler.Enabled {
		schedTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, schedTaskChan)
		go func() {
			defer close(schedTaskChan)
			select {
			case sched = <-internalSchedulerChan:
				internalSchedulerChan <- sched
			case <-time.After(cfg.General.InternalTtl.D()):
				utils.Logger.Panic("<Rater>: Internal scheduler connection timeout.")
				exitChan <- true
				return
			}

		}()
	}
	var bal *balancer2go.Balancer
	if *cfg.Rals.Balancer != "" { // Connection to balancer
		balTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, balTaskChan)
		go func() {
			defer close(balTaskChan)
			if *cfg.Rals.Balancer == utils.MetaInternal {
				select {
				case bal = <-internalBalancerChan:
					internalBalancerChan <- bal // Put it back if someone else is interested about
				case <-time.After(cfg.General.InternalTtl.D()):
					utils.Logger.Panic("<Rater>: Internal balancer connection timeout.")
					exitChan <- true
					return
				}
			} else {
				go registerToBalancer(exitChan)
				go stopRaterSignalHandler(internalCdrStatSChan, exitChan)
				*stopHandled = true
			}
		}()
	}
	var cdrStats *rpcclient.RpcClientPool
	if len(cfg.Rals.CdrStatsConns) != 0 { // Connections to CDRStats
		cdrstatTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, cdrstatTaskChan)
		go func() {
			defer close(cdrstatTaskChan)
			cdrStats, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
				cfg.Rals.CdrStatsConns, internalCdrStatSChan, cfg.General.InternalTtl.D())
			if err != nil {
				utils.Logger.Panic("<RALs> Could not connect to CDRStatS, error:", zap.Error(err))
				exitChan <- true
				return
			}
		}()
	}
	if len(cfg.Rals.HistorysConns) != 0 { // Connection to HistoryS,
		histTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, histTaskChan)
		go func() {
			defer close(histTaskChan)
			if historySConns, err := engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
				cfg.Rals.HistorysConns, internalHistorySChan, cfg.General.InternalTtl.D()); err != nil {
				utils.Logger.Panic("<RALs> Could not connect HistoryS, error:", zap.Error(err))
				exitChan <- true
				return
			} else {
				engine.SetHistoryScribe(historySConns)
			}
		}()
	}
	if len(cfg.Rals.PubsubsConns) != 0 { // Connection to pubsubs
		pubsubTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, pubsubTaskChan)
		go func() {
			defer close(pubsubTaskChan)
			if pubSubSConns, err := engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
				cfg.Rals.PubsubsConns, internalPubSubSChan, cfg.General.InternalTtl.D()); err != nil {
				utils.Logger.Panic("<RALs> Could not connect to PubSubS:", zap.Error(err))
				exitChan <- true
				return
			} else {
				engine.SetPubSub(pubSubSConns)
			}
		}()
	}
	if len(cfg.Rals.AliasesConns) != 0 { // Connection to AliasService
		aliasesTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, aliasesTaskChan)
		go func() {
			defer close(aliasesTaskChan)
			if aliaseSCons, err := engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
				cfg.Rals.AliasesConns, internalAliaseSChan, cfg.General.InternalTtl.D()); err != nil {
				utils.Logger.Panic("<RALs> Could not connect to AliaseS:", zap.Error(err))
				exitChan <- true
				return
			} else {
				engine.SetAliasService(aliaseSCons)
			}
		}()
	}
	var usersConns rpcclient.RpcClientConnection
	if len(cfg.Rals.UsersConns) != 0 { // Connection to UserService
		usersTaskChan := make(chan struct{})
		waitTasks = append(waitTasks, usersTaskChan)
		go func() {
			defer close(usersTaskChan)
			if usersConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
				cfg.Rals.UsersConns, internalUserSChan, cfg.General.InternalTtl.D()); err != nil {
				utils.Logger.Panic("<RALs> Could not connect UserS:", zap.Error(err))
				exitChan <- true
				return
			}
			engine.SetUserService(usersConns)
		}()
	}
	// Wait for all connections to complete before going further
	for _, chn := range waitTasks {
		<-chn
	}
	responder := &engine.Responder{Bal: bal, ExitChan: exitChan}
	responder.SetTimeToLive(cfg.General.ResponseCacheTtl.D(), nil)

	apiRpcV1 := v1.NewAPIV1(ratingDb, accountDb, cdrDb, sched, cfg, responder, cdrStats, usersConns)
	if cdrStats != nil { // ToDo: Fix here properly the init of stats
		responder.Stats = cdrStats
	}

	// internalSchedulerChan shared here
	server.RPCRegister(responder)
	server.RPCRegister(apiRpcV1)

	utils.RegisterRpcParams("", &engine.Stats{})
	utils.RegisterRpcParams("", &v1.CDRStatsV1{})
	utils.RegisterRpcParams("ScribeV1", &history.FileScribe{})
	utils.RegisterRpcParams("PubSubV1", &engine.PubSub{})
	utils.RegisterRpcParams("AliasesV1", &engine.AliasHandler{})
	utils.RegisterRpcParams("UsersV1", &engine.UserMap{})
	utils.RegisterRpcParams("", &v1.CdrsV1{})
	utils.RegisterRpcParams("", &v1.SessionManagerV1{})
	utils.RegisterRpcParams("", &v1.SMGenericV1{})
	utils.RegisterRpcParams("", responder)
	utils.RegisterRpcParams("", apiRpcV1)
	utils.GetRpcParams("")
	internalRaterChan <- responder // Rater done
}
