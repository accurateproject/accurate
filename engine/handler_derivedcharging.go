package engine

import "github.com/accurateproject/accurate/utils"

// Handles retrieving of DerivedChargers profile based on longest match from AccountingDb
func HandleGetDerivedChargers(ratingStorage RatingStorage, attrs *utils.AttrDerivedChargers) (*utils.DerivedChargerGroup, error) {
	result := &utils.DerivedChargerGroup{}
	if dcsDb, err := ratingStorage.GetDerivedChargers(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject, utils.CACHED); err != nil && err != utils.ErrNotFound {
		return nil, err
	} else if dcsDb != nil {
		for _, dcs := range dcsDb {
			if DerivedChargersMatchesDest(dcs, attrs.Destination) {
				result = dcs
				break
			}
		}
	}
	return result, nil
}

func DerivedChargersMatchesDest(dcs *utils.DerivedChargerGroup, dest string) bool {
	if len(dcs.DestinationIDs) == 0 || dcs.DestinationIDs[utils.ANY] {
		return true
	}
	// check destination ids
	if dests, err := ratingStorage.GetDestinations(dcs.Tenant, dest, "", utils.DestMatching, utils.CACHED); err == nil {
		for _, dest := range dests {
			includeDest, found := dcs.DestinationIDs[dest.Name]
			if found {
				return includeDest
			}
		}
	}
	return false
}
