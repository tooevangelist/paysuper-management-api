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
	merchantUsers        = "/merchants/users"
	merchantInvite       = "/merchants/invite"
	merchantInviteResend = "/merchants/users/resend"
	merchantUsersRole    = "/merchants/users/roles/:role_id"
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
	groups.AuthUser.DELETE(merchantUsersRole, h.deleteUser)
	groups.AuthUser.GET(merchantUsersRole, h.getUser)
}

func (h *MerchantUsersRoute) changeRole(ctx echo.Context) error {
	req := &grpc.ChangeRoleForMerchantUserRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.ChangeRoleForMerchantUser(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ChangeRoleForMerchantUser")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}

func (h *MerchantUsersRoute) getMerchantUsers(ctx echo.Context) error {
	req := &grpc.GetMerchantUsersRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetMerchantUsers(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetMerchantUsers")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Users)
}

func (h *MerchantUsersRoute) sendInvite(ctx echo.Context) error {
	req := &grpc.InviteUserMerchantRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.InviteUserMerchant(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "InviteUserMerchant")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *MerchantUsersRoute) resendInvite(ctx echo.Context) error {
	req := &grpc.ResendInviteMerchantRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.ResendInviteMerchant(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ResendInviteMerchant")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *MerchantUsersRoute) listRoles(ctx echo.Context) error {
	req := &grpc.GetRoleListRequest{Type: pkg.RoleTypeMerchant}
	res, err := h.dispatch.Services.Billing.GetRoleList(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetRoleList")
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *MerchantUsersRoute) deleteUser(ctx echo.Context) error {
	req := &grpc.MerchantRoleRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.DeleteMerchantUser(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "DeleteMerchantUser")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *MerchantUsersRoute) getUser(ctx echo.Context) error {
	req := &grpc.MerchantRoleRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetMerchantUserRole(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetMerchantUserRole")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}
