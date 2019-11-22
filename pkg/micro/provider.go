package micro

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/invoker"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/google/wire"
)

// Cfg
func Cfg(cfg config.Configurator) (*Config, func(), error) {
	c := &Config{
		invoker: invoker.NewInvoker(),
	}
	e := cfg.UnmarshalKeyOnReload(UnmarshalKey, c)
	return c, func() {}, e
}

// CfgTest
func CfgTest() (*Config, func(), error) {
	c := &Config{
		invoker: invoker.NewInvoker(),
	}
	return c, func() {}, nil
}

// Provider
func Provider(ctx context.Context, set provider.AwareSet, cfg *Config) (*Micro, func(), error) {
	micro := New(ctx, set, cfg)
	micro.Init()
	return micro, func() {}, nil
}

// ProviderTest
func ProviderTest(ctx context.Context, set provider.AwareSet, cfg *Config) (*Micro, func(), error) {
	micro := New(ctx, set, cfg)
	return micro, func() {}, nil
}

var (
	WireSet = wire.NewSet(
		Provider,
		Cfg,
	)
	WireTestSet = wire.NewSet(
		ProviderTest,
		CfgTest,
	)
)
