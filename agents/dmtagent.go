package agents

import (
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
	"go.uber.org/zap"
)

func NewDiameterAgent(cfg *config.Config, smg rpcclient.RpcClientConnection, pubsubs rpcclient.RpcClientConnection) (*DiameterAgent, error) {
	da := &DiameterAgent{cfg: cfg, smg: smg, pubsubs: pubsubs, connMux: new(sync.Mutex)}
	if reflect.ValueOf(da.pubsubs).IsNil() {
		da.pubsubs = nil // Empty it so we can check it later
	}
	dictsDir := *cfg.DiameterAgent.DictionariesDir
	if len(dictsDir) != 0 {
		if err := loadDictionaries(dictsDir, "DiameterAgent"); err != nil {
			return nil, err
		}
	}
	return da, nil
}

type DiameterAgent struct {
	cfg     *config.Config
	smg     rpcclient.RpcClientConnection // Connection towards CGR-SMG component
	pubsubs rpcclient.RpcClientConnection // Connection towards CGR-PubSub component
	connMux *sync.Mutex                   // Protect connection for read/write
}

// Creates the message handlers
func (self *DiameterAgent) handlers() diam.Handler {
	settings := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(*self.cfg.DiameterAgent.OriginHost),
		OriginRealm:      datatype.DiameterIdentity(*self.cfg.DiameterAgent.OriginRealm),
		VendorID:         datatype.Unsigned32(*self.cfg.DiameterAgent.VendorID),
		ProductName:      datatype.UTF8String(*self.cfg.DiameterAgent.ProductName),
		FirmwareRevision: datatype.Unsigned32(utils.DIAMETER_FIRMWARE_REVISION),
	}
	dSM := sm.New(settings)
	dSM.HandleFunc("CCR", self.handleCCR)
	dSM.HandleFunc("ALL", self.handleALL)
	go func() {
		for err := range dSM.ErrorReports() {
			utils.Logger.Error("<DiameterAgent> StateMachine error:", zap.Stringer("error", err))
		}
	}()
	return dSM
}

