package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/accurateproject/accurate/utils"
)

type Config struct {
	General       *General          `json:"general"`
	Cache         *Cache            `json:"cache"`
	Listen        *Listen           `json:"listen"`
	TariffPlanDb  *TariffPlanDb     `json:"tariffplan_db"`
	DataDb        *DataDb           `json:"data_db"`
	CdrDb         *CdrDb            `json:"cdr_db"`
	Balancer      *Balancer         `json:"balancer"`
	Rals          *Rals             `json:"rals"`
	Scheduler     *Scheduler        `json:"scheduler"`
	Cdrs          *Cdrs             `json:"cdrs"`
	CdrStats      *CdrStats         `json:"cdrstats"`
	Cdrc          *[]*Cdrc          `json:"cdrc"`
	Cdre          *map[string]*Cdre `json:"cdre"`
	SmGeneric     *SmGeneric        `json:"sm_generic"`
	SmFreeswitch  *SmFreeswitch     `json:"sm_freeswitch"`
	SmKamailio    *SmKamailio       `json:"sm_kamailio"`
	SmOpensips    *SmOpensips       `json:"sm_opensips"`
	SmAsterisk    *SmAsterisk       `json:"sm_asterisk"`
	DiameterAgent *DiameterAgent    `json:"diameter_agent"`
	Historys      *Historys         `json:"historys"`
	Pubsubs       *Pubsubs          `json:"pubsubs"`
	Aliases       *Aliases          `json:"aliases"`
	Users         *Users            `json:"users"`
	Rls           *Rls              `json:"rls"`
	Mailer        *Mailer           `json:"mailer"`
	SureTax       *SureTax          `json:"suretax"`
}

type General struct {
	InstanceID         *string // Identifier for this engine instance
	HttpSkipTlsVerify  *bool   `json:"http_skip_tls_verify"`      // if enabled Http Client will accept any TLS certificate
	RoundingDecimals   *int    `json:"rounding_decimals"`         // system level precision for floats
	TpexportDir        *string `json:"tpexport_dir"`              // path towards export folder for offline Tariff Plans
	HttpPosterAttempts *int    `json:"httpposter_attempts"`       // number of http attempts before considering request failed (eg: *call_url)
	HttpFailedDir      *string `json:"http_failed_dir"`           // directory path where we store failed http requests
	DefaultRequestType *string `json:"default_request_type"`      // default request type to consider when missing from requests: <""|*prepaid|*postpaid|*pseudoprepaid|*rated>
	DefaultCategory    *string `json:"default_category"`          // default category to consider when missing from requests
	DefaultTenant      *string `json:"default_tenant"`            // default tenant to consider when missing from requests
	DefaultTimezone    *string `json:"default_timezone"`          // default timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	ConnectAttempts    *int    `json:"connect_attempts"`          // initial server connect attempts
	Reconnects         *int    `json:"reconnects"`                // number of retries in case of connection lost
	ConnectTimeout     *dur    `json:"connect_timeout,string"`    // consider connection unsuccessful on timeout, 0 to disable the feature
	ReplyTimeout       *dur    `json:"reply_timeout,string"`      // consider connection down for replies taking longer than this value
	ResponseCacheTtl   *dur    `json:"response_cache_ttl,string"` // the life span of a cached response
	InternalTtl        *dur    `json:"internal_ttl,string"`       // maximum duration to wait for internal connections before giving up
	LockingTimeout     *dur    `json:"locking_timeout,string"`    // timeout internal locks to avoid deadlocks
	LogLevel           *int    `json:"log_level"`                 // control the level of messages logged (0-emerg to 7-debug)
	DataFolderPath     string  // Path towards data folder, for tests internal usage, not loading out of .json options
}

type Cache struct {
	Destinations   *CacheParam `json:"destinations"`
	RatingPlans    *CacheParam `json:"rating_plans"`
	RatingProfiles *CacheParam `json:"rating_profiles"`
	Lcr            *CacheParam `json:"lcr"`
	CdrStats       *CacheParam `json:"cdr_stats"`
	Actions        *CacheParam `json:"actions"`
	ActionPlans    *CacheParam `json:"action_plans"`
	ActionTriggers *CacheParam `json:"action_triggers"`
	SharedGroups   *CacheParam `json:"shared_groups"`
	Aliases        *CacheParam `json:"aliases"`
}

type Listen struct {
	RpcJson *string `json:"rpc_json"` // RPC JSON listening address
	RpcGob  *string `json:"rpc_gob"`  // RPC GOB listening address
	Http    *string `json:"http"`     // HTTP listening address
}

type TariffPlanDb struct { // database used to store active tariff plan configuration
	Host     *string `json:"db_host"`     // tariffplan_db host address
	Port     *string `json:"db_port"`     // port to reach the tariffplan_db
	Name     *string `json:"db_name"`     // tariffplan_db name to connect to
	User     *string `json:"db_user"`     // sername to use when connecting to tariffplan_db
	Password *string `json:"db_password"` // password to use when connecting to tariffplan_db
}

