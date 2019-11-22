// +build wireinject

package casbin

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/invoker"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/google/wire"
	"github.com/paysuper/paysuper-management-api/pkg/micro"
)

// Build
func Build(ctx context.Context, initial config.Initial, observer invoker.Observer) (*Casbin, func(), error) {
	panic(
		wire.Build(
			provider.Set,
			wire.Struct(new(provider.AwareSet), "*"),
			micro.WireSet,
			WireSet,
		),
	)
}
