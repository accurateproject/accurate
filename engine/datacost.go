package engine

import "github.com/accurateproject/accurate/dec"

// type used for showing sane data cost
type DataCost struct {
	Direction, Category, Tenant, Subject, Account, Destination, TOR string
	Cost                                                            *dec.Dec
	DataSpans                                                       []*DataSpan
	deductConnectFee                                                bool
}
type DataSpan struct {
	DataStart, DataEnd                                         float64
	Cost                                                       *dec.Dec
	ratingInfo                                                 *RatingInfo
	RateInterval                                               *RateInterval
	DataIndex                                                  float64 // the data transfer so far till DataEnd
	Increments                                                 *DataIncrements
	MatchedSubject, MatchedPrefix, MatchedDestID, RatingPlanID string
}

type DataIncrements struct {
	CompIncrement *DataIncrement
}

type DataIncrement struct {
	Amount         *dec.Dec
	Cost           *dec.Dec
	BalanceInfo    *DebitInfo // need more than one for units with cost
	CompressFactor int
	paid           int
}
