package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/accurateproject/accurate/config"
	"github.com/accurateproject/accurate/utils"
)

var (
	testMap  UserMap
	testMap2 UserMap
)

func init() {
	testMap = UserMap{
		table: map[string]*UserProfile{
			"test:user":   &UserProfile{Tenant: "test", Name: "user", Query: `{"T":{"$usr":"v"}}`, Index: map[string]string{"T": "v"}},
			"test1:user1": &UserProfile{Tenant: "test1", Name: "user1", Query: `{"T":{"$usr":"v"}, "X":{"$usr": "y"}}`, Index: map[string]string{"T": "v"}},
			"test:masked": &UserProfile{Tenant: "test", Name: "masked", Masked: true, Query: `{"T":{"$usr":"v"}}`, Index: map[string]string{"T": "v"}},
		},
		index: make(map[string]UserProfiles),
	}

	testMap2 = UserMap{
		table: map[string]*UserProfile{
			"an:u1": &UserProfile{Tenant: "an", Name: "u1", Query: `{"A":{"$usr": "b"}, "C":{"$usr":"d"}}`, Weight: 0},
			"an:u2": &UserProfile{Tenant: "an", Name: "u2", Query: `{"A":{"$usr": "b"}}`, Weight: 10},
		},
		index: make(map[string]UserProfiles),
	}
}

func TestUsersAdd(t *testing.T) {
	tm := newUserMap(accountingStorage, nil)
	var r string
	up := &UserProfile{Tenant: "test", Name: "user", Query: `{"t":"v", "t":{"$set":"v"}}`}
	if err := tm.SetUser(up, &r); err != nil {
		t.Fatal(err)
	}
	up, found := tm.table[up.FullID()]
	if r != utils.OK ||
		!found ||
		up.Tenant != "test" ||
		len(tm.table) != 1 {
		t.Error("Error setting user: ", tm, len(tm.table))
	}
}

func TestUsersUpdate(t *testing.T) {
	tm := newUserMap(accountingStorage, nil)
	var r string
	up := &UserProfile{
		Tenant: "test",
		Name:   "user",
		Query:  `{"t":{"$crepl":["^$|v","v"]}}`,
	}
	if err := tm.SetUser(up, &r); err != nil {
		t.Fatal(err)
	}
	p, found := tm.table[up.FullID()]
	if r != utils.OK ||
		!found ||
		p.Query != `{"t":{"$crepl":["^$|v","v"]}}` ||
		len(tm.table) != 1 ||
		p.getQuery() == nil {
		t.Errorf("Error setting user: %+v", p)
	}
	origQuery := p.query
	up = &UserProfile{
		Tenant: "test",
		Name:   "user",
		Query:  `{"x":{"$crepl":["^$|y","y"]}}`,
	}
	if err := tm.UpdateUser(up, &r); err != nil {
		t.Fatal(err)
	}
	p, found = tm.table[up.FullID()]
	if r != utils.OK ||
		!found ||
		p.Query != `{"x":{"$crepl":["^$|y","y"]}}` ||
		len(tm.table) != 1 ||
		p.getQuery() == origQuery {
		t.Errorf("Error updating user: %+v", p.getQuery() == origQuery)
	}
}

func TestUsersUpdateNotFound(t *testing.T) {
	tm := newUserMap(accountingStorage, nil)
	var r string
	up := &UserProfile{
		Tenant: "test",
		Name:   "user",
	}
	if err := tm.SetUser(up, &r); err != nil {
		t.Fatal(err)
	}
	up.Name = "test1"
	if err := tm.UpdateUser(up, &r); err != nil {
		t.Error("Error detecting user not found on update: ", err)
	}
}

func TestUsersUpdateInit(t *testing.T) {
	tm := newUserMap(accountingStorage, nil)
	var r string
	up := &UserProfile{
		Tenant: "test",
		Name:   "user",
	}
	if err := tm.SetUser(up, &r); err != nil {
		t.Fatal(err)
	}
	up = &UserProfile{
		Tenant: "test",
		Name:   "user",
		Query:  `{"t":{"$crepl":["^$|v","v"]}}`,
	}
	if err := tm.UpdateUser(up, &r); err != nil {
		t.Fatal(err)
	}
	p, found := tm.table[up.FullID()]
	if r != utils.OK ||
		!found ||
		p.Query != `{"t":{"$crepl":["^$|v","v"]}}` ||
		len(tm.table) != 1 ||
		p.getQuery() == nil {
		t.Error("Error updating user: ", tm)
	}
}

