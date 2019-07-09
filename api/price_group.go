package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type PriceGroup struct {
	*Api
}

func (api *Api) initPriceGroupRoutes() *Api {
	InitPriceGroup := PriceGroup{
		Api: api,
	}

	api.accessRouteGroup.GET("/price_group/country", InitPriceGroup.getPriceGroupByCountry)
	api.accessRouteGroup.GET("/price_group/currencies", InitPriceGroup.getCurrencyList)
	api.accessRouteGroup.GET("/price_group/region", InitPriceGroup.getCurrencyByRegion)
	api.accessRouteGroup.GET("/price_group/recommended", InitPriceGroup.getRecommendedPrice)

	return api
}

// Get currency and region by country code
// GET /api/v1/price_group/country
func (api *PriceGroup) getPriceGroupByCountry(ctx echo.Context) error {
	req := &grpc.PriceGroupByCountryRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = api.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	res, err := api.billingService.GetPriceGroupByCountry(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessagePriceGroupByCountry)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get a list of currencies with a list of countries and regions for them
// GET /api/v1/price_group/currencies
func (api *PriceGroup) getCurrencyList(ctx echo.Context) error {
	req := &grpc.EmptyRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = api.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	res, err := api.billingService.GetPriceGroupCurrencies(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessagePriceGroupCurrencyList)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get currency for a region and a list of countries in it
// GET /api/v1/price_group/region
func (api *PriceGroup) getCurrencyByRegion(ctx echo.Context) error {
	req := &grpc.PriceGroupByRegionRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = api.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	res, err := api.billingService.GetPriceGroupCurrencyByRegion(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessagePriceGroupCurrencyByRegion)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get a list of recommended prices for all regions
// GET /api/v1/price_group/recommended
func (api *PriceGroup) getRecommendedPrice(ctx echo.Context) error {
	req := &grpc.PriceGroupRecommendedPriceRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = api.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	res, err := api.billingService.GetPriceGroupRecommendedPrice(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessagePriceGroupRecommendedList)
	}

	return ctx.JSON(http.StatusOK, res)
}
