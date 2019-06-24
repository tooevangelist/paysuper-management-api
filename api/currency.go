package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/manager"
	"net/http"
	"strconv"
)

type CurrencyApiV1 struct {
	*Api
	currencyManager *manager.CurrencyManager
}

func (api *Api) InitCurrencyRoutes() *Api {
	cApiV1 := CurrencyApiV1{
		Api:             api,
		currencyManager: manager.InitCurrencyManager(api.database, api.logger),
	}

	api.Http.GET("/api/v1/currency", cApiV1.get)
	api.Http.GET("/api/v1/currency/:id", cApiV1.getById)

	return api
}

// @Summary Get list of currencies
// @Description Get full list of currencies or get list of currencies filtered by name
// @Tags Currency
// @Accept json
// @Produce json
// @Param name query string false "name or symbolic ISO 4217 code of currency"
// @Success 200 {array} model.Currency "OK"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/currency [get]
func (cApiV1 *CurrencyApiV1) get(ctx echo.Context) error {
	name := ctx.QueryParam("name")

	if name != "" {
		return ctx.JSON(http.StatusOK, cApiV1.currencyManager.FindByName(name))
	}

	return ctx.JSON(http.StatusOK, cApiV1.currencyManager.FindAll(cApiV1.limit, cApiV1.offset))
}

// @Summary Get currency by numeric ISO 4217 code
// @Description Get currency object by numeric ISO 4217 code
// @Tags Currency
// @Accept json
// @Produce json
// @Param id path int true "numeric ISO 4217 currency code"
// @Success 200 {object} model.Currency "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/currency/{id} [get]
func (cApiV1 *CurrencyApiV1) getById(ctx echo.Context) error {
	codeInt, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectCurrencyIdentifier)
	}

	c := cApiV1.currencyManager.FindByCodeInt(codeInt)

	if c == nil {
		return echo.NewHTTPError(http.StatusNotFound, errorCurrencyNotFound)
	}

	return ctx.JSON(http.StatusOK, c)
}
