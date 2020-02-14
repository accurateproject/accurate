package sessionmanager

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"go.uber.org/zap"
)

// One session handled by SM
type SMGSession struct {
	eventStart        SMGenericEvent // Event which started
	stopDebit         chan struct{}  // Channel to communicate with debit loops when closing the session
	runId             string         // Keep a reference for the derived run
	timezone          string
	rater             rpcclient.RpcClientConnection // Connector to Rater service
	cdrsrv            rpcclient.RpcClientConnection // Connector to CDRS service
	cd                *engine.CallDescriptor
	sessionCds        []*engine.CallDescriptor
	callCosts         []*engine.CallCost
	extraDuration     time.Duration                 // keeps the current duration debited on top of what heas been asked
	lastUsage         time.Duration                 // last requested Duration
	lastDebit         time.Duration                 // last real debited duration
	totalUsage        time.Duration                 // sum of lastUsage
	clntConn          rpcclient.RpcClientConnection // Reference towards client connection on SMG side so we can disconnect.
	postActionTrigger bool                          // execute triggered actions on debit confirmation
}

// Called in case of automatic debits
func (s *SMGSession) debitLoop(debitInterval time.Duration) {
	loopIndex := 0
	sleepDur := time.Duration(0) // start with empty duration for debit
	for {
		select {
		case <-s.stopDebit:
			return
		case <-time.After(sleepDur):
			if maxDebit, err := s.debit(debitInterval, nil); err != nil {
				utils.Logger.Error("<SMGeneric> Could not complete debit operation on session", zap.String("uuid", s.eventStart.GetUUID()), zap.Error(err))
				disconnectReason := SYSTEM_ERROR
				if err.Error() == utils.ErrUnauthorizedDestination.Error() {
					disconnectReason = err.Error()
				}
				if err := s.disconnectSession(disconnectReason); err != nil {
					utils.Logger.Error("<SMGeneric> Could not disconnect session", zap.String("uuid", s.eventStart.GetUUID()), zap.Error(err))
				}
				return
			} else if maxDebit < debitInterval {
				time.Sleep(maxDebit)
				if err := s.disconnectSession(INSUFFICIENT_FUNDS); err != nil {
					utils.Logger.Error("<SMGeneric> Could not disconnect session", zap.String("uuid", s.eventStart.GetUUID()), zap.Error(err))
				}
				return
			}
			sleepDur = debitInterval
			loopIndex++
		}
	}
}

// Attempts to debit a duration, returns maximum duration which can be debitted or error
func (s *SMGSession) debit(dur time.Duration, lastUsed *time.Duration) (time.Duration, error) {
	s.cd.PostActionTrigger = s.postActionTrigger
	requestedDuration := dur
	if lastUsed != nil { // we have a previous debit
		s.extraDuration = s.lastDebit - *lastUsed
		//utils.Logger.Debug(fmt.Sprintf("ExtraDuration LastUsed: %f", s.extraDuration.Seconds()))
		if *lastUsed != s.lastUsage {
			// total usage correction
			s.totalUsage -= s.lastUsage
			s.totalUsage += *lastUsed
			//utils.Logger.Debug(fmt.Sprintf("TotalUsage Correction: %f", s.totalUsage.Seconds()))
		}
		// fill the action triggers that were used and the ones that are not used
		if len(s.callCosts) > 0 {
			lastCC := s.callCosts[len(s.callCosts)-1]
			s.cd.ExeATIDs, s.cd.UnexeATIDs = lastCC.GetPostActionTriggers(*lastUsed)
		}
	}
	// apply correction from previous run
	if s.extraDuration < dur {
		dur -= s.extraDuration
	} else {
		s.lastUsage = requestedDuration
		s.totalUsage += s.lastUsage
		ccDuration := s.extraDuration // fake ccDuration
		s.lastDebit = ccDuration
		s.extraDuration -= dur
		// send a zero duration debit for the engine to process the action triggers
		if len(s.cd.ExeATIDs) > 0 || len(s.cd.UnexeATIDs) > 0 {
			fakeCD := s.cd.Clone()
			fakeCD.TimeEnd = fakeCD.TimeStart
			cc := &engine.CallCost{}
			if err := s.rater.Call("Responder.Debit", fakeCD, &cc); err != nil || cc.GetDuration() == 0 {
				return 0, err
			}
		}
		return ccDuration, nil
	}
	//utils.Logger.Debug(fmt.Sprintf("dur: %f", dur.Seconds()))
	initialExtraDuration := s.extraDuration
	s.extraDuration = 0
	if s.cd.LoopIndex > 0 {
		s.cd.TimeStart = s.cd.TimeEnd
	}
	s.cd.TimeEnd = s.cd.TimeStart.Add(dur)
	s.cd.DurationIndex += dur
	cc := &engine.CallCost{}
	if err := s.rater.Call("Responder.MaxDebit", s.cd, cc); err != nil {
		s.lastUsage = 0
		s.lastDebit = 0
		return 0, err
	}
	// cd corrections
	s.cd.TimeEnd = cc.GetEndTime() // set debited timeEnd
	// update call duration with real debited duration
	ccDuration := cc.GetDuration()
	//utils.Logger.Debug(fmt.Sprintf("CCDur: %f", ccDuration.Seconds()))
	if ccDuration != dur {
		s.extraDuration = ccDuration - dur
	}
	if ccDuration >= dur {
		s.lastUsage = requestedDuration
	} else {
		s.lastUsage = ccDuration
	}
	s.cd.DurationIndex -= dur
	s.cd.DurationIndex += ccDuration
	s.cd.GetMaxCostSoFar().AddS(cc.GetCost())
	s.cd.LoopIndex++
	s.sessionCds = append(s.sessionCds, s.cd.Clone())
	s.callCosts = append(s.callCosts, cc)
	s.lastDebit = initialExtraDuration + ccDuration
	s.totalUsage += s.lastUsage
	//utils.Logger.Debug(fmt.Sprintf("TotalUsage: %f", s.totalUsage.Seconds()))

	if ccDuration >= dur { // we got what we asked to be debited
		//utils.Logger.Debug(fmt.Sprintf("returning normal: %f", requestedDuration.Seconds()))
		return requestedDuration, nil
	}
	//utils.Logger.Debug(fmt.Sprintf("returning initialExtra: %f + ccDuration: %f", initialExtraDuration.Seconds(), ccDuration.Seconds()))
	return s.lastDebit, nil
}