type DataDb struct { // database used to store runtime data (eg: accounts, cdr stats)
	Host            *string `json:"db_host"`           // data_db host address
	Port            *string `json:"db_port"`           // data_db port to reach the database
	Name            *string `json:"db_name"`           // data_db database name to connect to
	User            *string `json:"db_user"`           // username to use when connecting to data_db
	Password        *string `json:"db_password"`       // password to use when connecting to data_db
	LoadHistorySize *int    `json:"load_history_size"` // Number of records in the load history
}

type CdrDb struct { // database used to store offline tariff plans and CDRs
	Host         *string  `json:"db_host"`        // the host to connect to
	Port         *string  `json:"db_port"`        // the port to reach the stordb
	Name         *string  `json:"db_name"`        // stor database name
	User         *string  `json:"db_user"`        // username to use when connecting to stordb
	Password     *string  `json:"db_password"`    // password to use when connecting to stordb
	MaxOpenConns *int     `json:"max_open_conns"` // maximum database connections opened
	MaxIdleConns *int     `json:"max_idle_conns"` // maximum database connections idle
	CdrsIndexes  []string `json:"cdrs_indexes"`   // indexes on cdrs table to speed up queries, used only in case of mongo
}

type Balancer struct {
	Enabled *bool `json:"enabled"` // start Balancer service: <true|false>
}

type Rals struct {
	Enabled                  *bool     `json:"enabled"`                     // enable Rater service: <true|false>
	Balancer                 *string   `json:"balancer"`                    // register to balancer as worker: <""|*internal|x.y.z.y:1234>
	CdrStatsConns            []*HaPool `json:"cdrstats_conns"`              // address where to reach the cdrstats service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	HistorysConns            []*HaPool `json:"historys_conns"`              // address where to reach the history service, empty to disable history functionality: <""|*internal|x.y.z.y:1234>
	PubsubsConns             []*HaPool `json:"pubsubs_conns"`               // address where to reach the pubusb service, empty to disable pubsub functionality: <""|*internal|x.y.z.y:1234>
	UsersConns               []*HaPool `json:"users_conns"`                 // address where to reach the user service, empty to disable user profile functionality: <""|*internal|x.y.z.y:1234>
	AliasesConns             []*HaPool `json:"aliases_conns"`               // address where to reach the aliases service, empty to disable aliases functionality: <""|*internal|x.y.z.y:1234>
	RpSubjectPrefixMatching  *bool     `json:"rp_subject_prefix_matching"`  // enables prefix matching for the rating profile subject
	LcrSubjectPrefixMatching *bool     `json:"lcr_subject_prefix_matching"` // enables prefix matching for the lcr subject
}

type Scheduler struct {
	Enabled *bool `json:"enabled"` // start Scheduler service: <true|false>
}

type Cdrs struct {
	Enabled        *bool             `json:"enabled"`         // start the CDR Server service:  <true|false>
	ExtraFields    RsrList           `json:"extra_fields"`    // extra fields to store in CDRs for non-generic CDRs
	StoreCdrs      *bool             `json:"store_cdrs"`      // store cdrs in storDb
	AccountSummary *bool             `json:"account_summary"` // add account information from dataDB
	SmCostRetries  *int              `json:"sm_cost_retries"` // number of queries to sm_costs before recalculating CDR
	RalsConns      []*HaPool         `json:"rals_conns"`      // address where to reach the Rater for cost calculation, empty to disable functionality: <""|*internal|x.y.z.y:1234>
	PubsubsConns   []*HaPool         `json:"pubsubs_conns"`   // address where to reach the pubusb service, empty to disable pubsub functionality: <""|*internal|x.y.z.y:1234>
	UsersConns     []*HaPool         `json:"users_conns"`     // address where to reach the user service, empty to disable user profile functionality: <""|*internal|x.y.z.y:1234>
	AliasesConns   []*HaPool         `json:"aliases_conns"`   // address where to reach the aliases service, empty to disable aliases functionality: <""|*internal|x.y.z.y:1234>
	CdrStatsConns  []*HaPool         `json:"cdrstats_conns"`  // address where to reach the cdrstats service, empty to disable stats functionality<""|*internal|x.y.z.y:1234>
	CdrReplication []*CdrReplication `json:"cdr_replication"` // replicate the raw CDR to a number of servers
}

type CdrStats struct {
	Enabled *bool `json:"enabled"` // starts the cdrstats service: <true|false>
}

