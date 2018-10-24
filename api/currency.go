package api

import (
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

type CurrencyApiV1 struct {
	*Api
	currencyManager *manager.CurrencyManager
}

func (api *Api) InitCurrencyRoutes() *Api {
	cApiV1 := CurrencyApiV1{
		Api: api,
		currencyManager: manager.InitCurrencyManager(api.database, api.logger),
	}

	api.Http.GET("/api/v1/currency", cApiV1.get)
	api.Http.GET("/api/v1/currency/:id", cApiV1.getById)

	return api
}

func (cApiV1 *CurrencyApiV1) get(ctx echo.Context) error {
	name := ctx.QueryParam("name")

	if name != "" {
		return ctx.JSON(http.StatusOK, cApiV1.currencyManager.FindByName(name))
	}

	return ctx.JSON(http.StatusOK, cApiV1.currencyManager.FindAll(cApiV1.limit, cApiV1.offset))
}

func (cApiV1 *CurrencyApiV1) getById(ctx echo.Context) error {
	codeInt, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Incorrect currency identifier")
	}

	c := cApiV1.currencyManager.FindByCodeInt(codeInt)

	if c == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Currency not found")
	}

	return ctx.JSON(http.StatusOK, c)
}