func TestUsersRemove(t *testing.T) {
	tm := newUserMap(accountingStorage, nil)
	var r string
	up := &UserProfile{
		Tenant: "test",
		Name:   "user",
		Query:  `{"t":{"$crepl":["^$|v","v"]}}`,
	}
	if err := tm.SetUser(up, &r); err != nil {
		t.Fatal(err)
	}
	p, found := tm.table[up.FullID()]
	if r != utils.OK ||
		!found ||
		p.Query != `{"t":{"$crepl":["^$|v","v"]}}` ||
		len(tm.table) != 1 ||
		p.getQuery() == nil {
		t.Error("Error setting user: ", tm)
	}
	if err := tm.RemoveUser(up, &r); err != nil {
		t.Fatal(err)
	}
	p, found = tm.table[up.FullID()]
	if r != utils.OK ||
		found ||
		len(tm.table) != 0 {
		t.Error("Error removing user: ", tm)
	}
}

func TestUsersGetFull(t *testing.T) {
	attr := AttrGetUsers{
		Object: struct {
			T string
			X string
		}{
			T: "v",
			X: "y",
		},
		Masked: false,
	}
	config.Get().Users.ComplexityMatch = utils.BoolPointer(true)
	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetFullMasked(t *testing.T) {
	attr := AttrGetUsers{
		Object: struct {
			T string
			X string
		}{
			T: "v",
			X: "y",
		},
		Masked: true,
	}
	config.Get().Users.ComplexityMatch = utils.BoolPointer(true)
	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetSingle(t *testing.T) {
	attr := AttrGetUsers{
		Object: struct {
			T string
			X string
		}{
			T: "v",
			X: "y",
		},
		Masked: true,
	}
	config.Get().Users.ComplexityMatch = utils.BoolPointer(false)
	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetTenant(t *testing.T) {
	attr := AttrGetUsers{
		Object: struct {
			Tenant, Name, T, X string
		}{
			Tenant: "testX",
			Name:   "user",
		},
	}
	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetName(t *testing.T) {
	attr := AttrGetUsers{
		Object: struct {
			Tenant, Name, T, X string
		}{
			Tenant: "testX",
			Name:   "user",
			T:      "a",
		},
	}
	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingIdTwo(t *testing.T) {
	attr := AttrGetUsers{
		Object: struct {
			Tenant, Name, T, X string
		}{
			T: "v",
			X: "y",
		},
	}
	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetMissingIdTwoSort(t *testing.T) {
	attr := AttrGetUsers{
		Object: struct {
			Tenant, Name, T, X string
		}{
			T: "v",
			X: "y",
		},
	}
	results := UserProfiles{}
	config.Get().Users.ComplexityMatch = utils.BoolPointer(true)
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Error("error getting users: ", results)
	}
	if results[0].FullID() != "test1:user1" {
		t.Errorf("Error sorting profiles: %+v", results[0])
	}
}

func TestUsersGetMissingIdTwoSortWeight(t *testing.T) {
	attr := AttrGetUsers{
		Object: struct {
			Tenant, Name, A, C string
		}{
			A: "b",
			C: "d",
		},
	}
	results := UserProfiles{}
	config.Get().Users.ComplexityMatch = utils.BoolPointer(true)
	if err := testMap2.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Error("error getting users: ", results)
	}
	if results[0].FullID() != "an:u2" {
		t.Errorf("Error sorting profiles: %+v", results[0])
	}
}

func TestUsersAddIndex(t *testing.T) {
	var r string
	if err := testMap.AddIndex([]string{"T"}, &r); err != nil {
		t.Fatal(err)
	}
	if r != utils.OK ||
		len(testMap.index) != 1 ||
		len(testMap.index[utils.ConcatKey("T", "v")]) != 3 {
		t.Error("error adding index: ", testMap.index)
	}
}

func TestUsersAddIndexFull(t *testing.T) {
	var r string
	testMap.index = make(map[string]UserProfiles) // reset index
	if err := testMap.AddIndex([]string{"T", "X", "Name", "Tenant"}, &r); err != nil {
		t.Fatal(err)
	}
	if r != utils.OK ||
		len(testMap.index) != 6 ||
		len(testMap.index[utils.ConcatKey("T", "v")]) != 3 {
		t.Error("error adding index: ", utils.ToIJSON(testMap.index))
	}
}

func TestUsersAddIndexNone(t *testing.T) {
	var r string
	testMap.index = make(map[string]UserProfiles) // reset index
	if err := testMap.AddIndex([]string{"test"}, &r); err != nil {
		t.Fatal(err)
	}
	if r != utils.OK ||
		len(testMap.index) != 0 {
		t.Error("error adding index: ", testMap.index)
	}
}

func TestUsersGetFullindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]UserProfiles) // reset index
	if err := testMap.AddIndex([]string{"T", "X", "Name", "Tenant"}, &r); err != nil {
		t.Fatal(err)
	}
	attr := AttrGetUsers{
		Object: struct {
			Tenant, Name, T, X string
		}{
			Tenant: "test",
			Name:   "user",
			T:      "v",
			X:      "y",
		},
	}
	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Error("error getting users: ", utils.ToIJSON(results))
	}
}

func TestUsersGetTenantindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]UserProfiles) // reset index
	if err := testMap.AddIndex([]string{"T", "X", "Name", "Tenant"}, &r); err != nil {
		t.Fatal(err)
	}

	attr := AttrGetUsers{
		Object: struct {
			Tenant, Name, T, X string
		}{
			Tenant: "testX",
			Name:   "user",
			T:      "v",
		},
	}

	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersGetNotFoundProfileindex(t *testing.T) {
	var r string
	testMap.index = make(map[string]UserProfiles) // reset index
	testMap.AddIndex([]string{"T", "X", "Name", "Tenant"}, &r)
	attr := AttrGetUsers{
		Object: struct {
			Tenant, Name, T, X string
		}{
			Tenant: "test",
			Name:   "user",
			T:      "o",
		},
	}

	results := UserProfiles{}
	if err := testMap.GetUsers(attr, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Error("error getting users: ", results)
	}
}

func TestUsersAddUpdateRemoveIndexes(t *testing.T) {
	tm := newUserMap(accountingStorage, nil)
	var r string
	if err := tm.AddIndex([]string{"T"}, &r); err != nil {
		t.Fatal(err)
	}
	if len(tm.index) != 0 {
		t.Error("error adding indexes: ", tm.index)
	}
	if err := tm.SetUser(&UserProfile{Tenant: "test", Name: "user", Index: map[string]string{"T": "v"}, Query: `{"T":{"$crepl":["^$|v","v"]}}`}, &r); err != nil {
		t.Fatal(err)
	}
	if len(tm.index) != 1 || len(tm.index["T:v"]) != 1 {
		t.Error("error adding indexes: ", tm.index)
	}
	if err := tm.SetUser(&UserProfile{Tenant: "test", Name: "best", Index: map[string]string{"T": "v"}, Query: `{"T":{"$crepl":["^$|v","v"]}}`}, &r); err != nil {
		t.Fatal(err)
	}
	if len(tm.index) != 1 || len(tm.index["T:v"]) != 2 {
		t.Error("error adding indexes: ", tm.index)
	}
	if err := tm.UpdateUser(&UserProfile{Tenant: "test", Name: "best", Index: map[string]string{"T": "v1"}, Query: `{"T":{"$crepl":["^$|v1","v1"]}}`}, &r); err != nil {
		t.Fatal(err)
	}

	if len(tm.index) != 2 ||
		len(tm.index["T:v"]) != 1 ||
		len(tm.index["T:v1"]) != 1 {
		t.Error("error adding indexes: ", tm.index)
	}
	if err := tm.UpdateUser(&UserProfile{Tenant: "test", Name: "best", Index: map[string]string{"T": "v"}, Query: `{"T":{"$crepl":["^$|v","v"]}}`}, &r); err != nil {
		t.Fatal(err)
	}
	if len(tm.index) != 1 || len(tm.index["T:v"]) != 2 {
		t.Error("error adding indexes: ", tm.index)
	}

	if err := tm.RemoveUser(&UserProfile{Tenant: "test", Name: "best", Index: map[string]string{"T": "v"}, Query: `{"T":{"$crepl":["^$|v","v"]}}`}, &r); err != nil {
		t.Fatal(err)
	}
	if len(tm.index) != 1 ||
		len(tm.index["T:v"]) != 1 {
		t.Error("error adding indexes: ", tm.index)
	}
	if err := tm.RemoveUser(&UserProfile{Tenant: "test", Name: "user", Index: map[string]string{"T": "v"}, Query: `{"T":{"$crepl":["^$|v","v"]}}`}, &r); err != nil {
		t.Fatal(err)
	}
	if len(tm.index) != 0 {
		t.Error("error adding indexes: ", tm.index)
	}
}

