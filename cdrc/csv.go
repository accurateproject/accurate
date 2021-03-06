package cdrc

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func NewCsvRecordsProcessor(csvReader *csv.Reader, timezone, fileName string,
	dfltCdrcCfg *config.Cdrc, cdrcCfgs []*config.Cdrc,
	httpSkipTlsCheck bool, unpairedRecordsCache *UnpairedRecordsCache, partialRecordsCache *PartialRecordsCache, cacheDumpFields []*config.CdrField) *CsvRecordsProcessor {
	return &CsvRecordsProcessor{csvReader: csvReader, timezone: timezone, fileName: fileName,
		dfltCdrcCfg: dfltCdrcCfg, cdrcCfgs: cdrcCfgs, httpSkipTlsCheck: httpSkipTlsCheck, unpairedRecordsCache: unpairedRecordsCache,
		partialRecordsCache: partialRecordsCache, partialCacheDumpFields: cacheDumpFields}

}

type CsvRecordsProcessor struct {
	csvReader              *csv.Reader
	timezone               string // Timezone for CDRs which are not clearly specifying it
	fileName               string
	dfltCdrcCfg            *config.Cdrc
	cdrcCfgs               []*config.Cdrc
	processedRecordsNr     int64 // Number of content records in file
	httpSkipTlsCheck       bool
	unpairedRecordsCache   *UnpairedRecordsCache // Shared by cdrc so we can cache for all files in a folder
	partialRecordsCache    *PartialRecordsCache  // Cache records which are of type "Partial"
	partialCacheDumpFields []*config.CdrField
}

func (self *CsvRecordsProcessor) ProcessedRecordsNr() int64 {
	return self.processedRecordsNr
}

func (self *CsvRecordsProcessor) ProcessNextRecord() ([]*engine.CDR, error) {
	record, err := self.csvReader.Read()
	if err != nil {
		return nil, err
	}
	self.processedRecordsNr += 1
	if utils.IsSliceMember([]string{utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE}, *self.dfltCdrcCfg.CdrFormat) {
		if record, err = self.processFlatstoreRecord(record); err != nil {
			return nil, err
		} else if record == nil {
			return nil, nil // Due to partial, none returned
		}
	}
	// Record was overwriten with complete information out of cache
	return self.processRecord(record)
}

// Processes a single partial record for flatstore CDRs
func (self *CsvRecordsProcessor) processFlatstoreRecord(record []string) ([]string, error) {
	if strings.HasPrefix(self.fileName, *self.dfltCdrcCfg.FailedCallsPrefix) { // Use the first index since they should be the same in all configs
		record = append(record, "0") // Append duration 0 for failed calls flatstore CDR and do not process it further
		return record, nil
	}
	pr, err := NewUnpairedRecord(record, self.timezone)
	if err != nil {
		return nil, err
	}
	// Retrieve and complete the record from cache
	cachedFilename, cachedPartial := self.unpairedRecordsCache.GetPartialRecord(pr.OriginID, self.fileName)
	if cachedPartial == nil { // Not cached, do it here and stop processing
		self.unpairedRecordsCache.CachePartial(self.fileName, pr)
		return nil, nil
	}
	pairedRecord, err := pairToRecord(cachedPartial, pr)
	if err != nil {
		return nil, err
	}
	self.unpairedRecordsCache.UncachePartial(cachedFilename, pr)
	return pairedRecord, nil
}

// Takes the record from a slice and turns it into StoredCdrs, posting them to the cdrServer
func (self *CsvRecordsProcessor) processRecord(record []string) ([]*engine.CDR, error) {
	recordCdrs := make([]*engine.CDR, 0)    // More CDRs based on the number of filters and field templates
	for _, cdrcCfg := range self.cdrcCfgs { // cdrFields coming from more templates will produce individual storCdr records
		// Make sure filters are matching
		filterBreak := false
		for _, rsrFilter := range cdrcCfg.CdrFilter {
			if rsrFilter == nil { // Nil filter does not need to match anything
				continue
			}
			if cfgFieldIdx, _ := strconv.Atoi(rsrFilter.Id); len(record) <= cfgFieldIdx {
				return nil, fmt.Errorf("Ignoring record: %v - cannot compile filter %+v", record, rsrFilter)
			} else if !rsrFilter.FilterPasses(record[cfgFieldIdx]) {
				filterBreak = true
				break
			}
		}
		if filterBreak { // Stop importing cdrc fields profile due to non matching filter
			continue
		}
		storedCdr, err := self.recordToStoredCdr(record, cdrcCfg)
		if err != nil {
			return nil, fmt.Errorf("Failed converting to StoredCdr, error: %s", err.Error())
		} else if *self.dfltCdrcCfg.CdrFormat == utils.PartialCSV {
			if storedCdr, err = self.partialRecordsCache.MergePartialCDRRecord(NewPartialCDRRecord(storedCdr, self.partialCacheDumpFields)); err != nil {
				return nil, fmt.Errorf("Failed merging PartialCDR, error: %s", err.Error())
			} else if storedCdr == nil { // CDR was absorbed by cache since it was partial
				continue
			}
		}
		recordCdrs = append(recordCdrs, storedCdr)
		if !*cdrcCfg.ContinueOnSuccess {
			break
		}
	}
	return recordCdrs, nil
}

