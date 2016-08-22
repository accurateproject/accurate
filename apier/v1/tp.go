
package v1

// Tariff plan related APIs

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetTPIds struct {
}

// Queries tarrif plan identities gathered from all tables.
func (self *ApierV1) GetTPIds(attrs AttrGetTPIds, reply *[]string) error {
	if ids, err := self.StorDb.GetTpIds(); err != nil {
		return utils.NewErrServerError(err)
	} else if ids == nil {
		return utils.ErrNotFound
	} else {
		*reply = ids
	}
	return nil
}

type AttrImportTPZipFile struct {
	TPid string
	File []byte
}

func (self *ApierV1) ImportTPZipFile(attrs AttrImportTPZipFile, reply *string) error {
	tmpDir, err := ioutil.TempDir("/tmp", "cgr_")
	if err != nil {
		*reply = "ERROR: creating temp directory!"
		return err
	}
	zipFile := filepath.Join(tmpDir, "/file.zip")
	if err = ioutil.WriteFile(zipFile, attrs.File, os.ModePerm); err != nil {
		*reply = "ERROR: writing zip file!"
		return err
	}
	if err = utils.Unzip(zipFile, tmpDir); err != nil {
		*reply = "ERROR: unziping file!"
		return err
	}
	csvfilesFound := false
	if err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		csvFiles, err := filepath.Glob(filepath.Join(path, "*csv"))
		if csvFiles != nil {
			if attrs.TPid == "" {
				*reply = "ERROR: missing TPid!"
				return err
			}
			csvImporter := engine.TPCSVImporter{
				TPid:     attrs.TPid,
				StorDb:   self.StorDb,
				DirPath:  path,
				Sep:      ',',
				Verbose:  false,
				ImportId: "",
			}
			if errImport := csvImporter.Run(); errImport != nil {
				return errImport
			}
			csvfilesFound = true
		}
		return err
	}); err != nil || !csvfilesFound {
		*reply = "ERROR: finding csv files!"
		return err
	}
	os.RemoveAll(tmpDir)
	*reply = "OK"
	return nil
}
