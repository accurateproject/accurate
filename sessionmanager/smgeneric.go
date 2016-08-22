package sessionmanager

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"github.com/cenkalti/rpc2"
	"github.com/cgrates/rpcclient"
)

var ErrPartiallyExecuted = errors.New("Partially executed")

func NewSMGeneric(cgrCfg *config.CGRConfig, rater rpcclient.RpcClientConnection, cdrsrv rpcclient.RpcClientConnection, timezone string, extconns *SMGExternalConnections) *SMGeneric {

	gsm := &SMGeneric{cgrCfg: cgrCfg, rater: rater, cdrsrv: cdrsrv, extconns: extconns, timezone: timezone,
		sessions: make(map[string][]*SMGSession), sessionTerminators: make(map[string]*smgSessionTerminator), sessionsMux: new(sync.RWMutex), guard: engine.Guardian}
	return gsm
}

type SMGeneric struct {
	cgrCfg             *config.CGRConfig // Separate from smCfg since there can be multiple
	rater              rpcclient.RpcClientConnection
	cdrsrv             rpcclient.RpcClientConnection
	timezone           string
	sessions           map[string][]*SMGSession         //Group sessions per sessionId, multiple runs based on derived charging
	sessionTerminators map[string]*smgSessionTerminator // terminate and cleanup the session if timer expires
	extconns           *SMGExternalConnections          // Reference towards external connections manager
	sessionsMux        *sync.RWMutex                    // Locks sessions map
	guard              *engine.GuardianLock             // Used to lock on uuid
}
type smgSessionTerminator struct {
	timer       *time.Timer
	endChan     chan bool
	ttl         time.Duration
	ttlLastUsed *time.Duration
	ttlUsage    *time.Duration
}

// Updates the timer for the session to a new ttl and terminate info
func (self *SMGeneric) resetTerminatorTimer(uuid string, ttl time.Duration, ttlLastUsed, ttlUsage *time.Duration) {
	self.sessionsMux.RLock()
	defer self.sessionsMux.RUnlock()
	if st, found := self.sessionTerminators[uuid]; found {
		if ttl != 0 {
			st.ttl = ttl
		}
		if ttlLastUsed != nil {
			st.ttlLastUsed = ttlLastUsed
		}
		if ttlUsage != nil {
			st.ttlUsage = ttlUsage
		}
		st.timer.Reset(st.ttl)
	}
}

// Called when a session timeouts
func (self *SMGeneric) ttlTerminate(s *SMGSession, tmtr *smgSessionTerminator) {
	debitUsage := tmtr.ttl
	if tmtr.ttlUsage != nil {
		debitUsage = *tmtr.ttlUsage
	}
	for _, s := range self.getSession(s.eventStart.GetUUID()) {
		s.debit(debitUsage, tmtr.ttlLastUsed)
	}
	self.sessionEnd(s.eventStart.GetUUID(), s.TotalUsage())
	cdr := s.eventStart.AsStoredCdr(self.cgrCfg, self.timezone)
	cdr.Usage = s.TotalUsage()
	var reply string
	self.cdrsrv.Call("CdrsV1.ProcessCDR", cdr, &reply)
}

func (self *SMGeneric) indexSession(uuid string, s *SMGSession) {
	self.sessionsMux.Lock()
	self.sessions[uuid] = append(self.sessions[uuid], s)
	if self.cgrCfg.SmGenericConfig.SessionTTL != 0 {
		if _, found := self.sessionTerminators[uuid]; !found {
			ttl := self.cgrCfg.SmGenericConfig.SessionTTL
			if ttlEv := s.eventStart.GetSessionTTL(); ttlEv != 0 {
				ttl = ttlEv
			}
			timer := time.NewTimer(ttl)
			endChan := make(chan bool, 1)
			terminator := &smgSessionTerminator{
				timer:       timer,
				endChan:     endChan,
				ttl:         ttl,
				ttlLastUsed: s.eventStart.GetSessionTTLLastUsed(),
				ttlUsage:    s.eventStart.GetSessionTTLUsage(),
			}
			self.sessionTerminators[uuid] = terminator
			go func() {
				select {
				case <-timer.C:
					self.ttlTerminate(s, terminator)
				case <-endChan:
					timer.Stop()
				}
			}()

		}
	}
	self.sessionsMux.Unlock()
}

