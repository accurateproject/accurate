package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
	"github.com/accurateproject/rpcclient"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type ActionGroup struct {
	Tenant  string  `bson:"tenant"`
	Name    string  `bson:"name"`
	Actions Actions `bson:"actions"`
}

// SetParentGroup populates parent to all actions
func (ag *ActionGroup) SetParentGroup() {
	for _, a := range ag.Actions {
		a.parentGroup = ag
	}
}

/*
Action is data to be filled for each tariff plan with the bonus value for received calls minutes.
*/
type Action struct {
	ActionType   string   `bson:"action_type"`
	TOR          string   `bson:"tor"`
	Params       string   `bson:"params"`
	ExecFilter   string   `bson:"exec_filter"`
	Filter1      string   `bson:"filter"`
	Weight       float64  `bson:"weight"`
	balanceValue *dec.Dec // balance value after action execution, used with cdrlog
	parentGroup  *ActionGroup
	filter       *utils.StructQ
	balance      *Balance
}

func (a *Action) getBalanceValue() *dec.Dec {
	if a.balanceValue == nil {
		a.balanceValue = dec.New()
	}
	return a.balanceValue
}

func (a *Action) getFilter() *utils.StructQ {
	if a == nil {
		return nil
	}
	if a.filter != nil {
		return a.filter
	}
	// ignore error as its hould be checked at load time
	a.filter, _ = utils.NewStructQ(a.Filter1)
	return a.filter
}

func (a *Action) getBalance(b *Balance) (*Balance, error) {
	if b == nil && a.balance != nil {
		return a.balance, nil
	}
	if a.Params == "" || !strings.Contains(a.Params, "Balance") {
		if b != nil {
			return b, nil
		}
		return &Balance{}, nil
	}
	param := struct {
		Balance *Balance
	}{}
	if b != nil {
		param.Balance = b
	}
	if err := json.Unmarshal([]byte(a.Params), &param); err != nil {
		//log.Print("err: ", err)
		return nil, err
	}
	a.balance = param.Balance
	// load action timings from tags
	if strings.Contains(a.Params, "TimingTags") {
		var x struct {
			Balance struct {
				TimingTags []string
			}
		}
		if err := json.Unmarshal([]byte(a.Params), &x); err != nil {
			return nil, err
		}
		if a.balance.TimingIDs == nil {
			a.balance.TimingIDs = utils.StringMap{}
		}
		for _, timingID := range x.Balance.TimingTags {
			if a.parentGroup == nil {
				return nil, fmt.Errorf("no parentgroup on: %s", utils.ToJSON(a))
			}
			timing, err := ratingStorage.GetTiming(a.parentGroup.Tenant, timingID)
			if err != nil {
				return nil, fmt.Errorf("error getting action timing id %s", timingID)
			}
			if timing != nil {
				a.balance.TimingIDs[timingID] = true
				a.balance.Timings = append(a.balance.Timings, &RITiming{
					Years:     timing.Years,
					Months:    timing.Months,
					MonthDays: timing.MonthDays,
					WeekDays:  timing.WeekDays,
					StartTime: timing.Time,
				})
			} else {
				return nil, fmt.Errorf("could not find timing: %v", timingID)
			}
		}
	}

	// load ExpiryTime
	if strings.Contains(a.Params, "ExpiryTime") {
		var x struct {
			Balance struct {
				ExpiryTime string
			}
		}
		if err := json.Unmarshal([]byte(a.Params), &x); err != nil {
			return nil, err
		}
		if expDate, err := utils.ParseDate(x.Balance.ExpiryTime); err != nil {
			return nil, err
		} else {
			a.balance.ExpirationDate = expDate
		}
	}
	// load ValueFormula
	if strings.Contains(a.Params, "ValueFormula") {
		var x struct {
			ValueFormula *utils.ValueFormula
		}
		if err := json.Unmarshal([]byte(a.Params), &x); err != nil {
			return nil, err
		}
		a.balance.GetValue().SetFloat64(x.ValueFormula.GetValue())
	}
	return a.balance, nil
}

