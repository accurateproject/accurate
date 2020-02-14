package utils

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/accurateproject/accurate/dec"
)

func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", " ")
	return string(b)
}

func TestStructQ(t *testing.T) {
	cl := &StructQ{}
	err := cl.Parse(`{"$or":[{"test":1},{"field":{"$gt":1}},{"best":"coco"}]}`)

	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}

	err = cl.Parse(`{"$has":["NAT","RET","EUR"]}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	err = cl.Parse(`{"Field":7, "Other":true}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	err = cl.Parse(``)
	if err != nil || cl.Complexity() != 0 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
}

func TestStructQKeyValue(t *testing.T) {
	o := struct {
		Test    string
		Field   float64
		Other   bool
		ExpDate time.Time
	}{
		Test:    "test",
		Field:   6.0,
		Other:   true,
		ExpDate: time.Date(2016, 1, 19, 20, 47, 0, 0, time.UTC),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Test":"test"}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6, "Other":true}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":7, "Other":true}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6, "Other":false}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true, "Field":{"$gt":5}}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"$not":[{"Other":true, "Field":{"$gt":5}}]}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true, "Field":{"$gt":7}}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(``)
	if err != nil || cl.Complexity() != 0 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"ExpDate":{"$exp":true}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"ExpDate":{"$exp":false}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"$and":[{"Field":{"$gte":50}},{"Test":{"$eq":"test"}}]}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"WrongFieldName":{"$eq":1}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQKeyValuePointer(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Test":"test"}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":6}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Other":true}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQOperatorValue(t *testing.T) {
	root := &operatorValue{operator: CondGT, value: 3.4}
	if check, err := root.checkStruct(3.5, false); !check.(bool) || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: 3.4}
	if check, err := root.checkStruct(3.5, false); check.(bool) || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: 3.4}
	if check, err := root.checkStruct(3.4, false); !check.(bool) || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
	root = &operatorValue{operator: CondEQ, value: "zinc"}
	if check, err := root.checkStruct("zinc", false); !check.(bool) || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
	root = &operatorValue{operator: CondHAS, value: []interface{}{"NAT", "RET", "EUR"}}
	if check, err := root.checkStruct(StringMap{"WOR": true, "EUR": true, "NAT": true, "RET": true, "ROM": true}, false); !check.(bool) || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(root))
	}
}

