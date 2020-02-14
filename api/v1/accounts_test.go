package v1

import (
	"log"
	"os"
	"testing"

	dockertest "gopkg.in/ory-am/dockertest.v3"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/dec"
	"github.com/accurateproject/accurate/engine"
	"github.com/accurateproject/accurate/utils"
)

var (
	apiObject         *ApiV1
	accountingStorage engine.AccountingStorage
	ratingStorage     engine.RatingStorage
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("mongo", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		port := resource.GetPort("27017/tcp")
		var err error
		ratingStorage, err = engine.NewMongoStorage("127.0.0.1", port, "tp_test", "", "", utils.TariffPlanDB, nil, &config.Cache{RatingPlans: &config.CacheParam{Precache: true}}, 10)
		if err != nil {
			log.Fatal(err)
		}
		accountingStorage, err = engine.NewMongoStorage("127.0.0.1", port, "acc_test", "", "", utils.DataDB, nil, &config.Cache{RatingPlans: &config.CacheParam{Precache: true}}, 10)
		if err != nil {
			log.Fatal(err)
		}
		return accountingStorage.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	cfg := config.NewDefault()
	apiObject = &ApiV1{accountDB: accountingStorage, ratingDB: ratingStorage, cfg: cfg}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestAccountsSetAccounts(t *testing.T) {
	tenant := "test"
	iscTenant := "iscTest"
	b10 := &engine.Balance{Value: dec.NewVal(10, 0), Weight: 10}
	cgrAcnt1 := &engine.Account{Tenant: tenant, Name: "account1",
		BalanceMap: map[string]engine.Balances{utils.MONETARY + utils.OUT: engine.Balances{b10}}}
	cgrAcnt2 := &engine.Account{Tenant: tenant, Name: "account2",
		BalanceMap: map[string]engine.Balances{utils.MONETARY + utils.OUT: engine.Balances{b10}}}
	cgrAcnt3 := &engine.Account{Tenant: tenant, Name: "account3",
		BalanceMap: map[string]engine.Balances{utils.MONETARY + utils.OUT: engine.Balances{b10}}}
	iscAcnt1 := &engine.Account{Tenant: iscTenant, Name: "account1",
		BalanceMap: map[string]engine.Balances{utils.MONETARY + utils.OUT: engine.Balances{b10}}}
	iscAcnt2 := &engine.Account{Tenant: iscTenant, Name: "account2",
		BalanceMap: map[string]engine.Balances{utils.MONETARY + utils.OUT: engine.Balances{b10}}}
	for _, account := range []*engine.Account{cgrAcnt1, cgrAcnt2, cgrAcnt3, iscAcnt1, iscAcnt2} {
		if err := accountingStorage.SetAccount(account); err != nil {
			t.Error(err)
		}
	}
}

func TestAccountsGetAccounts(t *testing.T) {
	var accounts []*engine.Account
	var attrs AttrGetMultiple
	if err := apiObject.GetAccounts(AttrGetMultiple{Tenant: "test"}, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 3 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
	attrs = AttrGetMultiple{Tenant: "iscTest"}
	if err := apiObject.GetAccounts(attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 2 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
	attrs = AttrGetMultiple{Tenant: "test", IDs: []string{"account1"}}
	if err := apiObject.GetAccounts(attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 1 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
	attrs = AttrGetMultiple{Tenant: "iscTest", IDs: []string{"INVALID"}}
	if err := apiObject.GetAccounts(attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 0 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
	attrs = AttrGetMultiple{Tenant: "INVALID"}
	if err := apiObject.GetAccounts(attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 0 {
		t.Errorf("Accounts returned: %+v", accounts)
	}
}

func TestAccountsSetGetAccount(t *testing.T) {
	var reply string
	if err := apiObject.SetAccount(AttrSetAccount{
		Tenant:        "test",
		Account:       "fast",
		AllowNegative: utils.BoolPointer(true),
		Disabled:      utils.BoolPointer(true),
	}, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Errorf("not set: %+v", reply)
	}
	var accounts []*engine.Account
	attrs := AttrGetMultiple{Tenant: "test", IDs: []string{"fast"}}
	if err := apiObject.GetAccounts(attrs, &accounts); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if len(accounts) != 1 || !accounts[0].Disabled || !accounts[0].AllowNegative {
		t.Errorf("Accounts returned: %+v", accounts)
	}

}