func (self DiameterAgent) processCCR(ccr *CCR, reqProcessor *config.RequestProcessor, processorVars map[string]string, cca *CCA) (bool, error) {
	passesAllFilters := true
	for _, fldFilter := range reqProcessor.RequestFilter {
		if passes, _ := passesFieldFilter(ccr.diamMessage, fldFilter, nil); !passes {
			passesAllFilters = false
			break
		}
	}
	if !passesAllFilters { // Not going with this processor further
		return false, nil
	}
	if *reqProcessor.DryRun { // DryRun should log the matching processor as well as the received CCR
		utils.Logger.Info("<DiameterAgent> RequestProcessor:", zap.String("id", *reqProcessor.ID))
		utils.Logger.Info("<DiameterAgent> CCR message:", zap.Stringer("message", ccr.diamMessage))
	}
	if !*reqProcessor.AppendCca {
		*cca = *NewBareCCAFromCCR(ccr, *self.cfg.DiameterAgent.OriginHost, *self.cfg.DiameterAgent.OriginRealm)
	}
	smgEv, err := ccr.AsSMGenericEvent(reqProcessor.CcrFields)
	if err != nil {
		*cca = *NewBareCCAFromCCR(ccr, *self.cfg.DiameterAgent.OriginHost, *self.cfg.DiameterAgent.OriginRealm)
		if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
			false, *self.cfg.DiameterAgent.Timezone); err != nil {
			return false, err
		}
		utils.Logger.Error("<DiameterAgent> Processing", zap.Stringer("message", ccr.diamMessage), zap.Error(err))
		return false, ErrDiameterRatingFailed
	}
	if len(reqProcessor.Flags) != 0 {
		smgEv[utils.CGRFlags] = strings.Join(reqProcessor.Flags, utils.INFIELD_SEP) // Populate CGRFlags automatically
	}
	if *reqProcessor.PublishEvent && self.pubsubs != nil {
		evt, err := smgEv.AsMapStringString()
		if err != nil {
			*cca = *NewBareCCAFromCCR(ccr, *self.cfg.DiameterAgent.OriginHost, *self.cfg.DiameterAgent.OriginRealm)
			if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
				false, *self.cfg.DiameterAgent.Timezone); err != nil {
				return false, err
			}
			utils.Logger.Error("<DiameterAgent> Processing message, failed converting SMGEvent to pubsub one", zap.Stringer("message", ccr.diamMessage), zap.Error(err))
			return false, ErrDiameterRatingFailed
		}
		var reply string
		if err := self.pubsubs.Call("PubSubV1.Publish", engine.CgrEvent(evt), &reply); err != nil {
			*cca = *NewBareCCAFromCCR(ccr, *self.cfg.DiameterAgent.OriginHost, *self.cfg.DiameterAgent.OriginRealm)
			if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
				false, *self.cfg.DiameterAgent.Timezone); err != nil {
				return false, err
			}
			utils.Logger.Error("<DiameterAgent> Processing message, failed publishing event", zap.Stringer("message", ccr.diamMessage), zap.Error(err))
			return false, ErrDiameterRatingFailed
		}
	}
	var maxUsage float64
	processorVars[CGRResultCode] = strconv.Itoa(diam.Success)
	processorVars[CGRError] = ""
	if *reqProcessor.DryRun { // DryRun does not send over network
		utils.Logger.Info("<DiameterAgent> SMGenericEvent", zap.Any("event", smgEv))
		processorVars[CGRResultCode] = strconv.Itoa(diam.LimitedSuccess)
	} else { // Find out maxUsage over APIs
		switch ccr.CCRequestType {
		case 1:
			err = self.smg.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage)
		case 2:
			err = self.smg.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage)
		case 3, 4: // Handle them together since we generate CDR for them
			var rpl string
			if ccr.CCRequestType == 3 {
				err = self.smg.Call("SMGenericV1.TerminateSession", smgEv, &rpl)
			} else if ccr.CCRequestType == 4 {
				err = self.smg.Call("SMGenericV1.ChargeEvent", smgEv.Clone(), &maxUsage)
				if maxUsage == 0 {
					smgEv[utils.USAGE] = 0 // For CDR not to debit
				}
			}
			if *self.cfg.DiameterAgent.CreateCdr &&
				(!*self.cfg.DiameterAgent.CdrRequiresSession || err == nil || !strings.HasSuffix(err.Error(), utils.ErrNoActiveSession.Error())) { // Check if CDR requires session
				if errCdr := self.smg.Call("SMGenericV1.ProcessCDR", smgEv, &rpl); errCdr != nil {
					err = errCdr
				}
			}
		}
		if err != nil {
			utils.Logger.Error("<DiameterAgent> Processing", zap.Stringer("message", ccr.diamMessage), zap.Error(err))
			switch { // Prettify some errors
			case strings.HasSuffix(err.Error(), utils.ErrAccountNotFound.Error()):
				processorVars[CGRError] = utils.ErrAccountNotFound.Error()
			case strings.HasSuffix(err.Error(), utils.ErrUserNotFound.Error()):
				processorVars[CGRError] = utils.ErrUserNotFound.Error()
			case strings.HasSuffix(err.Error(), utils.ErrInsufficientCredit.Error()):
				processorVars[CGRError] = utils.ErrInsufficientCredit.Error()
			case strings.HasSuffix(err.Error(), utils.ErrAccountDisabled.Error()):
				processorVars[CGRError] = utils.ErrAccountDisabled.Error()
			case strings.HasSuffix(err.Error(), utils.ErrRatingPlanNotFound.Error()):
				processorVars[CGRError] = utils.ErrRatingPlanNotFound.Error()
			case strings.HasSuffix(err.Error(), utils.ErrUnauthorizedDestination.Error()):
				processorVars[CGRError] = utils.ErrUnauthorizedDestination.Error()
			default: // Unknown error
				processorVars[CGRError] = err.Error()
				processorVars[CGRResultCode] = strconv.Itoa(DiameterRatingFailed)
			}
		}
		if maxUsage < 0 {
			maxUsage = 0
		}
		if prevMaxUsageStr, hasKey := processorVars[CGRMaxUsage]; hasKey {
			prevMaxUsage, _ := strconv.ParseFloat(prevMaxUsageStr, 64)
			if prevMaxUsage < maxUsage {
				maxUsage = prevMaxUsage
			}
		}
		processorVars[CGRMaxUsage] = strconv.FormatFloat(maxUsage, 'f', -1, 64)
	}
	if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, processorVars[CGRResultCode],
		false, *self.cfg.DiameterAgent.Timezone); err != nil {
		return false, err
	}
	if err := cca.SetProcessorAVPs(reqProcessor, processorVars); err != nil {
		if err := messageSetAVPsWithPath(cca.diamMessage, []interface{}{"Result-Code"}, strconv.Itoa(DiameterRatingFailed),
			false, *self.cfg.DiameterAgent.Timezone); err != nil {
			return false, err
		}
		utils.Logger.Error("<DiameterAgent> CCA SetProcessorAVPs", zap.Stringer("message", ccr.diamMessage), zap.Error(err))
		return false, ErrDiameterRatingFailed
	}
	return true, nil
}