// Remove session from session list, removes all related in case of multiple runs, true if item was found
func (self *SMGeneric) unindexSession(uuid string) bool {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	if _, found := self.sessions[uuid]; !found {
		return false
	}
	delete(self.sessions, uuid)
	if st, found := self.sessionTerminators[uuid]; found {
		st.endChan <- true
		delete(self.sessionTerminators, uuid)
	}
	return true
}

// Returns all sessions handled by the SM
func (self *SMGeneric) getSessions() map[string][]*SMGSession {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	return self.sessions
}

func (self *SMGeneric) getSessionIDsForPrefix(prefix string) []string {
	self.sessionsMux.Lock()
	defer self.sessionsMux.Unlock()
	sessionIDs := make([]string, 0)
	for sessionID := range self.sessions {
		if strings.HasPrefix(sessionID, prefix) {
			sessionIDs = append(sessionIDs, sessionID)
		}
	}
	return sessionIDs
}

// Returns sessions/derived for a specific uuid
func (self *SMGeneric) getSession(uuid string) []*SMGSession {
	self.sessionsMux.RLock()
	defer self.sessionsMux.RUnlock()
	return self.sessions[uuid]
}

// Handle a new session, pass the connectionId so we can communicate on disconnect request
func (self *SMGeneric) sessionStart(evStart SMGenericEvent, connId string) error {
	sessionId := evStart.GetUUID()
	processed, err := self.guard.Guard(func() (interface{}, error) { // Lock it on UUID level
		var sessionRuns []*engine.SessionRun
		if err := self.rater.Call("Responder.GetSessionRuns", evStart.AsStoredCdr(self.cgrCfg, self.timezone), &sessionRuns); err != nil {
			return true, err
		} else if len(sessionRuns) == 0 {
			return true, nil
		}
		stopDebitChan := make(chan struct{})
		for _, sessionRun := range sessionRuns {
			s := &SMGSession{eventStart: evStart, connId: connId, runId: sessionRun.DerivedCharger.RunID, timezone: self.timezone,
				rater: self.rater, cdrsrv: self.cdrsrv, cd: sessionRun.CallDescriptor}
			self.indexSession(sessionId, s)
			//utils.Logger.Info(fmt.Sprintf("<SMGeneric> Starting session: %s, runId: %s", sessionId, s.runId))
			if self.cgrCfg.SmGenericConfig.DebitInterval != 0 {
				s.stopDebit = stopDebitChan
				go s.debitLoop(self.cgrCfg.SmGenericConfig.DebitInterval)
			}
		}
		return true, nil
	}, self.cgrCfg.LockingTimeout, sessionId)
	if processed == nil || processed == false {
		utils.Logger.Err("<SMGeneric> Cannot start session, empty reply")
		return utils.ErrServerError
	}
	return err
}

// End a session from outside
func (self *SMGeneric) sessionEnd(sessionId string, usage time.Duration) error {
	_, err := self.guard.Guard(func() (interface{}, error) { // Lock it on UUID level
		ss := self.getSession(sessionId)
		if len(ss) == 0 { // Not handled by us
			return nil, nil
		}
		if !self.unindexSession(sessionId) { // Unreference it early so we avoid concurrency
			return nil, nil // Did not find the session so no need to close it anymore
		}
		for idx, s := range ss {
			s.totalUsage = usage // save final usage as totalUsage
			//utils.Logger.Info(fmt.Sprintf("<SMGeneric> Ending session: %s, runId: %s", sessionId, s.runId))
			if idx == 0 && s.stopDebit != nil {
				close(s.stopDebit) // Stop automatic debits
			}
			aTime, err := s.eventStart.GetAnswerTime(utils.META_DEFAULT, self.cgrCfg.DefaultTimezone)
			if err != nil || aTime.IsZero() {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not retrieve answer time for session: %s, runId: %s, aTime: %+v, error: %s",
					sessionId, s.runId, aTime, err.Error()))
			}
			if err := s.close(aTime.Add(usage)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not close session: %s, runId: %s, error: %s", sessionId, s.runId, err.Error()))
			}
			if err := s.saveOperations(sessionId); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not save session: %s, runId: %s, error: %s", sessionId, s.runId, err.Error()))
			}
		}
		return nil, nil
	}, time.Duration(2)*time.Second, sessionId)
	return err
}

