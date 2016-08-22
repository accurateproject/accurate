
package history

import (
	"io"
	"reflect"
	"sort"
)

const (
	DESTINATIONS_FN    = "destinations.json"
	RATING_PLANS_FN    = "rating_plans.json"
	RATING_PROFILES_FN = "rating_profiles.json"
)

type Record struct {
	Id       string
	Filename string
	Payload  []byte
	Deleted  bool
}

type records []*Record

var (
	recordsMap  = make(map[string]records)
	filenameMap = make(map[reflect.Type]string)
)

func (rs records) Len() int {
	return len(rs)
}

func (rs records) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs records) Less(i, j int) bool {
	return rs[i].Id < rs[j].Id
}

func (rs records) Sort() {
	sort.Sort(rs)
}

func (rs records) Modify(rec *Record) records {
	//rs.Sort()
	n := len(rs)
	i := sort.Search(n, func(i int) bool { return rs[i].Id >= rec.Id })
	if i < n && rs[i].Id == rec.Id {
		if rec.Deleted {
			// delete
			rs = append(rs[:i], rs[i+1:]...)
		} else {
			rs[i] = rec
		}
	} else {
		// i is the index where it would be inserted.
		rs = append(rs, nil)
		copy(rs[i+1:], rs[i:])
		rs[i] = rec
	}
	return rs
}

func format(b io.Writer, recs records) error {
	recs.Sort()
	b.Write([]byte("["))
	for i, r := range recs {
		b.Write(r.Payload)
		if i < len(recs)-1 {
			b.Write([]byte(",\n"))
		}
	}
	b.Write([]byte("]"))
	return nil
}