const (
	LOG                       = "*log"
	RESET_TRIGGERS            = "*reset_triggers"
	SET_RECURRENT             = "*set_recurrent"
	UNSET_RECURRENT           = "*unset_recurrent"
	ALLOW_NEGATIVE            = "*allow_negative"
	DENY_NEGATIVE             = "*deny_negative"
	RESET_ACCOUNT             = "*reset_account"
	REMOVE_ACCOUNT            = "*remove_account"
	SET_BALANCE               = "*set_balance"
	REMOVE_BALANCE            = "*remove_balance"
	TOPUP_RESET               = "*topup_reset"
	TOPUP                     = "*topup"
	DEBIT_RESET               = "*debit_reset"
	DEBIT                     = "*debit"
	RESET_COUNTERS            = "*reset_counters"
	ENABLE_ACCOUNT            = "*enable_account"
	DISABLE_ACCOUNT           = "*disable_account"
	CALL_URL                  = "*call_url"
	CALL_URL_ASYNC            = "*call_url_async"
	MAIL_ASYNC                = "*mail_async"
	UNLIMITED                 = "*unlimited"
	CDRLOG                    = "*cdrlog"
	SET_DDESTINATIONS         = "*set_ddestinations"
	TRANSFER_MONETARY_DEFAULT = "*transfer_monetary_default"
	CGR_RPC                   = "*cgr_rpc"
)

func (a *Action) Clone() *Action {
	return &Action{
		ActionType:  a.ActionType,
		ExecFilter:  a.ExecFilter,
		Params:      a.Params,
		Filter1:     a.Filter1,
		Weight:      a.Weight,
		TOR:         a.TOR,
		parentGroup: a.parentGroup,
	}
}

type actionTypeFunc func(*Account, *StatsQueueTriggered, *Action, Actions) error

func getActionFunc(typ string) (actionTypeFunc, bool) {
	actionFuncMap := map[string]actionTypeFunc{
		LOG:                       logAction,
		CDRLOG:                    cdrLogAction,
		RESET_TRIGGERS:            resetTriggersAction,
		SET_RECURRENT:             setRecurrentAction,
		UNSET_RECURRENT:           unsetRecurrentAction,
		ALLOW_NEGATIVE:            allowNegativeAction,
		DENY_NEGATIVE:             denyNegativeAction,
		RESET_ACCOUNT:             resetAccountAction,
		TOPUP_RESET:               topupResetAction,
		TOPUP:                     topupAction,
		DEBIT_RESET:               debitResetAction,
		DEBIT:                     debitAction,
		RESET_COUNTERS:            resetCountersAction,
		ENABLE_ACCOUNT:            enableAccountAction,
		DISABLE_ACCOUNT:           disableAccountAction,
		CALL_URL:                  callURL,
		CALL_URL_ASYNC:            callURLAsync,
		MAIL_ASYNC:                mailAsync,
		SET_DDESTINATIONS:         setddestinations,
		REMOVE_ACCOUNT:            removeAccountAction,
		REMOVE_BALANCE:            removeBalanceAction,
		SET_BALANCE:               setBalanceAction,
		TRANSFER_MONETARY_DEFAULT: transferMonetaryDefaultAction,
		CGR_RPC:                   cgrRPCAction,
	}
	f, exists := actionFuncMap[typ]
	return f, exists
}

func logAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if acc != nil {
		utils.Logger.Info("Threshold hit", zap.Any("balance", acc), zap.String("extra_info", a.Params))
	}
	if sq != nil {
		utils.Logger.Info("Threshold hit", zap.Any("stats queue", sq), zap.String("extra_info", a.Params))
	}
	return
}

// Used by cdrLogAction to dynamically parse values out of account and action
func parseTemplateValue(rsrFlds utils.RSRFields, acnt *Account, action *Action) string {
	var parsedValue string // Template values
	b, err := action.getBalance(nil)
	if err != nil {
		utils.Logger.Error("<parse Template value> error unmarshalling balance: ", zap.Error(err))
		return ""
	}
	for _, rsrFld := range rsrFlds {
		switch rsrFld.Id {
		case "AccountID":
			parsedValue += rsrFld.ParseValue(acnt.FullID())
		case "Directions":
			parsedValue += rsrFld.ParseValue(b.Directions.String())
		case utils.TENANT:
			parsedValue += rsrFld.ParseValue(acnt.Tenant)
		case utils.ACCOUNT:
			parsedValue += rsrFld.ParseValue(acnt.Name)
		case "ActionID":
			parsedValue += rsrFld.ParseValue(action.parentGroup.Name)
		case "ActionType":
			parsedValue += rsrFld.ParseValue(action.ActionType)
		case "ActionValue":
			parsedValue += rsrFld.ParseValue(b.GetValue().String())
		case "BalanceType":
			parsedValue += rsrFld.ParseValue(action.TOR)
		case "BalanceUUID":
			parsedValue += rsrFld.ParseValue(b.UUID)
		case "BalanceID":
			parsedValue += rsrFld.ParseValue(b.ID)
		case "BalanceValue":
			parsedValue += rsrFld.ParseValue(action.balanceValue.String())
		case "DestinationIDs":
			parsedValue += rsrFld.ParseValue(b.DestinationIDs.String())
		case "Params":
			parsedValue += rsrFld.ParseValue(action.Params)
		case "RatingSubject":
			parsedValue += rsrFld.ParseValue(b.RatingSubject)
		case utils.CATEGORY:
			parsedValue += rsrFld.ParseValue(b.Categories.String()) // TODO: check these
		case "SharedGroups":
			parsedValue += rsrFld.ParseValue(b.SharedGroups.String()) // TODO: check these
		default:
			parsedValue += rsrFld.ParseValue("") // Mostly for static values
		}
	}
	return parsedValue
}

func cdrLogAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	defaultTemplate := map[string]utils.RSRFields{
		utils.TOR:       utils.ParseRSRFieldsMustCompile("BalanceType", utils.INFIELD_SEP),
		utils.CDRHOST:   utils.ParseRSRFieldsMustCompile("^127.0.0.1", utils.INFIELD_SEP),
		utils.DIRECTION: utils.ParseRSRFieldsMustCompile("Directions", utils.INFIELD_SEP),
		utils.REQTYPE:   utils.ParseRSRFieldsMustCompile("^"+utils.META_PREPAID, utils.INFIELD_SEP),
		utils.TENANT:    utils.ParseRSRFieldsMustCompile(utils.TENANT, utils.INFIELD_SEP),
		utils.ACCOUNT:   utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP),
		utils.SUBJECT:   utils.ParseRSRFieldsMustCompile(utils.ACCOUNT, utils.INFIELD_SEP),
		utils.COST:      utils.ParseRSRFieldsMustCompile("ActionValue", utils.INFIELD_SEP),
	}

	// overwrite default template
	if a.Params != "" && strings.Contains(a.Params, "CdrLogTemplate") {
		x := struct {
			CdrLogTemplate map[string]string
		}{}
		if err = json.Unmarshal([]byte(a.Params), &x); err != nil {
			return
		}
		for field, rsr := range x.CdrLogTemplate {
			defaultTemplate[field], err = utils.ParseRSRFields(rsr, utils.INFIELD_SEP)
			if err != nil {
				return err
			}
		}
	}

	// set stored cdr values
	var cdrs []*CDR
	for _, action := range acs {
		if !utils.IsSliceMember([]string{DEBIT, DEBIT_RESET, TOPUP, TOPUP_RESET}, action.ActionType) {
			continue // Only log specific actions
		}
		cdr := &CDR{RunID: action.ActionType, Source: CDRLOG, SetupTime: time.Now(), AnswerTime: time.Now(), OriginID: utils.GenUUID(), ExtraFields: make(map[string]string)}
		cdr.UniqueID = utils.Sha1(cdr.OriginID, cdr.SetupTime.String())
		cdr.Usage = time.Duration(1) * time.Second
		elem := reflect.ValueOf(cdr).Elem()
		for key, rsrFlds := range defaultTemplate {
			parsedValue := parseTemplateValue(rsrFlds, acc, action)
			field := elem.FieldByName(key)
			if field.IsValid() && field.CanSet() {
				switch field.Kind() {
				case reflect.Float64:
					value, err := strconv.ParseFloat(parsedValue, 64)
					if err != nil {
						continue
					}
					field.SetFloat(value)
				case reflect.String:
					field.SetString(parsedValue)
				}
			} else { // invalid fields go in extraFields of CDR
				cdr.ExtraFields[key] = parsedValue
			}
		}
		cdrs = append(cdrs, cdr)
		if cdrStorage == nil { // Only save if the cdrStorage is defined
			continue
		}
		if err := cdrStorage.SetCDR(cdr, true); err != nil {
			return err
		}
	}
	b, _ := json.Marshal(cdrs)
	a.ExecFilter = string(b) // testing purpose only
	return
}

func resetTriggersAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.ResetActionTriggers(a)
	return
}

func setRecurrentAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.SetRecurrent(a, true)
	return
}

func unsetRecurrentAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.SetRecurrent(a, false)
	return
}

func allowNegativeAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.AllowNegative = true
	return
}

func denyNegativeAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	ub.AllowNegative = false
	return
}

func resetAccountAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	return genericReset(ub)
}

func topupResetAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]Balances, 0)
	}
	c := a.Clone()
	genericMakeNegative(c)
	err = genericDebit(ub, c, true)
	a.balanceValue = c.balanceValue
	return
}

func topupAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	c := a.Clone()
	genericMakeNegative(c)
	err = genericDebit(ub, c, false)
	a.balanceValue = c.balanceValue
	return
}

func debitResetAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil { // Init the map since otherwise will get error if nil
		ub.BalanceMap = make(map[string]Balances, 0)
	}
	return genericDebit(ub, a, true)
}

func debitAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	err = genericDebit(ub, a, false)
	return
}

func resetCountersAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.UnitCounters != nil {
		ub.UnitCounters.resetCounters(a)
	}
	return
}

func genericMakeNegative(a *Action) {
	b, _ := a.getBalance(nil)
	if b != nil && b.GetValue().GtZero() { // only apply if not allready negative
		b.GetValue().Neg(b.GetValue())
	}
}

func genericDebit(ub *Account, a *Action, reset bool) (err error) {
	if ub == nil {
		return errors.New("nil account")
	}
	if ub.BalanceMap == nil {
		ub.BalanceMap = make(map[string]Balances)
	}
	return ub.debitBalanceAction(a, reset)
}

func enableAccountAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if acc == nil {
		return errors.New("nil account")
	}
	acc.Disabled = false
	return
}

func disableAccountAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) (err error) {
	if acc == nil {
		return errors.New("nil account")
	}
	acc.Disabled = true
	return
}

func genericReset(ub *Account) error {
	for k, _ := range ub.BalanceMap {
		ub.BalanceMap[k] = Balances{&Balance{Value: dec.New()}}
	}
	ub.InitCounters()
	ub.ResetActionTriggers(nil)
	return nil
}

func callURL(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	var o interface{}
	if ub != nil {
		o = ub
	}
	if sq != nil {
		o = sq
	}
	jsn, err := json.Marshal(o)
	if err != nil {
		return err
	}
	cfg := config.Get()
	fallbackPath := path.Join(*cfg.General.HttpFailedDir, fmt.Sprintf("act_%s_%s_%s.json", a.ActionType, a.Params, utils.GenUUID()))
	_, err = utils.NewHTTPPoster(*cfg.General.HttpSkipTlsVerify,
		cfg.General.ReplyTimeout.D()).Post(a.Params, utils.CONTENT_JSON, jsn, *cfg.General.HttpPosterAttempts, fallbackPath)
	return err
}

// Does not block for posts, no error reports
func callURLAsync(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	var o interface{}
	if ub != nil {
		o = ub
	}
	if sq != nil {
		o = sq
	}
	jsn, err := json.Marshal(o)
	if err != nil {
		return err
	}
	cfg := config.Get()
	fallbackPath := path.Join(*cfg.General.HttpFailedDir, fmt.Sprintf("act_%s_%s_%s.json", a.ActionType, a.Params, utils.GenUUID()))
	_, err = utils.NewHTTPPoster(*cfg.General.HttpSkipTlsVerify,
		cfg.General.ReplyTimeout.D()).Post(a.Params, utils.CONTENT_JSON, jsn, *cfg.General.HttpPosterAttempts, fallbackPath)
	return nil
}

