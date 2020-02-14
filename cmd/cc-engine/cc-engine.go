package main

import (
	"flag"
	"fmt"
	"log"
	//	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/accurateproject/accurate/agents"
	"github.com/accurateproject/accurate/api/v1"
	"github.com/accurateproject/accurate/balancer2go"
	"github.com/accurateproject/accurate/cache2go"
	"github.com/accurateproject/accurate/cdrc"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/history"
	"github.com/accurateproject/accurate/scheduler"
	"github.com/accurateproject/accurate/sessionmanager"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"go.uber.org/zap"
)

const (
	JSON     = "json"
	GOB      = "gob"
	POSTGRES = "postgres"
	MYSQL    = "mysql"
	MONGO    = "mongo"
	REDIS    = "redis"
	SAME     = "same"
	FS       = "freeswitch"
	KAMAILIO = "kamailio"
	OSIPS    = "opensips"
)

var (
	cfgDir            = flag.String("cfgdir", utils.CONFIG_DIR, "Configuration directory path.")
	version           = flag.Bool("version", false, "Prints the application version.")
	raterEnabled      = flag.Bool("rater", false, "Enforce starting of the rater daemon overwriting config")
	schedEnabled      = flag.Bool("scheduler", false, "Enforce starting of the scheduler daemon .overwriting config")
	cdrsEnabled       = flag.Bool("cdrs", false, "Enforce starting of the cdrs daemon overwriting config")
	pidFile           = flag.String("pid", "", "Write pid file")
	cpuprofile        = flag.String("cpuprofile", "", "write cpu profile to file")
	blockprofile      = flag.Bool("blockprofile", false, "enable goroutine contention profiling")
	scheduledShutdown = flag.String("scheduled_shutdown", "", "shutdown the engine after this duration")
	singlecpu         = flag.Bool("singlecpu", false, "Run on single CPU core")
	logLevel          = flag.Int("log_level", -1, "Log level (0-emergency to 7-debug)")

	smRpc *v1.SessionManagerV1
	err   error
	cfg   *config.Config
)

func startCdrcs(internalCdrSChan, internalRaterChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	cdrcInitialized := false           // Control whether the cdrc was already initialized (so we don't reload in that case)
	var cdrcChildrenChan chan struct{} // Will use it to communicate with the children of one fork
	//for {
	select {
	case <-exitChan: // Stop forking CDRCs
		break
	default: //FIXME: Consume the load request and wait for a new one
		if cdrcInitialized {
			utils.Logger.Info("<CDRC> Configuration reload")
			close(cdrcChildrenChan) // Stop all the children of the previous run
		}
		cdrcChildrenChan = make(chan struct{})
	}
	// Start CDRCs
	for _, cdrcCfgs := range cfg.CdrcProfiles() {
		var enabledCfgs []*config.Cdrc
		for _, cdrcCfg := range cdrcCfgs { // Take a random config out since they should be the same
			if *cdrcCfg.Enabled {
				enabledCfgs = append(enabledCfgs, cdrcCfg)
			}
		}

		if len(enabledCfgs) != 0 {
			go startCdrc(internalCdrSChan, internalRaterChan, enabledCfgs, *cfg.General.HttpSkipTlsVerify, cdrcChildrenChan, exitChan)
		} else {
			utils.Logger.Info("<CDRC> No enabled CDRC clients")
		}
	}
	cdrcInitialized = true // Initialized
	//}
}

// Fires up a cdrc instance
func startCdrc(internalCdrSChan, internalRaterChan chan rpcclient.RpcClientConnection, cdrcCfgs []*config.Cdrc, httpSkipTlsCheck bool,
	closeChan chan struct{}, exitChan chan bool) {
	var cdrcCfg *config.Cdrc
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	cdrsConn, err := engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
		cdrcCfg.CdrsConns, internalCdrSChan, cfg.General.InternalTtl.D())
	if err != nil {
		utils.Logger.Panic("<CDRC> Could not connect to CDRS via RPC:", zap.Error(err))
		exitChan <- true
		return
	}
	cdrc, err := cdrc.NewCdrc(cdrcCfgs, httpSkipTlsCheck, cdrsConn, closeChan, *cfg.General.DefaultTimezone, *cfg.General.RoundingDecimals)
	if err != nil {
		utils.Logger.Panic("Cdrc config parsing error:", zap.Error(err))
		exitChan <- true
		return
	}
	if err := cdrc.Run(); err != nil {
		utils.Logger.Panic("Cdrc run error:", zap.Error(err))
		exitChan <- true // If run stopped, something is bad, stop the application
	}
}

