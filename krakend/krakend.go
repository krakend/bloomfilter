// Package krakend registers a bloomfilter given a config and registers the service with consul.
package krakend

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/devopsfaith/bloomfilter"
	bf_rpc "github.com/devopsfaith/bloomfilter/rpc"
	"github.com/devopsfaith/bloomfilter/rpc/server"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
)

// Namespace for bloomfilter
const Namespace = "github_com/devopsfaith/bloomfilter"

var (
	errNoConfig    = errors.New("no config for the bloomfilter")
	errWrongConfig = errors.New("invalid config for the bloomfilter")
)

// Config defines the configuration to be added to the KrakenD gateway
type Config struct {
	bf_rpc.Config
	TokenKeys []string
	Headers   []string
}

// Register registers a bloomfilter given a config and registers the service with consul
func Register(ctx context.Context, serviceName string, cfg config.ServiceConfig,
	logger logging.Logger, register func(n string, p int)) (Rejecter, error) {
	data, ok := cfg.ExtraConfig[Namespace]
	if !ok {
		logger.Debug(errNoConfig.Error())
		return nopRejecter, errNoConfig
	}

	raw, err := json.Marshal(data)
	if err != nil {
		logger.Debug(errWrongConfig.Error())
		return nopRejecter, errWrongConfig
	}

	var rpcConfig Config
	if err := json.Unmarshal(raw, &rpcConfig); err != nil {
		logger.Debug(err.Error(), string(raw))
		return nopRejecter, err
	}

	bf := server.New(ctx, rpcConfig.Config)
	register(serviceName, rpcConfig.Port)

	return Rejecter{
		BF:        bf.Bloomfilter(),
		TokenKeys: rpcConfig.TokenKeys,
		Headers:   rpcConfig.Headers,
	}, nil
}

type Rejecter struct {
	BF        bloomfilter.Bloomfilter
	TokenKeys []string
	Headers   []string
}

func (r *Rejecter) RejectToken(claims map[string]interface{}) bool {
	for _, k := range r.TokenKeys {
		v, ok := claims[k]
		if !ok {
			continue
		}
		data, ok := v.(string)
		if !ok {
			continue
		}
		if r.BF.Check([]byte(k + "-" + data)) {
			return true
		}
	}
	return false
}

func (r *Rejecter) RejectHeader(header http.Header) bool {
	for _, k := range r.Headers {
		data := header.Get(k)
		if data == "" {
			continue
		}
		if r.BF.Check([]byte(k + "-" + data)) {
			return true
		}
	}
	return false
}

var nopRejecter = Rejecter{BF: new(bloomfilter.EmptySet)}
