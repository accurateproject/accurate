package utils

import (
	"reflect"
	"testing"
)

func TestReflectFieldAsStringOnStruct(t *testing.T) {
	mystruct := struct {
		Title       string
		Count       int
		Count64     int64
		Val         float64
		ExtraFields map[string]interface{}
	}{"Title1", 5, 6, 7.3, map[string]interface{}{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3}}
	if strVal, err := ReflectFieldAsString(mystruct, "Title", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "Title1" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "Count", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "5" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "Count64", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "6" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "Val", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "7.3" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "a", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "Title2" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "b", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "15" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "c", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "16" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(mystruct, "d", "ExtraFields"); err != nil {
		t.Error(err)
	} else if strVal != "17.3" {
		t.Errorf("Received: %s", strVal)
	}
}

func TestReflectFieldAsStringOnMap(t *testing.T) {
	myMap := map[string]interface{}{"Title": "Title1", "Count": 5, "Count64": int64(6), "Val": 7.3,
		"a": "Title2", "b": 15, "c": int64(16), "d": 17.3}
	if strVal, err := ReflectFieldAsString(myMap, "Title", ""); err != nil {
		t.Error(err)
	} else if strVal != "Title1" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "Count", ""); err != nil {
		t.Error(err)
	} else if strVal != "5" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "Count64", ""); err != nil {
		t.Error(err)
	} else if strVal != "6" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "Val", ""); err != nil {
		t.Error(err)
	} else if strVal != "7.3" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "a", ""); err != nil {
		t.Error(err)
	} else if strVal != "Title2" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "b", ""); err != nil {
		t.Error(err)
	} else if strVal != "15" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "c", ""); err != nil {
		t.Error(err)
	} else if strVal != "16" {
		t.Errorf("Received: %s", strVal)
	}
	if strVal, err := ReflectFieldAsString(myMap, "d", ""); err != nil {
		t.Error(err)
	} else if strVal != "17.3" {
		t.Errorf("Received: %s", strVal)
	}
}

func TestReflectAsMapStringIface(t *testing.T) {
	mystruct := struct {
		Title       string
		Count       int
		Count64     int64
		Val         float64
		ExtraFields map[string]interface{}
	}{"Title1", 5, 6, 7.3, map[string]interface{}{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3}}
	expectOutMp := map[string]interface{}{
		"Title":       "Title1",
		"Count":       5,
		"Count64":     int64(6),
		"Val":         7.3,
		"ExtraFields": map[string]interface{}{"a": "Title2", "b": 15, "c": int64(16), "d": 17.3},
	}
	if outMp, err := AsMapStringIface(mystruct); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectOutMp, outMp) {
		t.Errorf("Expecting: %+v, received: %+v", expectOutMp, outMp)
	}
}
