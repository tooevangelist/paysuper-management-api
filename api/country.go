package api

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/manager"
	"net/http"
)

type CountryApiV1 struct {
	*Api
}

func (api *Api) InitCountryRoutes() *Api {
	cApiV1 := CountryApiV1{
		Api: api,
	}

	api.Http.GET("/api/v1/country", cApiV1.get)
	api.Http.GET("/api/v1/country/:code", cApiV1.getById)

	return api
}

// Get full list of currencies
// GET /api/v1/country
func (cApiV1 *CountryApiV1) get(ctx echo.Context) error {

	res, err := cApiV1.billingService.GetCountriesList(context.TODO(), &grpc.EmptyRequest{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Countries list error")
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get country by ISO 3166-1 alpha 2 country code
// GET /api/v1/country/{code}
func (cApiV1 *CountryApiV1) getById(ctx echo.Context) error {
	code := ctx.Param("code")

	if len(code) != 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "Incorrect country identifier")
	}

	req := &billing.GetCountryRequest{
		IsoCode: code,
	}
	err := cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	res, err := cApiV1.billingService.GetCountry(context.TODO(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Country not found")
	}

	return ctx.JSON(http.StatusOK, res)
}
