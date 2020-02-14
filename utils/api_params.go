package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/accurateproject/accurate/dec"
)

// Paginate stuff around items returned
type Paginator struct {
	Limit      *int   // Limit the number of items returned
	Offset     *int   // Offset of the first item returned (eg: use Limit*Page in case of PerPage items)
	SearchTerm string // Global matching pattern in items returned, partially used in some APIs
}

// CDRsFilter is a filter used to get records out of storDB
type CDRsFilter struct {
	UniqueIDs              []string          // If provided, it will filter based on the uniqueids present in list
	NotUniqueIDs           []string          // Filter specific UniqueIDs out
	RunIDs                 []string          // If provided, it will filter on mediation runid
	NotRunIDs              []string          // Filter specific runIds out
	OriginHosts            []string          // If provided, it will filter cdrhost
	NotOriginHosts         []string          // Filter out specific cdr hosts
	Sources                []string          // If provided, it will filter cdrsource
	NotSources             []string          // Filter out specific CDR sources
	ToRs                   []string          // If provided, filter on TypeOfRecord
	NotToRs                []string          // Filter specific TORs out
	RequestTypes           []string          // If provided, it will fiter reqtype
	NotRequestTypes        []string          // Filter out specific request types
	Directions             []string          // If provided, it will fiter direction
	NotDirections          []string          // Filter out specific directions
	Tenants                []string          // If provided, it will filter tenant
	NotTenants             []string          // If provided, it will filter tenant
	Categories             []string          // If provided, it will filter çategory
	NotCategories          []string          // Filter out specific categories
	Accounts               []string          // If provided, it will filter account
	NotAccounts            []string          // Filter out specific Accounts
	Subjects               []string          // If provided, it will filter the rating subject
	NotSubjects            []string          // Filter out specific subjects
	DestinationPrefixes    []string          // If provided, it will filter on destination prefix
	NotDestinationPrefixes []string          // Filter out specific destination prefixes
	Suppliers              []string          // If provided, it will filter the supplier
	NotSuppliers           []string          // Filter out specific suppliers
	DisconnectCauses       []string          // Filter for disconnect Cause
	NotDisconnectCauses    []string          // Filter out specific disconnect causes
	Costs                  []float64         // Query based on costs specified
	NotCosts               []float64         // Filter out specific costs out from result
	ExtraFields            map[string]string // Query based on extra fields content
	NotExtraFields         map[string]string // Filter out based on extra fields content
	OrderIDStart           *int64            // Export from this order identifier
	OrderIDEnd             *int64            // Export smaller than this order identifier
	SetupTimeStart         *time.Time        // Start of interval, bigger or equal than configured
	SetupTimeEnd           *time.Time        // End interval, smaller than setupTime
	AnswerTimeStart        *time.Time        // Start of interval, bigger or equal than configured
	AnswerTimeEnd          *time.Time        // End interval, smaller than answerTime
	CreatedAtStart         *time.Time        // Start of interval, bigger or equal than configured
	CreatedAtEnd           *time.Time        // End interval, smaller than
	UpdatedAtStart         *time.Time        // Start of interval, bigger or equal than configured
	UpdatedAtEnd           *time.Time        // End interval, smaller than
	MinUsage               string            // Start of the usage interval (>=)
	MaxUsage               string            // End of the usage interval (<)
	MinPDD                 string            // Start of the pdd interval (>=)
	MaxPDD                 string            // End of the pdd interval (<)
	MinCost                *float64          // Start of the cost interval (>=)
	MaxCost                *float64          // End of the usage interval (<)
	Unscoped               bool              // Include soft-deleted records in results
	Count                  bool              // If true count the items instead of returning data
	Paginator
}

type AttrRateCDRs struct {
	RPCCDRsFilter
	StoreCDRs     *bool
	SendToStatS   *bool // Set to true if the CDRs should be sent to stats server
	ReplicateCDRs *bool // Replicate results
}

type AttrDerivedChargers struct {
	Direction, Tenant, Category, Account, Subject, Destination string
}

