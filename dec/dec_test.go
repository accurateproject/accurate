package dec

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestDecRound(t *testing.T) {
	price := NewVal(2985, 4)
	if price.Round(3).Cmp(NewVal(298, 3)) != 0 {
		t.Error("round failed", price)
	}
}

func TestDecRoundBig(t *testing.T) {
	orig := NewVal(2985, 4)
	price := NewVal(2985, 4)
	if price.Round(6).Cmp(orig) != 0 {
		t.Error("round failed", price)
	}
}

func TestDecSetString(t *testing.T) {
	price := New()
	_, err := price.SetString("0.2985")
	if err != nil || price.Cmp(NewVal(2985, 4)) != 0 {
		t.Error("set string failed", price)
	}
}

func TestDecOne(t *testing.T) {
	price := NewVal(1, 0)
	one := New()
	one.SetFloat64(1.0)
	if price.Cmp(one) != 0 {
		t.Error("set float failed", price)
	}
}

func TestDecJSON(t *testing.T) {
	price := NewVal(2985, 4)
	b, err := json.Marshal(price)
	if err != nil {
		t.Fatal(err)
	}
	result := New()
	if err := json.Unmarshal(b, result); err != nil {
		t.Fatal(err)
	}
	if price.Cmp(result) != 0 {
		t.Error("JSON failed", result)
	}
}

func TestDecJSONStruct(t *testing.T) {
	price := &struct {
		Value *Dec
	}{Value: NewVal(2985, 4)}
	b, err := json.Marshal(price)
	if err != nil {
		t.Fatal(err)
	}
	result := &struct {
		Value *Dec
	}{Value: New()}
	if err := json.Unmarshal(b, result); err != nil {
		t.Fatal(err)
	}
	if price.Value.Cmp(result.Value) != 0 {
		t.Error("BSON failed", result)
	}
}

func TestDecBSONStruct(t *testing.T) {
	price := &struct {
		Value *Dec
	}{Value: NewVal(2985, 4)}
	b, err := bson.Marshal(price)
	if err != nil {
		t.Fatal(err)
	}
	result := &struct {
		Value *Dec
	}{}
	if err := bson.Unmarshal(b, result); err != nil {
		t.Fatal(err)
	}
	if price.Value.Cmp(result.Value) != 0 {
		t.Error("BSON failed", result)
	}
}

func TestDecComp(t *testing.T) {
	d := New()
	if !d.IsZero() {
		t.Error("error comp: ", d)
	}
	if !d.LteZero() {
		t.Error("error comp: ", d)
	}
	if !d.GteZero() {
		t.Error("error comp: ", d)
	}
	d.Set(NewVal(-1, 2))
	if d.IsZero() {
		t.Error("error comp: ", d)
	}
	if !d.LtZero() {
		t.Error("error comp: ", d)
	}
	if !d.LteZero() {
		t.Error("error comp: ", d)
	}
	if d.GtZero() {
		t.Error("error comp: ", d)
	}
	if d.GteZero() {
		t.Error("error comp: ", d)
	}
	d.Set(NewVal(1, 2))
	if d.IsZero() {
		t.Error("error comp: ", d)
	}
	if d.LtZero() {
		t.Error("error comp: ", d)
	}
	if d.LteZero() {
		t.Error("error comp: ", d)
	}
	if !d.GtZero() {
		t.Error("error comp: ", d)
	}
	if !d.GteZero() {
		t.Error("error comp: ", d)
	}
}

func TestDecDeepEqual(t *testing.T) {
	x1 := NewVal(70, 0)
	x2 := NewVal(70, 0)
	if !reflect.DeepEqual(x1, x2) {
		t.Error("deepequal does not work!")
	}
}

func TestDecDeepEqualFloat(t *testing.T) {
	x1 := NewVal(70, 0)
	x2 := NewFloat(70)
	if !reflect.DeepEqual(x1, x2) {
		t.Error("deepequal does not work: ", x1, x2)
	}
}

func TestDecCmpNeg(t *testing.T) {
	x := NewVal(-40000000, 6)
	y := NewVal(-20, 0)
	if x.Cmp(y) != -1 {
		t.Errorf("error in cmp x=%v, y=%v, x.Cmp(y)=%v", x, y, x.Cmp(y))
	}
}

func TestDecRoundMode(t *testing.T) {
	x := New()
	if _, err := x.SetString("1.2019"); err != nil {
		t.Fatal(err)
	}
	if x.Round(4).String() != "1.202" {
		t.Error("error rounding: ", x.Round(4).String())
	}
}

func TestDecInt64(t *testing.T) {
	x := NewVal(10240000000000, 0).MulS(NewFloat(0.000976563))
	if x.Int64() != 10000005120 {
		t.Error("error int64: ", x.Int64())
	}
}

//**************** bench ************************

func BenchmarkBigLen0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		price := 0.2985
		var unit float64 = 60
		result := price / unit
		var xunit float64 = 60
		_ = result * xunit
	}
}

func BenchmarkBigLen1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		price := NewVal(2985, 4)
		unit := NewVal(60, 0)
		result := New().Quo(price, unit)
		xunit := NewVal(60, 0)
		_ = New().Mul(result, xunit).String()
	}
}
