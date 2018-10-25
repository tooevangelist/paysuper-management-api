package api

import (
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
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

func (cApiV1 *CountryApiV1) get(ctx echo.Context) error {
	name := ctx.QueryParam("name")

	if name != "" {
		return ctx.JSON(http.StatusOK, cApiV1.countryManager.FindByName(name))
	}

	return ctx.JSON(http.StatusOK, cApiV1.countryManager.FindAll(cApiV1.limit, cApiV1.offset))
}

func (cApiV1 *CountryApiV1) getById(ctx echo.Context) error {
	codeInt, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Incorrect currency identifier")
	}

	c := cApiV1.countryManager.FindByCodeInt(codeInt)

	if c == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Currency not found")
	}

	return ctx.JSON(http.StatusOK, c)
}