// Used when an update will relocate an initial session (eg multiple data streams)
func (self *SMGeneric) sessionRelocate(sessionID, initialID string) error {
	_, err := self.guard.Guard(func() (interface{}, error) { // Lock it on initialID level
		if utils.IsSliceMember([]string{sessionID, initialID}, "") { // Not allowed empty params here
			return nil, utils.ErrMandatoryIeMissing
		}
		ssNew := self.getSession(sessionID) // Already relocated
		if len(ssNew) != 0 {
			return nil, nil
		}
		ss := self.getSession(initialID)
		if len(ss) == 0 { // No need of relocation
			return nil, utils.ErrNotFound
		}
		for i, s := range ss {
			s.eventStart[utils.ACCID] = sessionID // Overwrite initialSessionID with new one
			self.indexSession(sessionID, s)
			if i == 0 {
				self.unindexSession(initialID)
			}
		}
		return nil, nil
	}, time.Duration(2)*time.Second, initialID)
	return err
}

// Methods to apply on sessions, mostly exported through RPC/Bi-RPC
//Calculates maximum usage allowed for gevent
func (self *SMGeneric) MaxUsage(gev SMGenericEvent, clnt *rpc2.Client) (time.Duration, error) {
	gev[utils.EVENT_NAME] = utils.CGR_AUTHORIZATION
	storedCdr := gev.AsStoredCdr(config.CgrConfig(), self.timezone)
	var maxDur float64
	if err := self.rater.Call("Responder.GetDerivedMaxSessionTime", storedCdr, &maxDur); err != nil {
		return time.Duration(0), err
	}
	return time.Duration(maxDur), nil
}

func (self *SMGeneric) LCRSuppliers(gev SMGenericEvent, clnt *rpc2.Client) ([]string, error) {
	gev[utils.EVENT_NAME] = utils.CGR_LCR_REQUEST
	cd, err := gev.AsLcrRequest().AsCallDescriptor(self.timezone)
	cd.CgrID = gev.GetCgrId(self.timezone)
	if err != nil {
		return nil, err
	}
	var lcr engine.LCRCost
	if err = self.rater.Call("Responder.GetLCR", &engine.AttrGetLcr{CallDescriptor: cd}, &lcr); err != nil {
		return nil, err
	}
	if lcr.HasErrors() {
		lcr.LogErrors()
		return nil, errors.New("LCR_COMPUTE_ERROR")
	}
	return lcr.SuppliersSlice()
}

// Called on session start
func (self *SMGeneric) InitiateSession(gev SMGenericEvent, clnt *rpc2.Client) (time.Duration, error) {
	if err := self.sessionStart(gev, getClientConnId(clnt)); err != nil {
		self.sessionEnd(gev.GetUUID(), 0)
		return nilDuration, err
	}
	d, err := self.UpdateSession(gev, clnt)
	if err != nil || d == 0 {
		self.sessionEnd(gev.GetUUID(), 0)
	}
	return d, err
}

