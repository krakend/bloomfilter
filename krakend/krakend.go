package krakend

import (
	"context"
	"encoding/json"
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
	raw, err := json.Marshal(data)
	if err != nil {
		logger.Info(errWrongConfig.Error(), cfg.ExtraConfig)
		return new(bloomfilter.EmptySet), errWrongConfig
	}

	var rpcConfig bf_rpc.Config
	if err := json.Unmarshal(raw, &rpcConfig); err != nil {
		logger.Info(err.Error(), string(raw))
		return new(bloomfilter.EmptySet), err
	}

	bf := server.New(ctx, rpcConfig)
	return bf.Bloomfilter(), nil
}
