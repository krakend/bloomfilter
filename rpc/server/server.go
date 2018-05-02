// Package server implements an rpc server for the bloomfilter, registering a bloomfilter and accepting a tcp listener.
package server

import (
	"context"
	"fmt"
	"net"
	"net/rpc"

	rpc_bf "github.com/devopsfaith/bloomfilter/rpc"
)

// New creates an rpc bloomfilter and launches a serving goroutine
func New(ctx context.Context, cfg rpc_bf.Config) *rpc_bf.Bloomfilter {
	bf := rpc_bf.New(ctx, cfg)

	go Serve(ctx, cfg.Port, bf)

	return bf
}

// Serve creates an rpc server, registers a bloomfilter, accepts a tcp listener and closes when catching context done
func Serve(ctx context.Context, port int, bf *rpc_bf.Bloomfilter) {
	s := rpc.NewServer()

	if err := s.Register(&bf.BloomfilterRPC); err != nil {
		fmt.Println("server register error:", err.Error())
		return
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("unable to setup RPC listener:", err.Error())
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				l.Close()
				bf.Close()
				return
			}
		}
	}()

	s.Accept(l)
}
