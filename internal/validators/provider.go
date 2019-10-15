package validators

import (
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/google/wire"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
)

// Provider
func Provider(services common.Services, set provider.AwareSet) (*ValidatorSet, func(), error) {
	g := New(services, set)
	return g, func() {}, nil
}

var (
	WireSet = wire.NewSet(
		Provider,
	)
	WireTestSet = wire.NewSet(
		Provider,
	)
)