// RPCCDRsFilter is a filter used in Rpc calls
// RPCCDRsFilter is slightly different than CDRsFilter by using string instead of Time filters
type RPCCDRsFilter struct {
	UniqueIDs              []string          // If provided, it will filter based on the uniqueids present in list
	NotUniqueIDs           []string          // Filter specific UniqueIDs out
	RunIDs                 []string          // If provided, it will filter on mediation runid
	NotRunIDs              []string          // Filter specific runIds out
	OriginHosts            []string          // If provided, it will filter cdrhost
	NotOriginHosts         []string          // Filter out specific cdr hosts
	Sources                []string          // If provided, it will filter cdrsource
	NotSources             []string          // Filter out specific CDR sources
	ToRs                   []string          // If provided, filter on TypeOfRecord
	NotToRs                []string          // Filter specific TORs out
	RequestTypes           []string          // If provided, it will fiter reqtype
	NotRequestTypes        []string          // Filter out specific request types
	Directions             []string          // If provided, it will fiter direction
	NotDirections          []string          // Filter out specific directions
	Tenants                []string          // If provided, it will filter tenant
	NotTenants             []string          // If provided, it will filter tenant
	Categories             []string          // If provided, it will filter çategory
	NotCategories          []string          // Filter out specific categories
	Accounts               []string          // If provided, it will filter account
	NotAccounts            []string          // Filter out specific Accounts
	Subjects               []string          // If provided, it will filter the rating subject
	NotSubjects            []string          // Filter out specific subjects
	DestinationPrefixes    []string          // If provided, it will filter on destination prefix
	NotDestinationPrefixes []string          // Filter out specific destination prefixes
	Suppliers              []string          // If provided, it will filter the supplier
	NotSuppliers           []string          // Filter out specific suppliers
	DisconnectCauses       []string          // Filter for disconnect Cause
	NotDisconnectCauses    []string          // Filter out specific disconnect causes
	Costs                  []float64         // Query based on costs specified
	NotCosts               []float64         // Filter out specific costs out from result
	ExtraFields            map[string]string // Query based on extra fields content
	NotExtraFields         map[string]string // Filter out based on extra fields content
	OrderIDStart           *int64            // Export from this order identifier
	OrderIDEnd             *int64            // Export smaller than this order identifier
	SetupTimeStart         string            // Start of interval, bigger or equal than configured
	SetupTimeEnd           string            // End interval, smaller than setupTime
	AnswerTimeStart        string            // Start of interval, bigger or equal than configured
	AnswerTimeEnd          string            // End interval, smaller than answerTime
	CreatedAtStart         string            // Start of interval, bigger or equal than configured
	CreatedAtEnd           string            // End interval, smaller than
	UpdatedAtStart         string            // Start of interval, bigger or equal than configured
	UpdatedAtEnd           string            // End interval, smaller than
	MinUsage               string            // Start of the usage interval (>=)
	MaxUsage               string            // End of the usage interval (<)
	MinPDD                 string            // Start of the pdd interval (>=)
	MaxPDD                 string            // End of the pdd interval (<)
	MinCost                *float64          // Start of the cost interval (>=)
	MaxCost                *float64          // End of the usage interval (<)
	Paginator                                // Add pagination
}

