package rpc

import (
	"context"
	"net"
	"net/http"
	"net/rpc"
	"testing"

	"github.com/letgoapp/go-bloomfilter/rotate"
	"github.com/letgoapp/go-bloomfilter/testutils"
)

func TestBFType_ok(t *testing.T) {
	bf := New(context.Background(), 5, testutils.TestCfg)
	err := rpc.Register(bf)
	if err != nil {
		t.Errorf("register error: %s", err.Error())
		return
	}
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		t.Errorf("listen error: %s", err.Error())
		return
	}
	go http.Serve(l, nil)
	client, err := rpc.DialHTTP("tcp", ":1234")
	if err != nil {
		t.Errorf("dialing error: %s", err.Error())
		return
	}

	var (
		addOutput   AddOutput
		checkOutput CheckOutput
		unionOutput UnionOutput
		elems1      = [][]byte{[]byte("elem1"), []byte("elem2")}
		elems2      = [][]byte{[]byte("house")}
		elems3      = [][]byte{[]byte("house"), []byte("mouse")}
	)

	err = client.Call("BFType.Add", AddInput{elems1}, &addOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	err = client.Call("BFType.Check", CheckInput{elems1}, &checkOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		t.Errorf("checks error, expected true elements")
		return
	}

	divCall := client.Go("BFType.Check", elems1, &checkOutput, nil)
	<-divCall.Done

	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		t.Errorf("checks error, expected true elements")
		return
	}

	var b = rotate.New(context.Background(), 5, testutils.TestCfg)
	b.Add([]byte("house"))

	err = client.Call("BFType.Union", UnionInput{b}, &unionOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	err = client.Call("BFType.Check", CheckInput{elems2}, &checkOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if len(checkOutput.Checks) != 1 || !checkOutput.Checks[0] {
		t.Error("checks error, expected true element")
		return
	}

	var b2 = rotate.New(context.Background(), 5, testutils.TestCfg)
	b2.Add([]byte("mouse"))

	divCall = client.Go("BFType.Union", UnionInput{b2}, &unionOutput, nil)
	<-divCall.Done

	err = client.Call("BFType.Check", CheckInput{elems3}, &checkOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		t.Error("checks error, expected true element")
		return
	}

	bf.Close()
	return
}
