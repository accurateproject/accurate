package utils

/*
The query syntax is a json encoded string similar to mongodb query language.

Examples:
- {"Weight":{"*gt":50}} checks for a balance with weight greater than 50
- {"$or":[{"Value":{"$eq":0}},{"Value":{"$gte":100}}] checks for a balance with value equal to 0 or equal or highr than 100

Available operators:
- $eq: equal
- $empty: empty
- $gt: greater than
- $gte: greater or equal than
- $lt: less then
- $lte: less or equal than
- $btw: between [>=,<]
- $exp: expired
- $or: logical or
- $and: logical and
- $not: logical not
- $has: receives a list of elements and checks that the elements are present in the specified field (StringMap type)
- $in: receives a list of elements and check that the field is one of the elements
- $re: regular expression
- $sw: starts with
- $ew: ends with
- $set: sets a static value to the field (true if success)
- $rpl: replaces a static value to the field if original equals the given static value
- $repl: replaces using regular expression groups (true)
- $crepl: same as repl but returns false if no match
- $usr: checks if the fields is *users/empty/or given value and sets the value if passes
- $usrpl: checks if the fields is *users/empty/or matches given value and replaces the value if passes

Equal (*eq) and local and (*and) operators are implicit for shortcuts. In this way:

{"*and":[{"Value":{"*eq":3}},{"Weight":{"*eq":10}}]} is equivalent to: {"Value":3, "Weight":10}.
*/

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/accurateproject/accurate/dec"
)

const (
	Operator  = "$"
	CondEQ    = "$eq"
	CondEMPTY = "$empty"
	CondGT    = "$gt"
	CondGTE   = "$gte"
	CondLT    = "$lt"
	CondLTE   = "$lte"
	CondEXP   = "$exp"
	CondOR    = "$or"
	CondAND   = "$and"
	CondNOT   = "$not"
	CondHAS   = "$has"
	CondIN    = "$in"
	CondRE    = "$re"
	CondBTW   = "$btw"
	CondSW    = "$sw"
	CondEW    = "$ew"
	CondRSR   = "$rsr"
	CondSET   = "$set"
	CondRPL   = "$rpl"
	CondREPL  = "$repl"
	CondCREPL = "$crepl"
	CondUSR   = "$usr"
	CondUSRPL = "$usrpl"
)

func NewErrInvalidArgument(arg interface{}) error {
	return fmt.Errorf("INVALID_ARGUMENT: %v", arg)
}