// Takes the record out of csv and turns it into storedCdr which can be processed by CDRS
func (self *CsvRecordsProcessor) recordToStoredCdr(record []string, cdrcCfg *config.Cdrc) (*engine.CDR, error) {
	storedCdr := &engine.CDR{OriginHost: "0.0.0.0", Source: *cdrcCfg.CdrSourceID, ExtraFields: make(map[string]string), Cost: dec.NewVal(-1, 0)}
	var err error
	var lazyHttpFields []*config.CdrField
	for _, cdrFldCfg := range cdrcCfg.ContentFields {
		filterBreak := false
		for _, rsrFilter := range cdrFldCfg.FieldFilter {
			if rsrFilter == nil { // Nil filter does not need to match anything
				continue
			}
			if cfgFieldIdx, _ := strconv.Atoi(rsrFilter.Id); len(record) <= cfgFieldIdx {
				return nil, fmt.Errorf("Ignoring record: %v - cannot compile field filter %+v", record, rsrFilter)
			} else if !rsrFilter.FilterPasses(record[cfgFieldIdx]) {
				filterBreak = true
				break
			}
		}
		if filterBreak { // Stop processing this field template since it's filters are not matching
			continue
		}
		if utils.IsSliceMember([]string{utils.KAM_FLATSTORE, utils.OSIPS_FLATSTORE}, *self.dfltCdrcCfg.CdrFormat) { // Hardcode some values in case of flatstore
			switch cdrFldCfg.FieldID {
			case utils.ACCID:
				cdrFldCfg.Value = utils.ParseRSRFieldsMustCompile("3;1;2", utils.INFIELD_SEP) // in case of flatstore, accounting id is made up out of callid, from_tag and to_tag
			case utils.USAGE:
				cdrFldCfg.Value = utils.ParseRSRFieldsMustCompile(strconv.Itoa(len(record)-1), utils.INFIELD_SEP) // in case of flatstore, last element will be the duration computed by us
			}

		}
		var fieldVal string
		switch cdrFldCfg.Type {
		case utils.META_COMPOSED, utils.MetaUnixTimestamp:
			for _, cfgFieldRSR := range cdrFldCfg.Value {
				if cfgFieldRSR.IsStatic() {
					fieldVal += cfgFieldRSR.ParseValue("")
				} else { // Dynamic value extracted using index
					if cfgFieldIdx, _ := strconv.Atoi(cfgFieldRSR.Id); len(record) <= cfgFieldIdx {
						return nil, fmt.Errorf("Ignoring record: %v - cannot extract field %s", record, cdrFldCfg.Tag)
					} else {
						strVal := cfgFieldRSR.ParseValue(record[cfgFieldIdx])
						if cdrFldCfg.Type == utils.MetaUnixTimestamp {
							t, _ := utils.ParseTimeDetectLayout(strVal, self.timezone)
							strVal = strconv.Itoa(int(t.Unix()))
						}
						fieldVal += strVal
					}
				}
			}
		case utils.META_HTTP_POST:
			lazyHttpFields = append(lazyHttpFields, cdrFldCfg) // Will process later so we can send an estimation of storedCdr to http server
		default:
			return nil, fmt.Errorf("Unsupported field type: %s", cdrFldCfg.Type)
		}
		if err := storedCdr.ParseFieldValue(cdrFldCfg.FieldID, fieldVal, self.timezone); err != nil {
			return nil, err
		}
	}
	storedCdr.UniqueID = utils.Sha1(storedCdr.OriginID, storedCdr.SetupTime.UTC().String())
	if storedCdr.ToR == utils.DATA && *cdrcCfg.DataUsageMultiplyFactor != 0 {
		storedCdr.Usage = time.Duration(float64(storedCdr.Usage.Nanoseconds()) * *cdrcCfg.DataUsageMultiplyFactor)
	}
	for _, httpFieldCfg := range lazyHttpFields { // Lazy process the http fields
		var outValByte []byte
		var fieldVal, httpAddr string
		for _, rsrFld := range httpFieldCfg.Value {
			httpAddr += rsrFld.ParseValue("")
		}
		var jsn []byte
		jsn, err = json.Marshal(storedCdr)
		if err != nil {
			return nil, err
		}
		if outValByte, err = utils.HttpJsonPost(httpAddr, self.httpSkipTlsCheck, jsn); err != nil && httpFieldCfg.Mandatory {
			return nil, err
		} else {
			fieldVal = string(outValByte)
			if len(fieldVal) == 0 && httpFieldCfg.Mandatory {
				return nil, fmt.Errorf("MandatoryIeMissing: Empty result for http_post field: %s", httpFieldCfg.Tag)
			}
			if err := storedCdr.ParseFieldValue(httpFieldCfg.FieldID, fieldVal, self.timezone); err != nil {
				return nil, err
			}
		}
	}
	return storedCdr, nil
}
