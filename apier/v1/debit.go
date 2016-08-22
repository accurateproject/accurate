package v1

import (
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

func (self *ApierV1) DebitUsage(usageRecord engine.UsageRecord, reply *string) error {
	if missing := utils.MissingStructFields(&usageRecord, []string{"Account", "Destination", "Usage"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	err := engine.LoadUserProfile(usageRecord, "")
	if err != nil {
		*reply = err.Error()
		return err
	}
	if usageRecord.ToR == "" {
		usageRecord.ToR = utils.VOICE
	}
	if usageRecord.RequestType == "" {
		usageRecord.RequestType = self.Config.DefaultReqType
	}
	if usageRecord.Direction == "" {
		usageRecord.Direction = utils.OUT
	}
	if usageRecord.Tenant == "" {
		usageRecord.Tenant = self.Config.DefaultTenant
	}
	if usageRecord.Category == "" {
		usageRecord.Category = self.Config.DefaultCategory
	}
	if usageRecord.Subject == "" {
		usageRecord.Subject = usageRecord.Account
	}
	if usageRecord.AnswerTime == "" {
		usageRecord.AnswerTime = utils.META_NOW
	}
	cd, err := usageRecord.AsCallDescriptor(self.Config.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	var cc engine.CallCost
	if err := self.Responder.Debit(cd, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = OK
	return nil
}
