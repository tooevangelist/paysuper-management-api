package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	userInviteMerchantPath       = "/user/invite/merchant"
	userInviteAdminPath          = "/user/invite/member"
	userResendInviteMerchantPath = "/user/resend/merchant"
	userResendInviteAdminPath    = "/user/resend/member"
	userAcceptInviteMerchantPath = "/user/accept/merchant"
	userAcceptInviteAdminPath    = "/user/accept/member"
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
	groups.AuthUser.GET(userInviteMerchantPath, h.inviteMerchant)
	groups.AuthUser.GET(userInviteAdminPath, h.inviteAdmin)
	groups.AuthUser.GET(userResendInviteMerchantPath, h.resendInviteMerchant)
	groups.AuthUser.GET(userResendInviteAdminPath, h.resendInviteAdmin)
	groups.AuthUser.GET(userAcceptInviteMerchantPath, h.acceptInviteMerchant)
	groups.AuthUser.GET(userAcceptInviteAdminPath, h.acceptInviteAdmin)
}

func (h *UserRoute) inviteMerchant(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, nil)
}

func (h *UserRoute) inviteAdmin(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, nil)
}

func (h *UserRoute) resendInviteMerchant(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, nil)
}

func (h *UserRoute) resendInviteAdmin(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, nil)
}

func (h *UserRoute) acceptInviteMerchant(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, nil)
}

func (h *UserRoute) acceptInviteAdmin(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, nil)
}
