package v1

import (
	"strconv"
	"time"

	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (api *ApiV1) GetMaxUsage(usageRecord engine.UsageRecord, maxUsage *float64) error {
	err := engine.LoadUserProfile(&usageRecord, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
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
	if usageRecord.SetupTime == "" {
		usageRecord.SetupTime = utils.META_NOW
	}
	if usageRecord.Usage == "" {
		usageRecord.Usage = strconv.FormatFloat(api.cfg.SmGeneric.MaxCallDuration.D().Seconds(), 'f', -1, 64)
	}
	storedCdr, err := usageRecord.AsStoredCdr(*api.cfg.General.DefaultTimezone)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	var maxDur float64
	if err := api.responder.GetDerivedMaxSessionTime(storedCdr, &maxDur); err != nil {
		return err
	}
	if maxDur == -1.0 {
		*maxUsage = -1.0
		return nil
	}
	*maxUsage = time.Duration(maxDur).Seconds()
	return nil
}
