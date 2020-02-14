package utils

import (
	"reflect"
	"testing"
)

func TestMissingStructFieldsCorrect(t *testing.T) {
	var attr = struct {
		Tenant          string
		Direction       string
		Account         string
		Type            string
		ActionTimingsId string
	}{"bevoip.eu", "OUT", "danconns0001", META_PREPAID, "mama"}
	if missing := MissingStructFields(&attr,
		[]string{"Tenant", "Direction", "Account", "Type", "ActionTimingsId"}); len(missing) != 0 {
		t.Error("Found missing field on correct struct", missing)
	}
}

func TestStructFieldByName(t *testing.T) {
	type TestStruct struct {
		Name    string
		Surname string
		Address string
		Other   string
	}
	ts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	m := FieldByName(ts, NewStringMap("Surname"))
	if m["Surname"] != "2" {
		t.Errorf("expected %v got %v", ts.Surname, m)
	}
	m = FieldByName(*ts, NewStringMap("Address"))
	if m["Address"] != "3" {
		t.Errorf("expected %v got %v", ts.Surname, m)
	}

	m = FieldByName(*ts, NewStringMap("Nonexisting"))
	if len(m) != 0 {
		t.Errorf("expected %v got %v", ts.Surname, m)
	}
}

func TestStructFromMapStringInterface(t *testing.T) {
	ts := &struct {
		Name     string
		Class    *string
		List     []string
		Elements struct {
			Type  string
			Value float64
		}
	}{}
	s := "test2"
	m := map[string]interface{}{
		"Name":  "test1",
		"Class": &s,
		"List":  []string{"test3", "test4"},
		"Elements": struct {
			Type  string
			Value float64
		}{
			Type:  "test5",
			Value: 9.8,
		},
	}
	if err := FromMapStringInterface(m, ts); err != nil {
		t.Logf("ts: %+v", ToJSON(ts))
		t.Error("Error converting map to struct: ", err)
	}
}

func TestStructFromMapStringInterfaceValue(t *testing.T) {
	type T struct {
		Name     string
		Disabled *bool
		Members  []string
	}
	ts := &T{}
	vts := reflect.ValueOf(ts)
	x, err := FromMapStringInterfaceValue(map[string]interface{}{
		"Name":     "test",
		"Disabled": true,
		"Members":  []string{"1", "2", "3"},
	}, vts)
	rt := x.(T)
	if err != nil {
		t.Fatalf("error converting structure value: %v", err)
	}
	if rt.Name != "test" ||
		*rt.Disabled != true ||
		!reflect.DeepEqual(rt.Members, []string{"1", "2", "3"}) {
		t.Errorf("error converting structure value: %s", ToIJSON(rt))
	}
}

func TestStructMerge(t *testing.T) {
	type MergeTest struct {
		ID    *string
		Flag  *bool
		Slice []string
		Map   map[string]int
	}
	a := &MergeTest{
		ID:    StringPointer("orig id"),
		Slice: []string{"orig1", "orig2"},
		Map:   map[string]int{"o": 1},
	}
	b := &MergeTest{
		ID:    StringPointer("other id"),
		Flag:  BoolPointer(true),
		Slice: []string{"orig3"},
		Map:   map[string]int{"x": 2},
	}
	err := Merge(a, b, false)
	if err != nil {
		t.Error("Error merging structs: ", err)
	}
	expected := MergeTest{
		ID:    StringPointer("other id"),
		Flag:  BoolPointer(true),
		Slice: []string{"orig1", "orig2", "orig3"},
		Map:   map[string]int{"o": 1, "x": 2},
	}
	if *a.ID != *expected.ID ||
		*a.Flag != *expected.Flag ||
		!reflect.DeepEqual(a.Slice, expected.Slice) ||
		!reflect.DeepEqual(a.Map, expected.Map) {
		t.Error("Unexpected merge result: ", ToIJSON(a))
	}
}

func TestStructMergeNilDest(t *testing.T) {
	type MergeTest struct {
		ID    *string
		Flag  *bool
		Slice []string
		Map   map[string]int
	}
	a := &MergeTest{}
	b := &MergeTest{
		ID:    StringPointer("other id"),
		Flag:  BoolPointer(true),
		Slice: []string{"orig3"},
		Map:   map[string]int{"x": 2},
	}
	err := Merge(a, b, false)
	if err != nil {
		t.Error("Error merging structs: ", err)
	}
	expected := b
	if *a.ID != *expected.ID ||
		*a.Flag != *expected.Flag ||
		!reflect.DeepEqual(a.Slice, expected.Slice) ||
		!reflect.DeepEqual(a.Map, expected.Map) {
		t.Error("Unexpected merge result: ", ToIJSON(a))
	}
}

func TestStructMergeNilOther(t *testing.T) {
	type MergeTest struct {
		ID    *string
		Flag  *bool
		Slice []string
		Map   map[string]int
	}
	a := &MergeTest{
		ID:    StringPointer("orig id"),
		Slice: []string{"orig1", "orig2"},
		Map:   map[string]int{"o": 1},
	}
	b := &MergeTest{}
	err := Merge(a, b, true)
	if err != nil {
		t.Error("Error merging structs: ", err)
	}
	expected := a
	if *a.ID != *expected.ID ||
		a.Flag != expected.Flag ||
		!reflect.DeepEqual(a.Slice, expected.Slice) ||
		!reflect.DeepEqual(a.Map, expected.Map) {
		t.Error("Unexpected merge result: ", ToIJSON(a))
	}
}