func (self *DiameterAgent) handlerCCR(c diam.Conn, m *diam.Message) {
	ccr, err := NewCCRFromDiameterMessage(m, self.cfg.DiameterAgent.DebitInterval.D())
	if err != nil {
		utils.Logger.Error("<DiameterAgent> Unmarshaling", zap.Stringer("message", m), zap.Error(err))
		return
	}
	cca := NewBareCCAFromCCR(ccr, *self.cfg.DiameterAgent.OriginHost, *self.cfg.DiameterAgent.OriginRealm)
	var processed, lclProcessed bool
	processorVars := make(map[string]string) // Shared between processors
	for _, reqProcessor := range self.cfg.DiameterAgent.RequestProcessors {
		lclProcessed, err = self.processCCR(ccr, reqProcessor, processorVars, cca)
		if lclProcessed { // Process local so we don't overwrite globally
			processed = lclProcessed
		}
		if err != nil || (lclProcessed && !*reqProcessor.ContinueOnSuccess) {
			break
		}
	}
	if err != nil && err != ErrDiameterRatingFailed {
		utils.Logger.Error("<DiameterAgent> CCA SetProcessorAVPs", zap.Stringer("message", ccr.diamMessage), zap.Error(err))
		return
	} else if !processed {
		utils.Logger.Error("<DiameterAgent> No request processor enabled for CCR, ignoring request", zap.Stringer("message", ccr.diamMessage))
		return
	}
	self.connMux.Lock()
	defer self.connMux.Unlock()
	if _, err := cca.AsDiameterMessage().WriteTo(c); err != nil {
		utils.Logger.Error("<DiameterAgent> Failed to write message", zap.Stringer("address", c.RemoteAddr()), zap.Error(err), zap.Stringer("message", cca.AsDiameterMessage()))
		return
	}
}

// Simply dispatch the handling in goroutines
// Could be futher improved with rate control
func (self *DiameterAgent) handleCCR(c diam.Conn, m *diam.Message) {
	go self.handlerCCR(c, m)
}

func (self *DiameterAgent) handleALL(c diam.Conn, m *diam.Message) {
	utils.Logger.Warn("<DiameterAgent> Received unexpected message from", zap.Stringer("address", c.RemoteAddr()), zap.Stringer("message", m))
}

func (self *DiameterAgent) ListenAndServe() error {
	return diam.ListenAndServe(*self.cfg.DiameterAgent.Listen, self.handlers(), nil)
}