func (attr *RPCCDRsFilter) AsCDRsFilter(timezone string) (*CDRsFilter, error) {
	cdrFltr := &CDRsFilter{
		UniqueIDs:              attr.UniqueIDs,
		NotUniqueIDs:           attr.NotUniqueIDs,
		RunIDs:                 attr.RunIDs,
		NotRunIDs:              attr.NotRunIDs,
		ToRs:                   attr.ToRs,
		NotToRs:                attr.NotToRs,
		OriginHosts:            attr.OriginHosts,
		NotOriginHosts:         attr.NotOriginHosts,
		Sources:                attr.Sources,
		NotSources:             attr.NotSources,
		RequestTypes:           attr.RequestTypes,
		NotRequestTypes:        attr.NotRequestTypes,
		Directions:             attr.Directions,
		NotDirections:          attr.NotDirections,
		Tenants:                attr.Tenants,
		NotTenants:             attr.NotTenants,
		Categories:             attr.Categories,
		NotCategories:          attr.NotCategories,
		Accounts:               attr.Accounts,
		NotAccounts:            attr.NotAccounts,
		Subjects:               attr.Subjects,
		NotSubjects:            attr.NotSubjects,
		DestinationPrefixes:    attr.DestinationPrefixes,
		NotDestinationPrefixes: attr.NotDestinationPrefixes,
		Suppliers:              attr.Suppliers,
		NotSuppliers:           attr.NotSuppliers,
		DisconnectCauses:       attr.DisconnectCauses,
		NotDisconnectCauses:    attr.NotDisconnectCauses,
		Costs:                  attr.Costs,
		NotCosts:               attr.NotCosts,
		ExtraFields:            attr.ExtraFields,
		NotExtraFields:         attr.NotExtraFields,
		OrderIDStart:           attr.OrderIDStart,
		OrderIDEnd:             attr.OrderIDEnd,
		MinUsage:               attr.MinUsage,
		MaxUsage:               attr.MaxUsage,
		MinPDD:                 attr.MinPDD,
		MaxPDD:                 attr.MaxPDD,
		MinCost:                attr.MinCost,
		MaxCost:                attr.MaxCost,
		Paginator:              attr.Paginator,
	}
	if len(attr.SetupTimeStart) != 0 {
		if sTimeStart, err := ParseTimeDetectLayout(attr.SetupTimeStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.SetupTimeStart = &sTimeStart
		}
	}
	if len(attr.SetupTimeEnd) != 0 {
		if sTimeEnd, err := ParseTimeDetectLayout(attr.SetupTimeEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.SetupTimeEnd = &sTimeEnd
		}
	}
	if len(attr.AnswerTimeStart) != 0 {
		if aTimeStart, err := ParseTimeDetectLayout(attr.AnswerTimeStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeStart = &aTimeStart
		}
	}
	if len(attr.AnswerTimeEnd) != 0 {
		if aTimeEnd, err := ParseTimeDetectLayout(attr.AnswerTimeEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeEnd = &aTimeEnd
		}
	}
	if len(attr.CreatedAtStart) != 0 {
		if tStart, err := ParseTimeDetectLayout(attr.CreatedAtStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.CreatedAtStart = &tStart
		}
	}
	if len(attr.CreatedAtEnd) != 0 {
		if tEnd, err := ParseTimeDetectLayout(attr.CreatedAtEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.CreatedAtEnd = &tEnd
		}
	}
	if len(attr.UpdatedAtStart) != 0 {
		if tStart, err := ParseTimeDetectLayout(attr.UpdatedAtStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.UpdatedAtStart = &tStart
		}
	}
	if len(attr.UpdatedAtEnd) != 0 {
		if tEnd, err := ParseTimeDetectLayout(attr.UpdatedAtEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.UpdatedAtEnd = &tEnd
		}
	}
	return cdrFltr, nil
}

func NewDTCSFromRPKey(rpKey string) (*DirectionTenantCategorySubject, error) {
	rpSplt := strings.Split(rpKey, CONCATENATED_KEY_SEP)
	if len(rpSplt) != 4 {
		return nil, fmt.Errorf("Unsupported format for DirectionTenantCategorySubject: %s", rpKey)
	}
	return &DirectionTenantCategorySubject{rpSplt[0], rpSplt[1], rpSplt[2], rpSplt[3]}, nil
}

type DirectionTenantCategorySubject struct {
	Direction, Tenant, Category, Subject string
}

type TenantAccount struct {
	Tenant, Account string
}

type AttrRLsCache struct {
	LoadID           string
	ResourceLimitIDs []string
}

type AttrRLsResourceUsage struct {
	ResourceUsageID string
	Event           map[string]interface{}
	RequestedUnits  float64
}

type AttrExecuteAction struct {
	Tenant    string
	Account   string
	ActionsID string
}

type AttrSetActionGroups struct {
	ActionsID string           // Actions id
	Overwrite bool             // If previously defined, will be overwritten
	Actions   []*TpActionGroup // Set of actions this Actions profile will perform
}

// ValueOrDefault is used to populate empty values with *any or *default if value missing
func ValueOrDefault1(val string, deflt string) string {
	if len(val) == 0 {
		val = deflt
	}
	return val
}

// Attributes to send on SessionDisconnect by SMG
type AttrDisconnectSession struct {
	EventStart map[string]interface{}
	Reason     string
}

type AttrStatsQueueDisable struct {
	Tenant  string
	ID      string
	Disable bool
}

type AttrStatsQueueID struct {
	Tenant string
	ID     string
}

type AttrStatsQueueIDs struct {
	Tenant string
	IDs    []string
}

