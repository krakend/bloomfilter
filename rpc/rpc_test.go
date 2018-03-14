package rpc

import (
	"net"
	"net/http"
	"net/rpc"
	"testing"
)

func TestBFType_ok(t *testing.T) {
	bf := new(BFType)
	bf.bf = NewBloomfilter(testCfg)
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

	var (
		counterFloat float64
		counterInt   int
		elems1       = [][]byte{[]byte("elem1"), []byte("elem2")}
		elems2       = [][]byte{[]byte("house")}
		elems3       = [][]byte{[]byte("house"), []byte("mouse")}
	)

	err = client.Call("BFType.Add", elems1, &counterInt)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	checks := make([]bool, 2)

	err = client.Call("BFType.Check", elems1, &checks)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if !checks[0] || !checks[1] {
		t.Errorf("checks error, expected true elements")
	}

	checks[0] = false
	checks[1] = false
	divCall := client.Go("BFType.Check", elems1, &checks, nil)
	<-divCall.Done
	if !checks[0] || !checks[1] {
		t.Errorf("checks error, expected true elements")
	}

	var b = NewBloomfilter(testCfg)
	b.Add([]byte("house"))

	err = client.Call("BFType.Union", b, &counterFloat)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	checks[0] = false
	err = client.Call("BFType.Check", elems2, &checks)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if !checks[0] {
		t.Error("checks error, expected true element")
	}

	var b2 = NewBloomfilter(testCfg)
	b2.Add([]byte("mouse"))

	divCall = client.Go("BFType.Union", b2, &counterFloat, nil)
	<-divCall.Done

	checks[0] = false

	err = client.Call("BFType.Check", elems3, &checks)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if !checks[0] || !checks[1] {
		t.Error("checks error, expected true element")
	}
}
