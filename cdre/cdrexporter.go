package cdre

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

const (
	META_EXPORTID      = "*export_id"
	META_TIMENOW       = "*time_now"
	META_FIRSTCDRATIME = "*first_cdr_atime"
	META_LASTCDRATIME  = "*last_cdr_atime"
	META_NRCDRS        = "*cdrs_number"
	META_DURCDRS       = "*cdrs_duration"
	META_SMSUSAGE      = "*sms_usage"
	META_MMSUSAGE      = "*mms_usage"
	META_GENERICUSAGE  = "*generic_usage"
	META_DATAUSAGE     = "*data_usage"
	META_COSTCDRS      = "*cdrs_cost"
	META_FORMATCOST    = "*format_cost"
)

var err error

func NewCdrExporter(cdrs []*engine.CDR, cdrDb engine.CdrStorage, exportTpl *config.Cdre, cdrFormat string, fieldSeparator rune, exportId string,
	dataUsageMultiplyFactor, smsUsageMultiplyFactor, mmsUsageMultiplyFactor, genericUsageMultiplyFactor, costMultiplyFactor float64,
	cgrPrecision int, httpSkipTlsCheck bool) (*CdrExporter, error) {
	if len(cdrs) == 0 { // Nothing to export
		return nil, nil
	}
	cdre := &CdrExporter{
		cdrs:                       cdrs,
		cdrDb:                      cdrDb,
		exportTemplate:             exportTpl,
		cdrFormat:                  cdrFormat,
		fieldSeparator:             fieldSeparator,
		exportId:                   exportId,
		dataUsageMultiplyFactor:    dataUsageMultiplyFactor,
		smsUsageMultiplyFactor:     smsUsageMultiplyFactor,
		mmsUsageMultiplyFactor:     mmsUsageMultiplyFactor,
		genericUsageMultiplyFactor: genericUsageMultiplyFactor,
		costMultiplyFactor:         costMultiplyFactor,
		cgrPrecision:               cgrPrecision,
		httpSkipTlsCheck:           httpSkipTlsCheck,
		negativeExports:            make(map[string]string),
	}
	if err := cdre.processCdrs(); err != nil {
		return nil, err
	}
	return cdre, nil
}

type CdrExporter struct {
	cdrs           []*engine.CDR
	cdrDb          engine.CdrStorage // Used to extract cost_details if these are requested
	exportTemplate *config.Cdre
	cdrFormat      string // csv, fwv
	fieldSeparator rune
	exportId       string // Unique identifier or this export
	dataUsageMultiplyFactor,
	smsUsageMultiplyFactor, // Multiply the SMS usage (eg: some billing systems billing them as minutes)
	mmsUsageMultiplyFactor,
	genericUsageMultiplyFactor,
	costMultiplyFactor float64
	cgrPrecision                int
	httpSkipTlsCheck            bool
	header, trailer             []string   // Header and Trailer fields
	content                     [][]string // Rows of cdr fields
	firstCdrATime, lastCdrATime time.Time
	numberOfRecords             int
	totalDuration, totalDataUsage, totalSmsUsage,
	totalMmsUsage, totalGenericUsage time.Duration

	totalCost                       *dec.Dec
	firstExpOrderId, lastExpOrderId int64
	positiveExports                 []string          // UniqueIDs of successfully exported CDRs
	negativeExports                 map[string]string // UniqueIDs of failed exports
}

func (cdre *CdrExporter) GetTotalCost() *dec.Dec {
	if cdre.totalCost == nil {
		cdre.totalCost = dec.New()
	}
	return cdre.totalCost
}

// Handle various meta functions used in header/trailer
func (cdre *CdrExporter) metaHandler(tag, arg string) (string, error) {
	switch tag {
	case META_EXPORTID:
		return cdre.exportId, nil
	case META_TIMENOW:
		return time.Now().Format(arg), nil
	case META_FIRSTCDRATIME:
		return cdre.firstCdrATime.Format(arg), nil
	case META_LASTCDRATIME:
		return cdre.lastCdrATime.Format(arg), nil
	case META_NRCDRS:
		return strconv.Itoa(cdre.numberOfRecords), nil
	case META_DURCDRS:
		emulatedCdr := &engine.CDR{ToR: utils.VOICE, Usage: cdre.totalDuration}
		return emulatedCdr.FormatUsage(arg), nil
	case META_SMSUSAGE:
		emulatedCdr := &engine.CDR{ToR: utils.SMS, Usage: cdre.totalSmsUsage}
		return emulatedCdr.FormatUsage(arg), nil
	case META_MMSUSAGE:
		emulatedCdr := &engine.CDR{ToR: utils.MMS, Usage: cdre.totalMmsUsage}
		return emulatedCdr.FormatUsage(arg), nil
	case META_GENERICUSAGE:
		emulatedCdr := &engine.CDR{ToR: utils.GENERIC, Usage: cdre.totalGenericUsage}
		return emulatedCdr.FormatUsage(arg), nil
	case META_DATAUSAGE:
		emulatedCdr := &engine.CDR{ToR: utils.DATA, Usage: cdre.totalDataUsage}
		return emulatedCdr.FormatUsage(arg), nil
	case META_COSTCDRS:
		return cdre.GetTotalCost().Round(int32(cdre.cgrPrecision)).String(), nil
	default:
		return "", fmt.Errorf("Unsupported METATAG: %s", tag)
	}
}