var (
	operatorMap = map[string]func(field, value interface{}) (interface{}, error){
		CondEQ: func(field, value interface{}) (interface{}, error) {
			switch f := field.(type) {
			case *dec.Dec:
				x := dec.New()
				switch v := value.(type) {
				case string:
					if _, err := x.SetString(v); err != nil {
						return nil, err
					}
				case float64:
					x.SetFloat64(v)
				}
				return f.Cmp(x) == 0, nil
			}
			return value == field, nil
		},
		CondSET: func(field, value interface{}) (interface{}, error) {
			return value, nil
		},

		CondUSR: func(field, value interface{}) (interface{}, error) {
			if field != USERS && field != "" && field != value {
				return false, nil
			}
			return value, nil
		},
		CondEMPTY: func(field, value interface{}) (interface{}, error) {
			valueBool, ok := value.(bool)
			if !ok {
				return false, NewErrInvalidArgument(value)
			}
			switch f := field.(type) {
			case string:
				return (strings.TrimSpace(f) == "") == valueBool, nil
			case StringMap:
				return (len(f) == 0) == valueBool, nil
			case map[string]string:
				return (len(f) == 0) == valueBool, nil
			case []string:
				return (len(f) == 0) == valueBool, nil
			case float64:
				return (f == 0) == valueBool, nil
			case int:
				return (f == 0) == valueBool, nil
			case int32:
				return (f == 0) == valueBool, nil
			case *dec.Dec:
				return (f.Cmp(dec.Zero) == 0) == valueBool, nil
			}

			return false, nil
		},
		CondRPL: func(field, value interface{}) (interface{}, error) {
			var fieldStr string
			var ok bool
			if fieldStr, ok = field.(string); !ok {
				return false, NewErrInvalidArgument(field)
			}
			var valueSlice []interface{}
			if valueSlice, ok = value.([]interface{}); !ok || len(valueSlice) != 2 {
				return false, NewErrInvalidArgument(value)
			}
			var valueSearch, valueReplace string
			if valueSearch, ok = valueSlice[0].(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			if valueReplace, ok = valueSlice[1].(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			if valueSearch != ANY && valueSearch != "" && fieldStr != valueSearch {
				return false, nil
			}
			return valueReplace, nil
		},
		CondGT: func(field, value interface{}) (interface{}, error) {
			var ok bool
			switch f := field.(type) {
			case float64:
				var vf float64
				if vf, ok = value.(float64); !ok {
					return false, NewErrInvalidArgument(value)
				}
				return f > vf, nil
			case time.Time:
				var vf string
				if vf, ok = value.(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				t, err := ParseTimeDetectLayout(vf, time.UTC.String()) // must be converted on use
				if err != nil {
					return false, err
				}
				return f.After(t), nil
			case time.Duration:
				var vf string
				if vf, ok = value.(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				d, err := ParseDurationWithSecs(vf)
				if err != nil {
					return false, err
				}
				return f > d, nil
			case *dec.Dec:
				x := dec.New()
				switch v := value.(type) {
				case string:
					if _, err := x.SetString(v); err != nil {
						return nil, err
					}
				case float64:
					x.SetFloat64(v)
				}
				return f.Cmp(x) > 0, nil
			default:
				return false, NewErrInvalidArgument(value)
			}
		},
		CondGTE: func(field, value interface{}) (interface{}, error) {
			var ok bool
			switch f := field.(type) {
			case float64:
				var vf float64
				if vf, ok = value.(float64); !ok {
					return false, NewErrInvalidArgument(value)
				}
				return f >= vf, nil
			case time.Time:
				var vf string
				if vf, ok = value.(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				t, err := ParseTimeDetectLayout(vf, time.UTC.String()) // must be converted
				if err != nil {
					return false, err
				}
				return f.After(t) || f.Equal(t), nil
			case time.Duration:
				var vf string
				if vf, ok = value.(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				d, err := ParseDurationWithSecs(vf)
				if err != nil {
					return false, err
				}
				return f >= d, nil
			case *dec.Dec:
				x := dec.New()
				switch v := value.(type) {
				case string:
					if _, err := x.SetString(v); err != nil {
						return nil, err
					}
				case float64:
					x.SetFloat64(v)
				}
				return f.Cmp(x) >= 0, nil
			default:
				return false, NewErrInvalidArgument(value)
			}
		},
		CondLT: func(field, value interface{}) (interface{}, error) {
			var ok bool
			switch f := field.(type) {
			case float64:
				var vf float64
				if vf, ok = value.(float64); !ok {
					return false, NewErrInvalidArgument(value)
				}
				return f < vf, nil
			case time.Time:
				var vf string
				if vf, ok = value.(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				t, err := ParseTimeDetectLayout(vf, time.UTC.String()) // must be converted
				if err != nil {
					return false, err
				}
				return f.Before(t), nil
			case time.Duration:
				var vf string
				if vf, ok = value.(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				d, err := ParseDurationWithSecs(vf)
				if err != nil {
					return false, err
				}
				return f < d, nil
			case *dec.Dec:
				x := dec.New()
				switch v := value.(type) {
				case string:
					if _, err := x.SetString(v); err != nil {
						return nil, err
					}
				case float64:
					x.SetFloat64(v)
				}
				return f.Cmp(x) < 0, nil
			default:
				return false, NewErrInvalidArgument(value)
			}
		},
		CondLTE: func(field, value interface{}) (interface{}, error) {
			var ok bool
			switch f := field.(type) {
			case float64:
				var vf float64
				if vf, ok = value.(float64); !ok {
					return false, NewErrInvalidArgument(value)
				}
				return f <= vf, nil
			case time.Time:
				var vf string
				if vf, ok = value.(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				t, err := ParseTimeDetectLayout(vf, time.UTC.String()) // must be converted
				if err != nil {
					return false, err
				}
				return f.Before(t) || f.Equal(t), nil
			case time.Duration:
				var vf string
				if vf, ok = value.(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				d, err := ParseDurationWithSecs(vf)
				if err != nil {
					return false, err
				}
				return f <= d, nil
			case *dec.Dec:
				x := dec.New()
				switch v := value.(type) {
				case string:
					if _, err := x.SetString(v); err != nil {
						return nil, err
					}
				case float64:
					x.SetFloat64(v)
				}
				return f.Cmp(x) <= 0, nil
			default:
				return false, NewErrInvalidArgument(value)
			}
		},
		CondBTW: func(field, value interface{}) (interface{}, error) {
			var ok bool
			var vf []interface{}
			if vf, ok = value.([]interface{}); !ok || len(vf) != 2 {
				return false, NewErrInvalidArgument(value)
			}
			switch f := field.(type) {
			case float64:
				var vf1, vf2 float64
				if vf1, ok = vf[0].(float64); !ok {
					return false, NewErrInvalidArgument(value)
				}
				if vf2, ok = vf[1].(float64); !ok {
					return false, NewErrInvalidArgument(value)
				}

				return ((f >= vf1) && (f < vf2)), nil
			case time.Time:
				var vf1, vf2 string
				if vf1, ok = vf[0].(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				if vf2, ok = vf[1].(string); !ok {
					return false, NewErrInvalidArgument(value)
				}

				t1, err := ParseTimeDetectLayout(vf1, time.UTC.String()) // must be converted
				if err != nil {
					return false, err
				}
				t2, err := ParseTimeDetectLayout(vf2, time.UTC.String()) // must be converted
				if err != nil {
					return false, err
				}
				return (f.After(t1) || f.Equal(t1)) && f.Before(t2), nil
			case time.Duration:
				var vf1, vf2 string
				if vf1, ok = vf[0].(string); !ok {
					return false, NewErrInvalidArgument(value)
				}
				if vf2, ok = vf[1].(string); !ok {
					return false, NewErrInvalidArgument(value)
				}

				d1, err := ParseDurationWithSecs(vf1)
				if err != nil {
					return false, err
				}
				d2, err := ParseDurationWithSecs(vf2)
				if err != nil {
					return false, err
				}
				return ((f >= d1) && (f < d2)), nil
			case *dec.Dec:
				x1 := dec.New()
				x2 := dec.New()
				switch v := vf[0].(type) {
				case string:
					if _, err := x1.SetString(v); err != nil {
						return nil, err
					}
				case float64:
					x1.SetFloat64(v)
				}
				switch v := vf[1].(type) {
				case string:
					if _, err := x2.SetString(v); err != nil {
						return nil, err
					}
				case float64:
					x2.SetFloat64(v)
				}
				return ((f.Cmp(x1) >= 0) && (f.Cmp(x2) < 0)), nil
			default:
				return false, NewErrInvalidArgument(value)
			}
		},
		CondEXP: func(field, value interface{}) (interface{}, error) {
			var expDate time.Time
			var ok bool
			if expDate, ok = field.(time.Time); !ok {
				return false, NewErrInvalidArgument(field)
			}
			var expired bool
			if expired, ok = value.(bool); !ok {
				return false, NewErrInvalidArgument(value)
			}
			if expired { // check for expiration
				return !expDate.IsZero() && expDate.Before(time.Now()), nil
			} else { // check not expired
				return expDate.IsZero() || expDate.After(time.Now()), nil
			}
		},
		CondSW: func(field, value interface{}) (interface{}, error) {
			var of, vf string
			var ok bool
			if of, ok = field.(string); !ok {
				return false, NewErrInvalidArgument(field)
			}
			if vf, ok = value.(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			return strings.HasPrefix(of, vf), nil
		},
		CondEW: func(field, value interface{}) (interface{}, error) {
			var of, vf string
			var ok bool
			if of, ok = field.(string); !ok {
				return false, NewErrInvalidArgument(field)
			}
			if vf, ok = value.(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			return strings.HasSuffix(of, vf), nil
		},
		CondRE: func(field, value interface{}) (interface{}, error) {
			var fieldStr string
			var ok bool
			if fieldStr, ok = field.(string); !ok {
				return false, NewErrInvalidArgument(field)
			}
			var re string
			if re, ok = value.(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			validValue, err := regexp.Compile(re)
			if err != nil {
				return false, err
			}
			return validValue.MatchString(fieldStr), nil
		},
		CondHAS: func(field, value interface{}) (interface{}, error) {
			var ok bool
			var strSlice []interface{}
			if strSlice, ok = value.([]interface{}); !ok {
				return false, NewErrInvalidArgument(value)
			}
			switch f := field.(type) {
			case StringMap:
				for _, str := range strSlice {
					if !f[str.(string)] {
						return false, nil
					}
				}
			case []string:
				for _, str := range strSlice {
					x := NewStringMap(f...)
					if !x[str.(string)] {
						return false, nil
					}
				}
			}

			return true, nil
		},
		CondCREPL: func(field, value interface{}) (interface{}, error) {
			var fieldStr string
			var ok bool
			if fieldStr, ok = field.(string); !ok {
				return false, NewErrInvalidArgument(field)
			}
			var valueSlice []interface{}
			if valueSlice, ok = value.([]interface{}); !ok || len(valueSlice) != 2 {
				return false, NewErrInvalidArgument(value)
			}
			var valueSearch, valueReplace string
			if valueSearch, ok = valueSlice[0].(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			if valueReplace, ok = valueSlice[1].(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			re, err := regexp.Compile(valueSearch) // TODO: cache this
			if err != nil {
				return false, err
			}
			match := re.FindStringSubmatchIndex(fieldStr)
			if match == nil {
				return false, nil
			}
			res := []byte{}
			res = re.ExpandString(res, valueReplace, fieldStr, match)
			return string(res), nil
		},
		CondREPL: func(field, value interface{}) (interface{}, error) {
			var fieldStr string
			var ok bool
			if fieldStr, ok = field.(string); !ok {
				return false, NewErrInvalidArgument(field)
			}
			var valueSlice []interface{}
			if valueSlice, ok = value.([]interface{}); !ok || len(valueSlice) != 2 {
				return false, NewErrInvalidArgument(value)
			}
			var valueSearch, valueReplace string
			if valueSearch, ok = valueSlice[0].(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			if valueReplace, ok = valueSlice[1].(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			re, err := regexp.Compile(valueSearch)
			if err != nil {
				return false, err
			}
			match := re.FindStringSubmatchIndex(fieldStr)
			if match == nil {
				return fieldStr, nil
			}
			res := []byte{}
			res = re.ExpandString(res, valueReplace, fieldStr, match)
			return string(res), nil
		},
		CondUSRPL: func(field, value interface{}) (interface{}, error) {
			var fieldStr string
			var ok bool
			if fieldStr, ok = field.(string); !ok {
				return false, NewErrInvalidArgument(field)
			}
			var valueSlice []interface{}
			if valueSlice, ok = value.([]interface{}); !ok || len(valueSlice) != 2 {
				return false, NewErrInvalidArgument(value)
			}
			var valueSearch, valueReplace string
			if valueSearch, ok = valueSlice[0].(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			if valueReplace, ok = valueSlice[1].(string); !ok {
				return false, NewErrInvalidArgument(value)
			}
			var re *regexp.Regexp
			var err error
			if valueSearch != "" {
				re, err = regexp.Compile("\\*users|^$|" + valueSearch)
			} else {
				re, err = regexp.Compile("\\*users|^$")
			}
			if err != nil {
				return false, err
			}
			match := re.FindStringSubmatchIndex(fieldStr)
			if match == nil {
				return false, nil
			}
			res := []byte{}
			res = re.ExpandString(res, valueReplace, fieldStr, match)
			return string(res), nil
		},
		CondIN: func(field, value interface{}) (interface{}, error) {
			var elemSlice []interface{}
			var ok bool
			if elemSlice, ok = value.([]interface{}); !ok {
				return false, NewErrInvalidArgument(value)
			}
			switch f := field.(type) {
			case []string:
				for _, fieldEl := range f {
					found := false
					for _, el := range elemSlice {
						if fieldEl == el {
							found = true
							break
						}
					}
					if !found {
						return false, nil
					}

				}
			case StringMap:
				for fieldEl := range f {
					found := false
					for _, el := range elemSlice {
						if fieldEl == el {
							found = true
							break
						}
					}
					if !found {
						return false, nil
					}
				}
			default:
				for _, el := range elemSlice {
					if f == el {
						return true, nil
					}
				}
				return false, nil
			}

			return true, nil
		},
		CondRSR: func(field, value interface{}) (interface{}, error) {
			fltr, err := NewRSRFilter(value.(string))
			if err != nil {
				return false, err
			}
			return fltr.Pass(fmt.Sprintf("%v", field)), nil
		},
	}
)

type compositeElement interface {
	element
	addChild(element)
}

type element interface {
	checkStruct(o interface{}, change bool) (interface{}, error)
	complexity() int
}

type operatorSlice struct {
	operator string
	slice    []element
}

func (os *operatorSlice) addChild(ce element) {
	os.slice = append(os.slice, ce)
}

func (os *operatorSlice) complexity() (cx int) {
	for _, child := range os.slice {
		cx += child.complexity()
	}
	return
}

func (os *operatorSlice) checkStruct(o interface{}, change bool) (interface{}, error) {
	switch os.operator {
	case CondOR:
		for _, cond := range os.slice {
			check, err := cond.checkStruct(o, change)
			if err != nil {
				return false, err
			}
			if c, ok := check.(bool); ok {
				if c == true { // exit early
					return c, nil
				}
			}
		}
		return false, nil
	case CondAND, CondNOT:
		accumulator := true
		for _, cond := range os.slice {
			check, err := cond.checkStruct(o, change)
			if err != nil {
				return false, err
			}
			if c, ok := check.(bool); ok {
				if os.operator == CondAND && c == false {
					return false, nil // exit early for and
				}
				accumulator = accumulator && c
			}
		}
		if os.operator == CondAND {
			return accumulator, nil
		}
		return !accumulator, nil
	}
	return false, nil
}

type keyStruct struct {
	key  string
	elem element
}

func (ks *keyStruct) addChild(ce element) {
	ks.elem = ce
}

func (ks *keyStruct) complexity() (cx int) {
	return ks.elem.complexity()
}

func (ks *keyStruct) checkStruct(o interface{}, change bool) (interface{}, error) {
	obj := reflect.ValueOf(o)
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	value := obj.FieldByName(ks.key)
	var extraFields map[string]string
	var err error
	if !value.IsValid() {
		// try in extra fields
		extraFields, err = getExtraFields(obj, ks.key)
		if err != nil {
			return false, err
		}
		if extraFields == nil {
			return false, nil // no such field
		}
	}
	if extraFields == nil && value.CanInterface() {
		x, err := ks.elem.checkStruct(value.Interface(), change)
		if !change {
			if e, ok := x.(bool); ok {
				return e, err
			}
			// value to be set
			return true, err
		}

		switch e := x.(type) {
		case bool:
			return e, err
		case string:
			if value.CanSet() {
				if value.Kind() == reflect.String {
					value.SetString(e)
					return true, nil
				}
				if value.Kind() == reflect.Ptr {
					x := dec.New()
					if _, err := x.SetString(e); err != nil {
						return nil, err
					}
					value.Set(reflect.ValueOf(x))
					return true, nil
				}
			}
			return false, NewErrInvalidArgument(fmt.Sprintf("cannot set %s to %v", ks.key, e))
		case float64:
			if value.CanSet() {
				if value.Kind() == reflect.Float64 {
					value.SetFloat(e)
				}
				if value.Kind() == reflect.Int {
					value.SetInt(int64(e))
				}
				if value.Kind() == reflect.Ptr {
					x := dec.NewFloat(e)
					value.Set(reflect.ValueOf(x))
					return true, nil
				}

				return true, nil
			}
			return false, NewErrInvalidArgument(fmt.Sprintf("cannot set %s to %v", ks.key, e))
		default:
			return false, NewErrInvalidArgument(o)
		}
	} else {
		v, ok := extraFields[ks.key]
		if !ok {
			return false, NewErrInvalidArgument(ks.key)
		}
		x, err := ks.elem.checkStruct(v, change)
		if !change {
			if e, ok := x.(bool); ok {
				return e, err
			}
			// value to be set
			return true, err
		}
		switch e := x.(type) {
		case bool:
			return e, err
		case string:
			extraFields[ks.key] = e
			return true, nil
		default:
			return false, NewErrInvalidArgument(fmt.Sprintf("cannot set %s to %v", ks.key, e))
		}
	}
}

type operatorValue struct {
	operator string
	value    interface{}
}

func (ov *operatorValue) checkStruct(o interface{}, change bool) (interface{}, error) {
	if f, ok := operatorMap[ov.operator]; ok {
		op := ov.operator
		if !change && (op == CondSET || op == CondREPL) { // if is changing operator and change is false
			return true, nil
		}
		return f(o, ov.value)
	}
	return false, nil
}

func (ov *operatorValue) complexity() (cx int) {
	if ov.operator == CondSET || ov.operator == CondREPL {
		return 0
	}
	return 1
}

type keyValue struct {
	key   string
	value interface{}
}

func (kv *keyValue) complexity() (cx int) {
	return 1
}

func (kv *keyValue) checkStruct(o interface{}, change bool) (interface{}, error) {
	obj := reflect.ValueOf(o)
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	value := obj.FieldByName(kv.key)
	if !value.IsValid() {
		// try in extra fields too

		extraFields, err := getExtraFields(obj, kv.key)
		if err != nil {
			return false, err
		}
		v, ok := extraFields[kv.key]
		if !ok {
			return false, nil // nu such fields
		}
		if !change {
			return v == kv.value, nil
		}
		switch e := kv.value.(type) {
		case string:
			extraFields[kv.key] = e
			return true, nil
		default:
			return false, NewErrInvalidArgument(fmt.Sprintf("cannot set %s to %v", kv.key, e))
		}
	}
	if !change {
		valInterface := value.Interface()
		switch f := valInterface.(type) {
		case *dec.Dec:
			x := dec.New()
			switch v := kv.value.(type) {
			case string:
				if _, err := x.SetString(v); err != nil {
					return nil, err
				}
			case float64:
				x.SetFloat64(v)
			}
			return f.Cmp(x) == 0, nil
		}
		return valInterface == kv.value, nil
	}
	switch e := kv.value.(type) {
	case bool:
		if value.CanSet() && value.Kind() == reflect.Bool {
			value.SetBool(e)
			return true, nil
		}
		return false, NewErrInvalidArgument(fmt.Sprintf("cannot set %s to %v", kv.key, e))
	case string:
		if value.CanSet() {
			if value.Kind() == reflect.String {
				value.SetString(e)
				return true, nil
			}
			if value.Kind() == reflect.Ptr {
				x := dec.New()
				if _, err := x.SetString(e); err != nil {
					return nil, err
				}
				value.Set(reflect.ValueOf(x))
				return true, nil
			}
		}
		return false, NewErrInvalidArgument(fmt.Sprintf("cannot set %s to %v", kv.key, e))
	case float64:
		if value.CanSet() {
			if value.Kind() == reflect.Float64 {
				value.SetFloat(e)
			}
			if value.Kind() == reflect.Int {
				value.SetInt(int64(e))
			}
			if value.Kind() == reflect.Ptr {
				x := dec.NewFloat(e)
				value.Set(reflect.ValueOf(x))
				return true, nil
			}
			return true, nil
		}
		return false, NewErrInvalidArgument(fmt.Sprintf("cannot set %s to %v", kv.key, e))
	default:
		return false, NewErrInvalidArgument(o)
	}
}

type trueElement struct{}

func (te *trueElement) checkStruct(o interface{}, change bool) (interface{}, error) {
	return true, nil
}

func (te *trueElement) complexity() (cx int) {
	return 0
}

func isOperator(s string) bool {
	return strings.HasPrefix(s, Operator)
}

func notEmpty(x interface{}) bool {
	return !reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

type StructQ struct {
	complexity  int
	rootElement element
}

func getExtraFields(obj reflect.Value, key string) (map[string]string, error) {
	value := obj.FieldByName("ExtraFields")
	if !value.IsValid() || value.Kind() != reflect.Map {
		return nil, nil // no extra fields
	}
	mi := value.Interface()
	m, ok := mi.(map[string]string)
	if !ok {
		return nil, NewErrInvalidArgument(key)
	}
	return m, nil
}

func NewStructQ(q string) (sm *StructQ, err error) {
	sm = &StructQ{}
	err = sm.Parse(q)
	if err != nil {
		sm = nil
	}
	return
}

func (sm *StructQ) load(a map[string]interface{}, parentElement compositeElement) (element, error) {
	for key, value := range a {
		var currentElement element
		switch t := value.(type) {
		case []interface{}:
			if key == CondHAS || key == CondIN || key == CondBTW || key == CondREPL || key == CondCREPL || key == CondUSRPL || key == CondRPL {
				currentElement = &operatorValue{operator: key, value: t}
			} else {
				currentElement = &operatorSlice{operator: key}
				for _, e := range t {
					if el, err := sm.load(e.(map[string]interface{}), currentElement.(compositeElement)); err != nil {
						return el, err
					}
				}
			}
		case map[string]interface{}:
			currentElement = &keyStruct{key: key}
			//log.Print("map: ", t)
			if el, err := sm.load(t, currentElement.(compositeElement)); err != nil {
				return el, err
			}
		case interface{}:
			if isOperator(key) {
				currentElement = &operatorValue{operator: key, value: t}
			} else {
				currentElement = &keyValue{key: key, value: t}
			}
			//log.Print("generic interface: ", t)
		default:
			return nil, ErrParserError
		}
		if parentElement != nil { // normal recurrent action
			parentElement.addChild(currentElement)
		} else {
			if len(a) > 1 { // we have more keys in the map
				parentElement = &operatorSlice{operator: CondAND}
				parentElement.addChild(currentElement)
			} else { // it was only one key value
				return currentElement, nil
			}
		}
	}
	return parentElement, nil
}

func (sm *StructQ) Parse(s string) (err error) {
	a := make(map[string]interface{})
	if len(s) != 0 {
		if err := json.Unmarshal([]byte(s), &a); err != nil {
			return err
		}
		sm.rootElement, err = sm.load(a, nil)
	} else {
		sm.rootElement = &trueElement{}
	}
	sm.complexity = sm.rootElement.complexity()
	return
}

func (sm *StructQ) Query(o interface{}, change bool) (bool, error) {
	if sm == nil || sm.rootElement == nil {
		return false, ErrParserError
	}
	x, err := sm.rootElement.checkStruct(o, change)
	if c, ok := x.(bool); ok {
		return c, err
	}
	return false, NewErrInvalidArgument(x)
}

func (sm *StructQ) Complexity() int {
	return sm.complexity
}
