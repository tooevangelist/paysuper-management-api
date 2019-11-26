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
	priceGroupCountryPath    = "/price_group/country"
	priceGroupCurrenciesPath = "/price_group/currencies"
	priceGroupRegionPath     = "/price_group/region"
)

type PriceGroup struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPriceGroupRoute(set common.HandlerSet, cfg *common.Config) *PriceGroup {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PriceGroup"})
	return &PriceGroup{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *PriceGroup) Route(groups *common.Groups) {
	groups.Common.GET(priceGroupCountryPath, h.getPriceGroupByCountry)
	groups.Common.GET(priceGroupCurrenciesPath, h.getCurrencyList)
	groups.Common.GET(priceGroupRegionPath, h.getCurrencyByRegion)
}

func (h *PriceGroup) getPriceGroupByCountry(ctx echo.Context) error {
	req := &grpc.PriceGroupByCountryRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPriceGroupByCountry(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessagePriceGroupByCountry)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *PriceGroup) getCurrencyList(ctx echo.Context) error {
	req := &grpc.EmptyRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPriceGroupCurrencies(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessagePriceGroupCurrencyList)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *PriceGroup) getCurrencyByRegion(ctx echo.Context) error {
	req := &grpc.PriceGroupByRegionRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPriceGroupCurrencyByRegion(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessagePriceGroupCurrencyByRegion)
	}

	return ctx.JSON(http.StatusOK, res)
}
