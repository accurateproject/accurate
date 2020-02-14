package v1

import (
	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
)

type AttrReloadConfig struct {
	ConfigDir string
}

func (api *ApiV1) ReloadCdrcConfig(attrs AttrReloadConfig, reply *string) error {
	if attrs.ConfigDir == "" {
		attrs.ConfigDir = utils.CONFIG_DIR
	}
	// FIXME: this should be Reload config in general!
	if err := config.LoadPath(attrs.ConfigDir); err != nil {
		return utils.NewErrServerError(err)
	}
	//api.Config.CdrcProfiles = newCfg.CdrcProfiles // ToDo: Check if there can be any concurency involved here so we need to lock maybe
	//api.Config.ConfigReloads[utils.CDRC] <- struct{}{}
	*reply = OK
	return nil
}
