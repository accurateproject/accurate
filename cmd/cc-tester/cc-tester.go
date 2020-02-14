package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

var (
	cfg             = config.Get()
	cpuprofile      = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile      = flag.String("memprofile", "", "write memory profile to this file")
	runs            = flag.Int("runs", 10000, "stress cycle number")
	parallel        = flag.Int("parallel", 0, "run n requests in parallel")
	ratingdbHost    = flag.String("tp_host", *cfg.TariffPlanDb.Host, "The RatingDb host to connect to.")
	ratingdbPort    = flag.String("tp_port", *cfg.TariffPlanDb.Port, "The RatingDb port to bind to.")
	ratingdbName    = flag.String("tp_name", *cfg.TariffPlanDb.Name, "The name/number of the RatingDb to connect to.")
	ratingdbUser    = flag.String("tp_user", *cfg.TariffPlanDb.User, "The RatingDb user to sign in as.")
	ratingdbPass    = flag.String("tp_passwd", *cfg.TariffPlanDb.Password, "The RatingDb user's password.")
	accountdbHost   = flag.String("data_host", *cfg.DataDb.Host, "The AccountingDb host to connect to.")
	accountdbPort   = flag.String("data_port", *cfg.DataDb.Port, "The AccountingDb port to bind to.")
	accountdbName   = flag.String("data_name", *cfg.DataDb.Name, "The name/number of the AccountingDb to connect to.")
	accountdbUser   = flag.String("data_user", *cfg.DataDb.User, "The AccountingDb user to sign in as.")
	accountdbPass   = flag.String("data_passwd", *cfg.DataDb.Password, "The AccountingDb user's password.")
	raterAddress    = flag.String("rater_address", "", "Rater address for remote tests. Empty for internal rater.")
	tor             = flag.String("tor", utils.VOICE, "The type of record to use in queries.")
	category        = flag.String("category", "call", "The Record category to test.")
	tenant          = flag.String("tenant", "test", "The type of record to use in queries.")
	subject         = flag.String("subject", "1001", "The rating subject to use in queries.")
	destination     = flag.String("destination", "1002", "The destination to use in queries.")
	json            = flag.Bool("json", false, "Use JSON RPC")
	loadHistorySize = flag.Int("load_history_size", *cfg.DataDb.LoadHistorySize, "Limit the number of records in the load history")
	nilDuration     = time.Duration(0)
)

func durInternalRater(cd *engine.CallDescriptor) (time.Duration, error) {
	ratingDb, err := engine.ConfigureRatingStorage(*ratingdbHost, *ratingdbPort, *ratingdbName, *ratingdbUser, *ratingdbPass, cfg.Cache, *loadHistorySize)
	if err != nil {
		return nilDuration, fmt.Errorf("Could not connect to rating database: %s", err.Error())
	}
	defer ratingDb.Close()
	engine.SetRatingStorage(ratingDb)
	accountDb, err := engine.ConfigureAccountingStorage(*accountdbHost, *accountdbPort, *accountdbName, *accountdbUser, *accountdbPass, cfg.Cache, *loadHistorySize)
	if err != nil {
		return nilDuration, fmt.Errorf("Could not connect to accounting database: %s", err.Error())
	}
	defer accountDb.Close()
	engine.SetAccountingStorage(accountDb)
	if err := ratingDb.PreloadRatingCache(); err != nil {
		return nilDuration, fmt.Errorf("Cache rating error: %s", err.Error())
	}
	if err := accountDb.PreloadAccountingCache(); err != nil {
		return nilDuration, fmt.Errorf("Cache accounting error: %s", err.Error())
	}
	log.Printf("Runnning %d cycles...", *runs)
	var result *engine.CallCost
	j := 0
	start := time.Now()
	for i := 0; i < *runs; i++ {
		result, err = cd.GetCost()
		if *memprofile != "" {
			runtime.MemProfileRate = 1
			runtime.GC()
			f, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(f)
			f.Close()
			break
		}
		j = i
	}
	log.Print(result, j, err)
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	log.Printf("memstats before GC: Kbytes = %d footprint = %d",
		memstats.HeapAlloc/1024, memstats.Sys/1024)
	return time.Since(start), nil
}

func durRemoteRater(cd *engine.CallDescriptor) (time.Duration, error) {
	result := engine.CallCost{}
	var client *rpc.Client
	var err error
	if *json {
		client, err = jsonrpc.Dial("tcp", *raterAddress)
	} else {
		client, err = rpc.Dial("tcp", *raterAddress)
	}

	if err != nil {
		return nilDuration, fmt.Errorf("Could not connect to engine: %s", err.Error())
	}
	defer client.Close()
	start := time.Now()
	if *parallel > 0 {
		// var divCall *rpc.Call
		var sem = make(chan int, *parallel)
		var finish = make(chan int)
		for i := 0; i < *runs; i++ {
			go func() {
				sem <- 1
				client.Call("Responder.GetCost", cd, &result)
				<-sem
				finish <- 1
				// divCall = client.Go("Responder.GetCost", cd, &result, nil)
			}()
		}
		for i := 0; i < *runs; i++ {
			<-finish
		}
		// <-divCall.Done
	} else {
		for j := 0; j < *runs; j++ {
			client.Call("Responder.GetCost", cd, &result)
		}
	}
	log.Println(result)
	return time.Since(start), nil
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	cd := &engine.CallDescriptor{
		TimeStart:     time.Date(2014, time.December, 11, 55, 30, 0, 0, time.UTC),
		TimeEnd:       time.Date(2014, time.December, 11, 55, 31, 0, 0, time.UTC),
		DurationIndex: 60 * time.Second,
		Direction:     "*out",
		TOR:           *tor,
		Category:      *category,
		Tenant:        *tenant,
		Subject:       *subject,
		Destination:   *destination,
	}
	var duration time.Duration
	var err error
	if len(*raterAddress) == 0 {
		duration, err = durInternalRater(cd)
	} else {
		duration, err = durRemoteRater(cd)
	}
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Printf("Elapsed: %d resulted: %f req/s.", duration, float64(*runs)/duration.Seconds())
	}
}
