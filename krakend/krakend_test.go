package krakend

import (
	"context"
	"fmt"
	"testing"

	"github.com/devopsfaith/krakend-gologging"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/bloomfilter"
	"github.com/devopsfaith/bloomfilter/rotate"
	"github.com/devopsfaith/bloomfilter/rpc"
	"github.com/letgoapp/krakend-consul"
)

func TestRegister_ok(t *testing.T) {
	ctx := context.Background()
	cfgBloomFilter :=
		rpc.Config{
			Config: rotate.Config{
				Config: bloomfilter.Config{
					N:        10000000,
					P:        0.0000001,
					HashName: "optimal",
				},
				TTL: 1500,
			},
			Port: 1234,
		}
	cfgConsul := consul.Config{
		Address: "127.0.0.1:8500",
		Tags: []string{
			"1224", "2233",
		},
		Port: 1234,
		Name: "bloomfilter-test",
	}

	consulExtraConfig := config.ExtraConfig{
		"github_com/letgoapp/krakend-consul": cfgConsul,
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

	if _, err := Register(ctx, "bloomfilter-test", serviceConf, logger, func(name string, port int) {
		if err := consul.Register(ctx, consulExtraConfig, port, name, logger); err != nil {
			logger.Error(fmt.Sprintf("Couldn't register %s:%d in consul: %s", name, port, err.Error()))
		}
	}); err != nil {
		t.Errorf("got error when registering: %s", err.Error())
	}
}

func TestRegister_koNamespace(t *testing.T) {
	ctx := context.Background()
	cfgBloomFilter :=
		rpc.Config{
			Config: rotate.Config{
				Config: bloomfilter.Config{
					N:        10000000,
					P:        0.0000001,
					HashName: "optimal",
				},
				TTL: 1500,
			},
			Port: 1234,
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
		if err := consul.Register(ctx, map[string]interface{}{}, port, name, logger); err != nil {
			logger.Error(fmt.Sprintf("Couldn't register %s:%d in consul: %s", name, port, err.Error()))
		}
	}); err != errNoConfig {
		t.Errorf("didn't get error %s", errNoConfig)
	}
}
