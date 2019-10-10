package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	userEmptyPath = "/user"
)

type UserRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewUserRoute(set common.HandlerSet, cfg *common.Config) *UserRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "UserRoute"})
	return &UserRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *UserRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(userEmptyPath, h.temp)
}

func (h *UserRoute) temp(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, nil)
}