func TestUsersUsageRecordGetLoadUserProfile(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:user":   &UserProfile{Tenant: "test", Name: "user", Query: `{"ToR":{"$crepl":["\\*users|01", "01"]}, "RequestType":{"$crepl":["\\*users|1", "1"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c1", "c1"]}, "Account":{"$crepl":["\\*users|dan", "dan"]}, "Subject":{"$crepl":["\\*users|0723", "0723"]}, "Destination":{"$crepl":["\\*users|\\+401", "+401"]}, "SetupTime":{"$crepl":["\\*users|s1", "s1"]}, "AnswerTime":{"$crepl":["\\*users|t1", "t1"]}, "Usage":{"$crepl":["\\*users|10", "10"]}}`},
			":user":       &UserProfile{Tenant: "", Name: "user", Query: `{"ToR":{"$crepl":["\\*users|02", "02"]}, "RequestType":{"$crepl":["\\*users|2", "2"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c2", "c2"]}, "Account":{"$crepl":["\\*users|ivo", "ivo"]}, "Subject":{"$crepl":["\\*users|0724", "0724"]}, "Destination":{"$crepl":["\\*users|\\+402", "+402"]}, "SetupTime":{"$crepl":["\\*users|s2", "s2"]}, "AnswerTime":{"$crepl":["\\*users|t2", "t2"]}, "Usage":{"$crepl":["\\*users|11", "11"]}}`},
			"test:":       &UserProfile{Tenant: "test", Name: "", Query: `{"ToR":{"$crepl":["\\*users|03", "03"]}, "RequestType":{"$crepl":["\\*users|3", "3"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c3", "c3"]}, "Account":{"$crepl":["\\*users|elloy", "elloy"]}, "Subject":{"$crepl":["\\*users|0725", "0725"]}, "Destination":{"$crepl":["\\*users|\\+403", "+403"]}, "SetupTime":{"$crepl":["\\*users|s3", "s3"]}, "AnswerTime":{"$crepl":["\\*users|t3", "t3"]}, "Usage":{"$crepl":["\\*users|12", "11"]}}`},
			"test1:user1": &UserProfile{Tenant: "test1", Name: "user1", Query: `{"ToR":{"$crepl":["\\*users|04", "04"]}, "RequestType":{"$crepl":["\\*users|4", "4"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|call", "call"]}, "Account":{"$crepl":["\\*users|rif", "rif"]}, "Subject":{"$crepl":["\\*users|0726", "0726"]}, "Destination":{"$crepl":["\\*users|\\+404", "+404"]}, "SetupTime":{"$crepl":["\\*users|s4", "s4"]}, "AnswerTime":{"$crepl":["\\*users|t4", "t4"]}, "Usage":{"$crepl":["\\*users|13", "13"]}}`},
		},
		index: make(map[string]UserProfiles),
	}

	ur := &UsageRecord{
		ToR:         utils.USERS,
		RequestType: utils.USERS,
		Direction:   "*out",
		Tenant:      "",
		Category:    "call",
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: utils.USERS,
		SetupTime:   utils.USERS,
		AnswerTime:  utils.USERS,
		Usage:       "13",
	}

	err := LoadUserProfile(ur, false)
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	expected := &UsageRecord{
		ToR:         "04",
		RequestType: "4",
		Direction:   "*out",
		Tenant:      "",
		Category:    "call",
		Account:     "rif",
		Subject:     "0726",
		Destination: "+404",
		SetupTime:   "s4",
		AnswerTime:  "t4",
		Usage:       "13",
	}
	if !reflect.DeepEqual(ur, expected) {
		t.Errorf("Expected: %s got: %s", utils.ToIJSON(expected), utils.ToIJSON(ur))
	}
}

func TestUsersExternalCDRGetLoadUserProfileExtraFields(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:user":   &UserProfile{Tenant: "test", Name: "user", Query: `{"ToR":{"$crepl":["\\*users|01", "01"]}, "RequestType":{"$crepl":["\\*users|1", "1"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c1", "c1"]}, "Account":{"$crepl":["\\*users|dan", "dan"]}, "Subject":{"$crepl":["\\*users|0723", "0723"]}, "Destination":{"$crepl":["\\*users|\\+401", "+401"]}, "SetupTime":{"$crepl":["\\*users|s1", "s1"]}, "AnswerTime":{"$crepl":["\\*users|t1", "t1"]}, "Usage":{"$crepl":["\\*users|10", "10"]}}`},
			":user":       &UserProfile{Tenant: "", Name: "user", Query: `{"ToR":{"$crepl":["\\*users|02", "02"]}, "RequestType":{"$crepl":["\\*users|2", "2"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c2", "c2"]}, "Account":{"$crepl":["\\*users|ivo", "ivo"]}, "Subject":{"$crepl":["\\*users|0724", "0724"]}, "Destination":{"$crepl":["\\*users|\\+402", "+402"]}, "SetupTime":{"$crepl":["\\*users|s2", "s2"]}, "AnswerTime":{"$crepl":["\\*users|t2", "t2"]}, "Usage":{"$crepl":["\\*users|11", "11"]}}`},
			"test:":       &UserProfile{Tenant: "test", Name: "", Query: `{"ToR":{"$crepl":["\\*users|03", "03"]}, "RequestType":{"$crepl":["\\*users|3", "3"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c3", "c3"]}, "Account":{"$crepl":["\\*users|elloy", "elloy"]}, "Subject":{"$crepl":["\\*users|0725", "0725"]}, "Destination":{"$crepl":["\\*users|\\+403", "+403"]}, "SetupTime":{"$crepl":["\\*users|s3", "s3"]}, "AnswerTime":{"$crepl":["\\*users|t3", "t3"]}, "Usage":{"$crepl":["\\*users|12", "11"]}}`},
			"test1:user1": &UserProfile{Tenant: "test1", Name: "user1", Query: `{"ToR":{"$crepl":["\\*users|04", "04"]}, "RequestType":{"$crepl":["\\*users|4", "4"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|call", "call"]}, "Account":{"$crepl":["\\*users|rif", "rif"]}, "Subject":{"$crepl":["\\*users|0726", "0726"]}, "Destination":{"$crepl":["\\*users|\\+404", "+404"]}, "SetupTime":{"$crepl":["\\*users|s4", "s4"]}, "AnswerTime":{"$crepl":["\\*users|t4", "t4"]}, "Usage":{"$crepl":["\\*users|13", "13"]}, "Test":{"$crepl":["\\*users|1", "1"]}}`},
		},
		index: make(map[string]UserProfiles),
	}

	ur := &ExternalCDR{
		ToR:         utils.USERS,
		RequestType: utils.USERS,
		Direction:   "*out",
		Tenant:      "",
		Category:    "call",
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: utils.USERS,
		SetupTime:   utils.USERS,
		AnswerTime:  utils.USERS,
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
		},
	}

	err := LoadUserProfile(ur, false)
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	expected := &ExternalCDR{
		ToR:         "04",
		RequestType: "4",
		Direction:   "*out",
		Tenant:      "",
		Category:    "call",
		Account:     "rif",
		Subject:     "0726",
		Destination: "+404",
		SetupTime:   "s4",
		AnswerTime:  "t4",
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
		},
	}
	if !reflect.DeepEqual(ur, expected) {
		t.Errorf("Expected: %+v got: %+v", expected, ur)
	}
}

