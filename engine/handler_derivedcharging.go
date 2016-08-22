package engine

import "github.com/accurateproject/accurate/utils"

// Handles retrieving of DerivedChargers profile based on longest match from AccountingDb
func HandleGetDerivedChargers(ratingStorage RatingStorage, attrs *utils.AttrDerivedChargers) (*utils.DerivedChargers, error) {
	dcs := &utils.DerivedChargers{}
	strictKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, attrs.Subject)
	anySubjKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, attrs.Account, utils.ANY)
	anyAcntKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, attrs.Category, utils.ANY, utils.ANY)
	anyCategKey := utils.DerivedChargersKey(attrs.Direction, attrs.Tenant, utils.ANY, utils.ANY, utils.ANY)
	anyTenantKey := utils.DerivedChargersKey(attrs.Direction, utils.ANY, utils.ANY, utils.ANY, utils.ANY)
	for _, dcKey := range []string{strictKey, anySubjKey, anyAcntKey, anyCategKey, anyTenantKey} {
		if dcsDb, err := ratingStorage.GetDerivedChargers(dcKey, false); err != nil && err != utils.ErrNotFound {
			return nil, err
		} else if dcsDb != nil && DerivedChargersMatchesDest(dcsDb, attrs.Destination) {
			dcs = dcsDb
			break
		}
	}
	return dcs, nil
}

func DerivedChargersMatchesDest(dcs *utils.DerivedChargers, dest string) bool {
	if len(dcs.DestinationIDs) == 0 || dcs.DestinationIDs[utils.ANY] {
		return true
	}
	// check destination ids
	for _, p := range utils.SplitPrefix(dest, MIN_PREFIX_MATCH) {
		if destIDs, err := ratingStorage.GetReverseDestination(p, false); err == nil {
			for _, dId := range destIDs {
				includeDest, found := dcs.DestinationIDs[dId]
				if found {
					return includeDest
				}
			}
		}
	}
	return false
}