type Cdrc struct {
	ID                       *string         `json:"id"`                          // identifier of the CDRC runner
	Enabled                  *bool           `json:"enabled"`                     // enable CDR client functionality
	DryRun                   *bool           `json:"dry_run"`                     // do not send the CDRs to CDRS, just parse them
	CdrsConns                []*HaPool       `json:"cdrs_conns"`                  // address where to reach CDR server. <*internal|x.y.z.y:1234>
	CdrFormat                *string         `json:"cdr_format"`                  // CDR file format <csv|freeswitch_csv|fwv|opensips_flatstore>
	FieldSeparator           *string         `json:"field_separator"`             // separator used in case of csv files
	Timezone                 *string         `json:"timezone"`                    // timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	RunDelay                 *int            `json:"run_delay"`                   // sleep interval in seconds between consecutive runs, 0 to use automation via inotify
	MaxOpenFiles             *int            `json:"max_open_files"`              // maximum simultaneous files to process, 0 for unlimited
	DataUsageMultiplyFactor  *float64        `json:"data_usage_multiply_factor"`  // conversion factor for data usage
	CdrInDir                 *string         `json:"cdr_in_dir"`                  // absolute path towards the directory where the CDRs are stored
	CdrOutDir                *string         `json:"cdr_out_dir"`                 // absolute path towards the directory where processed CDRs will be moved
	FailedCallsPrefix        *string         `json:"failed_calls_prefix"`         // used in case of flatstore CDRs to avoid searching for BYE records
	CdrPath                  *string         `json:"cdr_path"`                    // path towards one CDR element in case of XML CDRs
	CdrSourceID              *string         `json:"cdr_source_id"`               // free form field, tag identifying the source of the CDRs within CDRS database
	CdrFilter                utils.RSRFields `json:"cdr_filter,string"`           // filter CDR records to import
	ContinueOnSuccess        *bool           `json:"continue_on_success"`         // continue to the next template if executed
	PartialRecordCache       *dur            `json:"partial_record_cache,string"` // duration to cache partial records when not pairing
	PartialCacheExpiryAction *string         `json:"partial_cache_expiry_action"` // action taken when cache when records in cache are timed-out <*dump_to_file|*post_cdr>
	HeaderFields             []*CdrField     `json:"header_fields"`               // template of the import header fields
	ContentFields            []*CdrField     `json:"content_fields"`              // import content_fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
	TrailerFields            []*CdrField     `json:"trailer_fields"`              // template of the import trailer fields
	CacheDumpFields          []*CdrField     `json:"cache_dump_fields"`
}

func (cdrc *Cdrc) RunDelayDuration() time.Duration {
	return time.Duration(*cdrc.RunDelay) * time.Second
}

func (cdrc *Cdrc) FieldSeparatorRune() rune {
	return rune((*cdrc.FieldSeparator)[0])
}

func (c *Config) CdrcProfiles() map[string][]*Cdrc {
	result := make(map[string][]*Cdrc)
	for _, cdrc := range *c.Cdrc {
		// here we can populate the defaults
		if cdrc.ID == nil || cdrc.CdrInDir == nil {
			continue // required data not populated, ignore
		}

		if _, found := result[*cdrc.CdrInDir]; !found {
			result[*cdrc.CdrInDir] = []*Cdrc{cdrc}
			continue
		}

		// check if we already have an cdrc with the same id in path
		// and replace it
		found := false
		for i, existingCdrc := range result[*cdrc.CdrInDir] {
			if existingCdrc.ID != nil && *existingCdrc.ID == *cdrc.ID {
				result[*cdrc.CdrInDir][i] = cdrc // found existing cdrc conf with same ID, owerwrite
				found = true
				break
			}
		}
		if !found {
			result[*cdrc.CdrInDir] = append(result[*cdrc.CdrInDir], cdrc)
		}
	}
	return result
}

type Cdre struct {
	CdrFormat                  *string     `json:"cdr_format"` // exported CDRs format <csv>
	FieldSeparator             *string     `json:"field_separator"`
	DataUsageMultiplyFactor    *float64    `json:"data_usage_multiply_factor"`    // multiply data usage before export (eg: convert from KBytes to Bytes)
	SmsUsageMultiplyFactor     *float64    `json:"sms_usage_multiply_factor"`     // multiply data usage before export (eg: convert from SMS unit to call duration in some billing systems)
	MmsUsageMultiplyFactor     *float64    `json:"mms_usage_multiply_factor"`     // multiply data usage before export (eg: convert from MMS unit to call duration in some billing systems)
	GenericUsageMultiplyFactor *float64    `json:"generic_usage_multiply_factor"` // multiply data usage before export (eg: convert from GENERIC unit to call duration in some billing systems)
	CostMultiplyFactor         *float64    `json:"cost_multiply_factor"`          // multiply cost before export, eg: add VAT
	CostRoundingDecimals       *int        `json:"cost_rounding_decimals"`        // rounding decimals for Cost values. -1 to disable rounding
	CostShiftDigits            *int        `json:"cost_shift_digits"`             // shift digits in the cost on export (eg: convert from EUR to cents)
	MaskDestinationID          *string     `json:"mask_destination_id"`           // destination id containing called addresses to be masked on export
	MaskLength                 *int        `json:"mask_length"`                   // length of the destination suffix to be masked
	ExportDirectory            *string     `json:"export_directory"`              // path where the exported CDRs will be placed
	HeaderFields               []*CdrField `json:"header_fields"`                 // template of the exported header fields
	ContentFields              []*CdrField `json:"content_fields"`
	TrailerFields              []*CdrField `json:"trailer_fields"` // template of the exported trailer fields
}

