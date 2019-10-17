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

type PermissionsRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

const permissionsRoute = "/permissions"

func NewPermissionsRoute(set common.HandlerSet, cfg *common.Config) *PermissionsRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PermissionsRoute"})
	return &PermissionsRoute{
		dispatch: set,
		cfg:      *cfg,
		LMT:      &set.AwareSet,
	}
}

func (h *PermissionsRoute) Route(groups *common.Groups) {
	groups.AuthProject.GET(permissionsRoute, h.getPermissions)
}

func (h *PermissionsRoute) getPermissions(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)

	res, err := h.dispatch.Services.Billing.GetPermissionsForUser(ctx.Request().Context(), &grpc.GetPermissionsForUserRequest{
		UserId: authUser.Id,
		MerchantId: authUser.MerchantId,
	})

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Permissions)
}



