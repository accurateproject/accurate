/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package sessionmanager

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var kamEv = KamEvent{KAM_TR_INDEX: "29223", KAM_TR_LABEL: "698469260", "callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", "from_tag": "eb082607", "to_tag": "4ea9687f", "cgr_account": "dan",
	"cgr_reqtype": "prepaid", "cgr_subject": "dan", "cgr_destination": "+4986517174963", "cgr_tenant": "itsyscom.com",
	"cgr_duration": "20", "extra1": "val1", "extra2": "val2"}

func TestKamailioEventInterface(t *testing.T) {
	var _ utils.Event = utils.Event(kamEv)
}

func TestNewKamEvent(t *testing.T) {
	evStr := `{"event":"CGR_CALL_END",
		"callid":"46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0",
		"from_tag":"bf71ad59",
		"to_tag":"7351fecf",
		"cgr_reqtype":"postpaid",
		"cgr_account":"1001", 
		"cgr_destination":"1002",
		"cgr_answertime":"1419839310",
		"cgr_duration":"3"}`
	eKamEv := KamEvent{"event": "CGR_CALL_END", "callid": "46c01a5c249b469e76333fc6bfa87f6a@0:0:0:0:0:0:0:0", "from_tag": "bf71ad59", "to_tag": "7351fecf",
		"cgr_reqtype": "postpaid", "cgr_account": "1001", "cgr_destination": "1002", "cgr_answertime": "1419839310", "cgr_duration": "3"}
	if kamEv, err := NewKamEvent([]byte(evStr)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eKamEv, kamEv) {
		t.Error("Received: ", kamEv)
	}
}

func TestKevAsKamAuthReply(t *testing.T) {
	expectedKar := &KamAuthReply{Event: CGR_AUTH_REPLY, TransactionIndex: 29223, TransactionLabel: 698469260, MaxSessionTime: 1200}
	if rcvKar, err := kamEv.AsKamAuthReply(1200.0, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedKar, rcvKar) {
		t.Error("Received KAR: ", rcvKar)
	}
}

func TestKevMissingParameter(t *testing.T) {
	kamEv := KamEvent{"event": "CGR_AUTH_REQUEST", "tr_index": "36045", "tr_label": "612369399", "cgr_reqtype": "postpaid",
		"cgr_account": "1001", "cgr_destination": "1002"}
	if !kamEv.MissingParameter() {
		t.Error("Failed detecting missing parameters")
	}
	kamEv["cgr_setuptime"] = "1419962256"
	if kamEv.MissingParameter() {
		t.Error("False detecting missing parameters")
	}
	kamEv = KamEvent{"event": "UNKNOWN"}
	if !kamEv.MissingParameter() {
		t.Error("Failed detecting missing parameters")
	}
	kamEv = KamEvent{"event": "CGR_CALL_START", "callid": "9d28ec3ee068babdfe036623f42c0969@0:0:0:0:0:0:0:0", "from_tag": "3131b566",
		"cgr_reqtype": "postpaid", "cgr_account": "1001", "cgr_destination": "1002"}
	if !kamEv.MissingParameter() {
		t.Error("Failed detecting missing parameters")
	}
	kamEv["h_entry"] = "463"
	kamEv["h_id"] = "2605"
	kamEv["cgr_answertime"] = "1419964961"
	if kamEv.MissingParameter() {
		t.Error("False detecting missing parameters")
	}
}