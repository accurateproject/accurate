package cdrc

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

const (
	CSV             = "csv"
	FS_CSV          = "freeswitch_csv"
	UNPAIRED_SUFFIX = ".unpaired"
)

// Understands and processes a specific format of cdr (eg: .csv or .fwv)
type RecordsProcessor interface {
	ProcessNextRecord() ([]*engine.CDR, error) // Process a single record in the CDR file, return a slice of CDRs since based on configuration we can have more templates
	ProcessedRecordsNr() int64
}

/*
One instance  of CDRC will act on one folder.
Common parameters within configs processed:
 * cdrS, cdrFormat, cdrInDir, cdrOutDir, runDelay
Parameters specific per config instance:
 * duMultiplyFactor, cdrSourceId, cdrFilter, cdrFields
*/
func NewCdrc(cdrcCfgs []*config.Cdrc, httpSkipTlsCheck bool, cdrs rpcclient.RpcClientConnection, closeChan chan struct{}, dfltTimezone string, roundDecimals int) (*Cdrc, error) {
	var cdrcCfg *config.Cdrc
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	cdrc := &Cdrc{httpSkipTlsCheck: httpSkipTlsCheck, cdrcCfgs: cdrcCfgs, dfltCdrcCfg: cdrcCfg, timezone: utils.FirstNonEmpty(*cdrcCfg.Timezone, dfltTimezone), cdrs: cdrs,
		closeChan: closeChan, maxOpenFiles: make(chan struct{}, *cdrcCfg.MaxOpenFiles),
	}
	var processFile struct{}
	for i := 0; i < *cdrcCfg.MaxOpenFiles; i++ {
		cdrc.maxOpenFiles <- processFile // Empty initiate so we do not need to wait later when we pop
	}
	var err error
	if cdrc.unpairedRecordsCache, err = NewUnpairedRecordsCache(cdrcCfg.PartialRecordCache.D(), *cdrcCfg.CdrOutDir, cdrcCfg.FieldSeparatorRune()); err != nil {
		return nil, err
	}
	if cdrc.partialRecordsCache, err = NewPartialRecordsCache(cdrcCfg.PartialRecordCache.D(), *cdrcCfg.PartialCacheExpiryAction, *cdrcCfg.CdrOutDir, rune((*cdrcCfg.FieldSeparator)[0]), roundDecimals, cdrc.timezone, cdrc.httpSkipTlsCheck, cdrc.cdrs); err != nil {
		return nil, err
	}
	// Before processing, make sure in and out folders exist
	for _, dir := range []string{*cdrcCfg.CdrInDir, *cdrcCfg.CdrOutDir} {
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			return nil, fmt.Errorf("Nonexistent folder: %s", dir)
		}
	}
	cdrc.httpClient = new(http.Client)
	return cdrc, nil
}

type Cdrc struct {
	httpSkipTlsCheck     bool
	cdrcCfgs             []*config.Cdrc // All cdrc config profiles attached to this CDRC (key will be profile instance name)
	dfltCdrcCfg          *config.Cdrc
	timezone             string
	cdrs                 rpcclient.RpcClientConnection
	httpClient           *http.Client
	closeChan            chan struct{}         // Used to signal config reloads when we need to span different CDRC-Client
	maxOpenFiles         chan struct{}         // Maximum number of simultaneous files processed
	unpairedRecordsCache *UnpairedRecordsCache // Shared between all files in the folder we process
	partialRecordsCache  *PartialRecordsCache
}

// When called fires up folder monitoring, either automated via inotify or manual by sleeping between processing
func (self *Cdrc) Run() error {
	if self.dfltCdrcCfg.RunDelayDuration() == time.Duration(0) { // Automated via inotify
		if err := self.processCdrDir(); err != nil { // process files already there
			utils.Logger.Warn("could not process exiting cdr files", zap.Error(err))
		}
		return self.trackCDRFiles()
	}
	// Not automated, process and sleep approach
	for {
		select {
		case <-self.closeChan: // Exit, reinject closeChan for other CDRCs
			utils.Logger.Info(fmt.Sprintf("<Cdrc> Shutting down CDRC on path %s.", *self.dfltCdrcCfg.CdrInDir))
			return nil
		default:
		}
		self.processCdrDir()
		time.Sleep(self.dfltCdrcCfg.RunDelayDuration())
	}
}

