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

func TestBloomfilter_ok(t *testing.T) {
	b := New(context.Background(), Config{rotate.Config{testutils.TestCfg, 5}, 1234})
	err := rpc.Register(b)
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

	err = client.Call("Bloomfilter.Add", AddInput{elems1}, &addOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	err = client.Call("Bloomfilter.Check", CheckInput{elems1}, &checkOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		t.Errorf("checks error, expected true elements")
		return
	}

	divCall := client.Go("Bloomfilter.Check", elems1, &checkOutput, nil)
	<-divCall.Done

	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		t.Errorf("checks error, expected true elements")
		return
	}

	var bf2 = rotate.New(context.Background(), rotate.Config{testutils.TestCfg, 5})
	bf2.Add([]byte("house"))

	err = client.Call("Bloomfilter.Union", UnionInput{bf2}, &unionOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	err = client.Call("Bloomfilter.Check", CheckInput{elems2}, &checkOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if len(checkOutput.Checks) != 1 || !checkOutput.Checks[0] {
		t.Error("checks error, expected true element")
		return
	}

	var bf3 = rotate.New(context.Background(), rotate.Config{testutils.TestCfg, 5})
	bf3.Add([]byte("mouse"))

	divCall = client.Go("Bloomfilter.Union", UnionInput{bf3}, &unionOutput, nil)
	<-divCall.Done

	err = client.Call("Bloomfilter.Check", CheckInput{elems3}, &checkOutput)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		t.Error("checks error, expected true element")
		return
	}

	b.Close()
	return
}
