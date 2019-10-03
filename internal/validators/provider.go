package validators

import (
	"github.com/google/wire"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
)

// Provider
func Provider(services common.Services) (*ValidatorSet, func(), error) {
	g := New(services)
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