// Execute debits for usage/maxUsage
func (self *SMGeneric) UpdateSession(gev SMGenericEvent, clnt *rpc2.Client) (time.Duration, error) {
	if initialID, err := gev.GetFieldAsString(utils.InitialOriginID); err == nil {
		err := self.sessionRelocate(gev.GetUUID(), initialID)
		if err == utils.ErrNotFound { // Session was already relocated, create a new  session with this update
			err = self.sessionStart(gev, getClientConnId(clnt))
		}
		if err != nil {
			return nilDuration, err
		}
	}
	self.resetTerminatorTimer(gev.GetUUID(), gev.GetSessionTTL(), gev.GetSessionTTLLastUsed(), gev.GetSessionTTLUsage())
	var lastUsed *time.Duration
	evLastUsed, err := gev.GetLastUsed(utils.META_DEFAULT)
	if err != nil && err != utils.ErrNotFound {
		return nilDuration, err
	}
	if err == nil {
		lastUsed = &evLastUsed
	}
	evMaxUsage, err := gev.GetMaxUsage(utils.META_DEFAULT, self.cgrCfg.MaxCallDuration)
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrMandatoryIeMissing
		}
		return nilDuration, err
	}
	aSessions := self.getSession(gev.GetUUID())
	if len(aSessions) == 0 {
		utils.Logger.Err(fmt.Sprintf("<SMGeneric> SessionUpdate with no active sessions for event: <%s>", gev.GetUUID()))
		return nilDuration, utils.ErrServerError
	}
	for _, s := range aSessions {
		if maxDur, err := s.debit(evMaxUsage, lastUsed); err != nil {
			return nilDuration, err
		} else if maxDur < evMaxUsage {
			evMaxUsage = maxDur
		}
	}
	return evMaxUsage, nil
}

// Called on session end, should stop debit loop
func (self *SMGeneric) TerminateSession(gev SMGenericEvent, clnt *rpc2.Client) error {
	if initialID, err := gev.GetFieldAsString(utils.InitialOriginID); err == nil {
		err := self.sessionRelocate(gev.GetUUID(), initialID)
		if err == utils.ErrNotFound { // Session was already relocated, create a new  session with this update
			err = self.sessionStart(gev, getClientConnId(clnt))
		}
		if err != nil && err != utils.ErrMandatoryIeMissing {
			return err
		}
	}
	sessionIDs := []string{gev.GetUUID()}
	if sessionIDPrefix, err := gev.GetFieldAsString(utils.OriginIDPrefix); err == nil { // OriginIDPrefix is present, OriginID will not be anymore considered
		sessionIDs = self.getSessionIDsForPrefix(sessionIDPrefix)
	}
	usage, errUsage := gev.GetUsage(utils.META_DEFAULT)
	var lastUsed time.Duration
	if errUsage != nil {
		if errUsage != utils.ErrNotFound {
			return errUsage
		}
		var err error
		lastUsed, err = gev.GetLastUsed(utils.META_DEFAULT)
		if err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrMandatoryIeMissing
			}
			return err
		}
	}
	var interimError error
	for _, sessionID := range sessionIDs {
		if errUsage != nil {
			var s *SMGSession
			for _, s = range self.getSession(sessionID) {
				break
			}
			if s == nil {
				continue // No session active, will not be able to close it anyway
			}
			usage = s.TotalUsage() - s.lastUsage + lastUsed
		}
		if err := self.sessionEnd(sessionID, usage); err != nil {
			interimError = err // Last error will be the one returned as API result
		}
	}
	return interimError
}

