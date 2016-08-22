
package v1

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type AttrReloadConfig struct {
	ConfigDir string
}

// Retrieves the callCost out of CGR logDb
func (apier *ApierV1) ReloadCdrcConfig(attrs AttrReloadConfig, reply *string) error {
	if attrs.ConfigDir == "" {
		attrs.ConfigDir = utils.CONFIG_DIR
	}
	newCfg, err := config.NewCGRConfigFromFolder(attrs.ConfigDir)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	apier.Config.CdrcProfiles = newCfg.CdrcProfiles // ToDo: Check if there can be any concurency involved here so we need to lock maybe
	apier.Config.ConfigReloads[utils.CDRC] <- struct{}{}
	*reply = OK
	return nil
}