type SmGeneric struct {
	Enabled            *bool     `json:"enabled"`                      // starts SessionManager service: <true|false>
	ListenBijson       *string   `json:"listen_bijson"`                // address where to listen for bidirectional JSON-RPC requests
	RalsConns          []*HaPool `json:"rals_conns"`                   // address where to reach the Rater <""|*internal|127.0.0.1:2013>
	CdrsConns          []*HaPool `json:"cdrs_conns"`                   // address where to reach CDR Server, empty to disable CDR capturing <*internal|x.y.z.y:1234>
	DebitInterval      *dur      `json:"debit_interval,string"`        // interval to perform debits on.
	MinCallDuration    *dur      `json:"min_call_duration,string"`     // only authorize calls with allowed duration higher than this
	MaxCallDuration    *dur      `json:"max_call_duration,string"`     // maximum call duration a prepaid call can last
	SessionTtl         *dur      `json:"session_ttl,string"`           // time after a session with no updates is terminated, not defined by default
	SessionTtlLastUsed *dur      `json:"session_ttl_last_used,string"` // tweak LastUsed for sessions timing-out, not defined by default
	SessionTtlUsage    *dur      `json:"session_ttl_usage,string"`     // tweak Usage for sessions timing-out, not defined by default
	SessionIndexes     []string  `json:"session_indexes"`              // index sessions based on these fields for GetActiveSessions API
	PostActionTrigger  *bool     `json:"post_action_trigger"`          // execute triggered actions on debit confirmation
}

type SmFreeswitch struct {
	Enabled             *bool     `json:"enabled"`                      // starts SessionManager service: <true|false>
	RalsConns           []*HaPool `json:"rals_conns"`                   // address where to reach the Rater <""|*internal|127.0.0.1:2013>
	CdrsConns           []*HaPool `json:"cdrs_conns"`                   // address where to reach CDR Server, empty to disable CDR capturing <*internal|x.y.z.y:1234>
	RlsConns            []*HaPool `json:"rls_conns"`                    // address where to reach the ResourceLimiter service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	CreateCdr           *bool     `json:"create_cdr"`                   // create CDR out of events and sends them to CDRS component
	ExtraFields         RsrList   `json:"extra_fields"`                 // extra fields to store in auth/CDRs when creating them
	DebitInterval       *dur      `json:"debit_interval,string"`        // interval to perform debits on.
	MinCallDuration     *dur      `json:"min_call_duration,string"`     // only authorize calls with allowed duration higher than this
	MaxCallDuration     *dur      `json:"max_call_duration,string"`     // maximum call duration a prepaid call can last
	MinDurLowBalance    *dur      `json:"min_dur_low_balance,string"`   // threshold which will trigger low balance warnings for prepaid calls (needs to be lower than debit_interval)
	LowBalanceAnnFile   *string   `json:"low_balance_ann_file"`         // file to be played when low balance is reached for prepaid calls
	EmptyBalanceContext *string   `json:"empty_balance_context"`        // if defined, prepaid calls will be transfered to this context on empty balance
	EmptyBalanceAnnFile *string   `json:"empty_balance_ann_file"`       // file to be played before disconnecting prepaid calls on empty balance (applies only if no context defined)
	SubscribePark       *bool     `json:"subscribe_park"`               // subscribe via fsock to receive park events
	ChannelSyncInterval *dur      `json:"channel_sync_interval,string"` // sync channels with freeswitch regularly
	MaxWaitConnection   *dur      `json:"max_wait_connection,string"`   // maximum duration to wait for a connection to be retrieved from the pool
	EventSocketConns    []*HaPool `json:"event_socket_conns"`           // Instantiate connections to multiple FreeSWITCH servers
}

type SmKamailio struct {
	Enabled         *bool     `json:"enabled"`                  // starts SessionManager service: <true|false>
	RalsConns       []*HaPool `json:"rals_conns"`               // address where to reach the Rater <""|*internal|127.0.0.1:2013>
	CdrsConns       []*HaPool `json:"cdrs_conns"`               // address where to reach CDR Server, empty to disable CDR capturing <*internal|x.y.z.y:1234>
	CreateCdr       *bool     `json:"create_cdr"`               // create CDR out of events and sends them to CDRS component
	DebitInterval   *dur      `json:"debit_interval,string"`    // interval to perform debits on.
	MinCallDuration *dur      `json:"min_call_duration,string"` // only authorize calls with allowed duration higher than this
	MaxCallDuration *dur      `json:"max_call_duration,string"` // maximum call duration a prepaid call can last
	EvapiConns      []*HaPool `json:"evapi_conns"`              // instantiate connections to multiple Kamailio servers
}

type SmOpensips struct {
	Enabled                 *bool     `json:"enabled"`                          // starts SessionManager service: <true|false>
	ListenUdp               *string   `json:"listen_udp"`                       // address where to listen for datagram events coming from OpenSIPS
	RalsConns               []*HaPool `json:"rals_conns"`                       // address where to reach the Rater <""|*internal|127.0.0.1:2013>
	CdrsConns               []*HaPool `json:"cdrs_conns"`                       // address where to reach CDR Server, empty to disable CDR capturing <*internal|x.y.z.y:1234>
	Reconnects              *int      `json:"reconnects"`                       // number of reconnects if connection is lost
	CreateCdr               *bool     `json:"create_cdr"`                       // create CDR out of events and sends it to CDRS component
	DebitInterval           *dur      `json:"debit_interval,string"`            // interval to perform debits on.
	MinCallDuration         *dur      `json:"min_call_duration,string"`         // only authorize calls with allowed duration higher than this
	MaxCallDuration         *dur      `json:"max_call_duration,string"`         // maximum call duration a prepaid call can last
	EventsSubscribeInterval *dur      `json:"events_subscribe_interval,string"` // automatic events subscription to OpenSIPS, 0 to disable it
	MiAddr                  *string   `json:"mi_addr"`                          // address where to reach OpenSIPS MI to send session disconnects
}

