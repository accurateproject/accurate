
package engine

import (
	"sync"
	"time"
)

// global package variable
var Guardian = &GuardianLock{locksMap: make(map[string]chan bool)}

type GuardianLock struct {
	locksMap map[string]chan bool
	mu       sync.RWMutex
}

func (cm *GuardianLock) Guard(handler func() (interface{}, error), timeout time.Duration, names ...string) (reply interface{}, err error) {
	var locks []chan bool // take existing locks out of the mutex
	cm.mu.Lock()
	for _, name := range names {
		if lock, exists := Guardian.locksMap[name]; !exists {
			lock = make(chan bool, 1)
			Guardian.locksMap[name] = lock
			lock <- true
		} else {
			locks = append(locks, lock)
		}
	}
	cm.mu.Unlock()

	for _, lock := range locks {
		lock <- true
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
	cm.mu.RLock()
	for _, name := range names {
		<-Guardian.locksMap[name]
	}
	cm.mu.RUnlock()
	return
}
