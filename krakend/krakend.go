package krakend

import (
	"context"
	"errors"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	bloomfilter "github.com/letgoapp/go-bloomfilter"
	bf_rpc "github.com/letgoapp/go-bloomfilter/rpc"
	"github.com/letgoapp/go-bloomfilter/rpc/server"
)

const Namespace = "bloomfilter"

var (
	errNoConfig    = errors.New("no config for the bloomfilter")
	errWrongConfig = errors.New("invalid config for the bloomfilter")
)

func Register(ctx context.Context, cfg config.ServiceConfig, logger logging.Logger) (bloomfilter.Bloomfilter, error) {
	data, ok := cfg.ExtraConfig[Namespace]
	if !ok {
		logger.Info(errNoConfig.Error(), cfg.ExtraConfig)
		return new(bloomfilter.EmptySet), errNoConfig
	}
	v, ok := data.(bf_rpc.Config)
	if !ok {
		logger.Info(errWrongConfig.Error(), data)
		return new(bloomfilter.EmptySet), errWrongConfig
	}

	bf := server.New(ctx, v)
	return bf.Bloomfilter(), nil
}
