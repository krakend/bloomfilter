package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	rpc_bf "github.com/letgoapp/go-bloomfilter/rpc"
)

func New(ctx context.Context, cfg rpc_bf.Config) *rpc_bf.Bloomfilter {
	bf := rpc_bf.New(ctx, cfg)

	go Serve(ctx, cfg.Port, bf)

	return bf
}

func Serve(ctx context.Context, port int, bf *rpc_bf.Bloomfilter) error {
	err := rpc.Register(bf)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
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

	return http.Serve(l, nil)
}