// Processes one time events (eg: SMS)
func (self *SMGeneric) ChargeEvent(gev SMGenericEvent, clnt *rpc2.Client) (maxDur time.Duration, err error) {
	var sessionRuns []*engine.SessionRun
	if err := self.rater.Call("Responder.GetSessionRuns", gev.AsStoredCdr(self.cgrCfg, self.timezone), &sessionRuns); err != nil {
		return nilDuration, err
	} else if len(sessionRuns) == 0 {
		return nilDuration, nil
	}
	var maxDurInit bool // Avoid differences between default 0 and received 0
	for _, sR := range sessionRuns {
		cc := new(engine.CallCost)
		if err = self.rater.Call("Responder.MaxDebit", sR.CallDescriptor, cc); err != nil {
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not Debit CD: %+v, RunID: %s, error: %s", sR.CallDescriptor, sR.DerivedCharger.RunID, err.Error()))
			break
		}
		sR.CallCosts = append(sR.CallCosts, cc) // Save it so we can revert on issues
		if ccDur := cc.GetDuration(); ccDur == 0 {
			err = utils.ErrInsufficientCredit
			break
		} else if !maxDurInit || ccDur < maxDur {
			maxDur = ccDur
		}
	}
	if err != nil { // Refund the ones already taken since we have error on one of the debits
		for _, sR := range sessionRuns {
			if len(sR.CallCosts) == 0 {
				continue
			}
			cc := sR.CallCosts[0]
			if len(sR.CallCosts) > 1 {
				for _, ccSR := range sR.CallCosts {
					cc.Merge(ccSR)
				}
			}
			// collect increments
			var refundIncrements engine.Increments
			cc.Timespans.Decompress()
			for _, ts := range cc.Timespans {
				refundIncrements = append(refundIncrements, ts.Increments...)
			}
			// refund cc
			if len(refundIncrements) > 0 {
				cd := cc.CreateCallDescriptor()
				cd.Increments = refundIncrements
				cd.CgrID = sR.CallDescriptor.CgrID
				cd.RunID = sR.CallDescriptor.RunID
				cd.Increments.Compress()
				//utils.Logger.Info(fmt.Sprintf("Refunding session run callcost: %s", utils.ToJSON(cd)))
				var response float64
				err := self.rater.Call("Responder.RefundIncrements", cd, &response)
				if err != nil {
					return nilDuration, err
				}
			}
		}
		return nilDuration, err
	}
	var withErrors bool
	for _, sR := range sessionRuns {
		if len(sR.CallCosts) == 0 {
			continue
		}
		cc := sR.CallCosts[0]
		if len(sR.CallCosts) > 1 {
			for _, ccSR := range sR.CallCosts[1:] {
				cc.Merge(ccSR)
			}
		}
		cc.Round()
		roundIncrements := cc.GetRoundIncrements()
		if len(roundIncrements) != 0 {
			cd := cc.CreateCallDescriptor()
			cd.Increments = roundIncrements
			var response float64
			if err := self.rater.Call("Responder.RefundRounding", cd, &response); err != nil {
				utils.Logger.Err(fmt.Sprintf("<SM> ERROR failed to refund rounding: %v", err))
			}
		}
		var reply string
		smCost := &engine.SMCost{
			CGRID:       gev.GetCgrId(self.timezone),
			CostSource:  utils.SESSION_MANAGER_SOURCE,
			RunID:       sR.DerivedCharger.RunID,
			OriginHost:  gev.GetOriginatorIP(utils.META_DEFAULT),
			OriginID:    gev.GetUUID(),
			CostDetails: cc,
		}
		if err := self.cdrsrv.Call("CdrsV1.StoreSMCost", engine.AttrCDRSStoreSMCost{Cost: smCost, CheckDuplicate: true}, &reply); err != nil && err != utils.ErrExists {
			withErrors = true
			utils.Logger.Err(fmt.Sprintf("<SMGeneric> Could not save CC: %+v, RunID: %s error: %s", cc, sR.DerivedCharger.RunID, err.Error()))
		}
	}
	if withErrors {
		return nilDuration, ErrPartiallyExecuted
	}
	return maxDur, nil
}

func (self *SMGeneric) ProcessCDR(gev SMGenericEvent) error {
	var reply string
	if err := self.cdrsrv.Call("CdrsV1.ProcessCDR", gev.AsStoredCdr(self.cgrCfg, self.timezone), &reply); err != nil {
		return err
	}
	return nil
}

func (self *SMGeneric) Connect() error {
	return nil
}

// Used by APIer to retrieve sessions
func (self *SMGeneric) Sessions() map[string][]*SMGSession {
	return self.getSessions()
}

func (self *SMGeneric) Timezone() string {
	return self.timezone
}

// System shutdown
func (self *SMGeneric) Shutdown() error {
	for ssId := range self.getSessions() { // Force sessions shutdown
		self.sessionEnd(ssId, time.Duration(self.cgrCfg.MaxCallDuration))
	}
	return nil
}
