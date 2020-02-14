package utils

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestLoaderJSONDestinationStringSingle(t *testing.T) {
	reader := strings.NewReader(`
{"Tenant":"x", "Code":"0723", "Tag": "VODAFONE"}
`)
	newDest := func() interface{} { return &TpDestination{} }
	callback := func(el interface{}) error {
		if !reflect.DeepEqual(el, &TpDestination{Tenant: "x", Code: "0723", Tag: "VODAFONE"}) {
			return fmt.Errorf("%+v", el)
		}
		return nil
	}
	if err := LoadJSON(reader, newDest, callback); err != nil {
		t.Error("error loading destination: ", err)
	}
}

func TestLoaderJSONDestinationStringMultiple(t *testing.T) {
	reader := strings.NewReader(`
{"Tenant": "x", "Code":"0723", "Tag": "VODAFONE"}
{"Tenant": "y", "Code":"+0724", "Tag": "VODAFONE"}
`)
	result := make([]*TpDestination, 0)
	newDest := func() interface{} { return &TpDestination{} }
	callback := func(el interface{}) error {
		result = append(result, el.(*TpDestination))
		return nil
	}
	if err := LoadJSON(reader, newDest, callback); err != nil {
		t.Error("error loading destinations: ", err)
	}
	if len(result) != 2 ||
		!reflect.DeepEqual(result[0], &TpDestination{Tenant: "x", Code: "0723", Tag: "VODAFONE"}) ||
		!reflect.DeepEqual(result[1], &TpDestination{Tenant: "y", Code: "+0724", Tag: "VODAFONE"}) {
		t.Error("error loading destinations: ", ToIJSON(result))
	}
}

func TestLoaderJSONDestinationFile(t *testing.T) {
	reader, err := os.Open("../data/tariffplans/tutorial/Destinations.json")
	if err != nil {
		t.Fatal("error opening destinations json file: ", err)
	}
	result := make([]*TpDestination, 0)
	newDest := func() interface{} { return &TpDestination{} }
	callback := func(el interface{}) error {
		result = append(result, el.(*TpDestination))
		return nil
	}
	if err := LoadJSON(reader, newDest, callback); err != nil {
		t.Error("error loading destinations: ", err)
	}
	if len(result) != 7 ||
		!reflect.DeepEqual(result[0], &TpDestination{Tenant: "tut", Code: "1002", Tag: "DST_1002"}) ||
		!reflect.DeepEqual(result[4], &TpDestination{Tenant: "tut", Code: "+49151", Tag: "DST_DE_MOBILE"}) {
		t.Error("error loading destinations: ", ToIJSON(result))
	}
}
