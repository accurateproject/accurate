package cache2go

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

func TestRemKey(t *testing.T) {
	Set("t", "t01_mm", "test", "")
	if t1, ok := Get("t", "t01_mm"); !ok || t1 != "test" {
		t.Error("Error setting cache: ", ok, t1)
	}
	RemKey("t", "t01_mm", "")
	if t1, ok := Get("t", "t01_mm"); ok || t1 == "test" {
		t.Error("Error removing cached key")
	}
}

func TestTransaction(t *testing.T) {
	transID := BeginTransaction()
	Set("t", "t11_mm", "test", transID)
	if t1, ok := Get("t", "t11_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("t", "t12_mm", "test", transID)
	RemKey("t", "t11_mm", transID)
	CommitTransaction(transID)
	if t1, ok := Get("t", "t12_mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t", "t11_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRem(t *testing.T) {
	transID := BeginTransaction()
	Set("t", "t21_mm", "test", transID)
	Set("t", "t21_nn", "test", transID)
	RemPrefixKey("t", "t21_", transID)
	CommitTransaction(transID)
	if t1, ok := Get("t", "t21_mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t", "t21_nn"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRollback(t *testing.T) {
	transID := BeginTransaction()
	Set("t", "t31_mm", "test", transID)
	if t1, ok := Get("t", "t31_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("t", "t32_mm", "test", transID)
	RollbackTransaction(transID)
	if t1, ok := Get("t", "t32_mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t", "t31_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRemBefore(t *testing.T) {
	transID := BeginTransaction()
	RemPrefixKey("t", "t41_", transID)
	Set("t", "t41_mm", "test", transID)
	Set("t", "t41_nn", "test", transID)
	CommitTransaction(transID)
	if t1, ok := Get("t", "t41_mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t", "t41_nn"); !ok || t1 != "test" {
		t.Error("Error in transaction cache")
	}
}

func TestRemPrefixKey(t *testing.T) {
	Set("t", "xxx_t1", "test", "")
	Set("t", "yyy_t1", "test", "")
	RemPrefixKey("t", "xxx_", "")
	_, okX := Get("t", "xxx_t1")
	_, okY := Get("t", "yyy_t1")
	if okX || !okY {
		t.Error("Error removing prefix: ", okX, okY)
	}
}

func TestTransactionMultiple(t *testing.T) {
	transID1 := BeginTransaction()
	transID2 := BeginTransaction()
	Set("t", "t51_mm", "test", transID1)
	if t5, ok := Get("t", "t51_mm"); ok || t5 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("t", "t52_mm", "test", transID1)
	RemKey("t", "t51_mm", transID1)
	RemKey("t", "t52_mm", transID2)
	CommitTransaction(transID1)
	if t5, ok := Get("t", "t52_mm"); !ok || t5 != "test" {
		t.Error("t", "Error commiting transaction")
	}
	if t5, ok := Get("t", "t51_mm"); ok || t5 == "test" {
		t.Error("Error in transaction cache")
	}
	CommitTransaction(transID2)
	if t5, ok := Get("t", "t52_mm"); ok || t5 == "test" {
		t.Error("Error commiting transaction")
	}
}

/***************** benchmarks ********************/

func BenchmarkSetGetParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		cache := NewLRUTTL(0, 10*time.Second)
		for pb.Next() {
			cache.Set(fmt.Sprintf("xxx_%d", rand.Intn(100)), "x")
			cache.Get(fmt.Sprintf("xxx_%d", rand.Intn(100)))
		}
	})
}

func BenchmarkSetGetParallelOther(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		cache, _ := lru.New(10000)
		for pb.Next() {
			cache.Add(fmt.Sprintf("xxx_%d", rand.Intn(100)), "x")
			cache.Get(fmt.Sprintf("xxx_%d", rand.Intn(100)))
		}
	})
}
