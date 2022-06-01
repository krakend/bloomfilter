package krakend

import (
	"context"
	"testing"

	"github.com/krakendio/bloomfilter/v2"
	"github.com/krakendio/bloomfilter/v2/rotate"
	"github.com/krakendio/bloomfilter/v2/rpc"
	gologging "github.com/krakendio/krakend-gologging/v2"
	"github.com/luraproject/lura/v2/config"
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
			"github_com/devopsfaith/bloomfilter": cfgBloomFilter,
		},
	}

	logger, err := gologging.NewLogger(config.ExtraConfig{
		gologging.Namespace: map[string]interface{}{
			"level":  "DEBUG",
			"stdout": true,
		},
	})
	if err != nil {
		t.Error(err.Error())
		return
	}

	registered := false

	if _, err := Register(ctx, "bloomfilter-test", serviceConf, logger, func(name string, port int) {
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
	logger, err := gologging.NewLogger(config.ExtraConfig{
		gologging.Namespace: map[string]interface{}{
			"level":  "DEBUG",
			"stdout": true,
		},
	})
	if err != nil {
		t.Error(err.Error())
		return
	}

	if _, err := Register(ctx, "bloomfilter-test", serviceConf, logger, func(name string, port int) {
		t.Error("this error should never been called")
	}); err != ErrNoConfig {
		t.Errorf("didn't get error %s", ErrNoConfig)
	}

}