type SmAsterisk struct {
	Enabled        *bool     `json:"enabled"`          // starts Asterisk SessionManager service: <true|false>
	SmGenericConns []*HaPool `json:"sm_generic_conns"` // connection towards SMG component for session management
	CreateCdr      *bool     `json:"create_cdr"`       // create CDR out of events and sends it to CDRS component
	AsteriskConns  []*HaPool `json:"asterisk_conns"`   // instantiate connections to multiple Asterisk servers
}

type RequestProcessor struct {
	ID                *string         `json:"id"`                    // formal identifier of this processor
	DryRun            *bool           `json:"dry_run"`               // do not send the events to SMG, just log them
	PublishEvent      *bool           `json:"publish_event"`         // if enabled, it will publish internal event to pubsub
	RequestFilter     utils.RSRFields `json:"request_filter,string"` // filter requests processed by this processor
	Flags             []string        `json:"flags"`                 // flags to influence processing behavior
	ContinueOnSuccess *bool           `json:"continue_on_success"`   // continue to the next template if executed
	AppendCca         *bool           `json:"append_cca"`            // when continuing will append cca fields to the previous ones
	CcrFields         []*CdrField     `json:"ccr_fields"`            // import content_fields template, tag will match internally CDR field, in case of .csv value will be represented by index of the field value
	CcaFields         []*CdrField     `json:"cca_fields"`            // fields returned in CCA
}

func (rp *RequestProcessor) FlagsMap() utils.StringMap {
	return utils.StringMapFromSlice(rp.Flags)
}

type DiameterAgent struct {
	Enabled            *bool               `json:"enabled"`               // enables the diameter agent: <true|false>
	Listen             *string             `json:"listen"`                // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesDir    *string             `json:"dictionaries_dir"`      // path towards directory holding additional dictionaries to load
	SmGenericConns     []*HaPool           `json:"sm_generic_conns"`      // connection towards SMG component for session management
	PubsubsConns       []*HaPool           `json:"pubsubs_conns"`         // address where to reach the pubusb service, empty to disable pubsub functionality: <""|*internal|x.y.z.y:1234>
	CreateCdr          *bool               `json:"create_cdr"`            // create CDR out of CCR terminate and send it to SMG component
	CdrRequiresSession *bool               `json:"cdr_requires_session"`  // only create CDR if there is an active session at terminate
	DebitInterval      *dur                `json:"debit_interval,string"` // interval for CCR updates
	Timezone           *string             `json:"timezone"`              // timezone for timestamps where not specified, empty for general defaults <""|UTC|Local|$IANA_TZ_DB>
	OriginHost         *string             `json:"origin_host"`           // diameter Origin-Host AVP used in replies
	OriginRealm        *string             `json:"origin_realm"`          // diameter Origin-Realm AVP used in replies
	VendorID           *int                `json:"vendor_id"`             // diameter Vendor-Id AVP used in replies
	ProductName        *string             `json:"product_name"`          // diameter Product-Name AVP used in replies
	RequestProcessors  []*RequestProcessor `json:"request_processors"`
}

type Historys struct {
	Enabled      *bool   `json:"enabled"`              // starts History service: <true|false>.
	HistoryDir   *string `json:"history_dir"`          // location on disk where to store history files.
	SaveInterval *dur    `json:"save_interval,string"` // interval to save changed cache into .git archive
}

type Pubsubs struct {
	Enabled *bool `json:"enabled"` // starts PubSub service: <true|false>.
}

type Aliases struct {
	Enabled *bool `json:"enabled"` // starts Aliases service: <true|false>.
}

type Users struct {
	Enabled         *bool    `json:"enabled"`          // starts User service: <true|false>.
	Indexes         []string `json:"indexes"`          // user profile field indexes
	ComplexityMatch *bool    `json:"complexity_match"` // stop after first matched users (no match complexity levels)
}

type Rls struct {
	Enabled           *bool     `json:"enabled"`                    // starts ResourceLimiter service: <true|false>.
	CdrStatsConns     []*HaPool `json:"cdrstats_conns"`             // address where to reach the cdrstats service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	CacheDumpInterval *dur      `json:"cache_dump_interval,string"` // dump cache regularly to dataDB, 0 - dump at start/shutdown: <""|*never|$dur>
	UsageTtl          *dur      `json:"usage_ttl,string"`           // expire usage records if older than this duration <""|*never|$dur>
}

type Mailer struct {
	Server       *string `json:"server"`        // the server to use when sending emails out
	AuthUser     *string `json:"auth_user"`     // authenticate to email server using this user
	AuthPassword *string `json:"auth_password"` // authenticate to email server with this password
	FromAddress  *string `json:"from_address"`  // from address used when sending emails out
}

