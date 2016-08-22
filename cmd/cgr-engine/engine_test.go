
package main

import (
	"net/rpc"
	"testing"
)

/*func TestRPCGet(t *testing.T) {
	client, err := rpc.DialHTTPPath("tcp", "localhost:2000", "/rpc")
	if err != nil {
		t.Error("Balancer server not started!")
		t.FailNow()
	}
	var reply string
	client.Call("Responder.Get", "test", &reply)
	const expect = "12223"
	if reply != expect {
		t.Errorf("replay == %v, want %v", reply, expect)
	}
}*/

func BenchmarkRPCGet(b *testing.B) {
	b.StopTimer()
	client, _ := rpc.DialHTTPPath("tcp", "localhost:2000", "/rpc")
	b.StartTimer()
	var reply string
	for i := 0; i < b.N; i++ {
		client.Call("Responder.Get", "test", &reply)
	}
}