func TestUsersExternalCDRGetLoadUserProfileExtraFieldsNotFound(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:user":   &UserProfile{Tenant: "test", Name: "user", Query: `{"ToR":{"$crepl":["\\*users|01", "01"]}, "RequestType":{"$crepl":["\\*users|1", "1"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c1", "c1"]}, "Account":{"$crepl":["\\*users|dan", "dan"]}, "Subject":{"$crepl":["\\*users|0723", "0723"]}, "Destination":{"$crepl":["\\*users|\\+401", "+401"]}, "SetupTime":{"$crepl":["\\*users|s1", "s1"]}, "AnswerTime":{"$crepl":["\\*users|t1", "t1"]}, "Usage":{"$crepl":["\\*users|10", "10"]}}`},
			":user":       &UserProfile{Tenant: "", Name: "user", Query: `{"ToR":{"$crepl":["\\*users|02", "02"]}, "RequestType":{"$crepl":["\\*users|2", "2"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c2", "c2"]}, "Account":{"$crepl":["\\*users|ivo", "ivo"]}, "Subject":{"$crepl":["\\*users|0724", "0724"]}, "Destination":{"$crepl":["\\*users|\\+402", "+402"]}, "SetupTime":{"$crepl":["\\*users|s2", "s2"]}, "AnswerTime":{"$crepl":["\\*users|t2", "t2"]}, "Usage":{"$crepl":["\\*users|11", "11"]}}`},
			"test:":       &UserProfile{Tenant: "test", Name: "", Query: `{"ToR":{"$crepl":["\\*users|03", "03"]}, "RequestType":{"$crepl":["\\*users|3", "3"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c3", "c3"]}, "Account":{"$crepl":["\\*users|elloy", "elloy"]}, "Subject":{"$crepl":["\\*users|0725", "0725"]}, "Destination":{"$crepl":["\\*users|\\+403", "+403"]}, "SetupTime":{"$crepl":["\\*users|s3", "s3"]}, "AnswerTime":{"$crepl":["\\*users|t3", "t3"]}, "Usage":{"$crepl":["\\*users|12", "11"]}}`},
			"test1:user1": &UserProfile{Tenant: "test1", Name: "user1", Query: `{"ToR":{"$crepl":["\\*users|04", "04"]}, "RequestType":{"$crepl":["\\*users|4", "4"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|call", "call"]}, "Account":{"$crepl":["\\*users|rif", "rif"]}, "Subject":{"$crepl":["\\*users|0726", "0726"]}, "Destination":{"$crepl":["\\*users|\\+404", "+404"]}, "SetupTime":{"$crepl":["\\*users|s4", "s4"]}, "AnswerTime":{"$crepl":["\\*users|t4", "t4"]}, "Usage":{"$crepl":["\\*users|13", "13"]}, "Test":{"$crepl":["\\*users|2", "2"]}}`},
		},
		index: make(map[string]UserProfiles),
	}

	ur := &ExternalCDR{
		ToR:         utils.USERS,
		RequestType: utils.USERS,
		Direction:   "*out",
		Tenant:      "",
		Category:    "call",
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: utils.USERS,
		SetupTime:   utils.USERS,
		AnswerTime:  utils.USERS,
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
		},
	}

	err := LoadUserProfile(ur, true)
	if err != utils.ErrUserNotFound {
		t.Error("Error detecting err in loading user profile: ", err)
	}
}

