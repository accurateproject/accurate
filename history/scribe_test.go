
package history

import (
	"strconv"
	"testing"
)

func TestHistorySet(t *testing.T) {
	rs := records{&Record{Id: "first"}}
	second := &Record{Id: "first"}
	rs.Modify(second)
	if len(rs) != 1 || rs[0] != second {
		t.Error("error setting new value: ", rs[0])
	}
}

func TestHistoryAdd(t *testing.T) {
	rs := records{&Record{Id: "first"}}
	second := &Record{Id: "second"}
	rs = rs.Modify(second)
	if len(rs) != 2 || rs[1] != second {
		t.Error("error setting new value: ", rs)
	}
}

func TestHistoryRemove(t *testing.T) {
	rs := records{&Record{Id: "first"}, &Record{Id: "second"}}
	rs = rs.Modify(&Record{Id: "first", Deleted: true})
	if len(rs) != 1 || rs[0].Id != "second" {
		t.Error("error deleting record: ", rs)
	}
}

func BenchmarkModify(b *testing.B) {
	var rs records
	for i := 0; i < 1000; i++ {
		rs = rs.Modify(&Record{Id: strconv.Itoa(i)})
	}
	for i := 0; i < b.N; i++ {
		rs.Modify(&Record{Id: "400"})
	}
}
