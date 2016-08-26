package cache2go

import "testing"

func TestRemKey(t *testing.T) {
	Set("t01_mm", "test", "")
	if t1, ok := Get("t01_mm"); !ok || t1 != "test" {
		t.Error("Error setting cache: ", ok, t1)
	}
	RemKey("t01_mm", "")
	if t1, ok := Get("t01_mm"); ok || t1 == "test" {
		t.Error("Error removing cached key")
	}
}

func TestTransaction(t *testing.T) {
	transID := BeginTransaction()
	Set("t11_mm", "test", transID)
	if t1, ok := Get("t11_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("t12_mm", "test", transID)
	RemKey("t11_mm", transID)
	CommitTransaction(transID)
	if t1, ok := Get("t12_mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t11_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRem(t *testing.T) {
	transID := BeginTransaction()
	Set("t21_mm", "test", transID)
	Set("t21_nn", "test", transID)
	RemPrefixKey("t21_", transID)
	CommitTransaction(transID)
	if t1, ok := Get("t21_mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t21_nn"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRollback(t *testing.T) {
	transID := BeginTransaction()
	Set("t31_mm", "test", transID)
	if t1, ok := Get("t31_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("t32_mm", "test", transID)
	RollbackTransaction(transID)
	if t1, ok := Get("t32_mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t31_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRemBefore(t *testing.T) {
	transID := BeginTransaction()
	RemPrefixKey("t41_", transID)
	Set("t41_mm", "test", transID)
	Set("t41_nn", "test", transID)
	CommitTransaction(transID)
	if t1, ok := Get("t41_mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t41_nn"); !ok || t1 != "test" {
		t.Error("Error in transaction cache")
	}
}

func TestRemPrefixKey(t *testing.T) {
	Set("xxx_t1", "test", "")
	Set("yyy_t1", "test", "")
	RemPrefixKey("xxx_", "")
	_, okX := Get("xxx_t1")
	_, okY := Get("yyy_t1")
	if okX || !okY {
		t.Error("Error removing prefix: ", okX, okY)
	}
}

func TestTransactionMultiple(t *testing.T) {
	transID1 := BeginTransaction()
	transID2 := BeginTransaction()
	Set("t51_mm", "test", transID1)
	if t5, ok := Get("t51_mm"); ok || t5 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("t52_mm", "test", transID1)
	RemKey("t51_mm", transID1)
	RemKey("t52_mm", transID2)
	CommitTransaction(transID1)
	if t5, ok := Get("t52_mm"); !ok || t5 != "test" {
		t.Error("Error commiting transaction")
	}
	if t5, ok := Get("t51_mm"); ok || t5 == "test" {
		t.Error("Error in transaction cache")
	}
	CommitTransaction(transID2)
	if t5, ok := Get("t52_mm"); ok || t5 == "test" {
		t.Error("Error commiting transaction")
	}
}

/*func TestCount(t *testing.T) {
	Set("dst_A1", "1")
	Set("dst_A2", "2")
	Set("rpf_A3", "3")
	Set("dst_A4", "4")
	Set("dst_A5", "5")
	if CountEntries("dst_") != 4 {
		t.Error("Error countiong entries: ", CountEntries("dst_"))
	}
}*/
