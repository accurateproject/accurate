package sessionmanager

import (
	"reflect"
	"testing"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
)

var smgCfg *config.Config

func init() {
	config.Reset()
	smgCfg = config.Get()
	smgCfg.SmGeneric.SessionIndexes = []string{"Tenant", "Account", "Extra3", "Extra4"}

}

func TestSMGSessionIndexing(t *testing.T) {
	smg := NewSMGeneric(smgCfg, nil, nil, "UTC")
	smGev := SMGenericEvent{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.TOR:              "*voice",
		utils.ACCID:            "12345",
		utils.DIRECTION:        "*out",
		utils.ACCOUNT:          "account1",
		utils.SUBJECT:          "subject1",
		utils.DESTINATION:      "+4986517174963",
		utils.CATEGORY:         "call",
		utils.TENANT:           "cgrates.org",
		utils.REQTYPE:          "*prepaid",
		utils.SETUP_TIME:       "2015-11-09 14:21:24",
		utils.ANSWER_TIME:      "2015-11-09 14:22:02",
		utils.USAGE:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.CDRHOST:          "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	}
	// Index first session
	smgSession := &SMGSession{eventStart: smGev}
	uuid := smGev.GetUUID()
	smg.indexSession(uuid, smgSession)
	eIndexes := map[string]map[string]utils.StringMap{
		"Tenant": map[string]utils.StringMap{
			"cgrates.org": utils.StringMap{
				uuid: true,
			},
		},
		"Account": map[string]utils.StringMap{
			"account1": utils.StringMap{
				uuid: true,
			},
		},
		"Extra3": map[string]utils.StringMap{
			utils.MetaEmpty: utils.StringMap{
				uuid: true,
			},
		},
		"Extra4": map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				uuid: true,
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, smg.sessionIndexes) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.sessionIndexes)
	}
	// Index seccond session
	smGev2 := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.ACCID:       "12346",
		utils.DIRECTION:   "*out",
		utils.ACCOUNT:     "account2",
		utils.DESTINATION: "+4986517174964",
		utils.TENANT:      "itsyscom.com",
		"Extra3":          "",
		"Extra4":          "info2",
	}
	uuid2 := smGev2.GetUUID()
	smgSession2 := &SMGSession{eventStart: smGev2}
	smg.indexSession(uuid2, smgSession2)
	eIndexes = map[string]map[string]utils.StringMap{
		"Tenant": map[string]utils.StringMap{
			"cgrates.org": utils.StringMap{
				uuid: true,
			},
			"itsyscom.com": utils.StringMap{
				uuid2: true,
			},
		},
		"Account": map[string]utils.StringMap{
			"account1": utils.StringMap{
				uuid: true,
			},
			"account2": utils.StringMap{
				uuid2: true,
			},
		},
		"Extra3": map[string]utils.StringMap{
			utils.MetaEmpty: utils.StringMap{
				uuid:  true,
				uuid2: true,
			},
		},
		"Extra4": map[string]utils.StringMap{
			utils.NOT_AVAILABLE: utils.StringMap{
				uuid: true,
			},
			"info2": utils.StringMap{
				uuid2: true,
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, smg.sessionIndexes) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.sessionIndexes)
	}
	// Unidex first session
	smg.unindexSession(uuid)
	eIndexes = map[string]map[string]utils.StringMap{
		"Tenant": map[string]utils.StringMap{
			"itsyscom.com": utils.StringMap{
				uuid2: true,
			},
		},
		"Account": map[string]utils.StringMap{
			"account2": utils.StringMap{
				uuid2: true,
			},
		},
		"Extra3": map[string]utils.StringMap{
			utils.MetaEmpty: utils.StringMap{
				uuid2: true,
			},
		},
		"Extra4": map[string]utils.StringMap{
			"info2": utils.StringMap{
				uuid2: true,
			},
		},
	}
	if !reflect.DeepEqual(eIndexes, smg.sessionIndexes) {
		t.Errorf("Expecting: %+v, received: %+v", eIndexes, smg.sessionIndexes)
	}
}

func TestSMGActiveSessions(t *testing.T) {
	smg := NewSMGeneric(smgCfg, nil, nil, "UTC")
	smGev1 := SMGenericEvent{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.TOR:              "*voice",
		utils.ACCID:            "111",
		utils.DIRECTION:        "*out",
		utils.ACCOUNT:          "account1",
		utils.SUBJECT:          "subject1",
		utils.DESTINATION:      "+4986517174963",
		utils.CATEGORY:         "call",
		utils.TENANT:           "cgrates.org",
		utils.REQTYPE:          "*prepaid",
		utils.SETUP_TIME:       "2015-11-09 14:21:24",
		utils.ANSWER_TIME:      "2015-11-09 14:22:02",
		utils.USAGE:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier1",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.CDRHOST:          "127.0.0.1",
		"Extra1":               "Value1",
		"Extra2":               5,
		"Extra3":               "",
	}
	smg.recordSession(smGev1.GetUUID(), &SMGSession{eventStart: smGev1})
	smGev2 := SMGenericEvent{
		utils.EVENT_NAME:       "TEST_EVENT",
		utils.TOR:              "*voice",
		utils.ACCID:            "222",
		utils.DIRECTION:        "*out",
		utils.ACCOUNT:          "account2",
		utils.DESTINATION:      "+4986517174963",
		utils.CATEGORY:         "call",
		utils.TENANT:           "itsyscom.com",
		utils.REQTYPE:          "*prepaid",
		utils.ANSWER_TIME:      "2015-11-09 14:22:02",
		utils.USAGE:            "1m23s",
		utils.LastUsed:         "21s",
		utils.PDD:              "300ms",
		utils.SUPPLIER:         "supplier2",
		utils.DISCONNECT_CAUSE: "NORMAL_DISCONNECT",
		utils.CDRHOST:          "127.0.0.1",
		"Extra1":               "Value1",
		"Extra3":               "extra3",
	}
	smg.recordSession(smGev2.GetUUID(), &SMGSession{eventStart: smGev2})
	if aSessions, _, err := smg.ActiveSessions(nil, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.ActiveSessions(map[string]string{"Tenant": "itsyscom.com"}, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.ActiveSessions(map[string]string{utils.TOR: "*voice"}, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.ActiveSessions(map[string]string{"Extra3": utils.MetaEmpty}, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
	if aSessions, _, err := smg.ActiveSessions(map[string]string{utils.SUPPLIER: "supplier2"}, false); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Received sessions: %+v", aSessions)
	}
}
