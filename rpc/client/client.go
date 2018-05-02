// Package client implements an rpc client for the bloomfilter, along with Add and Check methods.
package client

import (
	"errors"
	"fmt"
	"net/rpc"

	"github.com/devopsfaith/bloomfilter/rotate"
	rpc_bf "github.com/devopsfaith/bloomfilter/rpc"
)

// Bloomfilter rpc client type
type Bloomfilter struct {
	client *rpc.Client
}

// New creates a new bloomfilter rpc client with address
func New(address string) (*Bloomfilter, error) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Bloomfilter{client}, nil
}

// Add element through bloomfilter rpc client
func (b *Bloomfilter) Add(elem []byte) {
	var addOutput rpc_bf.AddOutput
	if err := b.client.Call("BloomfilterRPC.Add", rpc_bf.AddInput{[][]byte{elem}}, &addOutput); err != nil {
		fmt.Println("error on adding bloomfilter:", err.Error())
	}
}

// Check present element through bloomfilter rpc client
func (b *Bloomfilter) Check(elem []byte) bool {
	var checkOutput rpc_bf.CheckOutput
	if err := b.client.Call("BloomfilterRPC.Check", rpc_bf.CheckInput{[][]byte{elem}}, &checkOutput); err != nil {
		fmt.Println("error on check bloomfilter:", err.Error())
		return false
	}
	for _, v := range checkOutput.Checks {
		if !v {
			return false
		}
	}
	return true
}

// Union element through bloomfilter rpc client with sliding bloomfilters
func (b *Bloomfilter) Union(that interface{}) (float64, error) {
	v, ok := that.(*rotate.Bloomfilter)
	if !ok {
		return -1.0, errors.New("invalide argument to Union, expected rotate.Bloomfilter")
	}
	var unionOutput rpc_bf.UnionOutput
	if err := b.client.Call("BloomfilterRPC.Union", rpc_bf.UnionInput{v}, &unionOutput); err != nil {
		fmt.Println("error on union bloomfilter:", err.Error())
		return -1.0, err
	}

	return unionOutput.Capacity, nil
}

// Close bloomfilter rpc client
func (b *Bloomfilter) Close() {
	b.client.Close()
}
