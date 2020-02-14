package engine

import (
	"net/http"
	"strconv"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

func NewCgrCdrFromHttpReq(req *http.Request, timezone string) (CgrCdr, error) {
	if req.Form == nil {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
	}
	cgrCdr := make(CgrCdr)
	cgrCdr[utils.CDRSOURCE] = req.RemoteAddr
	for k, vals := range req.Form {
		cgrCdr[k] = vals[0] // We only support the first value for now, if more are provided it is considered remote's fault
	}
	return cgrCdr, nil
}

type CgrCdr map[string]string

func (cgrCdr CgrCdr) getUniqueID(timezone string) string {
	if UniqueID, hasIt := cgrCdr[utils.UniqueID]; hasIt {
		return UniqueID
	}
	setupTime, _ := utils.ParseTimeDetectLayout(cgrCdr[utils.SETUP_TIME], timezone)
	return utils.Sha1(cgrCdr[utils.ACCID], setupTime.UTC().String())
}

func (cgrCdr CgrCdr) getExtraFields() map[string]string {
	extraFields := make(map[string]string)
	for k, v := range cgrCdr {
		if !utils.IsSliceMember(utils.PrimaryCdrFields, k) {
			extraFields[k] = v
		}
	}
	return extraFields
}

func (cgrCdr CgrCdr) AsStoredCdr(timezone string) *CDR {
	storCdr := new(CDR)
	storCdr.UniqueID = cgrCdr.getUniqueID(timezone)
	storCdr.ToR = cgrCdr[utils.TOR]
	storCdr.OriginID = cgrCdr[utils.ACCID]
	storCdr.OriginHost = cgrCdr[utils.CDRHOST]
	storCdr.Source = cgrCdr[utils.CDRSOURCE]
	storCdr.RequestType = cgrCdr[utils.REQTYPE]
	storCdr.Direction = utils.OUT
	storCdr.Tenant = cgrCdr[utils.TENANT]
	storCdr.Category = cgrCdr[utils.CATEGORY]
	storCdr.Account = cgrCdr[utils.ACCOUNT]
	storCdr.Subject = cgrCdr[utils.SUBJECT]
	storCdr.Destination = cgrCdr[utils.DESTINATION]
	storCdr.SetupTime, _ = utils.ParseTimeDetectLayout(cgrCdr[utils.SETUP_TIME], timezone) // Not interested to process errors, should do them if necessary in a previous step
	storCdr.PDD, _ = utils.ParseDurationWithSecs(cgrCdr[utils.PDD])
	storCdr.AnswerTime, _ = utils.ParseTimeDetectLayout(cgrCdr[utils.ANSWER_TIME], timezone)
	storCdr.Usage, _ = utils.ParseDurationWithSecs(cgrCdr[utils.USAGE])
	storCdr.Supplier = cgrCdr[utils.SUPPLIER]
	storCdr.DisconnectCause = cgrCdr[utils.DISCONNECT_CAUSE]
	storCdr.ExtraFields = cgrCdr.getExtraFields()
	storCdr.Cost = dec.NewVal(-1, 0)
	if costStr, hasIt := cgrCdr[utils.COST]; hasIt {
		storCdr.Cost, _ = dec.New().SetString(costStr)
	}
	if ratedStr, hasIt := cgrCdr[utils.RATED]; hasIt {
		storCdr.Rated, _ = strconv.ParseBool(ratedStr)
	}
	return storCdr
}
