package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Detects missing field values based on mandatory field names, s should be a pointer to a struct
func MissingStructFields(s interface{}, mandatories []string) []string {
	missing := []string{}
	for _, fieldName := range mandatories {
		fld := reflect.ValueOf(s).Elem().FieldByName(fieldName)
		// sanitize the string fields before checking
		if fld.Kind() == reflect.String && fld.CanSet() {
			fld.SetString(strings.TrimSpace(fld.String()))
		}
		if (fld.Kind() == reflect.String && fld.String() == "") ||
			((fld.Kind() == reflect.Slice || fld.Kind() == reflect.Map) && fld.Len() == 0) ||
			(fld.Kind() == reflect.Int && fld.Int() == 0) {
			missing = append(missing, fieldName)
		}
	}
	return missing
}

func FieldByName(in interface{}, fieldNames StringMap) map[string]interface{} {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	m := make(map[string]interface{})
	for fieldName := range fieldNames {
		x := v.FieldByName(fieldName)
		if x.IsValid() && x.CanInterface() {
			m[fieldName] = x.Interface()
		}
	}
	return m
}

// Detects nonempty struct fields, s should be a pointer to a struct
// Useful to not overwrite db fields with non defined params in api
func NonemptyStructFields(s interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	numField := reflect.ValueOf(s).Elem().NumField()
	for i := 0; i < numField; i++ {
		fld := reflect.ValueOf(s).Elem().Field(i)
		switch fld.Kind() {
		case reflect.Bool:
			fields[reflect.TypeOf(s).Elem().Field(i).Name] = fld.Bool()
		case reflect.Int:
			fieldVal := fld.Int()
			if fieldVal != 0 {
				fields[reflect.TypeOf(s).Elem().Field(i).Name] = fieldVal
			}
		case reflect.String:
			fieldVal := fld.String()
			if fieldVal != "" {
				fields[reflect.TypeOf(s).Elem().Field(i).Name] = fieldVal
			}
		}
	}
	return fields
}

// Converts a struct to map[string]interface{}
func ToMapMapStringInterface(in interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		out[v.Type().Field(i).Name] = v.Field(i).Interface()
	}
	return out
}

func FromMapStringInterface(m map[string]interface{}, in interface{}) error {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for fieldName, fieldValue := range m {
		field := v.FieldByName(fieldName)
		if field.IsValid() {
			if !field.IsValid() || !field.CanSet() {
				continue
			}
			structFieldType := field.Type()
			val := reflect.ValueOf(fieldValue)
			if structFieldType != val.Type() {
				return errors.New("Provided value type didn't match obj field type")
			}
			field.Set(val)
		}
	}
	return nil
}

// initial intent was to use it with *cgr_rpc but does not handle slice and structure fields
func FromMapStringInterfaceValue(m map[string]interface{}, v reflect.Value) (interface{}, error) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for fieldName, fieldValue := range m {
		field := v.FieldByName(fieldName)
		if field.IsValid() {
			if !field.IsValid() || !field.CanSet() {
				continue
			}
			val := reflect.ValueOf(fieldValue)
			structFieldType := field.Type()
			if structFieldType.Kind() == reflect.Ptr {
				field.Set(reflect.New(field.Type().Elem()))
				field = field.Elem()
			}
			structFieldType = field.Type()
			if structFieldType != val.Type() {
				return nil, fmt.Errorf("provided value type didn't match obj field type: %v vs %v (%v vs %v)", structFieldType, val.Type(), structFieldType.Kind(), val.Type().Kind())
			}
			field.Set(val)
		}
	}
	return v.Interface(), nil
}

// Update struct with map fields, returns not matching map keys, s is a struct to be updated
func UpdateStructWithStrMap(s interface{}, m map[string]string) []string {
	notMatched := []string{}
	elem := reflect.ValueOf(s).Elem()
	for key, val := range m {
		fld := elem.FieldByName(key)
		if fld.IsValid() {
			switch fld.Kind() {
			case reflect.Bool:
				if valBool, err := strconv.ParseBool(val); err != nil {
					notMatched = append(notMatched, key)
				} else {
					fld.SetBool(valBool)
				}
			case reflect.Int:
				if valInt, err := strconv.ParseInt(val, 10, 64); err != nil {
					notMatched = append(notMatched, key)
				} else {
					fld.SetInt(valInt)
				}
			case reflect.String:
				fld.SetString(val)
			}
		} else {
			notMatched = append(notMatched, key)
		}
	}
	return notMatched
}

// Merges two objects appending to slices and maps and overwritting non nil fields
func Merge(dest, other interface{}, overwriteDefault bool) error {
	if dest == nil || other == nil {
		return nil
	}
	destType := reflect.TypeOf(dest)
	if destType != reflect.TypeOf(other) {
		return errors.New("different types")
	}
	destElem := reflect.ValueOf(dest)
	if destElem.Kind() == reflect.Ptr {
		destElem = destElem.Elem()
	}

	otherElem := reflect.ValueOf(other)
	if otherElem.Kind() == reflect.Ptr {
		otherElem = otherElem.Elem()
	}

	if !otherElem.IsValid() {
		return nil
	}
	switch destElem.Kind() {
	case reflect.Slice:
		destElem.Set(reflect.AppendSlice(destElem, otherElem))
		return nil
	case reflect.Map:
		for _, key := range otherElem.MapKeys() {
			destElem.SetMapIndex(key, otherElem.MapIndex(key))
		}
		return nil
	}

	for i := 0; i < destElem.NumField(); i++ {
		destField := destElem.Field(i)
		if !otherElem.IsValid() {
			continue
		}
		otherField := otherElem.Field(i)

		if otherField.Kind() == reflect.Ptr && otherField.IsNil() {
			continue
		}
		if destField.Kind() == reflect.Ptr && destField.IsNil() {
			destField.Set(otherField)
			continue
		}
		switch destField.Kind() {
		case reflect.Ptr:
			destField.Set(otherField)
		case reflect.Slice:
			if otherField.Len() > 0 {
				if overwriteDefault {
					destField.Set(otherField) // overwrite so we loose the defaults
				} else {
					destField.Set(reflect.AppendSlice(destField, otherField))
				}
			}
		case reflect.Map:
			for _, key := range otherField.MapKeys() {
				if destField.IsNil() {
					destField.Set(reflect.MakeMap(otherField.Type()))
				}
				destField.SetMapIndex(key, otherField.MapIndex(key))
			}
		}
	}
	return nil
}
