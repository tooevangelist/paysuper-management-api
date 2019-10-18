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

const (
	merchantListRoles    = "/merchants/roles"
	merchantUsers        = "/merchants/:merchant_id/users"
	merchantInvite       = "/merchants/:merchant_id/invite"
	merchantInviteResend = "/merchants/:merchant_id/users/:role_id/resend"
	merchantDeleteUser   = "/merchants/:merchant_id/users/:role_id/role"
	merchantUsersRole    = "/merchants/:merchant_id/users/:role_id/role"
)

type MerchantUsersRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewMerchantUsersRoute(set common.HandlerSet, cfg *common.Config) *MerchantUsersRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "MerchantUsersRoute"})
	return &MerchantUsersRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *MerchantUsersRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(merchantUsers, h.getMerchantUsers)
	groups.AuthUser.PUT(merchantUsersRole, h.changeRole)
	groups.AuthUser.POST(merchantInvite, h.sendInvite)
	groups.AuthUser.POST(merchantInviteResend, h.resendInvite)
	groups.AuthUser.GET(merchantListRoles, h.listRoles)
	groups.AuthUser.DELETE(merchantDeleteUser, h.deleteUser)
}

func (h *MerchantUsersRoute) changeRole(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	roleId := ctx.Param(common.RequestParameterRoleId)

	req := &grpc.ChangeRoleForMerchantUserRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.PerformerId = authUser.Id
	req.RoleId = roleId

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ChangeRoleForMerchantUser(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}

func (h *MerchantUsersRoute) getMerchantUsers(ctx echo.Context) error {
	req := &grpc.GetMerchantUsersRequest{}
	err := h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetMerchantUsers(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantUsers", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Users)
}

func (h *MerchantUsersRoute) sendInvite(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	merchantId := ctx.Param(common.RequestParameterMerchantId)

	req := &grpc.InviteUserMerchantRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.PerformerId = authUser.Id
	req.MerchantId = merchantId

	err := h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.InviteUserMerchant(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "InviteUserMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUnableToSendInvite)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *MerchantUsersRoute) resendInvite(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	roleId := ctx.Param(common.RequestParameterRoleId)

	req := &grpc.ResendInviteMerchantRequest{
		PerformerId: authUser.Id,
		RoleId:      roleId,
	}

	err := h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ResendInviteMerchant(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ResendInviteMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUnableToSendInvite)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *MerchantUsersRoute) listRoles(ctx echo.Context) error {
	req := &grpc.GetRoleListRequest{Type: pkg.RoleTypeMerchant}
	res, err := h.dispatch.Services.Billing.GetRoleList(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetRoleList", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageInvalidRoleType)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *MerchantUsersRoute) deleteUser(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	roleId := ctx.Param(common.RequestParameterRoleId)

	req := &grpc.DeleteMerchantUserRequest{
		PerformerId: authUser.Id,
		RoleId:      roleId,
	}

	err := h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeleteMerchantUser(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "DeleteMerchantUser", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUnableToDeleteUser)
	}

	return ctx.JSON(http.StatusOK, res)
}
