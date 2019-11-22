package casbin

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/invoker"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/paysuper/casbin-server/pkg/generated/api/proto/casbinpb"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Casbin struct {
	ctx    context.Context
	cfg    Config
	appSet AppSet
	provider.LMT
}

// ImportPolicy
func (c *Casbin) ImportPolicy(path string) error {
	b, e := ioutil.ReadFile(path)
	if e != nil {
		return errors.WithMessage(e, Prefix)
	}
	in := &casbinpb.ImportPolicyRequest{Data: b}
	_, e = c.appSet.CasbinService.ImportPolicy(c.ctx, in)
	if e != nil {
		return e
	}
	_, e = c.appSet.CasbinService.SavePolicy(c.ctx, &casbinpb.Empty{})
	return e
}

// Config
type Config struct {
	Debug   bool `fallback:"shared.debug"`
	WorkDir string

	invoker *invoker.Invoker
}

// OnReload
func (c *Config) OnReload(callback func(ctx context.Context)) {
	c.invoker.OnReload(callback)
}

// Reload
func (c *Config) Reload(ctx context.Context) {
	c.invoker.Reload(ctx)
}

// AppSet
type AppSet struct {
	CasbinService casbinpb.CasbinService
}

// New
func New(ctx context.Context, set provider.AwareSet, appSet AppSet, cfg *Config) *Casbin {
	set.Logger = set.Logger.WithFields(logger.Fields{"service": common.Prefix})
	return &Casbin{
		ctx:    ctx,
		cfg:    *cfg,
		appSet: appSet,
		LMT:    &set,
	}
}
