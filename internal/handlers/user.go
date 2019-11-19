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
	inviteCheck   = "/user/invite/check"
	inviteApprove = "/user/invite/approve"
	getMerchants  = "/user/merchants"
)

func NewUserRoute(set common.HandlerSet, cfg *common.Config) *UserRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "UsersRoute"})
	return &UserRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *UserRoute) Route(groups *common.Groups) {
	groups.AuthProject.POST(inviteCheck, h.checkInvite)
	groups.AuthProject.POST(inviteApprove, h.approveInvite)
	groups.AuthProject.GET(getMerchants, h.getMerchants)

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

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
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

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
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

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}