// Attempts to refund a duration, error on failure
func (s *SMGSession) refund(refundDuration, postATSplitDuration time.Duration) error {
	//initialRefundDuration := refundDuration
	if refundDuration == 0 { // Nothing to refund
		return nil
	}
	firstCC := s.callCosts[0] // use merged cc (from close function)
	lastCC := s.callCosts[len(s.callCosts)-1]
	exe, unexe := lastCC.GetPostActionTriggers(postATSplitDuration)
	refundIncrements := firstCC.TruncateTimespansAtDuration(refundDuration)
	// show only what was actualy refunded (stopped in timespan)
	// utils.Logger.Info(fmt.Sprintf("Refund duration: %v", initialRefundDuration-refundDuration))
	if len(refundIncrements) > 0 {
		cd := firstCC.CreateCallDescriptor()
		cd.Increments = refundIncrements
		cd.UniqueID = s.cd.UniqueID
		cd.RunID = s.cd.RunID
		cd.ExeATIDs, cd.UnexeATIDs = exe, unexe
		//utils.Logger.Info(fmt.Sprintf("Refunding %s duration %v with cd: %s", cd.UniqueID, initialRefundDuration, utils.ToJSON(cd)))
		var response float64
		err := s.rater.Call("Responder.RefundIncrements", cd, &response)
		if err != nil {
			return err
		}
	} else {
		if len(exe) > 0 || len(unexe) > 0 {
			// send a fake debit for execution of the last post action triggers
			fakeCD := firstCC.CreateCallDescriptor()
			fakeCD.ExeATIDs, fakeCD.UnexeATIDs = exe, unexe
			err := s.rater.Call("Responder.Debit", fakeCD, &engine.CallCost{})
			if err != nil {
				return err
			}
		}
	}
	//firstCC.Cost -= refundIncrements.GetTotalCost() // use updateCost instead
	firstCC.UpdateCost()
	firstCC.UpdateRatedUsage()
	firstCC.Timespans.Compress()
	return nil
}

// mergeCCs will merge the CallCosts recorded for this session
func (s *SMGSession) mergeCCs() {
	if len(s.callCosts) != 0 { // at least one cost calculation
		firstCC := s.callCosts[0]
		for _, cc := range s.callCosts[1:] {
			firstCC.Merge(cc)
		}
	}
}

