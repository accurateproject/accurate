package engine

import (
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/accurateproject/accurate/utils"
)

const (
	MINE_PREFIX           = "*mine_"
	STRATEGY_MINE_LOWEST  = "*mine_lowest"
	STRATEGY_MINE_HIGHEST = "*mine_highest"
	STRATEGY_MINE_RANDOM  = "*mine_random"
	STRATEGY_LOWEST       = "*lowest"
	STRATEGY_HIGHEST      = "*highest"
	STRATEGY_RANDOM       = "*random"
)

type SharedGroup struct {
	Tenant            string                   `bson:"tenant"`
	Name              string                   `bson:"name"`
	AccountParameters map[string]*SharingParam `bson:"account_parameters"`
	MemberIDs         utils.StringMap          `bson:"member_i_ds"`
}

type SharingParam struct {
	Strategy      string `bson:"strategy"`
	RatingSubject string `bson:"rating_subject"`
}

func (sg *SharedGroup) SortBalancesByStrategy(myBalance *Balance, bc Balances) Balances {
	sharingParameters := sg.AccountParameters[utils.ANY]
	if sp, hasParamsForAccount := sg.AccountParameters[myBalance.account.Name]; hasParamsForAccount {
		sharingParameters = sp
	}

	strategy := STRATEGY_MINE_RANDOM
	if sharingParameters != nil && sharingParameters.Strategy != "" {
		strategy = sharingParameters.Strategy
	}
	switch strategy {
	case STRATEGY_LOWEST, STRATEGY_MINE_LOWEST:
		sort.Sort(LowestBalancesSorter(bc))
	case STRATEGY_HIGHEST, STRATEGY_MINE_HIGHEST:
		sort.Sort(HighestBalancesSorter(bc))
	case STRATEGY_RANDOM, STRATEGY_MINE_RANDOM:
		rbc := RandomBalancesSorter(bc)
		(&rbc).Sort()
		bc = Balances(rbc)
	default: // use mine random for anything else
		strategy = STRATEGY_MINE_RANDOM
		rbc := RandomBalancesSorter(bc)
		(&rbc).Sort()
		bc = Balances(rbc)
	}
	if strings.HasPrefix(strategy, MINE_PREFIX) {
		// find index of my balance
		index := 0
		for i, b := range bc {
			if b.UUID == myBalance.UUID {
				index = i
				break
			}
		}
		// move my balance first
		bc[0], bc[index] = bc[index], bc[0]
	}
	return bc
}

// Returns all shared group's balances collected from user accounts'
func (sg *SharedGroup) GetBalances(destination, category, direction, balanceType string, ub *Account) (bc Balances) {
	//	if len(sg.members) == 0 {
	for ubId := range sg.MemberIDs {
		var nUb *Account
		if ubId == ub.Name { // skip the initiating user
			nUb = ub
		} else {
			nUb, _ = accountingStorage.GetAccount(sg.Tenant, ubId)
			if nUb == nil || nUb.Disabled {
				continue
			}
		}
		//sg.members = append(sg.members, nUb)
		sb := nUb.getBalancesForPrefix(destination, category, direction, balanceType, sg.Name)
		bc = append(bc, sb...)
	}
	/*	} else {
		for _, m := range sg.members {
			sb := m.getBalancesForPrefix(destination, m.BalanceMap[balanceType], sg.Id)
			bc = append(bc, sb...)
		}
	}*/
	return
}

type LowestBalancesSorter []*Balance

func (lbcs LowestBalancesSorter) Len() int {
	return len(lbcs)
}

func (lbcs LowestBalancesSorter) Swap(i, j int) {
	lbcs[i], lbcs[j] = lbcs[j], lbcs[i]
}

func (lbcs LowestBalancesSorter) Less(i, j int) bool {
	return lbcs[i].GetValue().Cmp(lbcs[j].GetValue()) < 0
}

type HighestBalancesSorter []*Balance

func (hbcs HighestBalancesSorter) Len() int {
	return len(hbcs)
}

func (hbcs HighestBalancesSorter) Swap(i, j int) {
	hbcs[i], hbcs[j] = hbcs[j], hbcs[i]
}

func (hbcs HighestBalancesSorter) Less(i, j int) bool {
	return hbcs[i].GetValue().Cmp(hbcs[j].GetValue()) > 0
}

type RandomBalancesSorter []*Balance

func (rbcs *RandomBalancesSorter) Sort() {
	src := *rbcs
	// randomize balance chain
	dest := make([]*Balance, len(src))
	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}
	*rbcs = dest
}
