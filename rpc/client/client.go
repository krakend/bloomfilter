package client

import (
	"errors"
	"fmt"
	"net/rpc"

	"github.com/letgoapp/go-bloomfilter/rotate"
	rpc_bf "github.com/letgoapp/go-bloomfilter/rpc"
)

type Bloomfilter struct {
	client []*rpc.Client
}

func New(address ...string) (*Bloomfilter, error) {
	clients := make([]*rpc.Client, len(address))
	for i, addr := range address {
		client, err := rpc.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		clients[i] = client
	}
	return &Bloomfilter{clients}, nil
}

func (b *Bloomfilter) Add(elem []byte) {
	for _, c := range b.client {
		var addOutput rpc_bf.AddOutput
		if err := c.Call("BloomfilterRPC.Add", rpc_bf.AddInput{[][]byte{elem}}, &addOutput); err != nil {
			fmt.Println("error on adding remote bloomfilter:", err.Error())
		}
	}
}

func (b *Bloomfilter) Check(elem []byte) bool {
	for _, c := range b.client {
		var checkOutput rpc_bf.CheckOutput
		if err := c.Call("BloomfilterRPC.Check", rpc_bf.CheckInput{[][]byte{elem}}, &checkOutput); err != nil {
			fmt.Println("error on check remote bloomfilter:", err.Error())
			return false
		}
		for _, v := range checkOutput.Checks {
			if !v {
				return false
			}
		}
	}
	return true
}

func (b *Bloomfilter) Union(that interface{}) (float64, error) {
	v, ok := that.(*rotate.Bloomfilter)
	if !ok {
		return -1.0, errors.New("invalide argument to Union, expected rotate.Bloomfilter")
	}
	var capacity float64
	for _, c := range b.client {
		var unionOutput rpc_bf.UnionOutput
		if err := c.Call("BloomfilterRPC.Union", rpc_bf.UnionInput{v}, &unionOutput); err != nil {
			fmt.Println("error on union remote bloomfilter:", err.Error())
			return -1.0, err
		}
		capacity += unionOutput.Capacity
	}
	return capacity / float64(len(b.client)), nil
}

func (b *Bloomfilter) Close() {
	for _, c := range b.client {
		c.Close()
	}
}