type AttrReloadCache struct {
	Tenants []string
}
type AttrExportCdrsToFile struct {
	CdrFormat                  *string  // Cdr output file format <utils.CdreCdrFormats>
	FieldSeparator             *string  // Separator used between fields
	ExportID                   *string  // Optional exportid
	ExportDirectory            *string  // If provided it overwrites the configured export directory
	ExportFileName             *string  // If provided the output filename will be set to this
	ExportTemplate             *string  // Exported fields template  <""|fld1,fld2|>
	DataUsageMultiplyFactor    *float64 // Multiply data usage before export (eg: convert from KBytes to Bytes)
	SMSUsageMultiplyFactor     *float64 // Multiply sms usage before export (eg: convert from SMS unit to call duration for some billing systems)
	MMSUsageMultiplyFactor     *float64 // Multiply mms usage before export (eg: convert from MMS unit to call duration for some billing systems)
	GenericUsageMultiplyFactor *float64 // Multiply generic usage before export (eg: convert from GENERIC unit to call duration for some billing systems)
	CostMultiplyFactor         *float64 // Multiply the cost before export, eg: apply VAT
	Verbose                    bool     // Disable UniqueIDs reporting in reply/ExportedUniqueIDs and reply/UnexportedUniqueIDs
	RPCCDRsFilter                       // Inherit the CDR filter attributes
}

type ExportedFileCdrs struct {
	ExportedFilePath          string            // Full path to the newly generated export file
	TotalRecords              int               // Number of CDRs to be exported
	TotalCost                 *dec.Dec          // Sum of all costs in exported CDRs
	FirstOrderId, LastOrderId int64             // The order id of the last exported CDR
	ExportedUniqueIDs         []string          // List of successfuly exported uniqueids in the file
	UnexportedUniqueIDs       map[string]string // Map of errored CDRs, map key is uniqueid, value will be the error string
}
type AttrExpFileCdrs struct {
	CdrFormat                  *string  // Cdr output file format <CdreCdrFormats>
	FieldSeparator             *string  // Separator used between fields
	ExportId                   *string  // Optional exportid
	ExportDir                  *string  // If provided it overwrites the configured export directory
	ExportFileName             *string  // If provided the output filename will be set to this
	ExportTemplate             *string  // Exported fields template  <""|fld1,fld2|*xml:instance_name>
	DataUsageMultiplyFactor    *float64 // Multiply data usage before export (eg: convert from KBytes to Bytes)
	SmsUsageMultiplyFactor     *float64 // Multiply sms usage before export (eg: convert from SMS unit to call duration for some billing systems)
	MmsUsageMultiplyFactor     *float64 // Multiply mms usage before export (eg: convert from MMS unit to call duration for some billing systems)
	GenericUsageMultiplyFactor *float64 // Multiply generic usage before export (eg: convert from GENERIC unit to call duration for some billing systems)
	CostMultiplyFactor         *float64 // Multiply the cost before export, eg: apply VAT
	UniqueIDs                  []string // If provided, it will filter based on the uniqueids present in list
	MediationRunIds            []string // If provided, it will filter on mediation runid
	TORs                       []string // If provided, filter on TypeOfRecord
	CdrHosts                   []string // If provided, it will filter cdrhost
	CdrSources                 []string // If provided, it will filter cdrsource
	ReqTypes                   []string // If provided, it will fiter reqtype
	Directions                 []string // If provided, it will fiter direction
	Tenants                    []string // If provided, it will filter tenant
	Categories                 []string // If provided, it will filter çategory
	Accounts                   []string // If provided, it will filter account
	Subjects                   []string // If provided, it will filter the rating subject
	DestinationPrefixes        []string // If provided, it will filter on destination prefix
	OrderIdStart               *int64   // Export from this order identifier
	OrderIdEnd                 *int64   // Export smaller than this order identifier
	TimeStart                  string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd                    string   // If provided, it will represent the end of the CDRs interval (<)
	SkipErrors                 bool     // Do not export errored CDRs
	SkipRated                  bool     // Do not export rated CDRs
	SuppressUniqueIDs          bool     // Disable UniqueIDs reporting in reply/ExportedUniqueIDs and reply/UnexportedUniqueIDs
	Paginator
}

func (attr *AttrExpFileCdrs) AsCDRsFilter(timezone string) (*CDRsFilter, error) {
	cdrFltr := &CDRsFilter{
		UniqueIDs:           attr.UniqueIDs,
		RunIDs:              attr.MediationRunIds,
		NotRunIDs:           []string{MetaRaw}, // In exportv1 automatically filter out *raw CDRs
		ToRs:                attr.TORs,
		OriginHosts:         attr.CdrHosts,
		Sources:             attr.CdrSources,
		RequestTypes:        attr.ReqTypes,
		Directions:          attr.Directions,
		Tenants:             attr.Tenants,
		Categories:          attr.Categories,
		Accounts:            attr.Accounts,
		Subjects:            attr.Subjects,
		DestinationPrefixes: attr.DestinationPrefixes,
		OrderIDStart:        attr.OrderIdStart,
		OrderIDEnd:          attr.OrderIdEnd,
		Paginator:           attr.Paginator,
	}
	if len(attr.TimeStart) != 0 {
		if answerTimeStart, err := ParseTimeDetectLayout(attr.TimeStart, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeStart = &answerTimeStart
		}
	}
	if len(attr.TimeEnd) != 0 {
		if answerTimeEnd, err := ParseTimeDetectLayout(attr.TimeEnd, timezone); err != nil {
			return nil, err
		} else {
			cdrFltr.AnswerTimeEnd = &answerTimeEnd
		}
	}
	if attr.SkipRated {
		cdrFltr.MaxCost = Float64Pointer(-1.0)
	} else if attr.SkipRated {
		cdrFltr.MinCost = Float64Pointer(0.0)
		cdrFltr.MaxCost = Float64Pointer(-1.0)
	}
	return cdrFltr, nil
}

