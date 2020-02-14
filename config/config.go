package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/DisposaBoy/JsonConfigReader"
	"github.com/accurateproject/accurate/utils"
)

var defaultConfig *Config

func init() {
	Reset()
}

func LoadBytes(data []byte, overwriteDefault bool) (err error) {
	// replace all env variables
	rp := regexp.MustCompile(`(\$\$\w+)`)
	env_replaced := rp.ReplaceAllStringFunc(string(data), func(s string) string {
		return os.Getenv(strings.Trim(s, "$$"))
	})
	//log.Print(env_replaced)
	r := JsonConfigReader.New(bytes.NewBufferString(env_replaced))

	newConf := &Config{}
	if err = json.NewDecoder(r).Decode(newConf); err != nil {
		return err
	}
	defaultMap := utils.ToMapMapStringInterface(defaultConfig)
	newMap := utils.ToMapMapStringInterface(newConf)

	for key, value := range newMap {
		if err = utils.Merge(defaultMap[key], value, overwriteDefault); err != nil {
			return err
		}
	}

	return
}

func LoadFile(filePath string, overwriteDefault bool) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return LoadBytes(data, overwriteDefault)
}

func LoadPath(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	overwriteDefault := true
	if fi.IsDir() {
		if err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				if filepath.Ext(walkPath) != ".json" {
					return nil
				}
				od := overwriteDefault
				overwriteDefault = false
				return LoadFile(walkPath, od)
			}
			return nil
		}); err != nil {
			return err
		}
	} else {
		od := overwriteDefault
		overwriteDefault = false
		return LoadFile(path, od)
	}
	return nil
}

func Get() *Config {
	return defaultConfig
}

