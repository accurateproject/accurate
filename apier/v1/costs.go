
package v1

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetDataCost struct {
	Direction                string
	Category                 string
	Tenant, Account, Subject string
	StartTime                time.Time
	Usage                    int64 // the call duration so far (till TimeEnd)
}

func (apier *ApierV1) GetDataCost(attrs AttrGetDataCost, reply *engine.DataCost) error {
	usageAsDuration := time.Duration(attrs.Usage) * time.Second // Convert to seconds to match the loaded rates
	cd := &engine.CallDescriptor{
		Direction:     attrs.Direction,
		Category:      attrs.Category,
		Tenant:        attrs.Tenant,
		Account:       attrs.Account,
		Subject:       attrs.Subject,
		TimeStart:     attrs.StartTime,
		TimeEnd:       attrs.StartTime.Add(usageAsDuration),
		DurationIndex: usageAsDuration,
		TOR:           utils.DATA,
	}
	var cc engine.CallCost
	if err := apier.Responder.GetCost(cd, &cc); err != nil {
		return utils.NewErrServerError(err)
	}
	if dc, err := cc.ToDataCost(); err != nil {
		return utils.NewErrServerError(err)
	} else if dc != nil {
		*reply = *dc
	}
	return nil
}
