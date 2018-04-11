package krakend

import (
	"context"
	"testing"

	gologging "github.com/devopsfaith/krakend-gologging"
	"github.com/devopsfaith/krakend/config"
	bloomfilter "github.com/letgoapp/go-bloomfilter"
	"github.com/letgoapp/go-bloomfilter/rotate"
	"github.com/letgoapp/go-bloomfilter/rpc"
)

func TestRegister_ok(t *testing.T) {
	ctx := context.Background()
	cfg :=
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
			"github_com/letgoapp/go-bloomfilter": cfg,
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

	if _, err := Register(ctx, serviceConf, logger); err != nil {
		t.Errorf("got error when registering: %s", err.Error())
	}

}

func TestRegister_koNamespace(t *testing.T) {
	ctx := context.Background()
	cfg :=
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
			"wrongnamespace": cfg,
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

	if _, err := Register(ctx, serviceConf, logger); err != errNoConfig {
		t.Errorf("didn't get error %s", errNoConfig)
	}

}
