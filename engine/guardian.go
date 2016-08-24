
package engine

import (
	"sync"
	"time"
)

var lockPool = &sync.Pool{New: func() interface{} {
	return make(chan bool, 1)
}}

// global package variable
var Guardian = &GuardianLock{locksMap: make(map[string]chan bool)}

type GuardianLock struct {
	locksMap map[string]chan bool
	mu       sync.Mutex
}

func (cm *GuardianLock) Guard(handler func() (interface{}, error), timeout time.Duration, names ...string) (reply interface{}, err error) {
	var locks []chan bool // take existing locks out of the mutex
	cm.mu.Lock()
	for _, name := range names {
		if lock, exists := Guardian.locksMap[name]; !exists {
			lock = lockPool.Get().(chan bool)
			Guardian.locksMap[name] = lock
			lock <- true
		} else {
			locks = append(locks, lock)
		}
	}
	cm.mu.Unlock()

	for _, lock := range locks {
		lock <- true // will block here if already locked
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
		}
	} else {
		<-funcWaiter
	}
	// release
	cm.mu.Lock()
	for _, name := range names {
		lock := Guardian.locksMap[name]
		<-lock
		delete(Guardian.locksMap, name)
		lockPool.Put(lock)
	}
	cm.mu.Unlock()
	return
}
