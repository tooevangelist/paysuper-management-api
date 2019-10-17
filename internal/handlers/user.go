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

type UserRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

const (
	inviteCheck      = "/user/invite/check"
	inviteApprove    = "/user/invite/approve"
	getMerchants     = "/user/merchants"
	permissionsRoute = "/permissions"
)

func NewUserRoute(set common.HandlerSet, cfg *common.Config) *UserRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "AdminUsersRoute"})
	return &UserRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *UserRoute) Route(groups *common.Groups) {
	groups.AuthUser.POST(inviteCheck, h.checkInvite)
	groups.AuthUser.POST(inviteApprove, h.approveInvite)
	groups.AuthUser.POST(getMerchants, h.getMerchants)
	groups.AuthProject.GET(permissionsRoute, h.getPermissions)

}

func (h *UserRoute) checkInvite(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)

	req := &grpc.CheckInviteTokenRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.Email = authUser.Email

	err := h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CheckInviteToken(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "CheckInviteToken", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUnableToCheckInviteToken)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *UserRoute) approveInvite(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)

	req := &grpc.AcceptInviteRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.UserId = authUser.Id
	req.Email = authUser.Email

	err := h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.AcceptInvite(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "AcceptInvite", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUnableToAcceptInvite)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *UserRoute) getMerchants(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)

	req := &grpc.GetMerchantsForUserRequest{UserId: authUser.Id}

	res, err := h.dispatch.Services.Billing.GetMerchantsForUser(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantsForUser", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *UserRoute) getPermissions(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)

	res, err := h.dispatch.Services.Billing.GetPermissionsForUser(ctx.Request().Context(), &grpc.GetPermissionsForUserRequest{
		UserId:     authUser.Id,
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