// Watch the specified folder for file moves and parse the files on events
func (self *Cdrc) trackCDRFiles() (err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()
	err = watcher.Add(*self.dfltCdrcCfg.CdrInDir)
	if err != nil {
		return
	}
	utils.Logger.Info("<Cdrc> Monitoring %s for file moves.", zap.String("indir", *self.dfltCdrcCfg.CdrInDir))
	for {
		select {
		case <-self.closeChan: // Exit, reinject closeChan for other CDRCs
			utils.Logger.Info(fmt.Sprintf("<Cdrc> Shutting down CDRC on path %s.", *self.dfltCdrcCfg.CdrInDir))
			return nil
		case ev := <-watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create && (*self.dfltCdrcCfg.CdrFormat != FS_CSV || path.Ext(ev.Name) != ".csv") {
				go func() { //Enable async processing here
					if err = self.processFile(ev.Name); err != nil {
						utils.Logger.Error("processing ", zap.String("file", ev.Name), zap.Error(err))
					}
				}()
			}
		case err := <-watcher.Errors:
			utils.Logger.Error("inotify:", zap.Error(err))
		}
	}
}

// One run over the CDR folder
func (self *Cdrc) processCdrDir() error {
	utils.Logger.Info("<Cdrc> Parsing folder %s for CDR files.", zap.String("indir", *self.dfltCdrcCfg.CdrInDir))
	filesInDir, _ := ioutil.ReadDir(*self.dfltCdrcCfg.CdrInDir)
	for _, file := range filesInDir {
		if *self.dfltCdrcCfg.CdrFormat != FS_CSV || path.Ext(file.Name()) != ".csv" {
			go func() { //Enable async processing here
				if err := self.processFile(path.Join(*self.dfltCdrcCfg.CdrInDir, file.Name())); err != nil {
					utils.Logger.Error("processing", zap.String("file", file.Name()), zap.Error(err))
				}
			}()
		}
	}
	return nil
}

// Processe file at filePath and posts the valid cdr rows out of it
func (self *Cdrc) processFile(filePath string) error {
	if cap(self.maxOpenFiles) != 0 { // 0 goes for no limit
		processFile := <-self.maxOpenFiles // Queue here for maxOpenFiles
		defer func() { self.maxOpenFiles <- processFile }()
	}
	_, fn := path.Split(filePath)
	utils.Logger.Info("<Cdrc> Parsing:", zap.String("path", filePath))
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		utils.Logger.Panic("error closing file:", zap.Error(err))
		return err
	}
	var recordsProcessor RecordsProcessor
	switch *self.dfltCdrcCfg.CdrFormat {
	case CSV, FS_CSV, utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE, utils.PartialCSV:
		csvReader := csv.NewReader(bufio.NewReader(file))
		csvReader.Comma = self.dfltCdrcCfg.FieldSeparatorRune()
		recordsProcessor = NewCsvRecordsProcessor(csvReader, self.timezone, fn, self.dfltCdrcCfg, self.cdrcCfgs,
			self.httpSkipTlsCheck, self.unpairedRecordsCache, self.partialRecordsCache, self.dfltCdrcCfg.CacheDumpFields)
	case utils.FWV:
		recordsProcessor = NewFwvRecordsProcessor(file, self.dfltCdrcCfg, self.cdrcCfgs, self.httpClient, self.httpSkipTlsCheck, self.timezone)
	case utils.XML:
		if recordsProcessor, err = NewXMLRecordsProcessor(file, utils.ParseHierarchyPath(*self.dfltCdrcCfg.CdrPath, ""), self.timezone, self.httpSkipTlsCheck, self.cdrcCfgs); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported CDR format: %s", *self.dfltCdrcCfg.CdrFormat)
	}
	rowNr := 0 // This counts the rows in the file, not really number of CDRs
	cdrsPosted := 0
	timeStart := time.Now()
	for {
		cdrs, err := recordsProcessor.ProcessNextRecord()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			utils.Logger.Error("<cdrc> ", zap.Int("row", rowNr), zap.Error(err))
			continue
		}
		for _, storedCdr := range cdrs { // Send CDRs to CDRS
			var reply string
			if *self.dfltCdrcCfg.DryRun {
				utils.Logger.Info("<Cdrc> DryRun", zap.Any("CDR", storedCdr))
				continue
			}
			if err := self.cdrs.Call("CdrsV1.ProcessCDR", storedCdr, &reply); err != nil {
				utils.Logger.Error("<cdrc> Failed sending ", zap.Any("CDR", storedCdr), zap.Error(err))
			} else if reply != "OK" {
				utils.Logger.Error("<cdrc> Received unexpected reply for ", zap.Any("CDR", storedCdr), zap.String("reply", reply))
			}
			cdrsPosted++
		}
	}
	// Finished with file, move it to processed folder
	newPath := path.Join(*self.dfltCdrcCfg.CdrOutDir, fn)
	if err := os.Rename(filePath, newPath); err != nil {
		utils.Logger.Error("rename", zap.Error(err))
		return err
	}
	utils.Logger.Info("finished processing",
		zap.String("file", fn), zap.String("moved", newPath), zap.Int64("processed", recordsProcessor.ProcessedRecordsNr()), zap.Int("posted", cdrsPosted), zap.Duration("duration", time.Now().Sub(timeStart)))
	return nil
}
