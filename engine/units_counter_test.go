package engine

import (
	"testing"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

func TestUnitsCounterAddBalance(t *testing.T) {
	uc := &UnitCounter{
		Counters: CounterFilters{&CounterFilter{Value: dec.NewFloat(1)}, &CounterFilter{Filter: `{"Weight": 20, "DestinationIDs": {"$in":["NAT"]}`}, &CounterFilter{Filter: `{"Weight": 10, "DestinationIDs": {"$in":["RET"]))`}},
	}
	UnitCounters{utils.SMS: []*UnitCounter{uc}}.addUnits(dec.NewFloat(20), utils.SMS, &CallCost{Destination: "test"}, nil)
	if len(uc.Counters) != 3 {
		t.Error("Error adding minute bucket: ", uc.Counters)
	}
}

func TestUnitsCounterAddBalanceExists(t *testing.T) {
	uc := &UnitCounter{
		Counters: CounterFilters{&CounterFilter{Value: dec.NewFloat(1)}, &CounterFilter{Value: dec.NewFloat(10), Filter: `{"Weight": 20, "DestinationIDs": {"$in":["NAT"]))`}, &CounterFilter{Filter: `{"Weight": 10, "DestinationIDs": {"$in":["RET"]))`}},
	}
	UnitCounters{utils.SMS: []*UnitCounter{uc}}.addUnits(dec.NewFloat(5), utils.SMS, &CallCost{Destination: "0723"}, nil)
	if len(uc.Counters) != 3 || uc.Counters[1].Value.String() != "15" {
		t.Error("Error adding minute bucket!")
	}
}

func TestUnitCountersCountAllMonetary(t *testing.T) {
	acc := &Account{
		triggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
		},
	}
	acc.InitCounters()
	acc.UnitCounters.addUnits(dec.NewFloat(10), utils.MONETARY, &CallCost{}, nil)

	if len(acc.UnitCounters) != 3 ||
		len(acc.UnitCounters[utils.MONETARY][0].Counters) != 2 ||
		acc.UnitCounters[utils.MONETARY][0].Counters[0].getValue().String() != "10" ||
		acc.UnitCounters[utils.MONETARY][0].Counters[1].getValue().String() != "10" {
		t.Log("UC: ", utils.ToIJSON(acc.UnitCounters))
		t.Errorf("Error Initializing adding unit counters: %v", len(acc.UnitCounters))
	}
}

func TestUnitCountersCountAllMonetaryId(t *testing.T) {
	a := &Account{
		triggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 20}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(dec.NewFloat(10), utils.MONETARY, nil, &Balance{Weight: 20, Directions: utils.NewStringMap(utils.OUT)})
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MONETARY][0].Counters) != 2 ||
		a.UnitCounters[utils.MONETARY][0].Counters[0].getValue().String() != "0" ||
		a.UnitCounters[utils.MONETARY][0].Counters[1].getValue().String() != "10" {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersCountAllVoiceDestinationEvent(t *testing.T) {
	a := &Account{
		triggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 20}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out"]}, "DestinationIDs":{"$in:["NAT"]"}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR22",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"DestinationIDs":{"$in:["RET"]"}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(dec.NewFloat(10), utils.VOICE, &CallCost{Destination: "0723045326"}, nil)

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.VOICE][0].Counters) != 2 ||
		a.UnitCounters[utils.VOICE][0].Counters[0].Value.String() != "10" ||
		a.UnitCounters[utils.VOICE][0].Counters[1].Value.String() != "10" {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error adding unit counters: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersKeepValuesAfterInit(t *testing.T) {
	a := &Account{
		triggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 20}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out"]}, "DestinationIDs":{"$in:["NAT"]"}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR22",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"DestinationIDs":{"$in:["RET"]"}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out"]}, "Weight": 10}`,
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(dec.NewFloat(10), utils.VOICE, &CallCost{Destination: "0723045326"}, nil)

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.VOICE][0].Counters) != 2 ||
		a.UnitCounters[utils.VOICE][0].Counters[0].Value.String() != "10" ||
		a.UnitCounters[utils.VOICE][0].Counters[1].Value.String() != "10" {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error adding unit counters: %v", len(a.UnitCounters))
	}
	a.InitCounters()

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.VOICE][0].Counters) != 2 ||
		a.UnitCounters[utils.VOICE][0].Counters[0].Value.String() != "10" ||
		a.UnitCounters[utils.VOICE][0].Counters[1].Value.String() != "10" {
		for key, counters := range a.UnitCounters {
			t.Log(key)
			for _, uc := range counters {
				t.Logf("UC: %+v", uc)
				for _, b := range uc.Counters {
					t.Logf("B: %+v", b)
				}
			}
		}
		t.Errorf("Error keeping counter values after init: %v", len(a.UnitCounters))
	}
}

func TestUnitCountersResetCounterById(t *testing.T) {
	a := &Account{
		triggers: ActionTriggers{
			&ActionTrigger{
				UniqueID:      "TestTR1",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR11",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.MONETARY,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR2",
				ThresholdType: utils.TRIGGER_MAX_EVENT_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR3",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.VOICE,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR4",
				ThresholdType: utils.TRIGGER_MAX_BALANCE_COUNTER,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
			&ActionTrigger{
				UniqueID:      "TestTR5",
				ThresholdType: utils.TRIGGER_MAX_BALANCE,
				TOR:           utils.SMS,
				Filter:        `{"Directions": {"$in":["*out", "*in"]}}`,
			},
		},
	}
	a.InitCounters()
	a.UnitCounters.addUnits(dec.NewFloat(10), utils.MONETARY, &CallCost{}, nil)

	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MONETARY][0].Counters) != 2 ||
		a.UnitCounters[utils.MONETARY][0].Counters[0].getValue().String() != "10" ||
		a.UnitCounters[utils.MONETARY][0].Counters[1].getValue().String() != "10" {
		t.Log("UC: ", utils.ToIJSON(a.UnitCounters))
		t.Errorf("Error Initializing adding unit counters: %v", len(a.UnitCounters))
	}
	a.UnitCounters.resetCounters(&Action{
		TOR:     utils.MONETARY,
		Filter1: `{"UniqueID":"TestTR11"}`,
	})
	if len(a.UnitCounters) != 3 ||
		len(a.UnitCounters[utils.MONETARY][0].Counters) != 2 ||
		a.UnitCounters[utils.MONETARY][0].Counters[0].Value.String() != "10" ||
		a.UnitCounters[utils.MONETARY][0].Counters[1].Value.String() != "0" {
		t.Log("UC: ", utils.ToIJSON(a.UnitCounters))
		t.Errorf("Error Initializing adding unit counters: %v", len(a.UnitCounters))
	}
}
