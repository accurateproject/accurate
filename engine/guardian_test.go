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
