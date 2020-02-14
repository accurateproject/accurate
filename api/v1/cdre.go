package v1

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/accurateproject/accurate/cdre"
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
)

// Export Cdrs to file
func (api *ApiV1) ExportCdrsToFile(attr utils.AttrExportCdrsToFile, reply *utils.ExportedFileCdrs) error {
	var err error
	//cdreReloadStruct := <-api.cfg.ConfigReloads[utils.CDRE]                  // Read the content of the channel, locking it
	//defer func() { api.cfg.ConfigReloads[utils.CDRE] <- cdreReloadStruct }() // Unlock reloads at exit
	exportTemplate := (*api.cfg.Cdre)[utils.META_DEFAULT]
	if attr.ExportTemplate != nil && len(*attr.ExportTemplate) != 0 { // Export template prefered, use it
		var hasIt bool
		if exportTemplate, hasIt = (*api.cfg.Cdre)[*attr.ExportTemplate]; !hasIt {
			return fmt.Errorf("%s:ExportTemplate", utils.ErrNotFound)
		}
	}
	cdrFormat := *exportTemplate.CdrFormat
	if attr.CdrFormat != nil && len(*attr.CdrFormat) != 0 {
		cdrFormat = strings.ToLower(*attr.CdrFormat)
	}
	if !utils.IsSliceMember(utils.CdreCdrFormats, cdrFormat) {
		return utils.NewErrMandatoryIeMissing("CdrFormat")
	}
	fieldSep, _ := utf8.DecodeRuneInString(*exportTemplate.FieldSeparator)
	if attr.FieldSeparator != nil && len(*attr.FieldSeparator) != 0 {
		fieldSep, _ = utf8.DecodeRuneInString(*attr.FieldSeparator)
		if fieldSep == utf8.RuneError {
			return fmt.Errorf("%s:FieldSeparator:%s", utils.ErrServerError, "Invalid")
		}
	}
	eDir := *exportTemplate.ExportDirectory
	if attr.ExportDirectory != nil && len(*attr.ExportDirectory) != 0 {
		eDir = *attr.ExportDirectory
	}
	ExportID := strconv.FormatInt(time.Now().Unix(), 10)
	if attr.ExportID != nil && len(*attr.ExportID) != 0 {
		ExportID = *attr.ExportID
	}
	fileName := fmt.Sprintf("cdre_%s.%s", ExportID, cdrFormat)
	if attr.ExportFileName != nil && len(*attr.ExportFileName) != 0 {
		fileName = *attr.ExportFileName
	}
	filePath := path.Join(eDir, fileName)
	if cdrFormat == utils.DRYRUN {
		filePath = utils.DRYRUN
	}
	dataUsageMultiplyFactor := *exportTemplate.DataUsageMultiplyFactor
	if attr.DataUsageMultiplyFactor != nil && *attr.DataUsageMultiplyFactor != 0.0 {
		dataUsageMultiplyFactor = *attr.DataUsageMultiplyFactor
	}
	SMSUsageMultiplyFactor := *exportTemplate.SmsUsageMultiplyFactor
	if attr.SMSUsageMultiplyFactor != nil && *attr.SMSUsageMultiplyFactor != 0.0 {
		SMSUsageMultiplyFactor = *attr.SMSUsageMultiplyFactor
	}
	MMSUsageMultiplyFactor := *exportTemplate.MmsUsageMultiplyFactor
	if attr.MMSUsageMultiplyFactor != nil && *attr.MMSUsageMultiplyFactor != 0.0 {
		MMSUsageMultiplyFactor = *attr.MMSUsageMultiplyFactor
	}
	genericUsageMultiplyFactor := *exportTemplate.GenericUsageMultiplyFactor
	if attr.GenericUsageMultiplyFactor != nil && *attr.GenericUsageMultiplyFactor != 0.0 {
		genericUsageMultiplyFactor = *attr.GenericUsageMultiplyFactor
	}
	costMultiplyFactor := *exportTemplate.CostMultiplyFactor
	if attr.CostMultiplyFactor != nil && *attr.CostMultiplyFactor != 0.0 {
		costMultiplyFactor = *attr.CostMultiplyFactor
	}
	cdrsFltr, err := attr.RPCCDRsFilter.AsCDRsFilter(*api.cfg.General.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	cdrs, _, err := api.cdrDB.GetCDRs(cdrsFltr, false)
	if err != nil {
		return err
	} else if len(cdrs) == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	cdrexp, err := cdre.NewCdrExporter(cdrs, api.cdrDB, exportTemplate, cdrFormat, fieldSep, ExportID, dataUsageMultiplyFactor, SMSUsageMultiplyFactor, MMSUsageMultiplyFactor, genericUsageMultiplyFactor, costMultiplyFactor, *api.cfg.General.RoundingDecimals, *api.cfg.General.HttpSkipTlsVerify)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if cdrexp.TotalExportedCdrs() == 0 {
		*reply = utils.ExportedFileCdrs{ExportedFilePath: ""}
		return nil
	}
	if err := cdrexp.WriteToFile(filePath); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.ExportedFileCdrs{ExportedFilePath: filePath, TotalRecords: len(cdrs), TotalCost: cdrexp.GetTotalCost(), FirstOrderId: cdrexp.FirstOrderId(), LastOrderId: cdrexp.LastOrderId()}
	if !attr.Verbose {
		reply.ExportedUniqueIDs = cdrexp.PositiveExports()
		reply.UnexportedUniqueIDs = cdrexp.NegativeExports()
	}
	return nil
}

/*
func (api *ApiV1) ExportCdrsToZipString(attr utils.AttrExpFileCdrs, reply *string) error {
	tmpDir := "/tmp"
	attr.ExportDir = &tmpDir // Enforce exporting to tmp always so we avoid cleanup issues
	efc := utils.ExportedFileCdrs{}
	if err := api.ExportCdrsToFile(attr, &efc); err != nil {
		return err
	} else if efc.TotalRecords == 0 || len(efc.ExportedFilePath) == 0 {
		return errors.New("No CDR records to export")
	}
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)
	// Create a new zip archive.
	w := zip.NewWriter(buf)
	// read generated file
	content, err := ioutil.ReadFile(efc.ExportedFilePath)
	if err != nil {
		return err
	}
	exportFileName := path.Base(efc.ExportedFilePath)
	f, err := w.Create(exportFileName)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	if err != nil {
		return err
	}
	// Write metadata into a separate file with extension .cgr
	medaData, err := json.MarshalIndent(efc, "", "  ")
	if err != nil {
		errors.New("Failed creating metadata content")
	}
	medatadaFileName := exportFileName[:len(path.Ext(exportFileName))] + ".cgr"
	mf, err := w.Create(medatadaFileName)
	if err != nil {
		return err
	}
	_, err = mf.Write(medaData)
	if err != nil {
		return err
	}
	// Make sure to check the error on Close.
	if err := w.Close(); err != nil {
		return err
	}
	if err := os.Remove(efc.ExportedFilePath); err != nil {
		fmt.Errorf("Failed removing exported file at path: %s", efc.ExportedFilePath)
	}
	*reply = base64.StdEncoding.EncodeToString(buf.Bytes())
	return nil
}
*/
// Reloads CDRE configuration out of folder specified
func (api *ApiV1) ReloadCdreConfig(attrs AttrReloadConfig, reply *string) error {
	if attrs.ConfigDir == "" {
		attrs.ConfigDir = utils.CONFIG_DIR
	}
	err := config.LoadPath(attrs.ConfigDir)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	//cdreReloadStruct := <-api.cfg.ConfigReloads[utils.CDRE] // Get the CDRE reload channel                     // Read the content of the channel, locking it
	//api.cfg.CdreProfiles = newCfg.CdreProfiles
	//api.cfg.ConfigReloads[utils.CDRE] <- cdreReloadStruct // Unlock reloads
	utils.Logger.Info("<CDRE> Configuration reloaded")
	*reply = OK
	return nil
}