// Compose and cache the header
func (cdre *CdrExporter) composeHeader() error {
	for _, cfgFld := range cdre.exportTemplate.HeaderFields {
		var outVal string
		switch cfgFld.Type {
		case utils.META_FILLER:
			outVal = cfgFld.Value.Id()
			cfgFld.Padding = "right"
		case utils.META_CONSTANT:
			outVal = cfgFld.Value.Id()
		case utils.META_HANDLER:
			outVal, err = cdre.metaHandler(cfgFld.Value.Id(), cfgFld.Layout)
		default:
			return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		}
		if err != nil {
			utils.Logger.Error("<CdreFw> Cannot export CDR header", zap.String("tag", cfgFld.Tag), zap.Error(err))
			return err
		}
		fmtOut := outVal
		if fmtOut, err = utils.FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			utils.Logger.Error("<CdreFw> Cannot export CDR header", zap.String("tag", cfgFld.Tag), zap.Error(err))
			return err
		}
		cdre.header = append(cdre.header, fmtOut)
	}
	return nil
}

// Compose and cache the trailer
func (cdre *CdrExporter) composeTrailer() error {
	for _, cfgFld := range cdre.exportTemplate.TrailerFields {
		var outVal string
		switch cfgFld.Type {
		case utils.META_FILLER:
			outVal = cfgFld.Value.Id()
			cfgFld.Padding = "right"
		case utils.META_CONSTANT:
			outVal = cfgFld.Value.Id()
		case utils.META_HANDLER:
			outVal, err = cdre.metaHandler(cfgFld.Value.Id(), cfgFld.Layout)
		default:
			return fmt.Errorf("Unsupported field type: %s", cfgFld.Type)
		}
		if err != nil {
			utils.Logger.Error("<CdreFw> Cannot export CDR trailer", zap.String("tag", cfgFld.Tag), zap.Error(err))
			return err
		}
		fmtOut := outVal
		if fmtOut, err = utils.FmtFieldWidth(outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
			utils.Logger.Error("<CdreFw> Cannot export CDR trailer", zap.String("tag", cfgFld.Tag), zap.Error(err))
			return err
		}
		cdre.trailer = append(cdre.trailer, fmtOut)
	}
	return nil
}

// Write individual cdr into content buffer, build stats
func (cdre *CdrExporter) processCdr(cdr *engine.CDR) error {
	if cdr == nil || len(cdr.UniqueID) == 0 { // We do not export empty CDRs
		return nil
	} else if cdr.ExtraFields == nil { // Avoid assignment in nil map if not initialized
		cdr.ExtraFields = make(map[string]string)
	}
	// Cost multiply
	if cdre.dataUsageMultiplyFactor != 0.0 && cdr.ToR == utils.DATA {
		cdr.UsageMultiply(cdre.dataUsageMultiplyFactor, cdre.cgrPrecision)
	} else if cdre.smsUsageMultiplyFactor != 0 && cdr.ToR == utils.SMS {
		cdr.UsageMultiply(cdre.smsUsageMultiplyFactor, cdre.cgrPrecision)
	} else if cdre.mmsUsageMultiplyFactor != 0 && cdr.ToR == utils.MMS {
		cdr.UsageMultiply(cdre.mmsUsageMultiplyFactor, cdre.cgrPrecision)
	} else if cdre.genericUsageMultiplyFactor != 0 && cdr.ToR == utils.GENERIC {
		cdr.UsageMultiply(cdre.genericUsageMultiplyFactor, cdre.cgrPrecision)
	}
	if cdre.costMultiplyFactor != 0.0 {
		cdr.CostMultiply(cdre.costMultiplyFactor, cdre.cgrPrecision)
	}
	cdrRow, err := cdr.AsExportRecord(cdre.exportTemplate.ContentFields, cdre.httpSkipTlsCheck, cdre.cdrs)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("<CdreFw> Cannot export CDR with UniqueID: %s and runid: %s, error: %s", cdr.UniqueID, cdr.RunID, err.Error()))
		return err
	}
	if len(cdrRow) == 0 { // No CDR data, most likely no configuration fields defined
		return nil
	} else {
		cdre.content = append(cdre.content, cdrRow)
	}
	// Done with writing content, compute stats here
	if cdre.firstCdrATime.IsZero() || cdr.AnswerTime.Before(cdre.firstCdrATime) {
		cdre.firstCdrATime = cdr.AnswerTime
	}
	if cdr.AnswerTime.After(cdre.lastCdrATime) {
		cdre.lastCdrATime = cdr.AnswerTime
	}
	cdre.numberOfRecords += 1
	if cdr.ToR == utils.VOICE { // Only count duration for non data cdrs
		cdre.totalDuration += cdr.Usage
	}
	if cdr.ToR == utils.SMS { // Count usage for SMS
		cdre.totalSmsUsage += cdr.Usage
	}
	if cdr.ToR == utils.MMS { // Count usage for MMS
		cdre.totalMmsUsage += cdr.Usage
	}
	if cdr.ToR == utils.GENERIC { // Count usage for GENERIC
		cdre.totalGenericUsage += cdr.Usage
	}
	if cdr.ToR == utils.DATA { // Count usage for DATA
		cdre.totalDataUsage += cdr.Usage
	}
	if cdr.GetCost().Cmp(dec.MinusOne) != 0 {
		cdre.GetTotalCost().AddS(cdr.Cost).Round(int32(cdre.cgrPrecision))
	}
	if cdre.firstExpOrderId > cdr.OrderID || cdre.firstExpOrderId == 0 {
		cdre.firstExpOrderId = cdr.OrderID
	}
	if cdre.lastExpOrderId < cdr.OrderID {
		cdre.lastExpOrderId = cdr.OrderID
	}
	return nil
}

