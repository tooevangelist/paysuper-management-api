package api

import (
	"github.com/ProtocolONE/payments.api/handler"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

type CurrencyApiV1 struct {
	*Api
	handler *handler.CurrencyHandler
}

func (api *Api) InitCurrencyRoutes() *Api {
	cApiV1 := CurrencyApiV1{
		Api: api,
		handler: api.handlers[handler.DBCollectionCurrency].(*handler.CurrencyHandler),
	}

	api.Http.GET("/api/v1/currency", cApiV1.get)
	api.Http.GET("/api/v1/currency/:limit/:offset", cApiV1.getAll)

	return api
}

func (cApiV1 *CurrencyApiV1) getAll(c echo.Context) error {
	limit, err := strconv.Atoi(c.Param("limit"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Incorrect limit value")
	}

	offset, err := strconv.Atoi(c.Param("offset"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Incorrect offset value")
	}

	q := c.QueryParam("q")

	err, currencies := cApiV1.handler.GetAll(handler.Conditions{"q": q}, limit, offset)

	return c.JSON(http.StatusOK, currencies)
}

func (cApiV1 *CurrencyApiV1) get(c echo.Context) error {
	return c.JSON(http.StatusOK, "a")
}
