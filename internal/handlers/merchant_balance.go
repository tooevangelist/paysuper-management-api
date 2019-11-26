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
	balancePath         = "/balance"
	balanceMerchantPath = "/balance/:merchant_id"
)

type BalanceRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewBalanceRoute(set common.HandlerSet, cfg *common.Config) *BalanceRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "BalanceRoute"})
	return &BalanceRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *BalanceRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(balancePath, h.getBalance)
	groups.SystemUser.GET(balanceMerchantPath, h.getBalance)
}

func (h *BalanceRoute) getBalance(ctx echo.Context) error {
	req := &grpc.GetMerchantBalanceRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	res, err := h.dispatch.Services.Billing.GetMerchantBalance(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBalance", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}
