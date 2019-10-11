package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

type AdminUsersRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

const (
	users = "/users"
)

func NewAdminUsersRoute(set common.HandlerSet, cfg *common.Config) *AdminUsersRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "AdminUsersRoute"})
	return &AdminUsersRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *AdminUsersRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(users, h.listUsers)
}

func (h *AdminUsersRoute) listUsers(ctx echo.Context) error {
	res, err := h.dispatch.Services.Billing.GetAdminUsers(ctx.Request().Context(), &grpc.EmptyRequest{})
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetAdminUsers", &grpc.EmptyRequest{})
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Users)
}