func TestUsersExternalCDRGetLoadUserProfileExtraFieldsSet(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:user":   &UserProfile{Tenant: "test", Name: "user", Query: `{"ToR":{"$crepl":["\\*users|01", "01"]}, "RequestType":{"$crepl":["\\*users|1", "1"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c1", "c1"]}, "Account":{"$crepl":["\\*users|dan", "dan"]}, "Subject":{"$crepl":["\\*users|0723", "0723"]}, "Destination":{"$crepl":["\\*users|\\+401", "+401"]}, "SetupTime":{"$crepl":["\\*users|s1", "s1"]}, "AnswerTime":{"$crepl":["\\*users|t1", "t1"]}, "Usage":{"$crepl":["\\*users|10", "10"]}}`},
			":user":       &UserProfile{Tenant: "", Name: "user", Query: `{"ToR":{"$crepl":["\\*users|02", "02"]}, "RequestType":{"$crepl":["\\*users|2", "2"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c2", "c2"]}, "Account":{"$crepl":["\\*users|ivo", "ivo"]}, "Subject":{"$crepl":["\\*users|0724", "0724"]}, "Destination":{"$crepl":["\\*users|\\+402", "+402"]}, "SetupTime":{"$crepl":["\\*users|s2", "s2"]}, "AnswerTime":{"$crepl":["\\*users|t2", "t2"]}, "Usage":{"$crepl":["\\*users|11", "11"]}}`},
			"test:":       &UserProfile{Tenant: "test", Name: "", Query: `{"ToR":{"$crepl":["\\*users|03", "03"]}, "RequestType":{"$crepl":["\\*users|3", "3"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|c3", "c3"]}, "Account":{"$crepl":["\\*users|elloy", "elloy"]}, "Subject":{"$crepl":["\\*users|0725", "0725"]}, "Destination":{"$crepl":["\\*users|\\+403", "+403"]}, "SetupTime":{"$crepl":["\\*users|s3", "s3"]}, "AnswerTime":{"$crepl":["\\*users|t3", "t3"]}, "Usage":{"$crepl":["\\*users|12", "11"]}}`},
			"test1:user1": &UserProfile{Tenant: "test1", Name: "user1", Query: `{"ToR":{"$crepl":["\\*users|04", "04"]}, "RequestType":{"$crepl":["\\*users|4", "4"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|call", "call"]}, "Account":{"$crepl":["\\*users|rif", "rif"]}, "Subject":{"$crepl":["\\*users|0726", "0726"]}, "Destination":{"$crepl":["\\*users|\\+404", "+404"]}, "SetupTime":{"$crepl":["\\*users|s4", "s4"]}, "AnswerTime":{"$crepl":["\\*users|t4", "t4"]}, "Usage":{"$crepl":["\\*users|13", "13"]}, "Test":{"$crepl":["\\*users|1", "1"]}, "Best":{"$crepl":["\\*users|BestValue", "BestValue"]}}`},
		},
		index: make(map[string]UserProfiles),
	}

	ur := &ExternalCDR{
		ToR:         utils.USERS,
		RequestType: utils.USERS,
		Direction:   "*out",
		Tenant:      "",
		Category:    "call",
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: utils.USERS,
		SetupTime:   utils.USERS,
		AnswerTime:  utils.USERS,
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
			"Best": utils.USERS,
		},
	}

	err := LoadUserProfile(ur, false)
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	expected := &ExternalCDR{
		ToR:         "04",
		RequestType: "4",
		Direction:   "*out",
		Tenant:      "",
		Category:    "call",
		Account:     "rif",
		Subject:     "0726",
		Destination: "+404",
		SetupTime:   "s4",
		AnswerTime:  "t4",
		Usage:       "13",
		ExtraFields: map[string]string{
			"Test": "1",
			"Best": "BestValue",
		},
	}
	if !reflect.DeepEqual(ur, expected) {
		t.Errorf("Expected: %+v got: %+v", expected, ur)
	}
}

