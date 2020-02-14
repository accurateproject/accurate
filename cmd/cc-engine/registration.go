package main

import (
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/accurateproject/accurate/balancer2go"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/scheduler"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"go.uber.org/zap"
)

/*
Listens for SIGTERM, SIGINT, SIGQUIT system signals and shuts down all the registered engines.
*/
func stopBalancerSignalHandler(bal *balancer2go.Balancer, exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c
	utils.Logger.Info("Caught signal sending shutdown to engines", zap.Stringer("sig", sig))
	bal.Shutdown("Responder.Shutdown")
	exitChan <- true
}

func generalSignalHandler(internalCdrStatSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-c
	utils.Logger.Info("Caught signal shuting down cgr-engine", zap.Stringer("sig", sig))
	var dummyInt int
	select {
	case cdrStats := <-internalCdrStatSChan:
		cdrStats.Call("CDRStatsV1.Stop", dummyInt, &dummyInt)
	default:
	}

	exitChan <- true
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and  gracefuly unregister from balancer and closes the storage before exiting.
*/
func stopRaterSignalHandler(internalCdrStatSChan chan rpcclient.RpcClientConnection, exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-c

	utils.Logger.Info("Caught signal, unregistering from balancer", zap.Stringer("sig", sig))
	unregisterFromBalancer(exitChan)
	var dummyInt int
	select {
	case cdrStats := <-internalCdrStatSChan:
		cdrStats.Call("CDRStatsV1.Stop", dummyInt, &dummyInt)
	default:
	}
	exitChan <- true
}

/*
Connects to the balancer and calls unregister RPC method.
*/
func unregisterFromBalancer(exitChan chan bool) {
	client, err := rpc.Dial("tcp", *cfg.Rals.Balancer)
	if err != nil {
		utils.Logger.Panic("Cannot contact the balancer!")
		exitChan <- true
		return
	}
	var reply int
	utils.Logger.Info("Unregistering from balancer", zap.String("bal", *cfg.Rals.Balancer))
	client.Call("Responder.UnRegisterRater", cfg.Listen.RpcGob, &reply)
	if err := client.Close(); err != nil {
		utils.Logger.Panic("Could not close balancer unregistration!")
		exitChan <- true
	}
}

/*
Connects to the balancer and rehisters the engine to the server.
*/
func registerToBalancer(exitChan chan bool) {
	client, err := rpc.Dial("tcp", *cfg.Rals.Balancer)
	if err != nil {
		utils.Logger.Panic("Cannot contact the balancer:", zap.Error(err))
		exitChan <- true
		return
	}
	var reply int
	utils.Logger.Info("Registering to balancer", zap.String("bal", *cfg.Rals.Balancer))
	client.Call("Responder.RegisterRater", cfg.Listen.RpcGob, &reply)
	if err := client.Close(); err != nil {
		utils.Logger.Panic("Could not close balancer registration!")
		exitChan <- true
	}
	utils.Logger.Info("Registration finished!")
}

// Listens for the HUP system signal and gracefuly reloads the timers from database.
func reloadSchedulerSingnalHandler(sched *scheduler.Scheduler, getter engine.RatingStorage) {
	for {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGHUP)
		sig := <-c

		utils.Logger.Info("Caught signal, reloading action timings.", zap.Stringer("sig", sig))
		sched.Reload(true)
	}
}

/*
Listens for the SIGTERM, SIGINT, SIGQUIT system signals and shuts down the session manager.
*/
func shutdownSessionmanagerSingnalHandler(exitChan chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-c
	if smRpc != nil {
		for _, sm := range smRpc.SMs {
			if err := sm.Shutdown(); err != nil {
				utils.Logger.Warn("<SessionManager> ", zap.Error(err))
			}
		}
	}
	exitChan <- true
}
