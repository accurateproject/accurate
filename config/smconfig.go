package config

import (
	"time"

	"github.com/accurateproject/accurate/utils"
)

// Returns the first cached default value for a SM-FreeSWITCH connection
func NewDfltHaPoolConfig() *HaPoolConfig {
	if dfltHaPoolConfig == nil {
		return new(HaPoolConfig) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltHaPoolConfig // Copy the value instead of it's pointer
	return &dfltVal
}

// One connection to Rater
type HaPoolConfig struct {
	Address   string
	Transport string
}

func (self *HaPoolConfig) loadFromJsonCfg(jsnCfg *HaPoolJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		self.Address = *jsnCfg.Address
	}
	if jsnCfg.Transport != nil {
		self.Transport = *jsnCfg.Transport
	}
	return nil
}

// Returns the first cached default value for a SM-FreeSWITCH connection
func NewDfltFsConnConfig() *FsConnConfig {
	if dfltFsConnConfig == nil {
		return new(FsConnConfig) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltFsConnConfig // Copy the value instead of it's pointer
	return &dfltVal
}

// One connection to FreeSWITCH server
type FsConnConfig struct {
	Address    string
	Password   string
	Reconnects int
}

func (self *FsConnConfig) loadFromJsonCfg(jsnCfg *FsConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		self.Address = *jsnCfg.Address
	}
	if jsnCfg.Password != nil {
		self.Password = *jsnCfg.Password
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

type SmGenericConfig struct {
	Enabled            bool
	ListenBijson       string
	RALsConns          []*HaPoolConfig
	CDRsConns          []*HaPoolConfig
	DebitInterval      time.Duration
	MinCallDuration    time.Duration
	MaxCallDuration    time.Duration
	SessionTTL         time.Duration
	SessionTTLLastUsed *time.Duration
	SessionTTLUsage    *time.Duration
}

func (self *SmGenericConfig) loadFromJsonCfg(jsnCfg *SmGenericJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_bijson != nil {
		self.ListenBijson = *jsnCfg.Listen_bijson
	}
	if jsnCfg.Rals_conns != nil {
		self.RALsConns = make([]*HaPoolConfig, len(*jsnCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rals_conns {
			self.RALsConns[idx] = NewDfltHaPoolConfig()
			self.RALsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Cdrs_conns != nil {
		self.CDRsConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrs_conns {
			self.CDRsConns[idx] = NewDfltHaPoolConfig()
			self.CDRsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Session_ttl != nil {
		if self.SessionTTL, err = utils.ParseDurationWithSecs(*jsnCfg.Session_ttl); err != nil {
			return err
		}
	}
	if jsnCfg.Session_ttl_last_used != nil {
		if sessionTTLLastUsed, err := utils.ParseDurationWithSecs(*jsnCfg.Session_ttl_last_used); err != nil {
			return err
		} else {
			self.SessionTTLLastUsed = &sessionTTLLastUsed
		}
	}
	if jsnCfg.Session_ttl_usage != nil {
		if sessionTTLUsage, err := utils.ParseDurationWithSecs(*jsnCfg.Session_ttl_usage); err != nil {
			return err
		} else {
			self.SessionTTLUsage = &sessionTTLUsage
		}
	}
	return nil
}

type SmFsConfig struct {
	Enabled             bool
	RALsConns           []*HaPoolConfig
	CDRsConns           []*HaPoolConfig
	RLsConns            []*HaPoolConfig
	CreateCdr           bool
	ExtraFields         []*utils.RSRField
	DebitInterval       time.Duration
	MinCallDuration     time.Duration
	MaxCallDuration     time.Duration
	MinDurLowBalance    time.Duration
	LowBalanceAnnFile   string
	EmptyBalanceContext string
	EmptyBalanceAnnFile string
	SubscribePark       bool
	ChannelSyncInterval time.Duration
	MaxWaitConnection   time.Duration
	EventSocketConns    []*FsConnConfig
}

func (self *SmFsConfig) loadFromJsonCfg(jsnCfg *SmFsJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Rals_conns != nil {
		self.RALsConns = make([]*HaPoolConfig, len(*jsnCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rals_conns {
			self.RALsConns[idx] = NewDfltHaPoolConfig()
			self.RALsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Cdrs_conns != nil {
		self.CDRsConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrs_conns {
			self.CDRsConns[idx] = NewDfltHaPoolConfig()
			self.CDRsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Rls_conns != nil {
		self.RLsConns = make([]*HaPoolConfig, len(*jsnCfg.Rls_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rls_conns {
			self.RLsConns[idx] = NewDfltHaPoolConfig()
			self.RLsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Create_cdr != nil {
		self.CreateCdr = *jsnCfg.Create_cdr
	}
	if jsnCfg.Extra_fields != nil {
		if self.ExtraFields, err = utils.ParseRSRFieldsFromSlice(*jsnCfg.Extra_fields); err != nil {
			return err
		}
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Min_dur_low_balance != nil {
		if self.MinDurLowBalance, err = utils.ParseDurationWithSecs(*jsnCfg.Min_dur_low_balance); err != nil {
			return err
		}
	}
	if jsnCfg.Low_balance_ann_file != nil {
		self.LowBalanceAnnFile = *jsnCfg.Low_balance_ann_file
	}
	if jsnCfg.Empty_balance_context != nil {
		self.EmptyBalanceContext = *jsnCfg.Empty_balance_context
	}
	if jsnCfg.Empty_balance_ann_file != nil {
		self.EmptyBalanceAnnFile = *jsnCfg.Empty_balance_ann_file
	}
	if jsnCfg.Subscribe_park != nil {
		self.SubscribePark = *jsnCfg.Subscribe_park
	}
	if jsnCfg.Channel_sync_interval != nil {
		if self.ChannelSyncInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Channel_sync_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Max_wait_connection != nil {
		if self.MaxWaitConnection, err = utils.ParseDurationWithSecs(*jsnCfg.Max_wait_connection); err != nil {
			return err
		}
	}
	if jsnCfg.Event_socket_conns != nil {
		self.EventSocketConns = make([]*FsConnConfig, len(*jsnCfg.Event_socket_conns))
		for idx, jsnConnCfg := range *jsnCfg.Event_socket_conns {
			self.EventSocketConns[idx] = NewDfltFsConnConfig()
			self.EventSocketConns[idx].loadFromJsonCfg(jsnConnCfg)
		}
	}
	return nil
}

// Returns the first cached default value for a SM-FreeSWITCH connection
func NewDfltKamConnConfig() *KamConnConfig {
	if dfltKamConnConfig == nil {
		return new(KamConnConfig) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltKamConnConfig
	return &dfltVal
}

// Represents one connection instance towards Kamailio
type KamConnConfig struct {
	Address    string
	Reconnects int
}

func (self *KamConnConfig) loadFromJsonCfg(jsnCfg *KamConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		self.Address = *jsnCfg.Address
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

// SM-Kamailio config section
type SmKamConfig struct {
	Enabled         bool
	RALsConns       []*HaPoolConfig
	CDRsConns       []*HaPoolConfig
	CreateCdr       bool
	DebitInterval   time.Duration
	MinCallDuration time.Duration
	MaxCallDuration time.Duration
	EvapiConns      []*KamConnConfig
}

func (self *SmKamConfig) loadFromJsonCfg(jsnCfg *SmKamJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Rals_conns != nil {
		self.RALsConns = make([]*HaPoolConfig, len(*jsnCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rals_conns {
			self.RALsConns[idx] = NewDfltHaPoolConfig()
			self.RALsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Cdrs_conns != nil {
		self.CDRsConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrs_conns {
			self.CDRsConns[idx] = NewDfltHaPoolConfig()
			self.CDRsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Create_cdr != nil {
		self.CreateCdr = *jsnCfg.Create_cdr
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Evapi_conns != nil {
		self.EvapiConns = make([]*KamConnConfig, len(*jsnCfg.Evapi_conns))
		for idx, jsnConnCfg := range *jsnCfg.Evapi_conns {
			self.EvapiConns[idx] = NewDfltKamConnConfig()
			self.EvapiConns[idx].loadFromJsonCfg(jsnConnCfg)
		}
	}
	return nil
}

// Represents one connection instance towards OpenSIPS, not in use for now but planned for future
type OsipsConnConfig struct {
	MiAddr     string
	Reconnects int
}

func (self *OsipsConnConfig) loadFromJsonCfg(jsnCfg *OsipsConnJsonCfg) error {
	if jsnCfg.Mi_addr != nil {
		self.MiAddr = *jsnCfg.Mi_addr
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

// SM-OpenSIPS config section
type SmOsipsConfig struct {
	Enabled                 bool
	ListenUdp               string
	RALsConns               []*HaPoolConfig
	CDRsConns               []*HaPoolConfig
	CreateCdr               bool
	DebitInterval           time.Duration
	MinCallDuration         time.Duration
	MaxCallDuration         time.Duration
	EventsSubscribeInterval time.Duration
	MiAddr                  string
}

func (self *SmOsipsConfig) loadFromJsonCfg(jsnCfg *SmOsipsJsonCfg) error {
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_udp != nil {
		self.ListenUdp = *jsnCfg.Listen_udp
	}
	if jsnCfg.Rals_conns != nil {
		self.RALsConns = make([]*HaPoolConfig, len(*jsnCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rals_conns {
			self.RALsConns[idx] = NewDfltHaPoolConfig()
			self.RALsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Cdrs_conns != nil {
		self.CDRsConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrs_conns {
			self.CDRsConns[idx] = NewDfltHaPoolConfig()
			self.CDRsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Create_cdr != nil {
		self.CreateCdr = *jsnCfg.Create_cdr
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithSecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Events_subscribe_interval != nil {
		if self.EventsSubscribeInterval, err = utils.ParseDurationWithSecs(*jsnCfg.Events_subscribe_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Mi_addr != nil {
		self.MiAddr = *jsnCfg.Mi_addr
	}

	return nil
}