type SureTax struct {
	Url                  *string         `json:"url"`                            // API url
	ClientNumber         *string         `json:"client_number"`                  // client number, provided by SureTax
	ValidationKey        *string         `json:"validation_key"`                 // validation key provided by SureTax
	BusinessUnit         *string         `json:"business_unit"`                  // client’s Business Unit
	Timezone             *string         `json:"timezone"`                       // convert the time of the events to this timezone before sending request out <UTC|Local|$IANA_TZ_DB>
	IncludeLocalCost     *bool           `json:"include_local_cost"`             // sum local calculated cost with tax one in final cost
	ReturnFileCode       *string         `json:"return_file_code"`               // default or Quote purposes <0|Q>
	ResponseGroup        *string         `json:"response_group"`                 // determines how taxes are grouped for the response <03|13>
	ResponseType         *string         `json:"response_type"`                  // determines the granularity of taxes and (optionally) the decimal precision for the tax calculations and amounts in the response
	RegulatoryCode       *string         `json:"regulatory_code"`                // provider type
	ClientTracking       utils.RSRFields `json:"client_tracking,string"`         // template extracting client information out of StoredCdr; <$RSRFields>
	CustomerNumber       utils.RSRFields `json:"customer_number,string"`         // template extracting customer number out of StoredCdr; <$RSRFields>
	OrigNumber           utils.RSRFields `json:"orig_number,string"`             // template extracting origination number out of StoredCdr; <$RSRFields>
	TermNumber           utils.RSRFields `json:"term_number,string"`             // template extracting termination number out of StoredCdr; <$RSRFields>
	BillToNumber         utils.RSRFields `json:"bill_to_number,string"`          // template extracting billed to number out of StoredCdr; <$RSRFields>
	Zipcode              utils.RSRFields `json:"zipcode,string"`                 // template extracting billing zip code out of StoredCdr; <$RSRFields>
	Plus4                utils.RSRFields `json:"plus4,string"`                   // template extracting billing zip code extension out of StoredCdr; <$RSRFields>
	P2PZipcode           utils.RSRFields `json:"p2pzipcode,string"`              // template extracting secondary zip code out of StoredCdr; <$RSRFields>
	P2PPlus4             utils.RSRFields `json:"p2pplus4,string"`                // template extracting secondary zip code extension out of StoredCdr; <$RSRFields>
	Units                utils.RSRFields `json:"units,string,string"`            // template extracting number of “lines” or unique charges contained within the revenue out of StoredCdr; <$RSRFields>
	UnitType             utils.RSRFields `json:"unit_type,string"`               // template extracting number of unique access lines out of StoredCdr; <$RSRFields>
	TaxIncluded          utils.RSRFields `json:"tax_included,string"`            // template extracting tax included in revenue out of StoredCdr; <$RSRFields>
	TaxSitusRule         utils.RSRFields `json:"tax_situs_rule,string"`          // template extracting tax situs rule out of StoredCdr; <$RSRFields>
	TransTypeCode        utils.RSRFields `json:"trans_type_code,string"`         // template extracting transaction type indicator out of StoredCdr; <$RSRFields>
	SalesTypeCode        utils.RSRFields `json:"sales_type_code,string"`         // template extracting sales type code out of StoredCdr; <$RSRFields>
	TaxExemptionCodeList utils.RSRFields `json:"tax_exemption_code_list,string"` // template extracting tax exemption code list out of StoredCdr; <$RSRFields>
}

