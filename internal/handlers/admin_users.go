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
	users             = "/users"
	adminListRoles    = "/users/roles"
	adminUserInvite   = "/users/invite"
	adminResendInvite = "/users/resend"
	adminUserRole     = "/users/roles/:role_id"
	adminUserDelete   = "/users/roles/:role_id"
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
	groups.SystemUser.GET(users, h.listUsers)
	groups.SystemUser.PUT(adminUserRole, h.changeRole)
	groups.SystemUser.POST(adminUserInvite, h.sendInvite)
	groups.SystemUser.POST(adminResendInvite, h.resendInvite)
	groups.SystemUser.GET(adminListRoles, h.listRoles)
	groups.SystemUser.DELETE(adminUserDelete, h.deleteUser)
}

func (h *AdminUsersRoute) changeRole(ctx echo.Context) error {
	req := &grpc.ChangeRoleForAdminUserRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.ChangeRoleForAdminUser(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
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

func (h *AdminUsersRoute) sendInvite(ctx echo.Context) error {
	req := &grpc.InviteUserAdminRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.InviteUserAdmin(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "InviteUserAdmin", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUnableToSendInvite)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *AdminUsersRoute) resendInvite(ctx echo.Context) error {
	req := &grpc.ResendInviteAdminRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.ResendInviteAdmin(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ResendInviteAdmin", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUnableToSendInvite)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *AdminUsersRoute) listRoles(ctx echo.Context) error {
	req := &grpc.GetRoleListRequest{Type: pkg.RoleTypeSystem}
	res, err := h.dispatch.Services.Billing.GetRoleList(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetRoleList", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageInvalidRoleType)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *AdminUsersRoute) deleteUser(ctx echo.Context) error {
	req := &grpc.DeleteAdminUserRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.DeleteAdminUser(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "DeleteAdminUser", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUnableToDeleteUser)
	}

	return ctx.JSON(http.StatusOK, res)
}
