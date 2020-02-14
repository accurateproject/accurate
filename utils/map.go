package utils

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Converts map[string]string into map[string]interface{}
func ConvertMapValStrIf(inMap map[string]string) map[string]interface{} {
	outMap := make(map[string]interface{})
	for field, val := range inMap {
		outMap[field] = val
	}
	return outMap
}

// Mirrors key/val
func MirrorMap(mapIn map[string]string) map[string]string {
	mapOut := make(map[string]string, len(mapIn))
	for key, val := range mapIn {
		mapOut[val] = key
	}
	return mapOut
}

// Returns mising keys in a map
func MissingMapKeys(inMap map[string]string, requiredKeys []string) []string {
	missingKeys := []string{}
	for _, reqKey := range requiredKeys {
		if val, hasKey := inMap[reqKey]; !hasKey || val == "" {
			missingKeys = append(missingKeys, reqKey)
		}
	}
	return missingKeys
}

// Return map keys
func MapKeys(m map[string]string) []string {
	n := make([]string, len(m))
	i := 0
	for k := range m {
		n[i] = k
		i++
	}
	return n
}

type StringMap map[string]bool

func NewStringMap(s ...string) StringMap {
	result := make(StringMap)
	if len(s) == 0 {
		return result
	}
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v != "" {
			if strings.HasPrefix(v, "!") {
				result[v[1:]] = false
			} else {
				result[v] = true
			}
		}
	}
	return result
}

func ParseStringMap(s, sep string) StringMap {
	if s == ZERO || s == "" {
		return make(StringMap)
	}
	return StringMapFromSlice(strings.Split(s, sep))
}

func (sm StringMap) Add(val string) {
	sm[val] = true
}

func (sm StringMap) Equal(om StringMap) bool {
	if sm == nil && om != nil {
		return false
	}
	if len(sm) != len(om) {
		return false
	}
	for key := range sm {
		if !om[key] {
			return false
		}
	}
	return true
}

func (sm StringMap) Includes(om StringMap) bool {
	if len(sm) < len(om) {
		return false
	}
	for key := range om {
		if !sm[key] {
			return false
		}
	}
	return true
}

func (sm StringMap) Slice() []string {
	result := make([]string, len(sm))
	i := 0
	for k := range sm {
		result[i] = k
		i++
	}
	return result
}

func (sm StringMap) IsEmpty() bool {
	return sm == nil ||
		len(sm) == 0 ||
		sm[ANY] == true
}

func StringMapFromSlice(s []string) StringMap {
	result := make(StringMap, len(s))
	if len(s) == 0 {
		return result
	}
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v != "" {
			if strings.HasPrefix(v, "!") {
				result[v[1:]] = false
			} else {
				result[v] = true
			}
		}
	}
	return result
}

func (sm StringMap) Copy(o StringMap) {
	for k, v := range o {
		sm[k] = v
	}
}

func (sm StringMap) CopySlice(s []string) {
	for _, k := range s {
		sm[k] = true
	}
}

func (sm StringMap) Clone() StringMap {
	result := make(StringMap, len(sm))
	result.Copy(sm)
	return result
}

func (sm StringMap) String() string {
	return strings.Join(sm.Slice(), ",")
}

func (sm StringMap) GetOne() string {
	for key := range sm {
		return key
	}
	return ""
}

func (sm StringMap) Join(mps ...StringMap) {
	for _, mp := range mps {
		for k, v := range mp {
			sm[k] = v
		}
	}
}

func (sm *StringMap) UnmarshalJSON(data []byte) (err error) {
	data = bytes.TrimSpace(data)
	switch data[0] {
	case []byte("[")[0]:
		s := make([]string, 0)
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*sm = StringMapFromSlice(s)

	case []byte("{")[0]:
		x := make(map[string]bool)
		if err := json.Unmarshal(data, &x); err != nil {
			return err
		}
		*sm = StringMap(x)
	default: // plain string comma separated
		var stringData string
		if err := json.Unmarshal(data, &stringData); err != nil {
			return err
		}
		x := ParseStringMap(stringData, ",")
		*sm = x
	}

	return
}

// Used to merge multiple maps (eg: output of struct having ExtraFields)
func MergeMapsStringIface(mps ...map[string]interface{}) (outMp map[string]interface{}) {
	outMp = make(map[string]interface{})
	for i, mp := range mps {
		if i == 0 {
			outMp = mp
			continue
		}
		for k, v := range mp {
			outMp[k] = v
		}
	}
	return

}