func (c *Config) checkConfigSanity() error {
	if *c.Rals.Enabled {
		if *c.Rals.Balancer == utils.MetaInternal && !*c.Balancer.Enabled {
			return errors.New("Balancer not enabled but requested by Rater component.")
		}
		for _, connCfg := range c.Rals.CdrStatsConns {
			if connCfg.Address == utils.MetaInternal && !*c.CdrStats.Enabled {
				return errors.New("CDRStats not enabled but requested by Rater component.")
			}
		}
		for _, connCfg := range c.Rals.HistorysConns {
			if connCfg.Address == utils.MetaInternal && !*c.Historys.Enabled {
				return errors.New("History server not enabled but requested by Rater component.")
			}
		}
		for _, connCfg := range c.Rals.PubsubsConns {
			if connCfg.Address == utils.MetaInternal && !*c.Pubsubs.Enabled {
				return errors.New("PubSub server not enabled but requested by Rater component.")
			}
		}
		for _, connCfg := range c.Rals.AliasesConns {
			if connCfg.Address == utils.MetaInternal && !*c.Aliases.Enabled {
				return errors.New("Alias server not enabled but requested by Rater component.")
			}
		}
		for _, connCfg := range c.Rals.UsersConns {
			if connCfg.Address == utils.MetaInternal && !*c.Users.Enabled {
				return errors.New("User service not enabled but requested by Rater component.")
			}
		}
	}
	if *c.Cdrs.Enabled {
		for _, cdrsRaterConn := range c.Cdrs.RalsConns {
			if cdrsRaterConn.Address == utils.MetaInternal && !*c.Rals.Enabled {
				return errors.New("RALs not enabled but requested by CDRS component.")
			}
		}
		for _, connCfg := range c.Cdrs.PubsubsConns {
			if connCfg.Address == utils.MetaInternal && !*c.Pubsubs.Enabled {
				return errors.New("PubSubS not enabled but requested by CDRS component.")
			}
		}
		for _, connCfg := range c.Cdrs.UsersConns {
			if connCfg.Address == utils.MetaInternal && !*c.Users.Enabled {
				return errors.New("UserS not enabled but requested by CDRS component.")
			}
		}
		for _, connCfg := range c.Cdrs.AliasesConns {
			if connCfg.Address == utils.MetaInternal && !*c.Aliases.Enabled {
				return errors.New("AliaseS not enabled but requested by CDRS component.")
			}
		}
		for _, connCfg := range c.Cdrs.CdrStatsConns {
			if connCfg.Address == utils.MetaInternal && !*c.CdrStats.Enabled {
				return errors.New("CDRStatS not enabled but requested by CDRS component.")
			}
		}
	}
	for _, cdrcInst := range *c.Cdrc {
		if !*cdrcInst.Enabled {
			continue
		}
		if len(cdrcInst.CdrsConns) == 0 {
			return fmt.Errorf("<CDRC> Instance: %s, CdrC enabled but no CDRS defined!", *cdrcInst.ID)
		}
		for _, conn := range cdrcInst.CdrsConns {
			if conn.Address == utils.MetaInternal && !*c.Cdrs.Enabled {
				return errors.New("CDRS not enabled but referenced from CDRC")
			}
		}
		if len(cdrcInst.ContentFields) == 0 {
			return errors.New("CdrC enabled but no fields to be processed defined!")
		}
		if *cdrcInst.CdrFormat == utils.CSV {
			for _, cdrFld := range cdrcInst.ContentFields {
				for _, rsrFld := range cdrFld.Value {
					if _, errConv := strconv.Atoi(rsrFld.Id); errConv != nil && !rsrFld.IsStatic() {
						return fmt.Errorf("CDR fields must be indices in case of .csv files, have instead: %s", rsrFld.Id)
					}
				}
			}
		}
	}
	if *c.SmGeneric.Enabled {
		if len(c.SmGeneric.RalsConns) == 0 {
			return errors.New("<SMGeneric> RALs definition is mandatory!")
		}
		for _, smgRALsConn := range c.SmGeneric.RalsConns {
			if smgRALsConn.Address == utils.MetaInternal && !*c.Rals.Enabled {
				return errors.New("<SMGeneric> RALs not enabled but requested by SMGeneric component.")
			}
		}
		if len(c.SmGeneric.CdrsConns) == 0 {
			return errors.New("<SMGeneric> CDRs definition is mandatory!")
		}
		for _, smgCDRSConn := range c.SmGeneric.CdrsConns {
			if smgCDRSConn.Address == utils.MetaInternal && !*c.Cdrs.Enabled {
				return errors.New("<SMGeneric> CDRS not enabled but referenced by SMGeneric component")
			}
		}
	}
	if *c.SmFreeswitch.Enabled {
		if len(c.SmFreeswitch.RalsConns) == 0 {
			return errors.New("<SMFreeSWITCH> RALs definition is mandatory!")
		}
		for _, smFSRaterConn := range c.SmFreeswitch.RalsConns {
			if smFSRaterConn.Address == utils.MetaInternal && !*c.Rals.Enabled {
				return errors.New("<SMFreeSWITCH> RALs not enabled but requested by SMFreeSWITCH component.")
			}
		}
		if len(c.SmFreeswitch.CdrsConns) == 0 {
			return errors.New("<SMFreeSWITCH> CDRS definition is mandatory!")
		}
		for _, smFSCDRSConn := range c.SmFreeswitch.CdrsConns {
			if smFSCDRSConn.Address == utils.MetaInternal && !*c.Cdrs.Enabled {
				return errors.New("CDRS not enabled but referenced by SMFreeSWITCH component")
			}
		}
		for _, smFSRLsConn := range c.SmFreeswitch.RlsConns {
			if smFSRLsConn.Address == utils.MetaInternal && !*c.Rls.Enabled {
				return errors.New("RLs not enabled but referenced by SMFreeSWITCH component")
			}
		}
	}
	if *c.SmKamailio.Enabled {
		if len(c.SmKamailio.RalsConns) == 0 {
			return errors.New("Rater definition is mandatory!")
		}
		for _, smKamRaterConn := range c.SmKamailio.RalsConns {
			if smKamRaterConn.Address == utils.MetaInternal && !*c.Rals.Enabled {
				return errors.New("Rater not enabled but requested by SM-Kamailio component.")
			}
		}
		if len(c.SmKamailio.CdrsConns) == 0 {
			return errors.New("Cdrs definition is mandatory!")
		}
		for _, smKamCDRSConn := range c.SmKamailio.CdrsConns {
			if smKamCDRSConn.Address == utils.MetaInternal && !*c.Cdrs.Enabled {
				return errors.New("CDRS not enabled but referenced by SM-Kamailio component")
			}
		}
	}
	if *c.SmOpensips.Enabled {
		if len(c.SmOpensips.RalsConns) == 0 {
			return errors.New("<SMOpenSIPS> Rater definition is mandatory!")
		}
		for _, smOsipsRaterConn := range c.SmOpensips.RalsConns {
			if smOsipsRaterConn.Address == utils.MetaInternal && !*c.Rals.Enabled {
				return errors.New("<SMOpenSIPS> RALs not enabled.")
			}
		}
		if len(c.SmOpensips.CdrsConns) == 0 {
			return errors.New("<SMOpenSIPS> CDRs definition is mandatory!")
		}

		for _, smOsipsCDRSConn := range c.SmOpensips.CdrsConns {
			if smOsipsCDRSConn.Address == utils.MetaInternal && !*c.Cdrs.Enabled {
				return errors.New("<SMOpenSIPS> CDRS not enabled.")
			}
		}
	}
	if *c.SmAsterisk.Enabled {
		if len(c.SmAsterisk.SmGenericConns) == 0 {
			return errors.New("<SMAsterisk> SMG definition is mandatory!")
		}
		for _, smAstSMGConn := range c.SmAsterisk.SmGenericConns {
			if smAstSMGConn.Address == utils.MetaInternal && !*c.SmGeneric.Enabled {
				return errors.New("<SMAsterisk> SMG not enabled.")
			}
		}
	}
	if *c.DiameterAgent.Enabled {
		for _, daSMGConn := range c.DiameterAgent.SmGenericConns {
			if daSMGConn.Address == utils.MetaInternal && !*c.SmGeneric.Enabled {
				return errors.New("SMGeneric not enabled but referenced by DiameterAgent component")
			}
		}
		for _, daPubSubSConn := range c.DiameterAgent.PubsubsConns {
			if daPubSubSConn.Address == utils.MetaInternal && !*c.Pubsubs.Enabled {
				return errors.New("PubSubS not enabled but requested by DiameterAgent component.")
			}
		}
	}
	if c.Rls != nil && *c.Rls.Enabled {
		for _, connCfg := range c.Rls.CdrStatsConns {
			if connCfg.Address == utils.MetaInternal && !*c.CdrStats.Enabled {
				return errors.New("CDRStats not enabled but requested by ResourceLimiter component.")
			}
		}
	}
	return nil
}

