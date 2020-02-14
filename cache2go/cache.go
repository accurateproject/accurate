//Simple caching library with expiration capabilities
package cache2go

import (
	"sync"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
)

const (
	PREFIX_LEN = 4
	KIND_ADD   = "ADD"
	KIND_REM   = "REM"
	KIND_PRF   = "PRF"
	COMMIT     = "commit"
)

var (
	mux         sync.RWMutex
	cfg         *config.Cache
	tenantCache = make(map[string]cacheStore)
	// transaction stuff
	transactionBuffer = make(map[string][]*transactionItem)
	transactionMux    sync.Mutex
)

type transactionItem struct {
	tenant string
	key    string
	value  interface{}
	kind   string
}

func init() {
	NewCache(nil)
}

func c(tenant string) cacheStore {
	c, ok := tenantCache[tenant]
	if !ok {
		c = newLruStore()
		tenantCache[tenant] = c
	}
	return c
}

func NewCache(cacheCfg *config.Cache) {
	cfg = cacheCfg
}

func BeginTransaction() string {
	transID := utils.GenUUID()
	transactionMux.Lock()
	defer transactionMux.Unlock()
	transactionBuffer[transID] = make([]*transactionItem, 0)
	return transID
}

func RollbackTransaction(transID string) {
	transactionMux.Lock()
	defer transactionMux.Unlock()
	if transactionBuffer != nil {
		delete(transactionBuffer, transID)
	}
}

func CommitTransaction(transID string) {
	transactionMux.Lock()
	defer transactionMux.Unlock()
	if transactionBuffer == nil {
		return
	}
	var items []*transactionItem
	if itm, transOK := transactionBuffer[transID]; transOK {
		items = itm
	} else {
		return
	}
	// apply all transactioned items
	mux.Lock()
	for _, item := range items {
		switch item.kind {
		case KIND_REM:
			RemKey(item.tenant, item.key, COMMIT)
		case KIND_PRF:
			RemPrefixKey(item.tenant, item.key, COMMIT)
		case KIND_ADD:
			Set(item.tenant, item.key, item.value, COMMIT)
		}
	}
	mux.Unlock()
	delete(transactionBuffer, transID)
}

// The function to be used to cache a key/value pair when expiration is not needed
func Set(tenant, key string, value interface{}, transID string) {
	if transID == "" || transID == COMMIT {
		if transID == "" {
			mux.Lock()
			defer mux.Unlock()
		}
		c(tenant).Put(key, value)
	} else {
		transactionMux.Lock()
		defer transactionMux.Unlock()
		if transactionBuffer == nil {
			return
		}
		items, trasOK := transactionBuffer[transID]
		if !trasOK {
			return
		}
		transactionBuffer[transID] = append(items, &transactionItem{tenant: tenant, key: key, value: value, kind: KIND_ADD})
	}
}

// The function to extract a value for a key that never expire
func Get(tenant, key string) (interface{}, bool) {
	mux.RLock()
	defer mux.RUnlock()
	return c(tenant).Get(key)
}

func RemKey(tenant, key, transID string) {
	if transID == "" || transID == COMMIT {
		if transID == "" {
			mux.Lock()
			defer mux.Unlock()
		}
		c(tenant).Delete(key)
	} else {
		transactionMux.Lock()
		defer transactionMux.Unlock()
		if transactionBuffer == nil {
			return
		}
		items, trasOK := transactionBuffer[transID]
		if !trasOK {
			return
		}
		transactionBuffer[transID] = append(items, &transactionItem{tenant: tenant, key: key, kind: KIND_REM})
	}
}

func RemPrefixKey(tenant, prefix, transID string) {
	if transID == "" || transID == COMMIT {
		if transID == "" {
			mux.Lock()
			defer mux.Unlock()
		}
		c(tenant).DeletePrefix(prefix)
	} else {
		transactionMux.Lock()
		defer transactionMux.Unlock()
		if transactionBuffer == nil {
			return
		}
		items, trasOK := transactionBuffer[transID]
		if !trasOK {
			return
		}
		transactionBuffer[transID] = append(items, &transactionItem{tenant: tenant, key: prefix, kind: KIND_PRF})
	}
}

// Delete all keys from cache
func Flush(tenant string) {
	mux.Lock()
	defer mux.Unlock()
	if tenant == "*any" {
		tenantCache = make(map[string]cacheStore)
	} else {
		tenantCache[tenant] = newLruStore()
	}
}
