/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"sync"
	"time"
)

var lockPool = &sync.Pool{New: func() interface{} { return &lockItem{c: make(chan struct{}, 1), i: 0} }}

// global package variable
var Guardian = &GuardianLock{locksMap: make(map[string]*lockItem)}

type lockItem struct {
	c chan struct{}
	i int // how many are queued to hold the lock
}

type GuardianLock struct {
	locksMap map[string]*lockItem
	mu       sync.Mutex
}

func (cm *GuardianLock) Guard(handler func() (interface{}, error), timeout time.Duration, names ...string) (reply interface{}, err error) {
	var locks []*lockItem // take existing locks out of the mutex
	cm.mu.Lock()
	for _, name := range names {
		var lock *lockItem
		var found bool
		if lock, found = Guardian.locksMap[name]; !found {
			lock = lockPool.Get().(*lockItem)
			Guardian.locksMap[name] = lock
			lock.c <- struct{}{}
		} else {
			locks = append(locks, lock)
		}
		lock.i++
	}
	cm.mu.Unlock()

	for _, lock := range locks {
		lock.c <- struct{}{} // will block here if already locked
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
		<-lock.c
		lock.i--
		if lock.i == 0 {
			delete(Guardian.locksMap, name)
			lockPool.Put(lock)
		}
	}
	cm.mu.Unlock()
	return
}