type CacheParam struct {
	Limit    int  `json:"limit"`
	Ttl      dur  `json:"ttl,string"`
	Precache bool `json:"precache"`
}

type dur time.Duration

func (d *dur) UnmarshalJSON(data []byte) (err error) {
	var x time.Duration
	if x, err = utils.ParseDurationWithSecs(string(data)); err == nil {
		*d = dur(x)
	}
	return
}

func durPointer(d time.Duration) *dur {
	ccd := dur(d)
	return &ccd
}

func (d *dur) D() time.Duration {
	return time.Duration(*d)
}

type HaPool struct {
	Address         string `json:"address"`
	Transport       string `json:"transport"`
	User            string `json:"user"`
	Password        string `json:"password"`
	ConnectAttempts int    `json:"connect_attempts"`
	Reconnects      int    `json:"reconnects"`
}

type CdrField struct {
	Tag              string          `json:"tag"`      // Identifier for the administrator
	Type             string          `json:"type"`     // Type of field
	FieldID          string          `json:"field_id"` // Field identifier
	HandlerID        string          `json:"handler_id"`
	Value            utils.RSRFields `json:"value,string"`
	Append           bool            `json:"append"`
	FieldFilter      utils.RSRFields `json:"field_filter,string"`
	Width            int             `json:"width"`
	Strip            string          `json:"strip"`
	Padding          string          `json:"padding"`
	Layout           string          `json:"layout"`
	Mandatory        bool            `json:"mandatory"`
	CostShiftDigits  int             `json:"cost_shift_digits"` // Used in exports
	RoundingDecimals int             `json:"rounding_decimals"`
	Timezone         string          `json:"timezone"`
	MaskLen          int             `json:"mask_len"`
	MaskDestID       string          `json:"mask_dest_id"`
}

type RsrList utils.RSRFields

func (r *RsrList) UnmarshalJSON(data []byte) (err error) {
	s := make([]string, 0)
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	var x utils.RSRFields
	if x, err = utils.ParseRSRFieldsFromSlice(s); err == nil {
		*r = RsrList(x)
	}
	return
}

func (d *RsrList) R() utils.RSRFields {
	return utils.RSRFields(*d)
}

type CdrReplication struct {
	Transport     string          `json:"transport"`
	Address       string          `json:"address"`
	Synchronous   bool            `json:"synchronous"`
	Attempts      int             `json:"attempts"`          // Number of attempts if not success
	CdrFilter     utils.RSRFields `json:"cdr_filter,string"` // Only replicate if the filters here are matching
	ContentFields []*CdrField     `json:"content_fields"`
}

func (rplCfg CdrReplication) FallbackFileName() string {
	return fmt.Sprintf("cdr_%s_%s_%s.form", rplCfg.Transport, url.QueryEscape(rplCfg.Address), utils.GenUUID())
}

type FSDefaultField struct {
	Direction   string `json:"direction"`
	Tenant      string `json:"tenant"`
	Subject     string `json:"subject"`
	Destination string `json:"destination"`
	Account     string `json:"account"`
	Category    string `json:"category"`
}
