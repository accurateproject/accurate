package v1

import (
	"time"

	"github.com/accurateproject/accurate/sessionmanager"
	"github.com/accurateproject/accurate/utils"
	"github.com/cenkalti/rpc2"
)

func NewSMGenericBiRpcV1(sm *sessionmanager.SMGeneric) *SMGenericBiRpcV1 {
	return &SMGenericBiRpcV1{sm: sm}
}

type SMGenericBiRpcV1 struct {
	sm *sessionmanager.SMGeneric
}

// Publishes methods exported by SMGenericBiRpcV1 as SMGenericV1 (so we can handle standard RPC methods via birpc socket)
func (self *SMGenericBiRpcV1) Handlers() map[string]interface{} {
	return map[string]interface{}{
		"SMGenericV1.MaxUsage":         self.MaxUsage,
		"SMGenericV1.LCRSuppliers":     self.LCRSuppliers,
		"SMGenericV1.InitiateSession":  self.InitiateSession,
		"SMGenericV1.UpdateSession":    self.UpdateSession,
		"SMGenericV1.TerminateSession": self.TerminateSession,
		"SMGenericV1.ProcessCDR":       self.ProcessCDR,
	}
}

/// Returns MaxUsage (for calls in seconds), -1 for no limit
func (self *SMGenericBiRpcV1) MaxUsage(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	maxUsageDur, err := self.sm.MaxUsage(ev, clnt)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if maxUsageDur == time.Duration(-1) {
		*maxUsage = -1.0
	} else {
		*maxUsage = maxUsageDur.Seconds()
	}
	return nil
}

/// Returns list of suppliers which can be used for the request
func (self *SMGenericBiRpcV1) LCRSuppliers(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, suppliers *[]string) error {
	if supls, err := self.sm.LCRSuppliers(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = supls
	}
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericBiRpcV1) InitiateSession(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.InitiateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericBiRpcV1) UpdateSession(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.UpdateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should stop debit loop
func (self *SMGenericBiRpcV1) TerminateSession(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.TerminateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Called on individual Events (eg SMS)
func (self *SMGenericBiRpcV1) ChargeEvent(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.ChargeEvent(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericBiRpcV1) ProcessCDR(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.ProcessCDR(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
