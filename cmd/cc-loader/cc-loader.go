package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
)

var (
	//separator = flag.String("separator", ",", "Default field separator")
	cfg      = config.Get()
	tpdbHost = flag.String("tp_host", *cfg.TariffPlanDb.Host, "The TariffPlan host to connect to.")
	tpdbPort = flag.String("tp_port", *cfg.TariffPlanDb.Port, "The TariffPlan port to bind to.")
	tpdbName = flag.String("tp_name", *cfg.TariffPlanDb.Name, "The name/number of the TariffPlan to connect to.")
	tpdbUser = flag.String("tp_user", *cfg.TariffPlanDb.User, "The TariffPlan user to sign in as.")
	tpdbPass = flag.String("tp_pass", *cfg.TariffPlanDb.Password, "The TariffPlan user's password.")

	datadbHost = flag.String("data_host", *cfg.DataDb.Host, "The DataDb host to connect to.")
	datadbPort = flag.String("data_port", *cfg.DataDb.Port, "The DataDb port to bind to.")
	datadbName = flag.String("data_name", *cfg.DataDb.Name, "The name/number of the DataDb to connect to.")
	datadbUser = flag.String("data_user", *cfg.DataDb.User, "The DataDb user to sign in as.")
	datadbPass = flag.String("data_pass", *cfg.DataDb.Password, "The DataDb user's password.")

	flush           = flag.Bool("flushdb", false, "Flush the database before importing")
	path            = flag.String("path", "./", "The path to folder containing the data files")
	version         = flag.Bool("version", false, "Prints the application version.")
	verbose         = flag.Bool("verbose", false, "Enable detailed verbose logging output")
	historyServer   = flag.String("history_server", *cfg.Listen.RpcGob, "The history server address:port, empty to disable automaticautomatic  history archiving")
	raterAddress    = flag.String("rater_address", *cfg.Listen.RpcGob, "Rater service to contact for cache reloads, empty to disable automatic cache reloads")
	cdrstatsAddress = flag.String("cdrstats_address", *cfg.Listen.RpcGob, "CDRStats service to contact for data reloads, empty to disable automatic data reloads")
	usersAddress    = flag.String("users_address", *cfg.Listen.RpcGob, "Users service to contact for data reloads, empty to disable automatic data reloads")
	//runID           = flag.String("runid", "", "Uniquely identify an import/load, postpended to some automatic fields")
	loadHistorySize = flag.Int("load_history_size", *cfg.DataDb.LoadHistorySize, "Limit the number of records in the load history")
	timezone        = flag.String("timezone", *cfg.General.DefaultTimezone, `Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>`)
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println("accuRate " + utils.VERSION)
		return
	}
	var errRatingDb, errAccDb, errStorDb, err error
	var ratingDb engine.RatingStorage
	var accountDb engine.AccountingStorage
	var rater, cdrstats, users *rpc.Client
	// Init necessary db connections, only if not already
	// load from csv files to dataDb
	ratingDb, errRatingDb = engine.ConfigureRatingStorage(*tpdbHost, *tpdbPort, *tpdbName, *tpdbUser, *tpdbPass, cfg.Cache, *loadHistorySize)
	accountDb, errAccDb = engine.ConfigureAccountingStorage(*datadbHost, *datadbPort, *datadbName, *datadbUser, *datadbPass, cfg.Cache, *loadHistorySize)
	// Defer databases opened to be closed when we are done
	for _, db := range []engine.Storage{ratingDb, accountDb} {
		if db != nil {
			defer db.Close()
		}
	}
	// Stop on db errors
	for _, err = range []error{errRatingDb, errAccDb, errStorDb} {
		if err != nil {
			log.Fatalf("Could not open database connection: %v", err)
		}
	}
	// load from csv files to dataDb
	tpReader, err := engine.LoadTariffPlanFromFolder(*path, *timezone, ratingDb, accountDb)
	if err != nil {
		log.Fatal(err)
	}
	if *historyServer != "" { // Init scribeAgent so we can store the differences
		if scribeAgent, err := rpcclient.NewRpcClient("tcp", *historyServer, 3, 3, time.Duration(1*time.Second), time.Duration(5*time.Minute), utils.GOB, nil, false); err != nil {
			log.Fatalf("Could not connect to history server, error: %s. Make sure you have properly configured it via -history_server flag.", err.Error())
			return
		} else {
			engine.SetHistoryScribe(scribeAgent)
			//defer scribeAgent.Client.Close()
		}
	} else {
		log.Print("WARNING: Rates history archiving is disabled!")
	}
	if *raterAddress != "" { // Init connection to rater so we can reload it's data
		rater, err = rpc.Dial("tcp", *raterAddress)
		if err != nil {
			log.Fatalf("Could not connect to rater: %s", err.Error())
			return
		}
	} else {
		log.Print("WARNING: Rates automatic cache reloading is disabled!")
	}
	if *cdrstatsAddress != "" { // Init connection to rater so we can reload it's data
		if *cdrstatsAddress == *raterAddress {
			cdrstats = rater
		} else {
			cdrstats, err = rpc.Dial("tcp", *cdrstatsAddress)
			if err != nil {
				log.Fatalf("Could not connect to CDRStats API: %s", err.Error())
				return
			}
		}
	} else {
		log.Print("WARNING: CDRStats automatic data reload is disabled!")
	}
	if *usersAddress != "" { // Init connection to rater so we can reload it's data
		if *usersAddress == *raterAddress {
			users = rater
		} else {
			users, err = rpc.Dial("tcp", *usersAddress)
			if err != nil {
				log.Fatalf("Could not connect to Users API: %s", err.Error())
				return
			}
		}
	} else {
		log.Print("WARNING: Users automatic data reload is disabled!")
	}

	if len(*historyServer) != 0 && *verbose {
		log.Print("Wrote history.")
	}
	loadStats := tpReader.LoadStats()
	// Reload scheduler and cache
	if rater != nil {
		reply := ""

		// Reload cache first since actions could be calling info from within
		if *verbose {
			log.Print("Reloading cache")
		}
		if err = rater.Call("ApierV1.ReloadCache", utils.AttrReloadCache{Tenants: loadStats.Tenants.Slice()}, &reply); err != nil {
			log.Printf("WARNING: Got error on cache reload: %s\n", err.Error())
		}

		if *verbose {
			log.Print("Reloading scheduler")
		}
		if err = rater.Call("ApierV1.ReloadScheduler", "", &reply); err != nil {
			log.Printf("WARNING: Got error on scheduler reload: %s\n", err.Error())
		}

	}
	if cdrstats != nil {
		for tenant, slice := range loadStats.CDRStats {
			statsQueueIDs := slice.Slice()
			if *flush {
				statsQueueIDs = []string{} // Force reload all
			}
			if *verbose {
				log.Print("Reloading CDRStats data for tenant: ", tenant)
			}
			var reply string
			if err := cdrstats.Call("CDRStatsV1.ReloadQueues", utils.AttrStatsQueueIDs{Tenant: tenant, IDs: statsQueueIDs}, &reply); err != nil {
				log.Printf("WARNING: Failed reloading stat queues, error: %s\n", err.Error())
			}
		}
	}

	if users != nil {
		for tenant := range loadStats.UserTenants {
			if *verbose {
				log.Print("Reloading Users data for tenant: ", tenant)
			}
			var reply string
			if err := users.Call("UsersV1.ReloadUsers", engine.AttrReloadUsers{Tenant: tenant}, &reply); err != nil {
				log.Printf("WARNING: Failed reloading users data, error: %s\n", err.Error())
			}
		}
	}
}