// Builds header, content and trailers
func (cdre *CdrExporter) processCdrs() error {
	for _, cdr := range cdre.cdrs {
		if err := cdre.processCdr(cdr); err != nil {
			cdre.negativeExports[cdr.UniqueID] = err.Error()
		} else {
			cdre.positiveExports = append(cdre.positiveExports, cdr.UniqueID)
		}
	}
	// Process header and trailer after processing cdrs since the metatag functions can access stats out of built cdrs
	if cdre.exportTemplate.HeaderFields != nil {
		if err := cdre.composeHeader(); err != nil {
			return err
		}
	}
	if cdre.exportTemplate.TrailerFields != nil {
		if err := cdre.composeTrailer(); err != nil {
			return err
		}
	}
	return nil
}

// Simple write method
func (cdre *CdrExporter) writeOut(ioWriter io.Writer) error {
	if len(cdre.header) != 0 {
		for _, fld := range append(cdre.header, "\n") {
			if _, err := io.WriteString(ioWriter, fld); err != nil {
				return err
			}
		}
	}
	for _, cdrContent := range cdre.content {
		for _, cdrFld := range append(cdrContent, "\n") {
			if _, err := io.WriteString(ioWriter, cdrFld); err != nil {
				return err
			}
		}
	}
	if len(cdre.trailer) != 0 {
		for _, fld := range append(cdre.trailer, "\n") {
			if _, err := io.WriteString(ioWriter, fld); err != nil {
				return err
			}
		}
	}
	return nil
}

// csvWriter specific method
func (cdre *CdrExporter) writeCsv(csvWriter *csv.Writer) error {
	csvWriter.Comma = cdre.fieldSeparator
	if len(cdre.header) != 0 {
		if err := csvWriter.Write(cdre.header); err != nil {
			return err
		}
	}
	for _, cdrContent := range cdre.content {
		if err := csvWriter.Write(cdrContent); err != nil {
			return err
		}
	}
	if len(cdre.trailer) != 0 {
		if err := csvWriter.Write(cdre.trailer); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return nil
}

// General method to write the content out to a file
func (cdre *CdrExporter) WriteToFile(filePath string) error {
	fileOut, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fileOut.Close()
	switch cdre.cdrFormat {
	case utils.DRYRUN:
		return nil
	case utils.CDRE_FIXED_WIDTH:
		if err := cdre.writeOut(fileOut); err != nil {
			return utils.NewErrServerError(err)
		}
	case utils.CSV:
		csvWriter := csv.NewWriter(fileOut)
		if err := cdre.writeCsv(csvWriter); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	return nil
}

// Return the first exported Cdr OrderId
func (cdre *CdrExporter) FirstOrderId() int64 {
	return cdre.firstExpOrderId
}

// Return the last exported Cdr OrderId
func (cdre *CdrExporter) LastOrderId() int64 {
	return cdre.lastExpOrderId
}

// Return total cost in the exported cdrs
/*func (cdre *CdrExporter) TotalCost() float64 {
	return cdre.totalCost
}*/

func (cdre *CdrExporter) TotalExportedCdrs() int {
	return cdre.numberOfRecords
}

// Return successfully exported UniqueIDs
func (cdre *CdrExporter) PositiveExports() []string {
	return cdre.positiveExports
}

// Return failed exported UniqueIDs together with the reason
func (cdre *CdrExporter) NegativeExports() map[string]string {
	return cdre.negativeExports
}
