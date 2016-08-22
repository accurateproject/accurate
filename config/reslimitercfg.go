
package config

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

type ResourceLimiterConfig struct {
	Enabled           bool
	CDRStatConns      []*HaPoolConfig // Connections towards CDRStatS
	CacheDumpInterval time.Duration   // Dump regularly from cache into dataDB
	UsageTTL          time.Duration   // Auto-Expire usage units older than this duration
}

func (rlcfg *ResourceLimiterConfig) loadFromJsonCfg(jsnCfg *ResourceLimiterServJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		rlcfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cdrstats_conns != nil {
		rlcfg.CDRStatConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrstats_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrstats_conns {
			rlcfg.CDRStatConns[idx] = NewDfltHaPoolConfig()
			rlcfg.CDRStatConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Cache_dump_interval != nil {
		if rlcfg.CacheDumpInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Cache_dump_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Usage_ttl != nil {
		if rlcfg.UsageTTL, err = utils.ParseDurationWithSecs(*jsnCfg.Usage_ttl); err != nil {
			return err
		}
	}
	return nil
}