// Mails the balance hitting the threshold towards predefined list of addresses
func mailAsync(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	cfg := config.Get()
	params := strings.Split(a.Params, string(utils.CSV_SEP))
	if len(params) == 0 {
		return errors.New("Unconfigured parameters for mail action")
	}
	toAddrs := strings.Split(params[0], string(utils.FALLBACK_SEP))
	toAddrStr := ""
	for idx, addr := range toAddrs {
		if idx != 0 {
			toAddrStr += ", "
		}
		toAddrStr += addr
	}
	var message []byte
	if ub != nil {
		balJsn, err := json.Marshal(ub)
		if err != nil {
			return err
		}
		message = []byte(fmt.Sprintf("To: %s\r\nSubject: [CGR Notification] Threshold hit on Balance: %s\r\n\r\nTime: \r\n\t%s\r\n\r\nBalance:\r\n\t%s\r\n\r\nYours faithfully,\r\nCGR Balance Monitor\r\n", toAddrStr, ub.FullID(), time.Now(), balJsn))
	} else if sq != nil {
		message = []byte(fmt.Sprintf("To: %s\r\nSubject: [CGR Notification] Threshold hit on StatsQueueId: %s\r\n\r\nTime: \r\n\t%s\r\n\r\nStatsQueueId:\r\n\t%s\r\n\r\nMetrics:\r\n\t%+v\r\n\r\nTrigger:\r\n\t%+v\r\n\r\nYours faithfully,\r\nCGR CDR Stats Monitor\r\n",
			toAddrStr, sq.Name, time.Now(), sq.Name, sq.Metrics, sq.Trigger))
	}
	auth := smtp.PlainAuth("", *cfg.Mailer.AuthUser, *cfg.Mailer.AuthPassword, strings.Split(*cfg.Mailer.Server, ":")[0]) // We only need host part, so ignore port
	go func() {
		for i := 0; i < 5; i++ { // Loop so we can increase the success rate on best effort
			if err := smtp.SendMail(*cfg.Mailer.Server, auth, *cfg.Mailer.FromAddress, toAddrs, message); err == nil {
				break
			} else if i == 4 {
				if ub != nil {
					utils.Logger.Warn("<Triggers> WARNING: Failed emailing", zap.String("params", a.Params), zap.Error(err), zap.String("BalanceID", ub.FullID()))
				} else if sq != nil {
					utils.Logger.Warn("<Triggers> WARNING: Failed emailing", zap.String("params", a.Params), zap.Error(err), zap.String("StatsQueueTriggeredId", sq.Name))
				}
				break
			}
			time.Sleep(time.Duration(i) * time.Minute)
		}
	}()
	return nil
}

func setddestinations(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	var ddcDestID string
	for _, bchain := range ub.BalanceMap {
		for _, b := range bchain {
			for destName := range b.DestinationIDs {
				if strings.HasPrefix(destName, "*ddc") {
					ddcDestID = destName
					break
				}
			}
			if ddcDestID != "" {
				break
			}
		}
		if ddcDestID != "" {
			break
		}
	}
	if ddcDestID != "" {
		// remove previous destinations with that name
		if err := ratingStorage.RemoveDestinations(ub.Tenant, "", ddcDestID); err != nil {
			return err
		}
		// make slice from prefixes
		destinations := make([]*Destination, len(sq.Metrics))
		i := 0
		for p := range sq.Metrics {
			destinations[i] = &Destination{Tenant: sq.Tenant, Code: p, Name: ddcDestID}
			i++
		}
		for _, dest := range destinations {
			// update dest in storage
			if err := ratingStorage.SetDestination(dest); err != nil {
				return err
			}
		}

	} else {
		return utils.ErrNotFound
	}
	return nil
}

func removeAccountAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	var accTenant, accName string
	if ub != nil {
		accTenant = ub.Tenant
		accName = ub.Name
	} else {
		accountInfo := struct {
			Tenant  string
			Account string
		}{}
		if a.Params != "" {
			if err := json.Unmarshal([]byte(a.Params), &accountInfo); err != nil {
				return err
			}
		}
		accTenant = accountInfo.Tenant
		accName = accountInfo.Account
	}
	if accTenant == "" || accName == "" {
		return utils.ErrInvalidKey
	}
	if err := accountingStorage.RemoveAccount(accTenant, accName); err != nil {
		utils.Logger.Error("Could not remove account Id: ", zap.String("tenant", accTenant), zap.String("ID", accName), zap.Error(err))
		return err
	}
	if err := ratingStorage.RemoveActionPlanBindings(ub.Tenant, ub.Name, ""); err != nil {
		return err
	}

	return nil
}

func removeBalanceAction(ub *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	if ub == nil {
		return fmt.Errorf("nil account for %s action", utils.ToJSON(a))
	}
	if _, exists := ub.BalanceMap[a.TOR]; !exists {
		return utils.ErrNotFound
	}
	bChain := ub.BalanceMap[a.TOR]
	found := false
	for i := 0; i < len(bChain); i++ {
		match, err := a.getFilter().Query(bChain[i], false)
		if err != nil {
			utils.Logger.Warn(fmt.Sprintf("<removeBalanceAction> action filter (%s) errored : (%s)", a.Filter1, err.Error()))
		}
		if match {
			// delete without preserving order
			bChain[i] = bChain[len(bChain)-1]
			bChain = bChain[:len(bChain)-1]
			i -= 1
			found = true
		}
	}
	ub.BalanceMap[a.TOR] = bChain
	if !found {
		return utils.ErrNotFound
	}
	return nil
}

func setBalanceAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	if acc == nil {
		return fmt.Errorf("nil account for %s action", utils.ToJSON(a))
	}
	err := acc.setBalanceAction(a)
	return err
}

func transferMonetaryDefaultAction(acc *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	if acc == nil {
		utils.Logger.Error("*transfer_monetary_default called without account")
		return utils.ErrAccountNotFound
	}
	if _, exists := acc.BalanceMap[utils.MONETARY]; !exists {
		return utils.ErrNotFound
	}
	defaultBalance := acc.GetDefaultMoneyBalance()
	bChain := acc.BalanceMap[utils.MONETARY]
	for _, balance := range bChain {
		match, err := a.getFilter().Query(balance, false)
		if err != nil {
			utils.Logger.Warn(fmt.Sprintf("<transferMonetaryDefaultAction> action filter (%s) errored : (%s)", a.Filter1, err.Error()))
		}
		if balance.UUID != defaultBalance.UUID &&
			balance.ID != defaultBalance.ID && // extra caution
			match {
			if balance.Value.GtZero() {
				defaultBalance.GetValue().AddS(balance.GetValue())
				balance.SetValue(dec.Zero)
			}
		}
	}
	return nil
}

type RPCRequest struct {
	Address   string
	Transport string
	Method    string
	Attempts  int
	Async     bool
	Params    map[string]interface{}
}

/*
<< .Object.Property >>

Property can be a attribute or a method both used without ()
Please also note the initial dot .

Currently there are following objects that can be used:

Account -  the account that this action is called on
Action - the action with all it's attributs
Actions - the list of actions in the current action set
Sq - StatsQueueTriggered object

We can actually use everythiong that go templates offer. You can read more here: https://golang.org/pkg/text/template/
*/
func cgrRPCAction(account *Account, sq *StatsQueueTriggered, a *Action, acs Actions) error {
	// parse template
	tmpl := template.New("extra_params")
	tmpl.Delims("<<", ">>")
	t, err := tmpl.Parse(a.Params)
	if err != nil {
		utils.Logger.Error(fmt.Sprintf("error parsing *cgr_rpc template: %s", err.Error()))
		return err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, struct {
		Account *Account
		Sq      *StatsQueueTriggered
		Action  *Action
		Actions Actions
	}{account, sq, a, acs}); err != nil {
		utils.Logger.Error(fmt.Sprintf("error executing *cgr_rpc template %s:", err.Error()))
		return err
	}
	processedExtraParam := buf.String()
	//utils.Logger.Info("Params: " + parsedParams)
	x := struct {
		RpcRequest *RPCRequest
	}{}
	if err := json.Unmarshal([]byte(processedExtraParam), &x); err != nil {
		return err
	}
	req := x.RpcRequest
	params, err := utils.GetRpcParams(req.Method)
	if err != nil {
		return err
	}
	var client rpcclient.RpcClientConnection
	if req.Address != utils.MetaInternal {
		if client, err = rpcclient.NewRpcClient("tcp", req.Address, req.Attempts, 0, config.Get().General.ConnectTimeout.D(), config.Get().General.ReplyTimeout.D(), req.Transport, nil, false); err != nil {
			return err
		}
	} else {
		client = params.Object.(rpcclient.RpcClientConnection)
	}
	in, out := params.InParam, params.OutParam
	//utils.Logger.Info("Params: " + utils.ToJSON(req.Params))
	//p, err := utils.FromMapStringInterfaceValue(req.Params, in)
	mapstructure.Decode(req.Params, in)
	if err != nil {
		utils.Logger.Info("<*cgr_rpc> err: " + err.Error())
		return err
	}
	if in == nil {
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> nil params err: req.Params: %+v params: %+v", req.Params, params))
		return utils.ErrParserError
	}
	utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> calling: %s with: %s and result %v", req.Method, utils.ToJSON(in), out))
	if !req.Async {
		err = client.Call(req.Method, in, out)
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> result: %s err: %v", utils.ToJSON(out), err))
		return err
	}
	go func() {
		err := client.Call(req.Method, in, out)
		utils.Logger.Info(fmt.Sprintf("<*cgr_rpc> result: %s err: %v", utils.ToJSON(out), err))
	}()
	return nil
}

// Structure to store actions according to weight
type Actions []*Action

func (apl Actions) Sort() {
	sort.Slice(apl, func(j, i int) bool {
		// we need higher weights earlyer in the list
		return apl[i].Weight < apl[j].Weight
	})
}