func TestStructQKeyStruct(t *testing.T) {
	o := struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Field":{"$gt": 5}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"$gt": 5}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || !strings.HasPrefix(err.Error(), "INVALID_ARGUMENT") {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"$gte": 6}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"$lt": 7}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"$lte": 6}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"$eq": 6}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"$eq": "test"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQKeyStructPointer(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Field":{"$gt": 5}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"$gt": 5}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || !strings.HasPrefix(err.Error(), "INVALID_ARGUMENT") {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"$gte": 6}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"$lt": 7}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"$lte": 6}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Field":{"$eq": 6}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Test":{"$eq": "test"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQOperatorSlice(t *testing.T) {
	o := &struct {
		Test  string
		Field float64
		Other bool
	}{
		Test:  "test",
		Field: 6.0,
		Other: true,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"$or":[{"Test":"test"},{"Field":{"$gt":5}},{"Other":true}]}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"$or":[{"Test":"test1"},{"Field":{"$gt":5}},{"Other":true}]}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"$or":[{"Test":"test"},{"Field":{"$gt":7}},{"Other":false}]}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"$not":[{"$or":[{"Test":"test"},{"Field":{"$gt":7}},{"Other":false}]}]}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"$and":[{"Test":"test"},{"Field":{"$gt":5}},{"Other":true}]}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"$not":[{"$and":[{"Test":"test"},{"Field":{"$gt":5}},{"Other":true}]}]}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"$and":[{"Test":"test"},{"Field":{"$gt":7}},{"Other":false}]}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQMixed(t *testing.T) {
	o := &struct {
		Test       string
		Field      float64
		Categories StringMap
		SliceStuff []string
		Other      bool
	}{
		Test:       "test",
		Field:      6.0,
		Categories: StringMap{"call": true, "data": true, "voice": true},
		SliceStuff: []string{"one", "two"},
		Other:      true,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"$and":[{"Test":"test"},{"Field":{"$gt":5}},{"Other":true},{"Categories":{"$has":["data", "call"]}}, {"SliceStuff":{"$has":["one"]}}]}`)
	if err != nil || cl.Complexity() != 5 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQBalanceType(t *testing.T) {
	type Balance struct {
		Value float64
	}

	o := &struct {
		BalanceType string
		Balance
	}{
		BalanceType: "*monetary",
		Balance:     Balance{Value: 10},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"BalanceType":"*monetary","Value":10}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQRSR(t *testing.T) {
	o := &struct {
		BalanceType string
	}{
		BalanceType: "*monetary",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"BalanceType":{"$rsr":"^*mon"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", !check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"BalanceType":{"$rsr":"^*min"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQIN(t *testing.T) {
	o := &struct {
		BalanceType  string
		Destinations StringMap
		AccountIDs   []string
	}{
		BalanceType:  "*monetary",
		Destinations: NewStringMap("NAT", "RET"),
		AccountIDs:   []string{"acc", "bac"},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"BalanceType":{"$in":["*data", "*sms", "*monetary"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Destinations":{"$in":["NAT", "GER", "RET"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"AccountIDs":{"$in":["bac", "acc"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"BalanceType":{"$in":["*data", "*sms", "*mms"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Destinations":{"$in":["NAT", "GER", "ITA"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"AccountIDs":{"$in":["dac", "hop"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"BalanceType":{"$in":["*data", "*sms", "*monetary"]}, "Destinations":{"$in":["NAT", "GER", "RET"]}, "AccountIDs":{"$in":["bac", "acc"]}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}

}

func TestStructQRE(t *testing.T) {
	o := &struct {
		BalanceType string
	}{
		BalanceType: "*monetary",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"BalanceType":{"$re":"^*mon.+$"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", !check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"BalanceType":{"$re":"^*mon.+x$"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQBTW(t *testing.T) {
	o := &struct {
		Value  float64
		Value1 time.Duration
		Value2 time.Time
	}{
		Value:  7.3,
		Value1: 10 * time.Second,
		Value2: time.Date(2016, 1, 18, 22, 30, 0, 0, time.UTC),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Value":{"$btw":[5,10]}, "Value1":{"$btw":["10s","11s"]}, "Value2":{"$btw":["2016-01-18T22:30:00Z", "2016-01-18T22:30:01Z"]}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"Value":{"$btw":[2,7.2]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQSW(t *testing.T) {
	o := &struct {
		BalanceType string
	}{
		BalanceType: "*monetary",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"BalanceType":{"$sw":"*mon"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"BalanceType":{"$sw":"*min"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQEW(t *testing.T) {
	o := &struct {
		BalanceType string
	}{
		BalanceType: "*monetary",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"BalanceType":{"$ew":"tary"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"BalanceType":{"$ew":"tory"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQGT(t *testing.T) {
	o := &struct {
		V1 float64
		V2 time.Duration
		V3 time.Time
	}{
		V1: 10.0,
		V2: 10 * time.Second,
		V3: time.Date(2016, 1, 18, 22, 30, 0, 0, time.UTC),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"V1":{"$gt":5.0}, "V2":{"$gt":"1s"}, "V3":{"$gt":"2016-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$gt":10.0}, "V2":{"$gt":"1s"}, "V3":{"$gt":"2016-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$gt":5.0}, "V2":{"$gt":"10s"}, "V3":{"$gt":"2016-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$gt":5.0}, "V2":{"$gt":"11s"}, "V3":{"$gt":"2016-01-18T22:30:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQLT(t *testing.T) {
	o := &struct {
		V1 float64
		V2 time.Duration
		V3 time.Time
	}{
		V1: 10.0,
		V2: 10 * time.Second,
		V3: time.Date(2016, 1, 18, 22, 30, 0, 0, time.UTC),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"V1":{"$lt":15.0}, "V2":{"$lt":"11s"}, "V3":{"$lt":"2017-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$lt":10.0}, "V2":{"$lt":"11s"}, "V3":{"$lt":"2017-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$lt":15.0}, "V2":{"$lt":"10s"}, "V3":{"$lt":"2017-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$lt":15.0}, "V2":{"$lt":"11s"}, "V3":{"$lt":"2016-01-18T22:30:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQGTE(t *testing.T) {
	o := &struct {
		V1 float64
		V2 time.Duration
		V3 time.Time
	}{
		V1: 10.0,
		V2: 10 * time.Second,
		V3: time.Date(2016, 1, 18, 22, 30, 0, 0, time.UTC),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"V1":{"$gte":10.0}, "V2":{"$gte":"10s"}, "V3":{"$gte":"2016-01-18T22:30:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$gte":15.0}, "V2":{"$gte":"1s"}, "V3":{"$gte":"2016-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$gte":5.0}, "V2":{"$gte":"11s"}, "V3":{"$gte":"2016-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$gte":5.0}, "V2":{"$gte":"11s"}, "V3":{"$gte":"2017-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQLTE(t *testing.T) {
	o := &struct {
		V1 float64
		V2 time.Duration
		V3 time.Time
	}{
		V1: 10.0,
		V2: 10 * time.Second,
		V3: time.Date(2016, 1, 18, 22, 30, 0, 0, time.UTC),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"V1":{"$lte":10.0}, "V2":{"$lte":"10s"}, "V3":{"$lte":"2016-01-18T22:30:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$lte":15.0}, "V2":{"$lte":"1s"}, "V3":{"$lte":"2016-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$lte":5.0}, "V2":{"$lte":"11s"}, "V3":{"$lte":"2016-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
	err = cl.Parse(`{"V1":{"$lte":5.0}, "V2":{"$lte":"11s"}, "V3":{"$lte":"2017-01-01T00:00:00Z"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || err != nil {
		t.Errorf("Error checking struct: %v %v  (%v)", check, err, toJSON(cl.rootElement))
	}
}

func TestStructQSET(t *testing.T) {
	o := &struct {
		V1 float64
		V2 string
		V3 int
	}{
		V1: 10.0,
		V2: "initial",
		V3: 0,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"V1":{"$set":11.0}, "V2":{"$set": "after"}, "V3":{"$set":1}}`)
	if err != nil || cl.Complexity() != 0 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || err != nil ||
		o.V1 != 11 ||
		o.V2 != "after" ||
		o.V3 != 1 {
		t.Errorf("$set failed: %s %v", ToIJSON(o), err)
	}
	err = cl.Parse(`{"V1":{"$set":11.0}, "V2":{"$set": ""}, "V3":{"$set":1}}`)
	if err != nil || cl.Complexity() != 0 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || err != nil ||
		o.V1 != 11 ||
		o.V2 != "" ||
		o.V3 != 1 {
		t.Errorf("$set failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQCheckAndSET(t *testing.T) {
	o := &struct {
		V1 float64
		V2 string
		V3 int
	}{
		V1: 10.0,
		V2: "initial",
		V3: 0,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"$and":[{"V1":{"$gt":5}}, {"V2":{"$set": "after"}}]}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || err != nil ||
		o.V1 != 10 ||
		o.V2 != "after" {
		t.Errorf("$set failed: %s %v", ToIJSON(o), err)
	}
	o = &struct {
		V1 float64
		V2 string
		V3 int
	}{
		V1: 10.0,
		V2: "initial",
		V3: 0,
	}
	err = cl.Parse(`{"$and":[{"V1":{"$gt":11}}, {"V2":{"$set": "after"}}]}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); check || err != nil ||
		o.V1 != 10 ||
		o.V2 != "initial" {
		t.Errorf("$set failed: %s %v", ToIJSON(o), err)
	}
	o = &struct {
		V1 float64
		V2 string
		V3 int
	}{
		V1: 10.0,
		V2: "initial",
		V3: 0,
	}
	err = cl.Parse(`{"$and":[{"V1":{"$gt":5}}, {"V2":{"$set": "after"}}, {"V2":"after"}]}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || err != nil {
		t.Errorf("$set failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQREPL(t *testing.T) {
	o := &struct {
		Cli string
	}{
		Cli: "+40723045326",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Cli":{"$repl":["\\+4(\\d+)", "004${1}"]}}`)
	if err != nil || cl.Complexity() != 0 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || err != nil ||
		o.Cli != "0040723045326" {
		t.Errorf("$repl failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQEmptyField(t *testing.T) {
	o := &struct {
		Account string
	}{}
	cl := &StructQ{}
	err := cl.Parse(`{"Account":"test"}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check {
		t.Errorf("check failed for empty string: %s %v", ToIJSON(o), err)
	}
}

func TestStructQMultiple(t *testing.T) {
	o := &struct {
		Subject string
		Account string
	}{
		Subject: "s1",
		Account: "initial_value",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Subject":"s1", "Account":{"$repl":["initial_value","changed"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || o.Account != "changed" {
		t.Errorf("check failed for multiple1: %s %v", ToIJSON(o), err)
	}
}

func TestStructQMultiple2(t *testing.T) {
	o := &struct {
		Subject string
		Account string
	}{
		Subject: "s1",
		Account: "initial_value",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Subject":"s1", "Account":{"$repl":["bad_initial_value","changed"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || o.Account != "initial_value" {
		t.Errorf("check failed for multiple2: %s %v", ToIJSON(o), err)
	}
}

func TestStructQMultiple3(t *testing.T) {
	o := &struct {
		Subject string
		Account string
	}{
		Subject: "s1",
		Account: "initial_value",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Subject":"s1", "Account":{"$crepl":["bad_initial_value","changed"]}}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || o.Account != "initial_value" {
		t.Errorf("check failed for multiple3: %s %v", ToIJSON(o), err)
	}
}

func TestStructQExtraFieldSimple(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject:     "s1",
		Account:     "some_value",
		ExtraFields: map[string]string{"niceField": "value"},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"niceField":"value"}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	err = cl.Parse(`{"niceField":"value1"}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQExtraFieldNil(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject: "s1",
		Account: "some_value",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"niceField":"value"}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQExtraFieldComplex(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject:     "s1",
		Account:     "some_value",
		ExtraFields: map[string]string{"niceField": "value"},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"niceField": {"$re":"^v\\w+e$"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	err = cl.Parse(`{"niceField": {"$re":"^x\\w+x$"}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQExtraFieldSet(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject:     "s1",
		Account:     "some_value",
		ExtraFields: map[string]string{"niceField": "value"},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"niceField": {"$set":"another"}}`)
	if err != nil || cl.Complexity() != 0 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || o.ExtraFields["niceField"] != "another" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQExtraFieldReplace(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject:     "s1",
		Account:     "some_value",
		ExtraFields: map[string]string{"niceField": "value"},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"niceField": {"$repl":["^v(\\w+)e$", "c${1}p"]}}`)
	if err != nil || cl.Complexity() != 0 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || o.ExtraFields["niceField"] != "calup" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQUsers(t *testing.T) {
	o := &struct {
		Subject string
		Account string
	}{
		Subject: "s1",
		Account: "*users",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Account": {"$crepl":["\\*users|test", "test"]}, "Subject":{"$crepl":["\\*users|s1", "s1"]}}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, true); !check || o.Account != "test" || o.Subject != "s1" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQUsrRepl(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		Direction   string
		Destination string
	}{
		Subject:     "s1",
		Direction:   "*in",
		Account:     "*users",
		Destination: "+0723045326",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Account": {"$usrpl":["", "test"]}, "Subject":{"$usrpl":["s1", "s2"]}, "Direction":{"$usrpl":["\\*in","*out"]}, "Destination":{"$usrpl":["\\+(\\d+)", "${1}"]}}`)
	if err != nil || cl.Complexity() != 4 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || o.Account != "*users" || o.Subject != "s1" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
	if check, err := cl.Query(o, true); !check || o.Account != "test" || o.Subject != "s2" || o.Direction != "*out" || o.Destination != "0723045326" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQUsrReplNotPass(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		Direction   string
		Destination string
	}{
		Subject:     "s1",
		Direction:   "*in",
		Account:     "*users",
		Destination: "+0723045326",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Account": {"$usrpl":["", "test"]}, "Subject":{"$usrpl":["s2", "s2"]}, "Direction":{"$usrpl":["\\*in","*out"]}, "Destination":{"$usrpl":["\\+(\\d+)", "${1}"]}}`)
	if err != nil || cl.Complexity() != 4 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || o.Account != "*users" || o.Subject != "s1" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQUsersNotPass(t *testing.T) {
	o := &struct {
		Subject string
		Account string
	}{
		Subject: "s1",
		Account: "*users",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Account": {"$crepl":["\\*users|test", "test"]}, "Subject":{"$crepl":["\\*users|s2", "s1"]}}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || o.Account == "test" || o.Subject != "s1" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQSimpleCheckOrSet(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject:     "s1",
		Account:     "some_value",
		ExtraFields: map[string]string{"niceField": "value"},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Subject": "s2", "niceField":"another", "Account":"extra_value"}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check || o.ExtraFields["niceField"] != "value" || o.Subject != "s1" || o.Account != "some_value" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); !check || o.ExtraFields["niceField"] != "another" || o.Subject != "s2" || o.Account != "extra_value" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQUsrNoStar(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject:     "s1",
		Account:     "some_value",
		ExtraFields: map[string]string{"niceField": "value"},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Subject": {"$usr":"s1"}, "niceField":{"$usr":"value"}, "Account":{"$usr":"some_value"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || o.ExtraFields["niceField"] != "value" || o.Subject != "s1" || o.Account != "some_value" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); !check || o.ExtraFields["niceField"] != "value" || o.Subject != "s1" || o.Account != "some_value" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQUsrSomeStar(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject:     USERS,
		Account:     "some_value",
		ExtraFields: map[string]string{"niceField": USERS},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Subject": {"$usr":"s1"}, "niceField":{"$usr":"value"}, "Account":{"$usr":"some_value"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || o.ExtraFields["niceField"] != USERS || o.Subject != USERS || o.Account != "some_value" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); !check || o.ExtraFields["niceField"] != "value" || o.Subject != "s1" || o.Account != "some_value" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQUsrAllStar(t *testing.T) {
	o := &struct {
		Subject     string
		Account     string
		ExtraFields map[string]string
	}{
		Subject:     USERS,
		Account:     USERS,
		ExtraFields: map[string]string{"niceField": USERS},
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Subject": {"$usr":"s1"}, "niceField":{"$usr":"value"}, "Account":{"$usr":"some_value"}}`)
	if err != nil || cl.Complexity() != 3 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || o.ExtraFields["niceField"] != USERS || o.Subject != USERS || o.Account != USERS {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); !check || o.ExtraFields["niceField"] != "value" || o.Subject != "s1" || o.Account != "some_value" {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQRPL(t *testing.T) {
	o := &struct {
		Subject string
		Account string
	}{
		Subject: "s1",
		Account: "x1",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Account": {"$rpl":["x1", "x2"]}, "Subject":{"$rpl":["s1", "s2"]}}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || o.Account != "x1" || o.Subject != "s1" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
	if check, err := cl.Query(o, true); !check || o.Account != "x2" || o.Subject != "s2" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQRPLAny(t *testing.T) {
	o := &struct {
		Subject string
		Account string
	}{
		Subject: "s1",
		Account: "x1",
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Account": {"$rpl":["*any", "x2"]}, "Subject":{"$rpl":["", "s2"]}}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check || o.Account != "x1" || o.Subject != "s1" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
	if check, err := cl.Query(o, true); !check || o.Account != "x2" || o.Subject != "s2" {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQEmptyTrue(t *testing.T) {
	o := &struct {
		Subject  string
		Accounts []string
		Test     StringMap
		X        float64
	}{
		Subject:  " ",
		Accounts: []string{},
		Test:     StringMap{},
		X:        0,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Accounts": {"$empty": true}, "Subject":{"$empty":true}, "Test": {"$empty": true}, "X": {"$empty": true}}`)
	if err != nil || cl.Complexity() != 4 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQEmptyFalse(t *testing.T) {
	o := &struct {
		Subject  string
		Accounts []string
		Test     StringMap
		X        int
	}{
		Subject:  "x",
		Accounts: []string{"x"},
		Test:     StringMap{"x": true},
		X:        1,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Accounts": {"$empty": false}, "Subject":{"$empty":false}, "Test": {"$empty": false}, "X": {"$empty": false}}`)
	if err != nil || cl.Complexity() != 4 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQNoFields(t *testing.T) {
	o := &struct {
		Subject  string
		Accounts []string
		Test     StringMap
		X        int
	}{
		Subject:  "x",
		Accounts: []string{"x"},
		Test:     StringMap{"x": true},
		X:        1,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Accounts1": "test", "RequestType":{"$crepl":["\\*users|\\*postpaid", "\\*postpaid"]}}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); err != nil || check {
		t.Errorf("extrafield check failed: %s %v %v", ToIJSON(o), err, check)
	}
}

func TestStructQDecimal(t *testing.T) {
	o := &struct {
		Subject string
		Account string
		Value   *dec.Dec
	}{
		Subject: "s1",
		Account: "some_value",
		Value:   dec.NewFloat(10),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Value": "20"}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); !check || o.Value.Cmp(dec.NewFloat(20)) != 0 {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQSetDecimal(t *testing.T) {
	o := &struct {
		Subject string
		Account string
		Value   *dec.Dec
	}{
		Subject: "s1",
		Account: "some_value",
		Value:   dec.NewFloat(10),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Value": {"$set": "20"}}`)
	if err != nil || cl.Complexity() != 0 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); !check || o.Value.Cmp(dec.NewFloat(20)) != 0 {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQGtDecimal(t *testing.T) {
	o := &struct {
		Subject string
		Account string
		Value   *dec.Dec
	}{
		Subject: "s1",
		Account: "some_value",
		Value:   dec.NewFloat(10),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Value": {"$gt": 5}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); !check || o.Value.Cmp(dec.NewFloat(10)) != 0 {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQBtwDecimal(t *testing.T) {
	o := &struct {
		Subject string
		Account string
		Value   *dec.Dec
	}{
		Subject: "s1",
		Account: "some_value",
		Value:   dec.NewFloat(10),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Value": {"$btw": [5,15]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); !check || o.Value.Cmp(dec.NewFloat(10)) != 0 {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}
func TestStructQBtwDecimal2(t *testing.T) {
	o := &struct {
		Subject string
		Account string
		Value   *dec.Dec
	}{
		Subject: "s1",
		Account: "some_value",
		Value:   dec.NewFloat(10),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"Value": {"$btw": ["5","10"]}}`)
	if err != nil || cl.Complexity() != 1 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	if check, err := cl.Query(o, true); check || o.Value.Cmp(dec.NewFloat(10)) != 0 {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQTriggerDecimal(t *testing.T) {
	o := &struct {
		TOR            string
		Filter         string
		ThresholdType  string
		ThresholdValue *dec.Dec
	}{
		TOR:            MONETARY,
		Filter:         `{"Directions":{"$has":["*out"]}}`,
		ThresholdType:  TRIGGER_MAX_BALANCE,
		ThresholdValue: dec.NewVal(2, 0),
	}
	cl := &StructQ{}
	err := cl.Parse(`{"ThresholdType":"*max_balance", "ThresholdValue": 2}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

func TestStructQSameField(t *testing.T) {
	o := &struct {
		Age float64
	}{
		Age: 10,
	}
	cl := &StructQ{}
	err := cl.Parse(`{"$and":[{"Age":{"$gt":5}}, {"Age":{"$lt":8}}]}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	err = cl.Parse(`{"$and":[{"Age":{"$gt":15}}, {"Age":{"$lt":11}}]}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
	err = cl.Parse(`{"$and":[{"Age":{"$gt":8}}, {"Age":{"$lt":11}}]}`)
	if err != nil || cl.Complexity() != 2 {
		t.Errorf("Error loading structure: %+v (%v), complexity: %d", toJSON(cl.rootElement), err, cl.Complexity())
	}
	if check, err := cl.Query(o, false); !check {
		t.Errorf("extrafield check failed: %s %v", ToIJSON(o), err)
	}
}

// ****************************** benchmarks ****************************
func BenchmarkBalanceQuery(b *testing.B) {
	type Balance struct {
		Value float64
	}

	o := &struct {
		BalanceType string
		Balance
	}{
		BalanceType: "*monetary",
		Balance:     Balance{Value: 10},
	}
	cl := &StructQ{}
	cl.Parse(`{"BalanceType":"*monetary","Value":10}`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cl.Query(o, false)
	}
}
