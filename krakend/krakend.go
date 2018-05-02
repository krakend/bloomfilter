// Package krakend registers a bloomfilter given a config and registers the service with consul.
package krakend

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/bloomfilter"
	bf_rpc "github.com/devopsfaith/bloomfilter/rpc"
	"github.com/devopsfaith/bloomfilter/rpc/server"
)

// Namespace for bloomfilter
const Namespace = "github_com/devopsfaith/bloomfilter"

var (
	errNoConfig    = errors.New("no config for the bloomfilter")
	errWrongConfig = errors.New("invalid config for the bloomfilter")
)

// Register registers a bloomfilter given a config and registers the service with consul
func Register(
	ctx context.Context, serviceName string, cfg config.ServiceConfig, logger logging.Logger, register func(n string, p int)) (bloomfilter.Bloomfilter, error) {
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
	register(serviceName, rpcConfig.Port)

	return bf.Bloomfilter(), nil
}
