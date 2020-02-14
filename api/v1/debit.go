package v1

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// DebitUsage will debit the balance for the usage cost, allowing the
// account to go negative if the cost calculated is greater than the balance
func (api *ApiV1) DebitUsage(usageRecord engine.UsageRecord, reply *string) error {
	return api.DebitUsageWithOptions(AttrDebitUsageWithOptions{
		UsageRecord:          &usageRecord,
		AllowNegativeAccount: true,
	}, reply)
}

// AttrDebitUsageWithOptions represents the DebitUsage request
type AttrDebitUsageWithOptions struct {
	UsageRecord          *engine.UsageRecord
	AllowNegativeAccount bool // allow account to go negative during debit
}

// DebitUsageWithOptions will debit the account based on the usage cost with
// additional options to control if the balance can go negative
func (api *ApiV1) DebitUsageWithOptions(args AttrDebitUsageWithOptions, reply *string) error {
	usageRecord := args.UsageRecord
	if missing := utils.MissingStructFields(usageRecord, []string{"Account", "Destination", "Usage"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	err := engine.LoadUserProfile(args.UsageRecord, false)
	if err != nil {
		*reply = err.Error()
		return err
	}

	// Set values for optional parameters
	if usageRecord.ToR == "" {
		usageRecord.ToR = utils.VOICE
	}
	if usageRecord.RequestType == "" {
		usageRecord.RequestType = *api.cfg.General.DefaultRequestType
	}
	if usageRecord.Direction == "" {
		usageRecord.Direction = utils.OUT
	}
	if usageRecord.Tenant == "" {
		usageRecord.Tenant = *api.cfg.General.DefaultTenant
	}
	if usageRecord.Category == "" {
		usageRecord.Category = *api.cfg.General.DefaultCategory
	}
	if usageRecord.Subject == "" {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.AnswerTime == "" {
		usageRecord.AnswerTime = utils.META_NOW
	}
	cd, err := usageRecord.AsCallDescriptor(*api.cfg.General.DefaultTimezone, !args.AllowNegativeAccount)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	var cc engine.CallCost
	if err := api.responder.Debit(cd, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}
