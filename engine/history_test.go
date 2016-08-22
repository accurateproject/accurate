
package engine

import (
	"strings"
	"testing"

	"github.com/cgrates/cgrates/history"
)

func TestHistoryRatinPlans(t *testing.T) {
	scribe := historyScribe.(*history.MockScribe)
	buf := scribe.GetBuffer(history.RATING_PROFILES_FN)
	if !strings.Contains(buf.String(), `{"Id":"*out:vdf:0:minu","RatingPlanActivations":[{"ActivationTime":"2012-01-01T00:00:00Z","RatingPlanId":"EVENING","FallbackKeys":null,"CdrStatQueueIds":[""]}]}`) {
		t.Error("Error in destination history content:", buf.String())
	}
}

func TestHistoryDestinations(t *testing.T) {
	scribe := historyScribe.(*history.MockScribe)
	buf := scribe.GetBuffer(history.DESTINATIONS_FN)
	expected := `{"Id":"ALL","Prefixes":["49","41","43"]},
{"Id":"DST_UK_Mobile_BIG5","Prefixes":["447956"]},
{"Id":"EU_LANDLINE","Prefixes":["444"]},
{"Id":"EXOTIC","Prefixes":["999"]},
{"Id":"GERMANY","Prefixes":["49"]},
{"Id":"GERMANY_O2","Prefixes":["41"]},
{"Id":"GERMANY_PREMIUM","Prefixes":["43"]},
{"Id":"NAT","Prefixes":["0256","0257","0723","+49"]},
{"Id":"PSTN_70","Prefixes":["+4970"]},
{"Id":"PSTN_71","Prefixes":["+4971"]},
{"Id":"PSTN_72","Prefixes":["+4972"]},
{"Id":"RET","Prefixes":["0723","0724"]},
{"Id":"SPEC","Prefixes":["0723045"]},
{"Id":"URG","Prefixes":["112"]}`
	if !strings.Contains(buf.String(), expected) {
		t.Error("Error in destination history content:", buf.String())
	}
}
