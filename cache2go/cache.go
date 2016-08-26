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
	mux   sync.RWMutex
	cache cacheStore
	cfg   *config.CacheConfig
	// transaction stuff
	transactionBuffer map[string][]*transactionItem
	transactionMux    sync.Mutex
)

type transactionItem struct {
	key   string
	value interface{}
	kind  string
}

func init() {
	NewCache(nil)
}

func NewCache(cacheCfg *config.CacheConfig) {
	cfg = cacheCfg
	cache = newLruStore()
}

func BeginTransaction() string {
	transID := utils.GenUUID()
	transactionMux.Lock()
	defer transactionMux.Unlock()
	if transactionBuffer == nil {
		transactionBuffer = make(map[string][]*transactionItem)
	}
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
			RemKey(item.key, COMMIT)
		case KIND_PRF:
			RemPrefixKey(item.key, COMMIT)
		case KIND_ADD:
			Set(item.key, item.value, COMMIT)
		}
	}
	mux.Unlock()
	delete(transactionBuffer, transID)
}

// The function to be used to cache a key/value pair when expiration is not needed
func Set(key string, value interface{}, transID string) {
	if transID == "" || transID == COMMIT {
		if transID == "" {
			mux.Lock()
			defer mux.Unlock()
		}
		cache.Put(key, value)
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
		transactionBuffer[transID] = append(items, &transactionItem{key: key, value: value, kind: KIND_ADD})
	}
}

// The function to extract a value for a key that never expire
func Get(key string) (interface{}, bool) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.Get(key)
}

func RemKey(key, transID string) {
	if transID == "" || transID == COMMIT {
		if transID == "" {
			mux.Lock()
			defer mux.Unlock()
		}
		cache.Delete(key)
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
		transactionBuffer[transID] = append(items, &transactionItem{key: key, kind: KIND_REM})
	}
}

func RemPrefixKey(prefix, transID string) {
	if transID == "" || transID == COMMIT {
		if transID == "" {
			mux.Lock()
			defer mux.Unlock()
		}
		cache.DeletePrefix(prefix)
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
		transactionBuffer[transID] = append(items, &transactionItem{key: prefix, kind: KIND_PRF})
	}
}

// Delete all keys from cache
func Flush() {
	mux.Lock()
	defer mux.Unlock()
	cache = newLruStore()
}

func CountEntries(prefix string) (result int) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.CountEntriesForPrefix(prefix)
}

func GetEntriesKeys(prefix string) (keys []string) {
	mux.RLock()
	defer mux.RUnlock()
	return cache.GetKeysForPrefix(prefix)
}
