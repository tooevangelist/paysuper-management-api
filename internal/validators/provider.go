package validators

import (
	"github.com/ProtocolONE/go-core/provider"
	"github.com/google/wire"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
)

// Provider
func Provider(services common.Services, lmt provider.LMT) (*ValidatorSet, func(), error) {
	g := New(services, lmt)
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
