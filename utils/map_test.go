package utils

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStringMapParse(t *testing.T) {
	sm := ParseStringMap("1;2;3;4", FALLBACK_SEP)
	if len(sm) != 4 {
		t.Error("Error pasring map: ", sm)
	}
}

func TestStringMapParseNegative(t *testing.T) {
	sm := ParseStringMap("1;2;!3;4", FALLBACK_SEP)
	if len(sm) != 4 {
		t.Error("Error pasring map: ", sm)
	}
	if sm["3"] != false {
		t.Error("Error parsing negative: ", sm)
	}
}

func TestStringMapCompare(t *testing.T) {
	sm := ParseStringMap("1;2;!3;4", FALLBACK_SEP)
	if include, found := sm["2"]; include != true && found != true {
		t.Error("Error detecting positive: ", sm)
	}
	if include, found := sm["3"]; include != false && found != true {
		t.Error("Error detecting negative: ", sm)
	}
	if include, found := sm["5"]; include != false && found != false {
		t.Error("Error detecting missing: ", sm)
	}
}

func TestMapMergeMapsStringIface(t *testing.T) {
	mp1 := map[string]interface{}{
		"Hdr1": "Val1",
		"Hdr2": "Val2",
		"Hdr3": "Val3",
	}
	mp2 := map[string]interface{}{
		"Hdr3": "Val4",
		"Hdr4": "Val4",
	}
	eMergedMap := map[string]interface{}{
		"Hdr1": "Val1",
		"Hdr2": "Val2",
		"Hdr3": "Val4",
		"Hdr4": "Val4",
	}
	if mergedMap := MergeMapsStringIface(mp1, mp2); !reflect.DeepEqual(eMergedMap, mergedMap) {
		t.Errorf("Expecting: %+v, received: %+v", eMergedMap, mergedMap)
	}
}

func TestStringMapEqual(t *testing.T) {
	t1 := NewStringMap("val1")
	t2 := NewStringMap("val2")
	result := t1.Equal(t2)
	expected := false
	if result != expected {
		t.Error("Expecting:", expected, ", received:", result)
	}
}

func TestStringMapIsEmpty(t *testing.T) {
	t1 := NewStringMap("val1")
	result := t1.IsEmpty()
	expected := false
	if result != expected {
		t.Error("Expecting:", expected, ", received:", result)
	}
}

func TestStringMapUnmarshallJSONMap(t *testing.T) {
	data := []byte(`{"*out":true, "NAT":true}`)
	x := StringMap{}
	err := json.Unmarshal(data, &x)
	if err != nil {
		t.Fatal(err)
	}
	if x["*out"] != true || x["NAT"] != true {
		t.Errorf("failed to unmarshall string map: %+v", x)
	}
}

func TestStringMapUnmarshallJSONSlice(t *testing.T) {
	data := []byte(`["*out","NAT"]`)
	x := StringMap{}
	err := json.Unmarshal(data, &x)
	if err != nil {
		t.Fatal(err)
	}
	if x["*out"] != true || x["NAT"] != true {
		t.Errorf("failed to unmarshall string map: %+v", x)
	}
}

func TestStringMapUnmarshallJSONString(t *testing.T) {
	data := []byte(`"*out,NAT"`)
	x := StringMap{}
	err := json.Unmarshal(data, &x)
	if err != nil {
		t.Fatal(err)
	}
	if x["*out"] != true || x["NAT"] != true {
		t.Errorf("failed to unmarshall string map: %+v", x)
	}
}

func TestStringMapUnmarshallJSONStringSingle(t *testing.T) {
	data := []byte(`"*out"`)
	x := StringMap{}
	err := json.Unmarshal(data, &x)
	if err != nil {
		t.Fatal(err)
	}
	if x["*out"] != true {
		t.Errorf("failed to unmarshall string map: %+v", x)
	}
}

func TestStringMapUnmarshallJSONInStruct(t *testing.T) {
	data := []byte(`{"Destinations":"*out"}`)
	x := struct {
		Destinations StringMap
	}{}
	err := json.Unmarshal(data, &x)
	if err != nil {
		t.Fatal(err)
	}
	if x.Destinations["*out"] != true {
		t.Errorf("failed to unmarshall string map: %+v", x)
	}
}