type AttrGetCallCost struct {
	UniqueID string // Unique id of the CDR
	RunID    string // Run Id
}

type AttrRateCdrs struct {
	UniqueIDs           []string // If provided, it will filter based on the uniqueids present in list
	MediationRunIds     []string // If provided, it will filter on mediation runid
	TORs                []string // If provided, filter on TypeOfRecord
	CdrHosts            []string // If provided, it will filter cdrhost
	CdrSources          []string // If provided, it will filter cdrsource
	ReqTypes            []string // If provided, it will fiter reqtype
	Directions          []string // If provided, it will fiter direction
	Tenants             []string // If provided, it will filter tenant
	Categories          []string // If provided, it will filter çategory
	Accounts            []string // If provided, it will filter account
	Subjects            []string // If provided, it will filter the rating subject
	DestinationPrefixes []string // If provided, it will filter on destination prefix
	OrderIdStart        *int64   // Export from this order identifier
	OrderIdEnd          *int64   // Export smaller than this order identifier
	TimeStart           string   // If provided, it will represent the starting of the CDRs interval (>=)
	TimeEnd             string   // If provided, it will represent the end of the CDRs interval (<)
	RerateErrors        bool     // Rerate previous CDRs with errors (makes sense for reqtype rated and pseudoprepaid
	RerateRated         bool     // Rerate CDRs which were previously rated (makes sense for reqtype rated and pseudoprepaid)
	SendToStats         bool     // Set to true if the CDRs should be sent to stats server
}

func (attrRateCDRs *AttrRateCdrs) AsCDRsFilter(timezone string) (*CDRsFilter, error) {
	cdrFltr := &CDRsFilter{
		UniqueIDs:           attrRateCDRs.UniqueIDs,
		RunIDs:              attrRateCDRs.MediationRunIds,
		OriginHosts:         attrRateCDRs.CdrHosts,
		Sources:             attrRateCDRs.CdrSources,
		ToRs:                attrRateCDRs.TORs,
		RequestTypes:        attrRateCDRs.ReqTypes,
		Directions:          attrRateCDRs.Directions,
		Tenants:             attrRateCDRs.Tenants,
		Categories:          attrRateCDRs.Categories,
		Accounts:            attrRateCDRs.Accounts,
		Subjects:            attrRateCDRs.Subjects,
		DestinationPrefixes: attrRateCDRs.DestinationPrefixes,
		OrderIDStart:        attrRateCDRs.OrderIdStart,
		OrderIDEnd:          attrRateCDRs.OrderIdEnd,
	}
	if aTime, err := ParseTimeDetectLayout(attrRateCDRs.TimeStart, timezone); err != nil {
		return nil, err
	} else if !aTime.IsZero() {
		cdrFltr.AnswerTimeStart = &aTime
	}
	if aTimeEnd, err := ParseTimeDetectLayout(attrRateCDRs.TimeEnd, timezone); err != nil {
		return nil, err
	} else if !aTimeEnd.IsZero() {
		cdrFltr.AnswerTimeEnd = &aTimeEnd
	}
	if attrRateCDRs.RerateErrors {
		cdrFltr.MinCost = Float64Pointer(-1.0)
		if !attrRateCDRs.RerateRated {
			cdrFltr.MaxCost = Float64Pointer(0.0)
		}
	} else if attrRateCDRs.RerateRated {
		cdrFltr.MinCost = Float64Pointer(0.0)
	}
	return cdrFltr, nil
}

type AttrRemCdrs struct {
	UniqueIDs []string // List of UniqueIDs to remove from storeDb
}

type AttrGetSMASessions struct {
	SessionManagerIndex int // Index of the session manager queried, defaults to first in the list
}
