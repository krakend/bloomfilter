package krakend

import (
	"context"
	"testing"

	"github.com/krakend/bloomfilter/v3"
	"github.com/krakend/bloomfilter/v3/rotate"
	"github.com/krakend/bloomfilter/v3/rpc"
	"github.com/luraproject/lura/v3/config"
	"github.com/luraproject/lura/v3/logging"
)

func TestRegister_ok(t *testing.T) {
	ctx := context.Background()
	cfgBloomFilter := Config{
		Config: rpc.Config{
			Config: rotate.Config{
				Config: bloomfilter.Config{
					N:        10000000,
					P:        0.0000001,
					HashName: "optimal",
				},
				TTL: 1500,
			},
			Port: 1234,
		},
	}

	serviceConf := config.ServiceConfig{
		ExtraConfig: map[string]interface{}{
			"auth/revoker": cfgBloomFilter,
		},
	}

	registered := false

	if _, err := Register(ctx, "bloomfilter-test", serviceConf, logging.NoOp, func(_ string, _ int) {
		registered = true
	}); err != nil {
		t.Errorf("got error when registering: %s", err.Error())
	}

	if !registered {
		t.Error("register function not called")
	}
}

func TestRegister_koNamespace(t *testing.T) {
	ctx := context.Background()
	cfgBloomFilter := Config{
		Config: rpc.Config{
			Config: rotate.Config{
				Config: bloomfilter.Config{
					N:        10000000,
					P:        0.0000001,
					HashName: "optimal",
				},
				TTL: 1500,
			},
			Port: 1234,
		},
	}
	serviceConf := config.ServiceConfig{
		ExtraConfig: config.ExtraConfig{
			"wrongnamespace": cfgBloomFilter,
		},
	}

	if _, err := Register(ctx, "bloomfilter-test", serviceConf, logging.NoOp, func(_ string, _ int) {
		t.Error("this error should never been called")
	}); err != ErrNoConfig {
		t.Errorf("didn't get error %s", ErrNoConfig)
	}
}
