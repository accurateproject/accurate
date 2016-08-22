
package utils

import (
	"reflect"
	"testing"
)

func TestNewDTCSFromRPKey(t *testing.T) {
	rpKey := "*out:tenant12:call:dan12"
	if dtcs, err := NewDTCSFromRPKey(rpKey); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dtcs, &DirectionTenantCategorySubject{"*out", "tenant12", "call", "dan12"}) {
		t.Error("Received: ", dtcs)
	}
}
