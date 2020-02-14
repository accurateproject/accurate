package engine

import (
	"encoding/json"
	"testing"

	"github.com/accurateproject/accurate/history"
	"github.com/accurateproject/accurate/utils"
)

func TestHistoryRatinPlans(t *testing.T) {
	scribe := historyScribe.(*history.MockScribe)
	buf := scribe.GetBuffer(history.RATING_PROFILES_FN)

	x := make([]*RatingProfile, 0)
	if err := json.Unmarshal(buf.Bytes(), &x); err != nil {
		t.Fatal(err)
	}
	if len(x) != 24 {
		t.Errorf("Error in rating profile history content: %d\n%s", len(x), utils.ToIJSON(x))
	}
}

func TestHistoryDestinations(t *testing.T) {
	scribe := historyScribe.(*history.MockScribe)
	buf := scribe.GetBuffer(history.DESTINATIONS_FN)

	x := make([]*Destination, 0)
	if err := json.Unmarshal(buf.Bytes(), &x); err != nil {
		t.Fatal(err)
	}
	if len(x) != 18 {
		t.Errorf("Error in destination history content: %d\n%s", len(x), utils.ToIJSON(x))
	}
}