func startSmGeneric(internalSMGChan chan *sessionmanager.SMGeneric, internalRaterChan, internalCDRSChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting accuRate SMGeneric service.")
	var ralsConns, cdrsConn *rpcclient.RpcClientPool
	if len(cfg.SmGeneric.RalsConns) != 0 {
		ralsConns, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmGeneric.RalsConns, internalRaterChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMGeneric> Could not connect to RALs:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmGeneric.CdrsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmGeneric.CdrsConns, internalCDRSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMGeneric> Could not connect to RALs:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	sm := sessionmanager.NewSMGeneric(cfg, ralsConns, cdrsConn, *cfg.General.DefaultTimezone)
	if err = sm.Connect(); err != nil {
		utils.Logger.Error("<SMGeneric> error:", zap.Error(err))
	}
	// Pass internal connection via BiRPCClient
	internalSMGChan <- sm
	// Register RPC handler
	smgRpc := v1.NewSMGenericV1(sm)
	server.RPCRegister(smgRpc)
	// Register BiRpc handlers
	//server.BiRPCRegister(v1.NewSMGenericBiRpcV1(sm))
	smgBiRpc := v1.NewSMGenericBiRpcV1(sm)
	for method, handler := range smgBiRpc.Handlers() {
		server.BiRPCRegisterName(method, handler)
	}
}

func startSmAsterisk(internalSMGChan chan *sessionmanager.SMGeneric, exitChan chan bool) {
	utils.Logger.Info("Starting accuRate SmAsterisk service.")
	/*
		var smgConn *rpcclient.RpcClientPool
		if len(cfg.SmAsteriskCfg().SMGConns) != 0 {
			smgConn, err = engine.NewRPCPool(rpcclient.POOL_BROADCAST, cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
				cfg.SmAsteriskCfg().SMGConns, internalSMGChan, cfg.General.InternalTtl.D())
			if err != nil {
				utils.Logger.Panic("<SmAsterisk> Could not connect to SMG: %s", zap.Error(err))
				exitChan <- true
				return
			}
		}
	*/
	smg := <-internalSMGChan
	internalSMGChan <- smg
	birpcClnt := utils.NewBiRPCInternalClient(smg)
	for connIdx := range cfg.SmAsterisk.AsteriskConns { // Instantiate connections towards asterisk servers
		sma, err := sessionmanager.NewSMAsterisk(cfg, connIdx, birpcClnt)
		if err != nil {
			utils.Logger.Error("<SmAsterisk> error:", zap.Error(err))
			exitChan <- true
			return
		}
		if err = sma.ListenAndServe(); err != nil {
			utils.Logger.Error("<SmAsterisk> runtime error:", zap.Error(err))
		}
	}
	exitChan <- true
}

func startDiameterAgent(internalSMGChan chan *sessionmanager.SMGeneric, internalPubSubSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	utils.Logger.Info("Starting accuRate DiameterAgent service.")
	smgChan := make(chan rpcclient.RpcClientConnection, 1) // Use it to pass smg
	go func(internalSMGChan chan *sessionmanager.SMGeneric, smgChan chan rpcclient.RpcClientConnection) {
		// Need this to pass from *sessionmanager.SMGeneric to rpcclient.RpcClientConnection
		smg := <-internalSMGChan
		internalSMGChan <- smg
		smgChan <- smg
	}(internalSMGChan, smgChan)
	var smgConn, pubsubConn *rpcclient.RpcClientPool

	if len(cfg.DiameterAgent.SmGenericConns) != 0 {
		smgConn, err = engine.NewRPCPool(rpcclient.POOL_BROADCAST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.DiameterAgent.SmGenericConns, smgChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<DiameterAgent> Could not connect to SMG:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.DiameterAgent.PubsubsConns) != 0 {
		pubsubConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.DiameterAgent.PubsubsConns, internalPubSubSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<DiameterAgent> Could not connect to PubSubS:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	da, err := agents.NewDiameterAgent(cfg, smgConn, pubsubConn)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("<DiameterAgent> error: %v", err))
		exitChan <- true
		return
	}
	if err = da.ListenAndServe(); err != nil {
		utils.Logger.Error("<DiameterAgent> error:", zap.Error(err))
	}
	exitChan <- true
}

func startSmFreeSWITCH(internalRaterChan, internalCDRSChan, rlsChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting accuRate SMFreeSWITCH service.")
	var ralsConn, cdrsConn, rlsConn *rpcclient.RpcClientPool
	if len(cfg.SmFreeswitch.RalsConns) != 0 {
		ralsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmFreeswitch.RalsConns, internalRaterChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMFreeSWITCH> Could not connect to RAL:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmFreeswitch.CdrsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmFreeswitch.CdrsConns, internalCDRSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMFreeSWITCH> Could not connect to RAL:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmFreeswitch.RlsConns) != 0 {
		rlsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmFreeswitch.RlsConns, rlsChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMFreeSWITCH> Could not connect to RLs:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	sm := sessionmanager.NewFSSessionManager(cfg.SmFreeswitch, ralsConn, cdrsConn, rlsConn, *cfg.General.DefaultTimezone)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err = sm.Connect(); err != nil {
		utils.Logger.Error("<SMFreeSWITCH> error:", zap.Error(err))
	}
	exitChan <- true
}

func startSmKamailio(internalRaterChan, internalCDRSChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting accuRate SMKamailio service.")
	var ralsConn, cdrsConn *rpcclient.RpcClientPool
	if len(cfg.SmKamailio.RalsConns) != 0 {
		ralsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmKamailio.RalsConns, internalRaterChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMKamailio> Could not connect to RAL:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmKamailio.CdrsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmKamailio.CdrsConns, internalCDRSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMKamailio> Could not connect to RAL:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	sm, _ := sessionmanager.NewKamailioSessionManager(cfg.SmKamailio, ralsConn, cdrsConn, *cfg.General.DefaultTimezone)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err = sm.Connect(); err != nil {
		utils.Logger.Error("<SMKamailio> error:", zap.Error(err))
	}
	exitChan <- true
}

func startSmOpenSIPS(internalRaterChan, internalCDRSChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, exitChan chan bool) {
	utils.Logger.Info("Starting accuRate SMOpenSIPS service.")
	var ralsConn, cdrsConn *rpcclient.RpcClientPool
	if len(cfg.SmOpensips.RalsConns) != 0 {
		ralsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmOpensips.RalsConns, internalRaterChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMOpenSIPS> Could not connect to RALs:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.SmOpensips.CdrsConns) != 0 {
		cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.SmOpensips.CdrsConns, internalCDRSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<SMOpenSIPS> Could not connect to CDRs:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	sm, _ := sessionmanager.NewOSipsSessionManager(cfg.SmOpensips, *cfg.General.Reconnects, ralsConn, cdrsConn, *cfg.General.DefaultTimezone)
	smRpc.SMs = append(smRpc.SMs, sm)
	if err := sm.Connect(); err != nil {
		utils.Logger.Error("<SM-OpenSIPS> error:", zap.Error(err))
	}
	exitChan <- true
}

func startCDRS(internalCdrSChan chan rpcclient.RpcClientConnection, cdrDb engine.CdrStorage, dataDB engine.AccountingStorage,
	internalRaterChan chan rpcclient.RpcClientConnection, internalPubSubSChan chan rpcclient.RpcClientConnection,
	internalUserSChan chan rpcclient.RpcClientConnection, internalAliaseSChan chan rpcclient.RpcClientConnection,
	internalCdrStatSChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting accuRate CDRS service.")
	var ralConn, pubSubConn, usersConn, aliasesConn, statsConn *rpcclient.RpcClientPool
	if len(cfg.Cdrs.RalsConns) != 0 { // Conn pool towards RAL
		ralConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.Cdrs.RalsConns, internalRaterChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<CDRS> Could not connect to RAL:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.Cdrs.PubsubsConns) != 0 { // Pubsub connection init
		pubSubConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.Cdrs.PubsubsConns, internalPubSubSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<CDRS> Could not connect to PubSubSystem:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.Cdrs.UsersConns) != 0 { // Users connection init
		usersConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.Cdrs.UsersConns, internalUserSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<CDRS> Could not connect to UserS:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	if len(cfg.Cdrs.AliasesConns) != 0 { // Aliases connection init
		aliasesConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.Cdrs.AliasesConns, internalAliaseSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<CDRS> Could not connect to AliaseS:", zap.Error(err))
			exitChan <- true
			return
		}
	}

	if len(cfg.Cdrs.CdrStatsConns) != 0 { // Stats connection init
		statsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.Cdrs.CdrStatsConns, internalCdrStatSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<CDRS> Could not connect to StatS:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	cdrServer, _ := engine.NewCdrServer(cfg, cdrDb, dataDB, ralConn, pubSubConn, usersConn, aliasesConn, statsConn)
	cdrServer.SetTimeToLive(cfg.General.ResponseCacheTtl.D(), nil)
	utils.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrServer.RegisterHandlersToServer(server)
	utils.Logger.Info("Registering CDRS RPC service.")
	cdrSrv := v1.CdrsV1{CdrSrv: cdrServer}
	server.RPCRegister(&cdrSrv)
	// Make the cdr server available for internal communication
	server.RPCRegister(cdrServer) // register CdrServer for internal usage (TODO: refactor this)
	internalCdrSChan <- cdrServer // Signal that cdrS is operational
}

func startScheduler(internalSchedulerChan chan *scheduler.Scheduler, cacheDoneChan chan struct{}, ratingDb engine.RatingStorage, exitChan chan bool) {
	// Wait for cache to load data before starting
	cacheDone := <-cacheDoneChan
	cacheDoneChan <- cacheDone
	utils.Logger.Info("Starting accuRate Scheduler.")
	sched := scheduler.NewScheduler(ratingDb)
	go reloadSchedulerSingnalHandler(sched, ratingDb)
	time.Sleep(1)
	internalSchedulerChan <- sched
	sched.Reload(true)
	sched.Loop()
	exitChan <- true // Should not get out of loop though
}

func startCdrStats(internalCdrStatSChan chan rpcclient.RpcClientConnection, ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, cdrDb engine.CdrStorage, server *utils.Server) {
	cdrStats := engine.NewStats(ratingDb, accountDb, cdrDb)
	server.RPCRegister(cdrStats)
	server.RPCRegister(&v1.CDRStatsV1{CdrStats: cdrStats}) // Public APIs
	internalCdrStatSChan <- cdrStats
}

func startHistoryServer(internalHistorySChan chan rpcclient.RpcClientConnection, server *utils.Server, exitChan chan bool) {
	scribeServer, err := history.NewFileScribe(*cfg.Historys.HistoryDir, cfg.Historys.SaveInterval.D())
	if err != nil {
		utils.Logger.Panic("<HistoryServer> Could not start, error:", zap.Error(err))
		exitChan <- true
	}
	server.RPCRegisterName("HistoryV1", scribeServer)
	internalHistorySChan <- scribeServer
}

func startPubSubServer(internalPubSubSChan chan rpcclient.RpcClientConnection, accountDb engine.AccountingStorage, server *utils.Server) {
	pubSubServer := engine.NewPubSub(accountDb, *cfg.General.HttpSkipTlsVerify)
	server.RPCRegisterName("PubSubV1", pubSubServer)
	internalPubSubSChan <- pubSubServer
}

// ToDo: Make sure we are caching before starting this one
func startAliasesServer(internalAliaseSChan chan rpcclient.RpcClientConnection, accountDb engine.AccountingStorage, server *utils.Server, exitChan chan bool) {
	aliasesServer := engine.NewAliasHandler(accountDb)
	server.RPCRegisterName("AliasesV1", aliasesServer)

	histIterator := accountDb.Iterator(engine.ColLht, "-$natural", nil)
	var loadHist utils.LoadInstance
	histIterator.Next(&loadHist)
	if err := histIterator.Close(); err != nil {
		utils.Logger.Info("could not get load history:", zap.Any("hist", loadHist), zap.Error(err))
		internalAliaseSChan <- aliasesServer
		return
	}

	if err := accountDb.PreloadAccountingCache(); err != nil {
		utils.Logger.Panic("<Aliases> Could not start, error:", zap.Error(err))
		exitChan <- true
		return
	}

	internalAliaseSChan <- aliasesServer
}

func startUsersServer(internalUserSChan chan rpcclient.RpcClientConnection, accountDb engine.AccountingStorage, server *utils.Server, exitChan chan bool) {
	userServer, err := engine.NewUserMap(accountDb, cfg.Users.Indexes)
	if err != nil {
		utils.Logger.Panic("<UsersService> Could not start, error:", zap.Error(err))
		exitChan <- true
		return
	}
	server.RPCRegisterName("UsersV1", userServer)
	internalUserSChan <- userServer
}

func startResourceLimiterService(internalRLSChan, internalCdrStatSChan chan rpcclient.RpcClientConnection, cfg *config.Config,
	accountDb engine.AccountingStorage, server *utils.Server, exitChan chan bool) {
	var statsConn *rpcclient.RpcClientPool
	if len(cfg.Rls.CdrStatsConns) != 0 { // Stats connection init
		statsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, *cfg.General.ConnectAttempts, *cfg.General.Reconnects, cfg.General.ConnectTimeout.D(), cfg.General.ReplyTimeout.D(),
			cfg.Rls.CdrStatsConns, internalCdrStatSChan, cfg.General.InternalTtl.D())
		if err != nil {
			utils.Logger.Panic("<RLs> Could not connect to StatS:", zap.Error(err))
			exitChan <- true
			return
		}
	}
	rls, err := engine.NewResourceLimiterService(cfg, accountDb, statsConn)
	if err != nil {
		utils.Logger.Panic("<RLs> Could not init, error:", zap.Error(err))
		exitChan <- true
		return
	}
	utils.Logger.Info("Starting ResourceLimiter service")
	if err := rls.ListenAndServe(); err != nil {
		utils.Logger.Panic("<RLs> Could not start, error:", zap.Error(err))
		exitChan <- true
		return
	}
	server.RPCRegisterName("RLsV1", rls)
	internalRLSChan <- rls
}

func startRpc(server *utils.Server, internalRaterChan,
	internalCdrSChan, internalCdrStatSChan, internalHistorySChan, internalPubSubSChan, internalUserSChan,
	internalAliaseSChan chan rpcclient.RpcClientConnection, internalSMGChan chan *sessionmanager.SMGeneric) {
	select { // Any of the rpc methods will unlock listening to rpc requests
	case resp := <-internalRaterChan:
		internalRaterChan <- resp
	case cdrs := <-internalCdrSChan:
		internalCdrSChan <- cdrs
	case cdrstats := <-internalCdrStatSChan:
		internalCdrStatSChan <- cdrstats
	case hist := <-internalHistorySChan:
		internalHistorySChan <- hist
	case pubsubs := <-internalPubSubSChan:
		internalPubSubSChan <- pubsubs
	case users := <-internalUserSChan:
		internalUserSChan <- users
	case aliases := <-internalAliaseSChan:
		internalAliaseSChan <- aliases
	case smg := <-internalSMGChan:
		internalSMGChan <- smg
	}
	go server.ServeJSON(*cfg.Listen.RpcJson)
	go server.ServeGOB(*cfg.Listen.RpcGob)
	go server.ServeHTTP(*cfg.Listen.Http)
}

func writePid() {
	utils.Logger.Info(*pidFile)
	f, err := os.Create(*pidFile)
	if err != nil {
		log.Fatal("Could not write pid file: ", err)
	}
	f.WriteString(strconv.Itoa(os.Getpid()))
	if err := f.Close(); err != nil {
		log.Fatal("Could not write pid file: ", err)
	}
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("accuRate " + utils.VERSION)
		return
	}
	if *pidFile != "" {
		writePid()
	}
	if *singlecpu {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}
	exitChan := make(chan bool)
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

	}
	if *blockprofile {
		runtime.SetBlockProfileRate(1)
	}
	if *scheduledShutdown != "" {
		shutdownDur, err := utils.ParseDurationWithSecs(*scheduledShutdown)
		if err != nil {
			log.Fatal(err)
		}
		go func() { // Schedule shutdown
			time.Sleep(shutdownDur)
			exitChan <- true
		}()
	}
	err = config.LoadPath(*cfgDir)
	if err != nil {
		utils.Logger.Panic("Could not parse config exiting!", zap.Error(err))
		return
	}
	cfg = config.Get()

	cache2go.NewCache(cfg.Cache)

	if *raterEnabled {
		*cfg.Rals.Enabled = *raterEnabled
	}
	if *schedEnabled {
		*cfg.Scheduler.Enabled = *schedEnabled
	}
	if *cdrsEnabled {
		*cfg.Cdrs.Enabled = *cdrsEnabled
	}
	var ratingDb engine.RatingStorage
	var accountDb engine.AccountingStorage
	var cdrDb engine.CdrStorage

	if *cfg.Rals.Enabled || *cfg.Scheduler.Enabled || *cfg.CdrStats.Enabled { // Only connect to dataDb if necessary
		ratingDb, err = engine.ConfigureRatingStorage(*cfg.TariffPlanDb.Host, *cfg.TariffPlanDb.Port,
			*cfg.TariffPlanDb.Name, *cfg.TariffPlanDb.User, *cfg.TariffPlanDb.Password, cfg.Cache, *cfg.DataDb.LoadHistorySize)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Panic("Could not configure ratingDB exiting!", zap.Error(err))
			return
		}
		defer ratingDb.Close()
		engine.SetRatingStorage(ratingDb)
	}
	if *cfg.Rals.Enabled || *cfg.CdrStats.Enabled || *cfg.Pubsubs.Enabled || *cfg.Aliases.Enabled || *cfg.Users.Enabled {
		accountDb, err = engine.ConfigureAccountingStorage(*cfg.DataDb.Host, *cfg.DataDb.Port,
			*cfg.DataDb.Name, *cfg.DataDb.User, *cfg.DataDb.Password, cfg.Cache, *cfg.DataDb.LoadHistorySize)
		if err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Panic("Could not configure dataDB exiting!", zap.Error(err))
			return
		}
		defer accountDb.Close()
		engine.SetAccountingStorage(accountDb)
		if err := engine.CheckVersion(nil); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	if *cfg.Rals.Enabled || *cfg.Cdrs.Enabled || *cfg.Scheduler.Enabled { // Only connect to storDb if necessary
		cdrDb, err = engine.ConfigureCdrStorage(*cfg.CdrDb.Host, *cfg.CdrDb.Port,
			*cfg.CdrDb.Name, *cfg.CdrDb.User, *cfg.CdrDb.Password, *cfg.CdrDb.MaxOpenConns, *cfg.CdrDb.MaxIdleConns, cfg.CdrDb.CdrsIndexes)
		if err != nil { // Cannot configure logger database, show stopper
			utils.Logger.Panic("Could not configure cdrDB exiting!", zap.Error(err))
			return
		}
		defer cdrDb.Close()
		engine.SetCdrStorage(cdrDb)

	}
	engine.SetRoundingDecimals(*cfg.General.RoundingDecimals)
	engine.SetRpSubjectPrefixMatching(*cfg.Rals.RpSubjectPrefixMatching)
	engine.SetLcrSubjectPrefixMatching(*cfg.Rals.LcrSubjectPrefixMatching)
	if err := engine.InitSimpleAccounts(); err != nil {
		utils.Logger.Panic("Could not configure simple accounts exiting!", zap.Error(err))
		return
	}
	stopHandled := false

	// Rpc/http server
	server := new(utils.Server)

	// Async starts here, will follow cgrates.json start order

	// Define internal connections via channels
	internalBalancerChan := make(chan *balancer2go.Balancer, 1)
	internalRaterChan := make(chan rpcclient.RpcClientConnection, 1)
	cacheDoneChan := make(chan struct{}, 1)
	internalSchedulerChan := make(chan *scheduler.Scheduler, 1)
	internalCdrSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalCdrStatSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalHistorySChan := make(chan rpcclient.RpcClientConnection, 1)
	internalPubSubSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalUserSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAliaseSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalSMGChan := make(chan *sessionmanager.SMGeneric, 1)
	internalRLSChan := make(chan rpcclient.RpcClientConnection, 1)
	// Start balancer service
	if *cfg.Balancer.Enabled {
		go startBalancer(internalBalancerChan, &stopHandled, exitChan) // Not really needed async here but to cope with uniformity
	}

	// Start rater service
	if *cfg.Rals.Enabled {
		go startRater(internalRaterChan, cacheDoneChan, internalBalancerChan, internalSchedulerChan, internalCdrStatSChan, internalHistorySChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan,
			server, ratingDb, accountDb, cdrDb, &stopHandled, exitChan)
	}

	// Start Scheduler
	if *cfg.Scheduler.Enabled {
		go startScheduler(internalSchedulerChan, cacheDoneChan, ratingDb, exitChan)
	}

	// Start CDR Server
	if *cfg.Cdrs.Enabled {
		go startCDRS(internalCdrSChan, cdrDb, accountDb,
			internalRaterChan, internalPubSubSChan, internalUserSChan, internalAliaseSChan, internalCdrStatSChan, server, exitChan)
	}

	// Start CDR Stats server
	if *cfg.CdrStats.Enabled {
		go startCdrStats(internalCdrStatSChan, ratingDb, accountDb, cdrDb, server)
	}

	// Start CDRC components if necessary
	go startCdrcs(internalCdrSChan, internalRaterChan, exitChan)

	// Start SM-Generic
	if *cfg.SmGeneric.Enabled {
		go startSmGeneric(internalSMGChan, internalRaterChan, internalCdrSChan, server, exitChan)
	}
	// Start SM-FreeSWITCH
	if *cfg.SmFreeswitch.Enabled {
		go startSmFreeSWITCH(internalRaterChan, internalCdrSChan, internalRLSChan, cdrDb, exitChan)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler(exitChan)
	}

	// Start SM-Kamailio
	if *cfg.SmKamailio.Enabled {
		go startSmKamailio(internalRaterChan, internalCdrSChan, cdrDb, exitChan)
	}

	// Start SM-OpenSIPS
	if *cfg.SmOpensips.Enabled {
		go startSmOpenSIPS(internalRaterChan, internalCdrSChan, cdrDb, exitChan)
	}

	// Register session manager service // FixMe: make sure this is thread safe
	if *cfg.SmGeneric.Enabled || *cfg.SmFreeswitch.Enabled || *cfg.SmKamailio.Enabled || *cfg.SmOpensips.Enabled || *cfg.SmAsterisk.Enabled { // Register SessionManagerV1 service
		smRpc = new(v1.SessionManagerV1)
		server.RPCRegister(smRpc)
	}

	if *cfg.SmAsterisk.Enabled {
		go startSmAsterisk(internalSMGChan, exitChan)
	}

	if *cfg.DiameterAgent.Enabled {
		go startDiameterAgent(internalSMGChan, internalPubSubSChan, exitChan)
	}

	// Start HistoryS service
	if *cfg.Historys.Enabled {
		go startHistoryServer(internalHistorySChan, server, exitChan)
	}

	// Start PubSubS service
	if *cfg.Pubsubs.Enabled {
		go startPubSubServer(internalPubSubSChan, accountDb, server)
	}

	// Start Aliases service
	if *cfg.Aliases.Enabled {
		go startAliasesServer(internalAliaseSChan, accountDb, server, exitChan)
	}

	// Start users service
	if *cfg.Users.Enabled {
		go startUsersServer(internalUserSChan, accountDb, server, exitChan)
	}

	// Start RL service
	if *cfg.Rls.Enabled {
		go startResourceLimiterService(internalRLSChan, internalCdrStatSChan, cfg, accountDb, server, exitChan)
	}

	// Serve rpc connections
	go startRpc(server, internalRaterChan, internalCdrSChan, internalCdrStatSChan, internalHistorySChan,
		internalPubSubSChan, internalUserSChan, internalAliaseSChan, internalSMGChan)
	<-exitChan

	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warn("Could not remove pid file:", zap.Error(err))
		}
	}
	utils.Logger.Info("Stopped all components. accuRate shutdown!")
}
