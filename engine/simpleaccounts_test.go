package engine

import (
	"testing"

	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/utils"
)

func TestSANewGet(t *testing.T) {
	sas := NewSimpleAccounts(accountingStorage)
	if err := sas.NewAccount("test", "first", false, nil); err != nil {
		t.Fatal(err)
	}
	if acc, err := sas.GetAccount("test", "first"); err != nil || acc == nil || acc.BalanceMap == nil {
		t.Errorf("error getting new account: %+v, %v", acc, err)
	}
}

func TestSADebit(t *testing.T) {
	sas := NewSimpleAccounts(accountingStorage)
	sas.LoadFromStorage("test")
	if err := sas.Debit("test", "first", "call_in", dec.NewVal(10, 0)); err != nil {
		t.Error("error debiting account:", err)
	}
	if acc, err := sas.GetAccount("test", "first"); err != nil ||
		acc == nil ||
		len(acc.BalanceMap) != 1 ||
		acc.BalanceMap["call_in"].Cmp(dec.NewVal(-10, 0)) != 0 {
		t.Errorf("error getting new account: %+v, %v", acc, err)
	}

	if err := sas.Debit("test", "first", "call_out", dec.NewFloat(0.002859)); err != nil {
		t.Error("error debiting account:", err)
	}
	if acc, err := sas.GetAccount("test", "first"); err != nil ||
		acc == nil ||
		len(acc.BalanceMap) != 2 ||
		acc.BalanceMap["call_in"].Cmp(dec.NewVal(-10, 0)) != 0 ||
		acc.BalanceMap["call_out"].Cmp(dec.NewFloat(-0.002859)) != 0 {
		t.Errorf("error getting new account: %+v, %v", acc, err)
	}
}

func TestSANegativeDebit(t *testing.T) {
	sas := NewSimpleAccounts(accountingStorage)
	sas.LoadFromStorage("test")
	if err := sas.Debit("test", "first", "call_in", dec.NewVal(-20, 0)); err != nil {
		t.Error("error debiting account:", err)
	}
	if acc, err := sas.GetAccount("test", "first"); err != nil ||
		acc == nil ||
		len(acc.BalanceMap) != 2 ||
		acc.BalanceMap["call_in"].Cmp(dec.NewVal(10, 0)) != 0 ||
		acc.BalanceMap["call_out"].Cmp(dec.NewFloat(-0.002859)) != 0 {
		t.Errorf("error getting new account: %+v, %v", acc, err)
	}
}

func TestSASet(t *testing.T) {
	sas := NewSimpleAccounts(accountingStorage)
	sas.LoadFromStorage("test")
	if err := sas.SetValue("test", "first", "call_in", dec.NewVal(5, 3)); err != nil {
		t.Error("error debiting account:", err)
	}
	if acc, err := sas.GetAccount("test", "first"); err != nil ||
		acc == nil ||
		len(acc.BalanceMap) != 2 ||
		acc.BalanceMap["call_in"].Cmp(dec.NewVal(5, 3)) != 0 ||
		acc.BalanceMap["call_out"].Cmp(dec.NewFloat(-0.002859)) != 0 {
		t.Errorf("error getting new account: %+v, %v", acc, err)
	}
}

func TestSASetDisabled(t *testing.T) {
	sas := NewSimpleAccounts(accountingStorage)
	sas.LoadFromStorage("test")
	if err := sas.SetDisabled("test", "first", true); err != nil {
		t.Error("error debiting account:", err)
	}
	if acc, err := sas.GetAccount("test", "first"); err != nil ||
		acc == nil ||
		len(acc.BalanceMap) != 2 ||
		acc.BalanceMap["call_in"].Cmp(dec.NewVal(5, 3)) != 0 ||
		acc.BalanceMap["call_out"].Cmp(dec.NewFloat(-0.002859)) != 0 ||
		acc.Disabled != true {
		t.Errorf("error getting new account: %+v, %v", acc, err)
	}
	if err := sas.Debit("test", "first", "call_in", dec.NewFloat(0.1)); err != utils.ErrAccountDisabled {
		t.Error("error debiting disabled account: ", err)
	}
	if err := sas.SetDisabled("test", "first", false); err != nil {
		t.Error("error debiting account:", err)
	}
	if err := sas.Debit("test", "first", "call_out", dec.NewFloat(0.1)); err != nil {
		t.Error("error debiting disabled account: ", err)
	}
}

func TestSASetMaxBalance(t *testing.T) {
	sas := NewSimpleAccounts(accountingStorage)
	sas.LoadFromStorage("test")
	if err := sas.SetMaxBalance("test", "first", dec.NewVal(20, 0)); err != nil {
		t.Error("error debiting account:", err)
	}
	if acc, err := sas.GetAccount("test", "first"); err != nil ||
		acc == nil ||
		len(acc.BalanceMap) != 2 ||
		acc.BalanceMap["call_in"].Cmp(dec.NewVal(5, 3)) != 0 ||
		acc.BalanceMap["call_out"].Cmp(dec.NewFloat(-0.102859)) != 0 ||
		acc.Disabled != false ||
		acc.MaxBalance.Cmp(dec.NewFloat(20)) != 0 {
		t.Errorf("error getting new account: %+v, %v", acc, err)
	}
	if err := sas.Debit("test", "first", "call_in", dec.NewFloat(-20)); err != nil {
		t.Error("error debiting max quota account: ", err)
	}
	if err := sas.Debit("test", "first", "call_out", dec.NewFloat(20)); err != nil {
		t.Error("error debiting max quota account: ", err)
	}
	// now the quota should be exceeeded
	if err := sas.Debit("test", "first", "call_in", dec.NewFloat(-20)); err != utils.ErrQuotaExceeded {
		t.Error("error debiting max quota account: ", err)
	}
	if err := sas.Debit("test", "first", "call_out", dec.NewFloat(20)); err != utils.ErrQuotaExceeded {
		t.Error("error debiting max quota account: ", err)
	}
	if err := sas.SetMaxBalance("test", "first", nil); err != nil {
		t.Error("error debiting account:", err)
	}
	if err := sas.Debit("test", "first", "call_out", dec.NewFloat(20)); err != nil {
		t.Error("error debiting disabled account: ", err)
	}
}
