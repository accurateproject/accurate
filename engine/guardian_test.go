
package engine

import (
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
	if _, ok := Guardian.locksMap["1"]; !ok {
		t.Error("Deleted after 11 milliseconds")
	}
	time.Sleep(11 * time.Millisecond)
	if _, ok := Guardian.locksMap["1"]; !ok {
		t.Error("Deleted after 22 milliseconds")
	}
	time.Sleep(11 * time.Millisecond)
	if _, ok := Guardian.locksMap["1"]; ok {
		t.Error("should be deleted by now")
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
