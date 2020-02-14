package v1

import (
	"reflect"
	"strings"
	"time"

	"github.com/accurateproject/accurate/sessionmanager"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
)

func NewSMGenericV1(sm *sessionmanager.SMGeneric) *SMGenericV1 {
	return &SMGenericV1{sm: sm}
}

// Exports RPC from SMGeneric
type SMGenericV1 struct {
	sm *sessionmanager.SMGeneric
}

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (self *SMGenericV1) MaxUsage(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	maxUsageDur, err := self.sm.MaxUsage(ev)
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

// Returns list of suppliers which can be used for the request
func (self *SMGenericV1) LCRSuppliers(ev sessionmanager.SMGenericEvent, suppliers *[]string) error {
	if supls, err := self.sm.LCRSuppliers(ev); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = supls
	}
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericV1) InitiateSession(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.InitiateSession(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericV1) UpdateSession(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.UpdateSession(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should stop debit loop
func (self *SMGenericV1) TerminateSession(ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.TerminateSession(ev, nil); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Called on individual Events (eg SMS)
func (self *SMGenericV1) ChargeEvent(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.ChargeEvent(ev); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericV1) ProcessCDR(ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.ProcessCDR(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

func (self *SMGenericV1) ActiveSessions(attrs map[string]string, reply *[]*sessionmanager.ActiveSession) error {
	aSessions, _, err := self.sm.ActiveSessions(attrs, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = aSessions
	return nil
}

func (self *SMGenericV1) ActiveSessionsCount(attrs map[string]string, reply *int) error {
	if _, count, err := self.sm.ActiveSessions(attrs, true); err != nil {
		return err
	} else {
		*reply = count
	}
	return nil
}

// rpcclient.RpcClientConnection interface
func (self *SMGenericV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(serviceMethod, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	method := reflect.ValueOf(self).MethodByName(methodSplit[1])
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
