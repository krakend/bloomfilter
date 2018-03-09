package bloomfilter

import (
	"net"
	"net/http"
	"net/rpc"
	"testing"
)

func TestBFType_Add(t *testing.T) {
	bf := new(BFType)
	rpc.Register(bf)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		t.Errorf("listen error: %s", err.Error())
	}
	go http.Serve(l, nil)
	client, err := rpc.DialHTTP("tcp", ":1234")
	if err != nil {
		t.Errorf("dialing error: %s", err.Error())
	}

	bf.bf = new(Bloomfilter)
	var (
		counter int
		elems   = [][]byte{[]byte("elem1"), []byte("elem2")}
	)
	err = client.Call("BFType.Add", elems, &counter)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	checks := make([]bool, 2)
	err = client.Call("BFType.Check", elems, &checks)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if len(checks) != 2 || !checks[0] || !checks[1] {
		if !checks[0] {
			t.Errorf("checks error, expected 1 true elem, got: %v", checks)
		}
	}
}
