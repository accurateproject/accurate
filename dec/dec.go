package dec

import (
	"bytes"
	"errors"
	"sort"

	"github.com/ericlagergren/decimal"
	"github.com/globalsign/mgo/bson"
)

var (
	Zero     = New()         // zero value
	MinusOne = NewVal(-1, 0) // -1 value
	z        = new(decimal.Big)
)

type Dec struct {
	*decimal.Big
}

func New() *Dec {
	return &Dec{&decimal.Big{}}
}

func NewVal(value int64, scale int32) *Dec {
	return &Dec{decimal.New(value, scale)}
}

func NewFloat(v float64) *Dec {
	return New().SetFloat64(v)
}

// GetBSON implements bson.Getter.
func (d *Dec) GetBSON() (interface{}, error) {
	if d == nil {
		d = New()
	}
	return d.String(), nil
}

// SetBSON implements bson.Setter.
func (d *Dec) SetBSON(raw bson.Raw) error {
	var s string
	if err := raw.Unmarshal(&s); err != nil {
		return err
	}
	if d.Big == nil {
		d.Big = new(decimal.Big)
	}
	if _, err := d.SetString(s); err != nil {
		return err
	}
	return nil
}

func (d *Dec) UnmarshalJSON(data []byte) (err error) {
	return d.UnmarshalBinary(data)
}

func (d *Dec) MarshalJSON() ([]byte, error) {
	x, err := d.MarshalText()
	return bytes.Trim(x, `"`), err
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface. As a string representation
// is already used when encoding to text, this method stores that string as []byte
func (d *Dec) UnmarshalBinary(data []byte) error {
	if d == nil {
		*d = Dec{new(decimal.Big)}
	}
	if d.Big == nil {
		d.Big = new(decimal.Big)
	}
	return d.Big.UnmarshalText(data)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (d *Dec) MarshalBinary() (data []byte, err error) {
	if d.Big == nil {
		d.Big = new(decimal.Big)
	}
	return d.Big.MarshalText()
}

func (d *Dec) SubS(o *Dec) *Dec {
	d.Big.Sub(d.Big, o.Big)
	return d
}

func (d *Dec) AddS(o *Dec) *Dec {
	d.Big.Add(d.Big, o.Big)
	return d
}

func (d *Dec) Sub(o1, o2 *Dec) *Dec {
	d.Big.Sub(o1.Big, o2.Big)
	return d
}

func (d *Dec) Add(o1, o2 *Dec) *Dec {
	d.Big.Add(o1.Big, o2.Big)
	return d
}

func (d *Dec) Set(o *Dec) *Dec {
	d.Big.Set(o.Big)
	return d
}

func (d *Dec) Neg(o *Dec) *Dec {
	d.Big.Neg(o.Big)
	return d
}

func (d *Dec) Round(n int32) *Dec {
	d.Big.Round(n)
	return d
}

func (d *Dec) Cmp(o *Dec) int {
	return d.Big.Cmp(o.Big)
}

func (d *Dec) SetFloat64(v float64) *Dec {
	d.Big.SetFloat64(v)
	return d
}

func (d *Dec) SetString(s string) (*Dec, error) {
	var err error
	if _, ok := d.Big.SetString(s); !ok {
		err = errors.New("malformed decimal: " + s)
	}
	return d, err
}

func (d *Dec) MulS(o *Dec) *Dec {
	d.Big.Mul(d.Big, o.Big)
	return d
}

func (d *Dec) QuoS(o *Dec) *Dec {
	d.Big.Quo(d.Big, o.Big)
	return d
}

func (d *Dec) Mul(o1, o2 *Dec) *Dec {
	d.Big.Mul(o1.Big, o2.Big)
	return d
}

func (d *Dec) Quo(o1, o2 *Dec) *Dec {
	d.Big.Quo(o1.Big, o2.Big)
	return d
}

func (d *Dec) IsZero() bool {
	return d.Big.Cmp(z) == 0
}

func (d *Dec) GtZero() bool {
	return d.Big.Cmp(z) > 0
}

func (d *Dec) GteZero() bool {
	return d.Big.Cmp(z) >= 0
}

func (d *Dec) LtZero() bool {
	return d.Big.Cmp(z) < 0
}

func (d *Dec) LteZero() bool {
	return d.Big.Cmp(z) <= 0
}

type DecSlice []*Dec

func (ds DecSlice) Sort() {
	sort.Slice(ds, func(j, i int) bool {
		return ds[i].Cmp(ds[j]) < 0
	})
}
