package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-currencies/pkg/proto/currencies"
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

// Get test settings for payment method
// GET /api/v1/price_group/country
func (api *PriceGroup) getPriceGroupByCountry(ctx echo.Context) error {
	req := &currencies.CountryRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = api.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	res, err := api.currencyService.GetPriceGroupByCountry(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get test settings for payment method
// GET /api/v1/payment_method/currencies
func (api *PriceGroup) getCurrencyList(ctx echo.Context) error {
	req := &currencies.EmptyRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = api.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	res, err := api.currencyService.GetCurrencyList(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get test settings for payment method
// GET /api/v1/payment_method/region
func (api *PriceGroup) getCurrencyByRegion(ctx echo.Context) error {
	req := &currencies.RegionRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = api.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	res, err := api.currencyService.GetCurrencyByRegion(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get test settings for payment method
// GET /api/v1/payment_method/recommended
func (api *PriceGroup) getRecommendedPrice(ctx echo.Context) error {
	req := &currencies.RecommendedPriceRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = api.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	res, err := api.currencyService.GetRecommendedPrice(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}
