package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/manager"
	"net/http"
)

type CountryApiV1 struct {
	*Api
	countryManager *manager.CountryManager
}

func (api *Api) InitCountryRoutes() *Api {
	cApiV1 := CountryApiV1{
		Api:            api,
		countryManager: manager.InitCountryManager(api.database, api.logger),
	}

	api.Http.GET("/api/v1/country", cApiV1.get)
	api.Http.GET("/api/v1/country/:id", cApiV1.getById)

	return api
}

// @Summary Get list of countries
// @Description Get full list of currencies or get list of currencies filtered by name
// @Tags Country
// @Accept json
// @Produce json
// @Param name query string false "ISO 3166-1 alpha 2 code of country"
// @Success 200 {array} model.Country "OK"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/country [get]
func (cApiV1 *CountryApiV1) get(ctx echo.Context) error {
	code := ctx.QueryParam("code")

	if code != "" {
		return ctx.JSON(http.StatusOK, cApiV1.countryManager.FindByIsoCodeA2(code))
	}

	return ctx.JSON(http.StatusOK, cApiV1.countryManager.FindAll(cApiV1.limit, cApiV1.offset))
}

// @Summary Get country by numeric ISO 3166-1 code
// @Description Get country object by numeric ISO 3166-1 code
// @Tags Country
// @Accept json
// @Produce json
// @Param id path string true "ISO 3166-1 alpha 2 country code"
// @Success 200 {object} model.Country "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/country/{id} [get]
func (cApiV1 *CountryApiV1) getById(ctx echo.Context) error {
	codeIsoA2 := ctx.Param("id")

	if len(codeIsoA2) != 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "Incorrect currency identifier")
	}

	c := cApiV1.countryManager.FindByIsoCodeA2(codeIsoA2)

	if c == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Currency not found")
	}

	return ctx.JSON(http.StatusOK, c)
}
