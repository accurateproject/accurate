package engine

import (
	"log"
	"sync"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

type SimpleAccount struct {
	Tenant     string              `bson:"tenant"`
	Name       string              `bson:"name"`
	BalanceMap map[string]*dec.Dec `bson:"balance_map"`
	MaxBalance *dec.Dec            `bson:"max_balance"`
	Disabled   bool                `bson:"disabled"`
}

func (sa *SimpleAccount) debit(category string, value *dec.Dec) {
	b, found := sa.BalanceMap[category]
	if !found {
		b = dec.New()
		sa.BalanceMap[category] = b
	}
	b.SubS(value)
}

func (sa *SimpleAccount) set(category string, value *dec.Dec) {
	b, found := sa.BalanceMap[category]
	if !found {
		b = dec.New()
		sa.BalanceMap[category] = b
	}
	b.Set(value)
}

type SimpleAccounts struct {
	saMap        map[string]*SimpleAccount
	accountingDB AccountingStorage
	sync.RWMutex
}

func NewSimpleAccounts(accountingStorage AccountingStorage) *SimpleAccounts {
	return &SimpleAccounts{
		saMap:        make(map[string]*SimpleAccount),
		accountingDB: accountingStorage,
	}
}

func (sas *SimpleAccounts) LoadFromStorage(tenant string) error {
	sas.Lock()
	defer sas.Unlock()
	if len(sas.saMap) > 0 {
		sas.saMap = make(map[string]*SimpleAccount)
	}
	accIter := sas.accountingDB.Iterator(ColSac, "", map[string]interface{}{"tenant": tenant})
	acc := &SimpleAccount{}
	for accIter.Next(acc) {
		sas.saMap[utils.ConcatKey(acc.Tenant, acc.Name)] = acc
		acc = &SimpleAccount{}
	}
	return accIter.Close()
}

func (sas *SimpleAccounts) Debit(tenant, name, category string, value *dec.Dec) error {
	sas.Lock()
	defer sas.Unlock()
	acc, found := sas.saMap[utils.ConcatKey(tenant, name)]
	if !found {
		return utils.ErrNotFound
	}
	if acc.Disabled {
		return utils.ErrAccountDisabled
	}
	if acc.MaxBalance != nil && acc.MaxBalance.Cmp(dec.Zero) > 0 {
		b, found := acc.BalanceMap[category]
		if found {
			if b.Cmp(dec.Zero) > 0 && b.Cmp(acc.MaxBalance) >= 0 {
				return utils.ErrQuotaExceeded
			}
			x := dec.New().Neg(acc.MaxBalance)
			log.Printf("b: %v, x: %v, b.Cmp(x): %v", b, x, b.Cmp(x))
			if b.Cmp(dec.Zero) < 0 && b.Cmp(x) <= 0 {
				return utils.ErrQuotaExceeded
			}
		}
	}
	acc.debit(category, value)
	return sas.accountingDB.SetSimpleAccount(acc)
}

func (sas *SimpleAccounts) SetValue(tenant, name, category string, value *dec.Dec) error {
	sas.Lock()
	defer sas.Unlock()
	acc, found := sas.saMap[utils.ConcatKey(tenant, name)]
	if !found {
		return utils.ErrNotFound
	}
	if acc.Disabled {
		return utils.ErrAccountDisabled
	}
	acc.set(category, value)
	return sas.accountingDB.SetSimpleAccount(acc)
}

func (sas *SimpleAccounts) SetDisabled(tenant, name string, disabled bool) error {
	sas.Lock()
	defer sas.Unlock()
	acc, found := sas.saMap[utils.ConcatKey(tenant, name)]
	if !found {
		return utils.ErrNotFound
	}
	acc.Disabled = disabled
	return sas.accountingDB.SetSimpleAccount(acc)
}

func (sas *SimpleAccounts) SetMaxBalance(tenant, name string, maxBalance *dec.Dec) error {
	sas.Lock()
	defer sas.Unlock()
	acc, found := sas.saMap[utils.ConcatKey(tenant, name)]
	if !found {
		return utils.ErrNotFound
	}
	if maxBalance != nil {
		if acc.MaxBalance == nil {
			acc.MaxBalance = dec.New()
		}
		acc.MaxBalance.Set(maxBalance)
	} else {
		acc.MaxBalance = nil
	}
	return sas.accountingDB.SetSimpleAccount(acc)
}

func (sas *SimpleAccounts) NewAccount(tenant, name string, disabled bool, maxBalance *dec.Dec) error {
	sas.Lock()
	defer sas.Unlock()
	acc := &SimpleAccount{
		Tenant:     tenant,
		Name:       name,
		BalanceMap: make(map[string]*dec.Dec),
		MaxBalance: maxBalance,
		Disabled:   disabled,
	}
	sas.saMap[utils.ConcatKey(tenant, name)] = acc
	return sas.accountingDB.SetSimpleAccount(acc)
}

func (sas *SimpleAccounts) RemoveAccount(tenant, name string) error {
	sas.Lock()
	defer sas.Unlock()
	delete(sas.saMap, utils.ConcatKey(tenant, name))
	return sas.accountingDB.RemoveSimpleAccount(tenant, name)
}

func (sas *SimpleAccounts) GetAccount(tenant, name string) (*SimpleAccount, error) {
	sas.RLock()
	defer sas.RUnlock()
	acc, found := sas.saMap[utils.ConcatKey(tenant, name)]
	if !found {
		return nil, utils.ErrNotFound
	}
	return acc, nil
}
