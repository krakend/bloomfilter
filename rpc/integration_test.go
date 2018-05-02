package rpc

import (
	"context"
	"fmt"
	"net/rpc"

	"github.com/devopsfaith/bloomfilter"
	"github.com/devopsfaith/bloomfilter/rotate"
)

func Example_integration() {
	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		fmt.Printf("dialing error: %s", err.Error())
		return
	}

	var (
		addOutput   AddOutput
		checkOutput CheckOutput
		unionOutput UnionOutput
		elems1      = [][]byte{[]byte("rrrr"), []byte("elem2")}
		elems2      = [][]byte{[]byte("house")}
		elems3      = [][]byte{[]byte("house"), []byte("mouse")}
		cfg         = rotate.Config{
			Config: bloomfilter.Config{
				N:        10000000,
				P:        0.0000001,
				HashName: "optimal",
			},
			TTL: 1500,
		}
	)

	err = client.Call("BloomfilterRPC.Add", AddInput{elems1}, &addOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}

	err = client.Call("BloomfilterRPC.Check", CheckInput{elems1}, &checkOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}
	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		fmt.Printf("checks error, expected true elements")
		return
	}

	divCall := client.Go("BloomfilterRPC.Check", elems1, &checkOutput, nil)
	<-divCall.Done

	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		fmt.Printf("checks error, expected true elements")
		return
	}

	var bf2 = rotate.New(context.Background(), cfg)
	bf2.Add([]byte("house"))

	err = client.Call("BloomfilterRPC.Union", UnionInput{bf2}, &unionOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}
	fmt.Println(unionOutput.Capacity < 1e-6)

	err = client.Call("BloomfilterRPC.Check", CheckInput{elems2}, &checkOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}
	fmt.Println(checkOutput.Checks)
	if len(checkOutput.Checks) != 1 || !checkOutput.Checks[0] {
		fmt.Println("checks error, expected true element")
		return
	}

	var bf3 = rotate.New(context.Background(), cfg)
	bf3.Add([]byte("mouse"))

	divCall = client.Go("BloomfilterRPC.Union", UnionInput{bf3}, &unionOutput, nil)
	<-divCall.Done

	err = client.Call("BloomfilterRPC.Check", CheckInput{elems3}, &checkOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}

	fmt.Println(unionOutput.Capacity < 1e-6)
	fmt.Println(checkOutput.Checks)
	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		fmt.Println("checks error, expected true element")
		return
	}

	// Output:
	// true
	// [true]
	// true
	// [true true]
}
