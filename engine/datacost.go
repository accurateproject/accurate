package engine

// type used for showing sane data cost
type DataCost struct {
	Direction, Category, Tenant, Subject, Account, Destination, TOR string
	Cost                                                            float64
	DataSpans                                                       []*DataSpan
	deductConnectFee                                                bool
}
type DataSpan struct {
	DataStart, DataEnd                                         float64
	Cost                                                       float64
	ratingInfo                                                 *RatingInfo
	RateInterval                                               *RateInterval
	DataIndex                                                  float64 // the data transfer so far till DataEnd
	Increments                                                 []*DataIncrement
	MatchedSubject, MatchedPrefix, MatchedDestId, RatingPlanId string
}

type DataIncrement struct {
	Amount         float64
	Cost           float64
	BalanceInfo    *DebitInfo // need more than one for units with cost
	CompressFactor int
	paid           bool
}
