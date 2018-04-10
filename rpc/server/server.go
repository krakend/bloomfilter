package server

import (
	"context"
	"fmt"
	"net"
	"net/rpc"

	rpc_bf "github.com/letgoapp/go-bloomfilter/rpc"
)

func New(ctx context.Context, cfg rpc_bf.Config) *rpc_bf.Bloomfilter {
	bf := rpc_bf.New(ctx, cfg)

	go Serve(ctx, cfg.Port, bf)

	return bf
}

func Serve(ctx context.Context, port int, bf *rpc_bf.Bloomfilter) {
	s := rpc.NewServer()
	s.Register(bf)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("unable to setup RPC listener:", err.Error())
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				bf.Close()
				return
			}
		}
	}()

	s.Accept(l)
}
