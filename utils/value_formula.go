package utils

import "time"

//for computing a dynamic value for Value field
type ValueFormula struct {
	Method string
	Args   map[string]interface{}
}

func (vf *ValueFormula) GetValue() float64 {
	formula, exists := ValueFormulas[vf.Method]
	if !exists {
		return 0.0
	}
	return formula(vf.Args)
}

type valueFormula func(map[string]interface{}) float64

const (
	INCREMENTAL = "*incremental"
)

var ValueFormulas = map[string]valueFormula{
	INCREMENTAL: incrementalFormula,
}

func (vf *ValueFormula) String() string {
	return ToJSON(vf)
}

func incrementalFormula(params map[string]interface{}) float64 {
	// check parameters
	unitsInterface, unitsFound := params["Value"]
	intervalInterface, intervalFound := params["Interval"]
	incrementInterface, incrementFound := params["Increment"]

	if !unitsFound || !intervalFound || !incrementFound {
		return 0.0
	}
	units, ok := unitsInterface.(float64)
	if !ok {
		return 0.0
	}
	var interval string
	switch intr := intervalInterface.(type) {
	case string:
		interval = intr
	case []byte:
		interval = string(intr)
	default:
		return 0.0
	}
	var increment string
	switch incr := incrementInterface.(type) {
	case string:
		increment = incr
	case []byte:
		increment = string(incr)
	default:
		return 0.0
	}
	now := time.Now()
	if increment == "day" {
		if interval == "week" {
			return units / 7
		}
		if interval == "month" {
			return units / DaysInMonth(now.Year(), now.Month())
		}
		if interval == "year" {
			return units / DaysInYear(now.Year())
		}
	}
	return 0.0
}
