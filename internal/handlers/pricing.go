package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	pricingRecommendedConversionPath = "/pricing/recommended/conversion"
	pricingRecommendedSteamPath      = "/pricing/recommended/steam"
	pricingRecommendedTablePath      = "/pricing/recommended/table"
)

type Pricing struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPricingRoute(set common.HandlerSet, cfg *common.Config) *Pricing {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PriceGroup"})
	return &Pricing{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *Pricing) Route(groups *common.Groups) {
	groups.Common.GET(pricingRecommendedConversionPath, h.getRecommendedByConversion)
	groups.Common.GET(pricingRecommendedSteamPath, h.getRecommendedBySteam)
	groups.Common.GET(pricingRecommendedTablePath, h.getRecommendedTable)
}

func (h *Pricing) getRecommendedByConversion(ctx echo.Context) error {
	req := &grpc.RecommendedPriceRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetRecommendedPriceByConversion(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessagePriceGroupRecommendedList)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *Pricing) getRecommendedBySteam(ctx echo.Context) error {
	req := &grpc.RecommendedPriceRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetRecommendedPriceByPriceGroup(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessagePriceGroupRecommendedList)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *Pricing) getRecommendedTable(ctx echo.Context) error {
	req := &grpc.RecommendedPriceTableRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetRecommendedPriceTable(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessagePriceGroupRecommendedList)
	}

	return ctx.JSON(http.StatusOK, res)
}
