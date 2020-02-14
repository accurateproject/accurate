package engine

import (
	"time"

	"github.com/accurateproject/accurate/utils"
)

type Timing struct {
	Tenant    string          `bson:"tenant"`
	Name      string          `bson:"name"`
	Years     utils.Years     `bson:"years"`
	Months    utils.Months    `bson:"months"`
	MonthDays utils.MonthDays `bson:"month_days"`
	WeekDays  utils.WeekDays  `bson:"week_days"`
	Time      string          `bson:"time"`
}

type Rate struct {
	Tenant string      `bson:"tenant"`
	Name   string      `bson:"name"`
	Slots  []*RateSlot `bson:"slots"`
}

type RateSlot struct {
	ConnectFee         float64       `bson:"connect_fee"`
	Rate               float64       `bson:"rate"`
	RateUnit           time.Duration `bson:"rate_unit"`
	RateIncrement      time.Duration `bson:"rate_increment"`
	GroupIntervalStart time.Duration `bson:"group_interval_start"`
}

type DestinationRate struct {
	Tenant   string                             `bson:"tenant"`
	Name     string                             `bson:"name"`
	Bindings map[string]*DestinationRateBinding `bson:"bindings"`
}

type DestinationRateBinding struct {
	DestinationCode string  `bson:"destination_code"`
	DestinationName string  `bson:"destination_name"`
	RateID          string  `bson:"rate_name"`
	MaxCost         float64 `bson:"max_cost"`
	MaxCostStrategy string  `bson:"max_cost_strategy"`
}
