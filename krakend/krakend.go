// Package krakend registers a bloomfilter given a config and registers the service with consul.
package krakend

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/krakendio/bloomfilter/v2"
	bf_rpc "github.com/krakendio/bloomfilter/v2/rpc"
	"github.com/krakendio/bloomfilter/v2/rpc/server"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

// Namespace for bloomfilter
const Namespace = "github_com/devopsfaith/bloomfilter"

var (
	ErrNoConfig    = errors.New("no config for the bloomfilter")
	errWrongConfig = errors.New("invalid config for the bloomfilter")
)

// Config defines the configuration to be added to the KrakenD gateway
type Config struct {
	bf_rpc.Config
	TokenKeys []string `json:"token_keys"`
	Headers   []string `json:"headers"`
}

// Registers a bloomfilter given a config
func Register(ctx context.Context, serviceName string, cfg config.ServiceConfig,
	logger logging.Logger, register func(n string, p int)) (Rejecter, error) {
	data, ok := cfg.ExtraConfig[Namespace]
	logPrefix := "[SERVICE: Bloomfilter]"
	if !ok {
		return nopRejecter, ErrNoConfig
	}

	raw, err := json.Marshal(data)
	if err != nil {
		logger.Error(logPrefix, "Unable to read the bloomfilter configuration:", errWrongConfig.Error())
		return nopRejecter, errWrongConfig
	}

	var rpcConfig Config
	if err := json.Unmarshal(raw, &rpcConfig); err != nil {
		logger.Error(logPrefix, "Unable to parse the bloomfilter configuration:", err.Error(), string(raw))
		return nopRejecter, err
	}

	bf := server.New(ctx, rpcConfig.Config)
	register(serviceName, rpcConfig.Port)

	logger.Debug(logPrefix, "Service registered successfully")

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
