package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	paymentMinLimitSystemPath = "/payment_min_limit_system"
)

type PaymentMinLimitSystemRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPaymentMinLimitSystemRoute(set common.HandlerSet, cfg *common.Config) *PaymentMinLimitSystemRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PaymentMinLimitSystemRoute"})
	return &PaymentMinLimitSystemRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *PaymentMinLimitSystemRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(paymentMinLimitSystemPath, h.getPaymentMinLimitSystemList)
	groups.AuthUser.POST(paymentMinLimitSystemPath, h.setPaymentMinLimitSystem)
}

func (h *PaymentMinLimitSystemRoute) getPaymentMinLimitSystemList(ctx echo.Context) error {
	req := &grpc.EmptyRequest{}

	res, err := h.dispatch.Services.Billing.GetOperatingCompaniesList(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetOperatingCompaniesList", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Items)
}

func (h *PaymentMinLimitSystemRoute) setPaymentMinLimitSystem(ctx echo.Context) error {
	req := &billing.PaymentMinLimitSystem{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.SetPaymentMinLimitSystem(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "AddPaymentMinLimitSystem", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}
