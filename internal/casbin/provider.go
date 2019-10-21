package casbin

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/invoker"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/google/wire"
	"github.com/paysuper/casbin-server/pkg/generated/api/proto/casbinpb"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/pkg/micro"
)

// ProviderCfg
func ProviderCfg(cfg config.Configurator) (*Config, func(), error) {
	c := &Config{
		WorkDir: cfg.WorkDir(),
		invoker: invoker.NewInvoker(),
	}
	e := cfg.UnmarshalKeyOnReload(common.UnmarshalKey, c)
	return c, func() {}, e
}

// ProviderCasbinService
func ProviderCasbinService(srv *micro.Micro) casbinpb.CasbinService {
	return casbinpb.NewCasbinService("", srv.Client())
}

// ProviderDispatcher
func ProviderCasbin(ctx context.Context, set provider.AwareSet, appSet AppSet, cfg *Config) (*Casbin, func(), error) {
	d := New(ctx, set, appSet, cfg)
	return d, func() {}, nil
}

var (
	WireSet = wire.NewSet(
		ProviderCasbin,
		ProviderCasbinService,
		ProviderCfg,
		wire.Struct(new(AppSet), "*"),
	)
	WireTestSet = wire.NewSet(
		ProviderCasbin,
		ProviderCasbinService,
		ProviderCfg,
		wire.Struct(new(AppSet), "*"),
	)
)