// Session has ended, check debits and refund the extra charged duration
func (s *SMGSession) close(endTime time.Time) (err error) {
	if len(s.callCosts) != 0 { // at least one cost calculation
		firstCC := s.callCosts[0]
		lastCC := s.callCosts[len(s.callCosts)-1]
		chargedEndTime := lastCC.GetEndTime()
		if endTime.After(chargedEndTime) { // we did not charge enough, make a manual debit here
			extraDur := endTime.Sub(chargedEndTime)
			if s.cd.LoopIndex > 0 {
				s.cd.TimeStart = s.cd.TimeEnd
			}
			s.cd.TimeEnd = s.cd.TimeStart.Add(extraDur)
			s.cd.DurationIndex += extraDur
			cc := &engine.CallCost{}
			if err = s.rater.Call("Responder.Debit", s.cd, cc); err == nil {
				s.callCosts = append(s.callCosts, cc)
				s.mergeCCs()
			}
		} else {
			s.mergeCCs()
			end := firstCC.GetEndTime()
			refundDuration := end.Sub(endTime)
			postActionTriggersSplitDuration := lastCC.GetDuration() - refundDuration
			err = s.refund(refundDuration, postActionTriggersSplitDuration)
		}
	}
	return
}

// Send disconnect order to remote connection
func (s *SMGSession) disconnectSession(reason string) error {
	if s.clntConn == nil || reflect.ValueOf(s.clntConn).IsNil() {
		return errors.New("Calling SMGClientV1.DisconnectSession requires bidirectional JSON connection")
	}
	var reply string
	if err := s.clntConn.Call("SMGClientV1.DisconnectSession", utils.AttrDisconnectSession{EventStart: s.eventStart, Reason: reason}, &reply); err != nil {
		return err
	} else if reply != utils.OK {
		return errors.New(fmt.Sprintf("Unexpected disconnect reply: %s", reply))
	}
	return nil
}

// Merge the sum of costs and sends it to CDRS for storage
// originID could have been changed from original event, hence passing as argument here
func (s *SMGSession) saveOperations(originID string) error {
	if len(s.callCosts) == 0 {
		return nil // There are no costs to save, ignore the operation
	}
	firstCC := s.callCosts[0] // was merged in close method
	smCost := &engine.SMCost{
		UniqueID:    s.eventStart.GetUniqueID(s.timezone),
		CostSource:  utils.SESSION_MANAGER_SOURCE,
		RunID:       s.runId,
		OriginHost:  s.eventStart.GetOriginatorIP(utils.META_DEFAULT),
		OriginID:    originID,
		Usage:       s.TotalUsage().Seconds(),
		CostDetails: firstCC,
	}
	var reply string
	if err := s.cdrsrv.Call("CdrsV1.StoreSMCost", engine.AttrCDRSStoreSMCost{Cost: smCost, CheckDuplicate: true}, &reply); err != nil {
		if err == utils.ErrExists {
			s.refund(s.cd.GetDuration(), 0) // Refund entire duration
		} else {
			return err
		}
	}
	return nil
}

func (s *SMGSession) TotalUsage() time.Duration {
	return s.totalUsage
}

func (s *SMGSession) AsActiveSession(timezone string) *ActiveSession {
	sTime, _ := s.eventStart.GetSetupTime(utils.META_DEFAULT, timezone)
	aTime, _ := s.eventStart.GetAnswerTime(utils.META_DEFAULT, timezone)
	pdd, _ := s.eventStart.GetPdd(utils.META_DEFAULT)
	aSession := &ActiveSession{
		UniqueID:    s.eventStart.GetUniqueID(timezone),
		TOR:         s.eventStart.GetTOR(utils.META_DEFAULT),
		RunId:       s.runId,
		OriginID:    s.eventStart.GetUUID(),
		CdrHost:     s.eventStart.GetOriginatorIP(utils.META_DEFAULT),
		CdrSource:   s.eventStart.GetCdrSource(),
		ReqType:     s.eventStart.GetReqType(utils.META_DEFAULT),
		Direction:   s.eventStart.GetDirection(utils.META_DEFAULT),
		Tenant:      s.eventStart.GetTenant(utils.META_DEFAULT),
		Category:    s.eventStart.GetCategory(utils.META_DEFAULT),
		Account:     s.eventStart.GetAccount(utils.META_DEFAULT),
		Subject:     s.eventStart.GetSubject(utils.META_DEFAULT),
		Destination: s.eventStart.GetDestination(utils.META_DEFAULT),
		SetupTime:   sTime,
		AnswerTime:  aTime,
		Usage:       s.TotalUsage(),
		Pdd:         pdd,
		ExtraFields: s.eventStart.GetExtraFields(),
		Supplier:    s.eventStart.GetSupplier(utils.META_DEFAULT),
		SMId:        "CGR-DA",
	}
	if s.cd != nil {
		aSession.LoopIndex = s.cd.LoopIndex
		aSession.DurationIndex = s.cd.DurationIndex
		aSession.MaxRate = s.cd.MaxRate
		aSession.MaxRateUnit = s.cd.MaxRateUnit
		aSession.MaxCostSoFar = s.cd.MaxCostSoFar
	}
	return aSession
}
