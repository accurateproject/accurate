package engine

import (
	"sync"
	"testing"
	"time"
)

func TestGuardianDelete(t *testing.T) {
	for i := 0; i < 3; i++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)
			return 0, nil
		}, 0, "1")
	}
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; !ok {
		t.Error("Deleted after 11 milliseconds")
	}
	Guardian.mu.Unlock()
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; !ok {
		t.Error("Deleted after 22 milliseconds")
	}
	Guardian.mu.Unlock()
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; ok {
		t.Error("should be deleted by now")
	}
	Guardian.mu.Unlock()
}

func TestGuardianDeleteReuse(t *testing.T) {
	for i := 0; i < 3; i++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)
			return 0, nil
		}, 0, "1")
	}
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; !ok {
		t.Error("Deleted after 11 milliseconds")
	}
	Guardian.mu.Unlock()
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; !ok {
		t.Error("Deleted after 22 milliseconds")
	}
	Guardian.mu.Unlock()
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; ok {
		t.Error("should be deleted by now")
	}
	Guardian.mu.Unlock()
	for i := 0; i < 3; i++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)
			return 0, nil
		}, 0, "1")
	}
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; !ok {
		t.Error("Deleted after 11 milliseconds")
	}
	Guardian.mu.Unlock()
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; !ok {
		t.Error("Deleted after 22 milliseconds")
	}
	Guardian.mu.Unlock()
	time.Sleep(11 * time.Millisecond)
	Guardian.mu.Lock()
	if _, ok := Guardian.locksMap["1"]; ok {
		t.Error("should be deleted by now")
	}
	Guardian.mu.Unlock()
}

func TestGuardianRace(t *testing.T) {
	account := make(map[string]int) // simplified data representing an account
	account["acc:test"] = 0         // acount balance is 0

	// this will make sure we only check account after all threads done their job
	wg := sync.WaitGroup{}

	// we run ten concurrent account modifications
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() { // start threads

			// protect the account  with the guradian locker
			// this should only allow one thread at a time
			Guardian.Guard(func() (interface{}, error) {

				// typical account operation
				value := account["acc:test"] // get the balance
				value += 1                   // modify the balance
				account["acc:test"] = value  // put the balance back

				wg.Done()
				return nil, nil
			}, 0, "acc:test")
		}()
	}
	wg.Wait() // here we use the waitgroup

	// if the final balance value is not ten we have a problem
	if x := account["acc:test"]; x != 10 {
		t.Error("Error setting right account value: ", x)
	}
}

func BenchmarkGuard(b *testing.B) {
	for i := 0; i < 100; i++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return 0, nil
		}, 0, "1")
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return 0, nil
		}, 0, "2")
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return 0, nil
		}, 0, "1")
	}

}

func BenchmarkGuardian(b *testing.B) {
	for i := 0; i < 100; i++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return 0, nil
		}, 0, "1")
	}
}
