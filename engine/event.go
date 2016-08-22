package engine

import (
	"time"

	"github.com/accurateproject/accurate/utils"
)

type Event interface {
	GetName() string
	GetCgrId(timezone string) string
	GetUUID() string
	GetSessionIds() []string // Returns identifiers needed to control a session (eg disconnect)
	GetDirection(string) string
	GetSubject(string) string
	GetAccount(string) string
	GetDestination(string) string
	GetCallDestNr(string) string
	GetOriginatorIP(string) string
	GetCategory(string) string
	GetTenant(string) string
	GetReqType(string) string
	GetSetupTime(string, string) (time.Time, error)
	GetAnswerTime(string, string) (time.Time, error)
	GetEndTime(string, string) (time.Time, error)
	GetDuration(string) (time.Duration, error)
	GetPdd(string) (time.Duration, error)
	GetSupplier(string) string
	GetDisconnectCause(string) string
	GetExtraFields() map[string]string
	MissingParameter(string) bool
	ParseEventValue(*utils.RSRField, string) string
	AsStoredCdr(timezone string) *CDR
	String() string
	AsEvent(string) Event
	ComputeLcr() bool
}