func NewDefault() *Config {
	return &Config{
		General: &General{
			InstanceID:         utils.StringPointer(utils.GenUUID()),
			HttpSkipTlsVerify:  utils.BoolPointer(false),
			RoundingDecimals:   utils.IntPointer(5),
			TpexportDir:        utils.StringPointer("/var/spool/accurate/tpe"),
			HttpPosterAttempts: utils.IntPointer(3),
			HttpFailedDir:      utils.StringPointer("/var/spool/accurate/http_failed"),
			DefaultRequestType: utils.StringPointer("*rated"),
			DefaultCategory:    utils.StringPointer("call"),
			DefaultTenant:      utils.StringPointer("accurate"),
			DefaultTimezone:    utils.StringPointer("Local"),
			ConnectAttempts:    utils.IntPointer(3),
			Reconnects:         utils.IntPointer(-1),
			ConnectTimeout:     durPointer(1 * time.Second),
			ReplyTimeout:       durPointer(2 * time.Second),
			ResponseCacheTtl:   durPointer(0 * time.Second),
			InternalTtl:        durPointer(2 * time.Minute),
			LockingTimeout:     durPointer(5 * time.Second),
			LogLevel:           utils.IntPointer(6),
		},

		Cache: &Cache{
			Destinations:   &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
			RatingPlans:    &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: true},
			RatingProfiles: &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
			Lcr:            &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
			CdrStats:       &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
			Actions:        &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
			ActionPlans:    &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
			ActionTriggers: &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
			SharedGroups:   &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
			Aliases:        &CacheParam{Limit: 10000, Ttl: dur(0 * time.Second), Precache: false},
		},

		Listen: &Listen{
			RpcJson: utils.StringPointer("127.0.0.1:2012"),
			RpcGob:  utils.StringPointer("127.0.0.1:2013"),
			Http:    utils.StringPointer("127.0.0.1:2080"),
		},

		TariffPlanDb: &TariffPlanDb{
			Host:     utils.StringPointer("127.0.0.1"),
			Port:     utils.StringPointer("27017"),
			Name:     utils.StringPointer("tpdb"),
			User:     utils.StringPointer("accurate"),
			Password: utils.StringPointer("accuRate"),
		},

		DataDb: &DataDb{
			Host:            utils.StringPointer("127.0.0.1"),
			Port:            utils.StringPointer("27017"),
			Name:            utils.StringPointer("datadb"),
			User:            utils.StringPointer("accurate"),
			Password:        utils.StringPointer("accuRate"),
			LoadHistorySize: utils.IntPointer(10),
		},

		CdrDb: &CdrDb{
			Host:         utils.StringPointer("127.0.0.1"),
			Port:         utils.StringPointer("27017"),
			Name:         utils.StringPointer("cdrdb"),
			User:         utils.StringPointer("accurate"),
			Password:     utils.StringPointer("accuRate"),
			MaxOpenConns: utils.IntPointer(100),
			MaxIdleConns: utils.IntPointer(10),
			CdrsIndexes:  []string{},
		},

		Balancer: &Balancer{
			Enabled: utils.BoolPointer(false),
		},

		Rals: &Rals{
			Enabled:                  utils.BoolPointer(false),
			Balancer:                 utils.StringPointer(""),
			CdrStatsConns:            []*HaPool{},
			HistorysConns:            []*HaPool{},
			PubsubsConns:             []*HaPool{},
			UsersConns:               []*HaPool{},
			AliasesConns:             []*HaPool{},
			RpSubjectPrefixMatching:  utils.BoolPointer(false),
			LcrSubjectPrefixMatching: utils.BoolPointer(false),
		},

		Scheduler: &Scheduler{
			Enabled: utils.BoolPointer(false),
		},

		Cdrs: &Cdrs{
			Enabled:        utils.BoolPointer(false),
			ExtraFields:    nil,
			StoreCdrs:      utils.BoolPointer(true),
			AccountSummary: utils.BoolPointer(false),
			SmCostRetries:  utils.IntPointer(5),
			RalsConns:      []*HaPool{&HaPool{Address: "*internal"}},
			PubsubsConns:   []*HaPool{},
			UsersConns:     []*HaPool{},
			AliasesConns:   []*HaPool{},
			CdrStatsConns:  []*HaPool{},
			CdrReplication: []*CdrReplication{},
		},

		CdrStats: &CdrStats{
			Enabled: utils.BoolPointer(false),
		},

		Cdrc: &[]*Cdrc{
			&Cdrc{
				ID:                       utils.StringPointer("*default"),
				Enabled:                  utils.BoolPointer(false),
				DryRun:                   utils.BoolPointer(false),
				CdrsConns:                []*HaPool{&HaPool{Address: "*internal"}},
				CdrFormat:                utils.StringPointer("csv"),
				FieldSeparator:           utils.StringPointer(","),
				Timezone:                 utils.StringPointer(""),
				RunDelay:                 utils.IntPointer(0),
				MaxOpenFiles:             utils.IntPointer(1024),
				DataUsageMultiplyFactor:  utils.Float64Pointer(1024),
				CdrInDir:                 utils.StringPointer("/var/spool/accurate/cdrc/in"),
				CdrOutDir:                utils.StringPointer("/var/spool/accurate/cdrc/out"),
				FailedCallsPrefix:        utils.StringPointer("missed_calls"),
				CdrPath:                  utils.StringPointer(""),
				CdrSourceID:              utils.StringPointer("freeswitch_csv"),
				CdrFilter:                utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
				ContinueOnSuccess:        utils.BoolPointer(false),
				PartialRecordCache:       durPointer(10 * time.Second),
				PartialCacheExpiryAction: utils.StringPointer("*dump_to_file"),
				HeaderFields:             []*CdrField{},
				ContentFields: []*CdrField{
					&CdrField{Tag: "TOR", FieldID: "ToR", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("2", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "OriginID", FieldID: "OriginID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("3", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "RequestType", FieldID: "RequestType", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("4", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "Direction", FieldID: "Direction", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("5", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "Tenant", FieldID: "Tenant", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("6", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "Category", FieldID: "Category", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("7", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "Account", FieldID: "Account", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("8", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "Subject", FieldID: "Subject", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("9", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "Destination", FieldID: "Destination", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("10", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "SetupTime", FieldID: "SetupTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("11", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "AnswerTime", FieldID: "AnswerTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("12", utils.INFIELD_SEP), Mandatory: true},
					&CdrField{Tag: "Usage", FieldID: "Usage", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("13", utils.INFIELD_SEP), Mandatory: true},
				},
				TrailerFields: []*CdrField{},
				CacheDumpFields: []*CdrField{
					&CdrField{Tag: "UniqueID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("UniqueID", utils.INFIELD_SEP)},
					&CdrField{Tag: "RunID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("RunID", utils.INFIELD_SEP)},
					&CdrField{Tag: "TOR", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("ToR", utils.INFIELD_SEP)},
					&CdrField{Tag: "OriginID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("OriginID", utils.INFIELD_SEP)},
					&CdrField{Tag: "RequestType", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("RequestType", utils.INFIELD_SEP)},
					&CdrField{Tag: "Direction", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Direction", utils.INFIELD_SEP)},
					&CdrField{Tag: "Tenant", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Tenant", utils.INFIELD_SEP)},
					&CdrField{Tag: "Category", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Category", utils.INFIELD_SEP)},
					&CdrField{Tag: "Account", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Account", utils.INFIELD_SEP)},
					&CdrField{Tag: "Subject", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP)},
					&CdrField{Tag: "Destination", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Destination", utils.INFIELD_SEP)},
					&CdrField{Tag: "SetupTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("SetupTime", utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
					&CdrField{Tag: "AnswerTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("AnswerTime", utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
					&CdrField{Tag: "Usage", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Usage", utils.INFIELD_SEP)},
					&CdrField{Tag: "Cost", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Cost", utils.INFIELD_SEP)},
				},
			},
		},

		Cdre: &map[string]*Cdre{
			"*default": &Cdre{
				CdrFormat:                  utils.StringPointer("csv"),
				FieldSeparator:             utils.StringPointer(","),
				DataUsageMultiplyFactor:    utils.Float64Pointer(1),
				SmsUsageMultiplyFactor:     utils.Float64Pointer(1),
				MmsUsageMultiplyFactor:     utils.Float64Pointer(1),
				GenericUsageMultiplyFactor: utils.Float64Pointer(1),
				CostMultiplyFactor:         utils.Float64Pointer(1),
				ExportDirectory:            utils.StringPointer("/var/spool/accurate/cdre"),
				HeaderFields:               []*CdrField{},
				ContentFields: []*CdrField{
					&CdrField{Tag: "UniqueID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("UniqueID", utils.INFIELD_SEP)},
					&CdrField{Tag: "RunID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("RunID", utils.INFIELD_SEP)},
					&CdrField{Tag: "TOR", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("ToR", utils.INFIELD_SEP)},
					&CdrField{Tag: "OriginID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("OriginID", utils.INFIELD_SEP)},
					&CdrField{Tag: "RequestType", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("RequestType", utils.INFIELD_SEP)},
					&CdrField{Tag: "Direction", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Direction", utils.INFIELD_SEP)},
					&CdrField{Tag: "Tenant", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Tenant", utils.INFIELD_SEP)},
					&CdrField{Tag: "Category", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Category", utils.INFIELD_SEP)},
					&CdrField{Tag: "Account", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Account", utils.INFIELD_SEP)},
					&CdrField{Tag: "Subject", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP)},
					&CdrField{Tag: "Destination", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Destination", utils.INFIELD_SEP)},
					&CdrField{Tag: "SetupTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("SetupTime", utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
					&CdrField{Tag: "AnswerTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("AnswerTime", utils.INFIELD_SEP), Layout: "2006-01-02T15:04:05Z07:00"},
					&CdrField{Tag: "Usage", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Usage", utils.INFIELD_SEP)},
					&CdrField{Tag: "Cost", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Cost", utils.INFIELD_SEP), RoundingDecimals: 4},
				},
				TrailerFields: []*CdrField{},
			},
		},

		SmGeneric: &SmGeneric{
			Enabled:            utils.BoolPointer(false),
			ListenBijson:       utils.StringPointer("127.0.0.1:2014"),
			RalsConns:          []*HaPool{&HaPool{Address: "*internal"}},
			CdrsConns:          []*HaPool{&HaPool{Address: "*internal"}},
			DebitInterval:      durPointer(0 * time.Second),
			MinCallDuration:    durPointer(0 * time.Second),
			MaxCallDuration:    durPointer(3 * time.Hour),
			SessionTtl:         durPointer(0 * time.Second),
			SessionTtlLastUsed: nil,
			SessionTtlUsage:    nil,
			SessionIndexes:     []string{},
			PostActionTrigger:  utils.BoolPointer(false),
		},

		SmFreeswitch: &SmFreeswitch{
			Enabled:             utils.BoolPointer(false),
			RalsConns:           []*HaPool{&HaPool{Address: "*internal"}},
			CdrsConns:           []*HaPool{&HaPool{Address: "*internal"}},
			RlsConns:            []*HaPool{},
			CreateCdr:           utils.BoolPointer(false),
			ExtraFields:         nil,
			DebitInterval:       durPointer(10 * time.Second),
			MinCallDuration:     durPointer(0 * time.Second),
			MaxCallDuration:     durPointer(3 * time.Hour),
			MinDurLowBalance:    durPointer(5 * time.Second),
			LowBalanceAnnFile:   utils.StringPointer(""),
			EmptyBalanceContext: utils.StringPointer(""),
			EmptyBalanceAnnFile: utils.StringPointer(""),
			SubscribePark:       utils.BoolPointer(true),
			ChannelSyncInterval: durPointer(5 * time.Minute),
			MaxWaitConnection:   durPointer(2 * time.Second),
			EventSocketConns:    []*HaPool{&HaPool{Address: "127.0.0.1:8021", Password: "ClueCon", Reconnects: 5}},
		},

		SmKamailio: &SmKamailio{
			Enabled:         utils.BoolPointer(false),
			RalsConns:       []*HaPool{&HaPool{Address: "*internal"}},
			CdrsConns:       []*HaPool{&HaPool{Address: "*internal"}},
			CreateCdr:       utils.BoolPointer(false),
			DebitInterval:   durPointer(10 * time.Second),
			MinCallDuration: durPointer(0 * time.Second),
			MaxCallDuration: durPointer(3 * time.Hour),
			EvapiConns:      []*HaPool{&HaPool{Address: "127.0.0.1:8448", Reconnects: 5}},
		},

		SmOpensips: &SmOpensips{
			Enabled:                 utils.BoolPointer(false),
			ListenUdp:               utils.StringPointer("127.0.0.1:2020"),
			RalsConns:               []*HaPool{&HaPool{Address: "*internal"}},
			CdrsConns:               []*HaPool{&HaPool{Address: "*internal"}},
			Reconnects:              utils.IntPointer(5),
			CreateCdr:               utils.BoolPointer(false),
			DebitInterval:           durPointer(10 * time.Second),
			MinCallDuration:         durPointer(0 * time.Second),
			MaxCallDuration:         durPointer(3 * time.Hour),
			EventsSubscribeInterval: durPointer(60 * time.Second),
			MiAddr:                  utils.StringPointer("127.0.0.1:8020"),
		},

		SmAsterisk: &SmAsterisk{
			Enabled:        utils.BoolPointer(false),
			SmGenericConns: []*HaPool{&HaPool{Address: "*internal"}},
			CreateCdr:      utils.BoolPointer(false),
			AsteriskConns:  []*HaPool{&HaPool{Address: "127.0.0.1:8088", User: "accurate", Password: "accuRate", ConnectAttempts: 3, Reconnects: 5}},
		},

		DiameterAgent: &DiameterAgent{
			Enabled:            utils.BoolPointer(false),
			Listen:             utils.StringPointer("127.0.0.1:3868"),
			DictionariesDir:    utils.StringPointer("/usr/share/accurate/diameter/dict/"),
			SmGenericConns:     []*HaPool{&HaPool{Address: "*internal"}},
			PubsubsConns:       []*HaPool{},
			CreateCdr:          utils.BoolPointer(true),
			CdrRequiresSession: utils.BoolPointer(true),
			DebitInterval:      durPointer(5 * time.Minute),
			Timezone:           utils.StringPointer(""),
			OriginHost:         utils.StringPointer("CC-DA"),
			OriginRealm:        utils.StringPointer("accurate"),
			VendorID:           utils.IntPointer(0),
			ProductName:        utils.StringPointer("accuRate"),
			RequestProcessors: []*RequestProcessor{
				&RequestProcessor{
					ID:                utils.StringPointer("*default"),
					DryRun:            utils.BoolPointer(false),
					PublishEvent:      utils.BoolPointer(false),
					RequestFilter:     utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Type(0)", utils.INFIELD_SEP),
					Flags:             []string{},
					ContinueOnSuccess: utils.BoolPointer(false),
					AppendCca:         utils.BoolPointer(true),
					CcrFields: []*CdrField{
						&CdrField{Tag: "TOR", FieldID: "ToR", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*voice", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "OriginID", FieldID: "OriginID", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Session-Id", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "RequestType", FieldID: "RequestType", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*users", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "Direction", FieldID: "Direction", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*out", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "Tenant", FieldID: "Tenant", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*users", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "Category", FieldID: "Category", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^call", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "Account", FieldID: "Account", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*users", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "Subject", FieldID: "Subject", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("^*users", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "Destination", FieldID: "Destination", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Service-Information>IN-Information>Real-Called-Number", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "SetupTime", FieldID: "SetupTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Event-Timestamp", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "AnswerTime", FieldID: "AnswerTime", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Event-Timestamp", utils.INFIELD_SEP), Mandatory: true},
						&CdrField{Tag: "Usage", FieldID: "Usage", Type: "*handler", HandlerID: "*ccr_usage", Mandatory: true},
						&CdrField{Tag: "SubscriberID", FieldID: "SubscriberId", Type: "*composed", Value: utils.ParseRSRFieldsMustCompile("Subscription-Id>Subscription-Id-Data", utils.INFIELD_SEP), Mandatory: true},
					},
					CcaFields: []*CdrField{
						&CdrField{Tag: "GrantedUnits", FieldID: "Granted-Service-Unit>CC-Time", Type: "*handler", HandlerID: "*cca_usage", Mandatory: true},
					},
				},
			},
		},

		Historys: &Historys{
			Enabled:      utils.BoolPointer(false),
			HistoryDir:   utils.StringPointer("/var/lib/accurate/history"),
			SaveInterval: durPointer(1 * time.Second),
		},

		Pubsubs: &Pubsubs{
			Enabled: utils.BoolPointer(false),
		},

		Aliases: &Aliases{
			Enabled: utils.BoolPointer(false),
		},

		Users: &Users{
			Enabled:         utils.BoolPointer(false),
			Indexes:         []string{},
			ComplexityMatch: utils.BoolPointer(false),
		},

		Rls: &Rls{
			Enabled:           utils.BoolPointer(false),
			CdrStatsConns:     []*HaPool{},
			CacheDumpInterval: durPointer(0 * time.Second),
			UsageTtl:          durPointer(3 * time.Hour),
		},

		Mailer: &Mailer{
			Server:       utils.StringPointer("localhost"),
			AuthUser:     utils.StringPointer("accurate"),
			AuthPassword: utils.StringPointer("accuRate"),
			FromAddress:  utils.StringPointer("cc-mailer@localhost.localdomain"),
		},

		SureTax: &SureTax{
			Url:                  utils.StringPointer(""),
			ClientNumber:         utils.StringPointer(""),
			ValidationKey:        utils.StringPointer(""),
			BusinessUnit:         utils.StringPointer(""),
			Timezone:             utils.StringPointer("Local"),
			IncludeLocalCost:     utils.BoolPointer(false),
			ReturnFileCode:       utils.StringPointer("0"),
			ResponseGroup:        utils.StringPointer("03"),
			ResponseType:         utils.StringPointer("D4"),
			RegulatoryCode:       utils.StringPointer("03"),
			ClientTracking:       utils.ParseRSRFieldsMustCompile("UniqueID", utils.INFIELD_SEP),
			CustomerNumber:       utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP),
			OrigNumber:           utils.ParseRSRFieldsMustCompile("Subject", utils.INFIELD_SEP),
			TermNumber:           utils.ParseRSRFieldsMustCompile("Destination", utils.INFIELD_SEP),
			BillToNumber:         utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			Zipcode:              utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			Plus4:                utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			P2PZipcode:           utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			P2PPlus4:             utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
			Units:                utils.ParseRSRFieldsMustCompile("^1", utils.INFIELD_SEP),
			UnitType:             utils.ParseRSRFieldsMustCompile("^00", utils.INFIELD_SEP),
			TaxIncluded:          utils.ParseRSRFieldsMustCompile("^0", utils.INFIELD_SEP),
			TaxSitusRule:         utils.ParseRSRFieldsMustCompile("^04", utils.INFIELD_SEP),
			TransTypeCode:        utils.ParseRSRFieldsMustCompile("^010101", utils.INFIELD_SEP),
			SalesTypeCode:        utils.ParseRSRFieldsMustCompile("^R", utils.INFIELD_SEP),
			TaxExemptionCodeList: utils.ParseRSRFieldsMustCompile("", utils.INFIELD_SEP),
		},
	}
}

func Reset() {
	defaultConfig = NewDefault()
}
