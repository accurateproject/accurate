package config

import (
	"fmt"
	"net/url"

	"github.com/accurateproject/accurate/utils"
)

type CdrReplicationCfg struct {
	Transport   string
	Address     string
	Synchronous bool
	Attempts    int             // Number of attempts if not success
	CdrFilter   utils.RSRFields // Only replicate if the filters here are matching
}

func (rplCfg CdrReplicationCfg) FallbackFileName() string {
	return fmt.Sprintf("cdr_%s_%s_%s.form", rplCfg.Transport, url.QueryEscape(rplCfg.Address), utils.GenUUID())
}