func TestUsersCallDescLoadUserProfile(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:dan":      &UserProfile{Tenant: "test", Name: "dan", Query: `{"Tenant":{"$crepl":["\\*users|test", "test"]}, "Category":{"$crepl":["\\*users|call1", "call1"]}, "Account":{"$crepl":["\\*users|dan", "dan"]}, "Subject":{"$crepl":["\\*users|dan", "dan"]}, "Cli":{"$crepl":["\\*users|\\+4986517174963", "+4986517174963"]}}`},
			"test:danvoice": &UserProfile{Tenant: "test", Name: "danvoice", Query: `{"TOR":{"$crepl":["\\*users|\\*voice", "*voice"]}, "RequestType":{"$crepl":["\\*users|\\*prepaid", "*prepaid"]}, "Category":{"$crepl":["\\*users|call1", "call1"]}, "Account":{"$crepl":["\\*users|dan", "dan"]}, "Subject":{"$crepl":["\\*users|0723", "0723"]}}`},
			"test:rif":      &UserProfile{Tenant: "test", Name: "rif", Query: `{"RequestType":{"$crepl":["\\*users|\\*postpaid", "*postpaid"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|call", "call"]}, "Account":{"$crepl":["\\*users|rif", "rif"]}, "Subject":{"$crepl":["\\*users|0726", "0726"]}}`},
		},
		index: make(map[string]UserProfiles),
	}
	startTime := time.Now()
	cd := &CallDescriptor{
		TOR:         "*sms",
		Tenant:      utils.USERS,
		Category:    utils.USERS,
		Subject:     utils.USERS,
		Account:     utils.USERS,
		Destination: "+4986517174963",
		TimeStart:   startTime,
		TimeEnd:     startTime.Add(time.Duration(1) * time.Minute),
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	expected := &CallDescriptor{
		TOR:         "*sms",
		Tenant:      "test",
		Category:    "call1",
		Account:     "dan",
		Subject:     "dan",
		Destination: "+4986517174963",
		TimeStart:   startTime,
		TimeEnd:     startTime.Add(time.Duration(1) * time.Minute),
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	err := LoadUserProfile(cd, false)
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	if !reflect.DeepEqual(expected, cd) {
		t.Errorf("Expected: %s got: %s", utils.ToIJSON(expected), utils.ToIJSON(cd))
	}
}

func TestUsersCDRLoadUserProfile(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:dan":      &UserProfile{Tenant: "test", Name: "dan", Query: `{"RequestType":{"$crepl":["\\*users|\\*prepaid", "*prepaid"]}, "Tenant":{"$crepl":["\\*users|test", "test"]}, "Category":{"$crepl":["\\*users|call1", "call1"]}, "Account":{"$crepl":["\\*users|dan", "dan"]}, "Subject":{"$crepl":["\\*users|dan", "dan"]}, "Cli":{"$crepl":["\\*users|\\+4986517174963", "+4986517174963"]}}`},
			"test:danvoice": &UserProfile{Tenant: "test", Name: "danvoice", Query: `{"ToR":{"$crepl":["\\*users|\\*voice", "*voice"]}, "RequestType":{"$crepl":["\\*users|\\*prepaid", "*prepaid"]}, "Category":{"$crepl":["\\*users|call1", "call1"]}, "Account":{"$crepl":["\\*users|dan", "dan"]}, "Subject":{"$crepl":["\\*users|0723", "0723"]}}`},
			"test:rif":      &UserProfile{Tenant: "test", Name: "rif", Query: `{"RequestType":{"$crepl":["\\*users|\\*postpaid", "*postpaid"]}, "Direction":{"$crepl":["\\*users|\\*out", "*out"]}, "Category":{"$crepl":["\\*users|call", "call"]}, "Account":{"$crepl":["\\*users|rif", "rif"]}, "Subject":{"$crepl":["\\*users|0726", "0726"]}}`},
		},
		index: make(map[string]UserProfiles),
	}
	startTime := time.Now()
	cdr := &CDR{
		ToR:         "*sms",
		RequestType: utils.USERS,
		Tenant:      utils.USERS,
		Category:    utils.USERS,
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: "+4986517174963",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	expected := &CDR{
		ToR:         "*sms",
		RequestType: "*prepaid",
		Tenant:      "test",
		Category:    "call1",
		Account:     "dan",
		Subject:     "dan",
		Destination: "+4986517174963",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	err := LoadUserProfile(cdr, false)
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	if !reflect.DeepEqual(expected, cdr) {
		t.Errorf("Expected: %s got: %s", utils.ToIJSON(expected), utils.ToIJSON(cdr))
	}
}

func TestUsersCDRLoadUserProfileAlternateQuery(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:dan":      &UserProfile{Tenant: "test", Name: "dan", Query: `{"RequestType":{"$usr":"*prepaid"}, "Tenant":{"$usr":"test"}, "Category":{"$usr":"call1"}, "Account":{"$usr":"dan"}, "Subject":{"$usr": "dan"}, "Cli":{"$usr": "+4986517174963"}}`},
			"test:danvoice": &UserProfile{Tenant: "test", Name: "danvoice", Query: `{"ToR":{"$usr": "*voice"}, "RequestType":{"$usr": "*prepaid"}, "Category":{"$usr": "call1"}, "Account":{"$usr": "dan"}, "Subject":{"$usr": "0723"}}`},
			"test:rif":      &UserProfile{Tenant: "test", Name: "rif", Query: `{"RequestType":{"$usr": "*postpaid"}, "Direction":{"$usr": "*out"}, "Category":{"$usr": "call"}, "Account":{"$usr": "rif"}, "Subject":{"$usr": "0726"}}`},
		},
		index: make(map[string]UserProfiles),
	}
	startTime := time.Now()
	cdr := &CDR{
		ToR:         "*sms",
		RequestType: utils.USERS,
		Tenant:      utils.USERS,
		Category:    utils.USERS,
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: "+4986517174963",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	expected := &CDR{
		ToR:         "*sms",
		RequestType: "*prepaid",
		Tenant:      "test",
		Category:    "call1",
		Account:     "dan",
		Subject:     "dan",
		Destination: "+4986517174963",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	err := LoadUserProfile(cdr, false)
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	if !reflect.DeepEqual(expected, cdr) {
		t.Errorf("Expected: %s got: %s", utils.ToIJSON(expected), utils.ToIJSON(cdr))
	}
}

func TestUsersCDRLoadUserProfileUsrRepl(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:rif": &UserProfile{Tenant: "test", Name: "rif", Query: `{"Tenant":{"$usr":"test"}, "RequestType":{"$usrpl": ["\\*prepaid", "*postpaid"]}, "Direction":{"$usrpl": ["", "*out"]}, "Category":{"$usr": "call"}, "Account":{"$usr": "rif"}, "Subject":{"$usr": "0726"}, "Destination":{"$usrpl":["\\+(\\d+)", "${1}"]}}`},
		},
		index: make(map[string]UserProfiles),
	}
	startTime := time.Now()
	cdr := &CDR{
		ToR:         "*sms",
		RequestType: utils.USERS,
		Tenant:      utils.USERS,
		Category:    utils.USERS,
		Account:     utils.USERS,
		Subject:     utils.USERS,
		Destination: "+4986517174963",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	expected := &CDR{
		ToR:         "*sms",
		RequestType: "*postpaid",
		Direction:   "*out",
		Tenant:      "test",
		Category:    "call",
		Account:     "rif",
		Subject:     "0726",
		Destination: "4986517174963",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{"Cli": "+4986517174963"},
	}
	err := LoadUserProfile(cdr, false)
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	if !reflect.DeepEqual(expected, cdr) {
		t.Errorf("Expected: %s got: %s", utils.ToIJSON(expected), utils.ToIJSON(cdr))
	}
}
func TestUsersPrefix(t *testing.T) {
	userService = &UserMap{
		table: map[string]*UserProfile{
			"test:rif": &UserProfile{Tenant: "t1", Name: "t1", Query: `{"sip_from_host":{"$in":["206.222.29.2","206.222.29.3","206.222.29.4","206.222.29.5","206.222.29.6"]}, "Destination":{"$crepl":["^9023(\\d+)","${1}"]}, "Account":{"$usr":"t1"}, "Tenant":{"$usr":"t1"}, "Category":{"$usr":"call"}, "RequestType":{"$usr":"*postpaid"}, "Direction":{"$usr":"*out"}, "direction":{"$usr": "outbound"}, "Subject":{"$repl":["^9023(\\d+)","${1}"]}}`},
		},
		index: make(map[string]UserProfiles),
	}
	startTime := time.Now()
	cdr := &CDR{
		ToR:         "*sms",
		RequestType: utils.USERS,
		Tenant:      utils.USERS,
		Category:    utils.USERS,
		Account:     utils.USERS,
		Subject:     "9023456789",
		Destination: "9023456789",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{
			"sip_from_host": "206.222.29.3",
			"direction":     "*users",
		},
	}
	expected := &CDR{
		ToR:         "*sms",
		RequestType: "*postpaid",
		Direction:   "*out",
		Tenant:      "t1",
		Category:    "call",
		Account:     "t1",
		Subject:     "456789",
		Destination: "456789",
		SetupTime:   startTime,
		AnswerTime:  startTime,
		Usage:       time.Duration(1) * time.Minute,
		ExtraFields: map[string]string{
			"sip_from_host": "206.222.29.3",
			"direction":     "outbound",
		},
	}
	err := LoadUserProfile(cdr, false)
	if err != nil {
		t.Error("Error loading user profile: ", err)
	}
	if !reflect.DeepEqual(expected, cdr) {
		t.Errorf("Expected: %s got: %s", utils.ToIJSON(expected), utils.ToIJSON(cdr))
	}
}
