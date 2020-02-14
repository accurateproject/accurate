package engine

import (
	"fmt"
	"time"

	"github.com/accurateproject/accurate/utils"
	"github.com/globalsign/mgo/bson"
)

func (ms *MongoStorage) SetSMCost(smc *SMCost) error {
	session, col := ms.conn(utils.TBLSMCosts)
	defer session.Close()
	return col.Insert(smc)
}

func (ms *MongoStorage) GetSMCosts(uniqueid, runid, originHost, originIDPrefix string) (smcs []*SMCost, err error) {
	filter := bson.M{UniqueIDLow: uniqueid, RunIDLow: runid}
	if originIDPrefix != "" {
		filter = bson.M{OriginIDLow: bson.M{"$regex": bson.RegEx{Pattern: fmt.Sprintf("^%s", originIDPrefix)}}, OriginHostLow: originHost, RunIDLow: runid}
	}
	// Execute query
	session, col := ms.conn(utils.TBLSMCosts)
	defer session.Close()
	iter := col.Find(filter).Iter()
	var smCost SMCost
	for iter.Next(&smCost) {
		smcs = append(smcs, &smCost)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return smcs, nil
}

func (ms *MongoStorage) SetCDR(cdr *CDR, update bool) (err error) {
	if cdr.OrderID == 0 {
		cdr.OrderID = time.Now().UnixNano()
	}
	session, col := ms.conn(utils.TBLCDRS)
	defer session.Close()
	if update {
		_, err = col.Upsert(bson.M{UniqueIDLow: cdr.UniqueID, RunIDLow: cdr.RunID}, cdr)
	} else {
		err = col.Insert(cdr)
	}
	return err
}

func (ms *MongoStorage) cleanEmptyFilters(filters bson.M) {
	for k, v := range filters {
		switch value := v.(type) {
		case *int64:
			if value == nil {
				delete(filters, k)
			}
		case *float64:
			if value == nil {
				delete(filters, k)
			}
		case *time.Time:
			if value == nil {
				delete(filters, k)
			}
		case *time.Duration:
			if value == nil {
				delete(filters, k)
			}
		case []string:
			if len(value) == 0 {
				delete(filters, k)
			}
		case bson.M:
			ms.cleanEmptyFilters(value)
			if len(value) == 0 {
				delete(filters, k)
			}
		}
	}
}

//  _, err := col(utils.TBLCDRS).UpdateAll(bson.M{UniqueIDLow: bson.M{"$in": uniqueids}}, bson.M{"$set": bson.M{"deleted_at": time.Now()}})
func (ms *MongoStorage) GetCDRs(qryFltr *utils.CDRsFilter, remove bool) ([]*CDR, int64, error) {
	var minPDD, maxPDD, minUsage, maxUsage *time.Duration
	if len(qryFltr.MinPDD) != 0 {
		if parsed, err := utils.ParseDurationWithSecs(qryFltr.MinPDD); err != nil {
			return nil, 0, err
		} else {
			minPDD = &parsed
		}
	}
	if len(qryFltr.MaxPDD) != 0 {
		if parsed, err := utils.ParseDurationWithSecs(qryFltr.MaxPDD); err != nil {
			return nil, 0, err
		} else {
			maxPDD = &parsed
		}
	}
	if len(qryFltr.MinUsage) != 0 {
		if parsed, err := utils.ParseDurationWithSecs(qryFltr.MinUsage); err != nil {
			return nil, 0, err
		} else {
			minUsage = &parsed
		}
	}
	if len(qryFltr.MaxUsage) != 0 {
		if parsed, err := utils.ParseDurationWithSecs(qryFltr.MaxUsage); err != nil {
			return nil, 0, err
		} else {
			maxUsage = &parsed
		}
	}
	filters := bson.M{
		UniqueIDLow:        bson.M{"$in": qryFltr.UniqueIDs, "$nin": qryFltr.NotUniqueIDs},
		RunIDLow:           bson.M{"$in": qryFltr.RunIDs, "$nin": qryFltr.NotRunIDs},
		OrderIDLow:         bson.M{"$gte": qryFltr.OrderIDStart, "$lt": qryFltr.OrderIDEnd},
		ToRLow:             bson.M{"$in": qryFltr.ToRs, "$nin": qryFltr.NotToRs},
		CDRHostLow:         bson.M{"$in": qryFltr.OriginHosts, "$nin": qryFltr.NotOriginHosts},
		CDRSourceLow:       bson.M{"$in": qryFltr.Sources, "$nin": qryFltr.NotSources},
		RequestTypeLow:     bson.M{"$in": qryFltr.RequestTypes, "$nin": qryFltr.NotRequestTypes},
		DirectionLow:       bson.M{"$in": qryFltr.Directions, "$nin": qryFltr.NotDirections},
		TenantLow:          bson.M{"$in": qryFltr.Tenants, "$nin": qryFltr.NotTenants},
		CategoryLow:        bson.M{"$in": qryFltr.Categories, "$nin": qryFltr.NotCategories},
		AccountLow:         bson.M{"$in": qryFltr.Accounts, "$nin": qryFltr.NotAccounts},
		SubjectLow:         bson.M{"$in": qryFltr.Subjects, "$nin": qryFltr.NotSubjects},
		SupplierLow:        bson.M{"$in": qryFltr.Suppliers, "$nin": qryFltr.NotSuppliers},
		DisconnectCauseLow: bson.M{"$in": qryFltr.DisconnectCauses, "$nin": qryFltr.NotDisconnectCauses},
		SetupTimeLow:       bson.M{"$gte": qryFltr.SetupTimeStart, "$lt": qryFltr.SetupTimeEnd},
		AnswerTimeLow:      bson.M{"$gte": qryFltr.AnswerTimeStart, "$lt": qryFltr.AnswerTimeEnd},
		CreatedAtLow:       bson.M{"$gte": qryFltr.CreatedAtStart, "$lt": qryFltr.CreatedAtEnd},
		UpdatedAtLow:       bson.M{"$gte": qryFltr.UpdatedAtStart, "$lt": qryFltr.UpdatedAtEnd},
		UsageLow:           bson.M{"$gte": minUsage, "$lt": maxUsage},
		PDDLow:             bson.M{"$gte": minPDD, "$lt": maxPDD},
		//CostDetailsLow + "." + AccountLow: bson.M{"$in": qryFltr.RatedAccounts, "$nin": qryFltr.NotRatedAccounts},
		//CostDetailsLow + "." + SubjectLow: bson.M{"$in": qryFltr.RatedSubjects, "$nin": qryFltr.NotRatedSubjects},
	}
	//file, _ := ioutil.TempFile(os.TempDir(), "debug")
	//file.WriteString(fmt.Sprintf("FILTER: %v\n", utils.ToIJSON(qryFltr)))
	//file.WriteString(fmt.Sprintf("BEFORE: %v\n", utils.ToIJSON(filters)))
	ms.cleanEmptyFilters(filters)
	if len(qryFltr.DestinationPrefixes) != 0 {
		var regexpRule string
		for _, prefix := range qryFltr.DestinationPrefixes {
			if len(prefix) == 0 {
				continue
			}
			if len(regexpRule) != 0 {
				regexpRule += "|"
			}
			regexpRule += "^(" + prefix + ")"
		}
		if _, hasIt := filters["$and"]; !hasIt {
			filters["$and"] = make([]bson.M, 0)
		}
		filters["$and"] = append(filters["$and"].([]bson.M), bson.M{DestinationLow: bson.RegEx{Pattern: regexpRule}}) // $and gathers all rules not fitting top level query
	}
	if len(qryFltr.NotDestinationPrefixes) != 0 {
		if _, hasIt := filters["$and"]; !hasIt {
			filters["$and"] = make([]bson.M, 0)
		}
		for _, prefix := range qryFltr.NotDestinationPrefixes {
			if len(prefix) == 0 {
				continue
			}
			filters["$and"] = append(filters["$and"].([]bson.M), bson.M{DestinationLow: bson.RegEx{Pattern: "^(?!" + prefix + ")"}})
		}
	}

	if len(qryFltr.ExtraFields) != 0 {
		var extrafields []bson.M
		for field, value := range qryFltr.ExtraFields {
			extrafields = append(extrafields, bson.M{"extrafields." + field: value})
		}
		filters["$or"] = extrafields
	}

	if len(qryFltr.NotExtraFields) != 0 {
		var extrafields []bson.M
		for field, value := range qryFltr.ExtraFields {
			extrafields = append(extrafields, bson.M{"extrafields." + field: value})
		}
		filters["$not"] = bson.M{"$or": extrafields}
	}

	if qryFltr.MinCost != nil {
		if qryFltr.MaxCost == nil {
			filters[CostLow] = bson.M{"$gte": *qryFltr.MinCost}
		} else if *qryFltr.MinCost == 0.0 && *qryFltr.MaxCost == -1.0 { // Special case when we want to skip errors
			filters["$or"] = []bson.M{
				bson.M{CostLow: bson.M{"$gte": 0.0}},
			}
		} else {
			filters[CostLow] = bson.M{"$gte": *qryFltr.MinCost, "$lt": *qryFltr.MaxCost}
		}
	} else if qryFltr.MaxCost != nil {
		if *qryFltr.MaxCost == -1.0 { // Non-rated CDRs
			filters[CostLow] = 0.0 // Need to include it otherwise all CDRs will be returned
		} else { // Above limited CDRs, since MinCost is empty, make sure we query also NULL cost
			filters[CostLow] = bson.M{"$lt": *qryFltr.MaxCost}
		}
	}
	//file.WriteString(fmt.Sprintf("AFTER: %v\n", utils.ToIJSON(filters)))
	//file.Close()
	session, col := ms.conn(utils.TBLCDRS)
	defer session.Close()
	if remove {
		chgd, err := col.RemoveAll(filters)
		if err != nil {
			return nil, 0, err
		}
		return nil, int64(chgd.Removed), nil
	}
	q := col.Find(filters)
	if qryFltr.Paginator.Limit != nil {
		q = q.Limit(*qryFltr.Paginator.Limit)
	}
	if qryFltr.Paginator.Offset != nil {
		q = q.Skip(*qryFltr.Paginator.Offset)
	}
	if qryFltr.Count {
		cnt, err := q.Count()
		if err != nil {
			return nil, 0, err
		}
		return nil, int64(cnt), nil
	}
	// Execute query
	var cdrs []*CDR
	if qryFltr.Paginator.Limit != nil {
		cdrs = make([]*CDR, *qryFltr.Paginator.Limit)
	} else {
		cdrs = make([]*CDR, 0)
	}
	if err := q.All(&cdrs); err != nil {
		return nil, 0, err
	}
	return cdrs, 0, nil
}
