package engine

import (
	"sync"
	"time"

	"github.com/accurateproject/accurate/utils"
	"go.uber.org/zap"
)

var lockPool = &sync.Pool{New: func() interface{} { return &lockItem{} }}

// global package variable
var Guardian = &GuardianLock{locksMap: make(map[string]*lockItem)}

func init() {
	//go Guardian.debugGuardianStatus()
}

type lockItem struct {
	mu sync.Mutex
	i  int // how many are queued to hold the lock
}

type GuardianLock struct {
	locksMap map[string]*lockItem
	mu       sync.Mutex
}

/*func (cm *GuardianLock) debugGuardianStatus() {
	for {
		cm.mu.Lock()
		for key, value := range cm.locksMap {
			log.Printf("Key: %s: %d", key, value.i)
		}
		cm.mu.Unlock()
		time.Sleep(30 * time.Millisecond)

	}
}*/

func (cm *GuardianLock) Guard(handler func() (interface{}, error), timeout time.Duration, names ...string) (reply interface{}, err error) {
	var locks []*lockItem // take existing locks out of the mutex
	cm.mu.Lock()
	for _, name := range names {
		var lock *lockItem
		var found bool
		if lock, found = Guardian.locksMap[name]; !found {
			lock = lockPool.Get().(*lockItem)
			Guardian.locksMap[name] = lock
			lock.mu.Lock()
		} else {
			locks = append(locks, lock)
		}
		lock.i++
	}
	cm.mu.Unlock()

	for _, lock := range locks {
		lock.mu.Lock()
	}

	funcWaiter := make(chan bool)
	go func() {
		// execute
		reply, err = handler()
		funcWaiter <- true
	}()
	// wait with timeout
	if timeout > 0 {
		select {
		case <-funcWaiter:
		case <-time.After(timeout):
			utils.Logger.Warn("<Guardian> Timeout on keys: ", zap.Strings("keys", names))
		}
	} else {
		<-funcWaiter
	}
	// release
	cm.mu.Lock()
	for _, name := range names {
		lock := Guardian.locksMap[name]
		lock.mu.Unlock()
		lock.i--
		if lock.i == 0 {
			delete(Guardian.locksMap, name)
			lockPool.Put(lock)
		}
	}
	cm.mu.Unlock()
	return
}
