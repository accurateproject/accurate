
package engine

// RawCDR is the original CDR received from external sources (eg: FreeSWITCH)
type RawCdr interface {
	AsStoredCdr(string) *CDR // Convert the inbound Cdr into internally used one, CgrCdr
}
