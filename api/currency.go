package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
	"strconv"
)

type CurrencyApiV1 struct {
	*Api
}

func (api *Api) InitCurrencyRoutes() *Api {
	cApiV1 := CurrencyApiV1{
		Api: api,
	}

	api.Http.GET("/api/v1/currency", cApiV1.get)
	api.Http.GET("/api/v1/currency/name", cApiV1.getByName)
	api.Http.GET("/api/v1/currency/:id", cApiV1.getById)

	return api
}

// get list of currencies
// GET /api/v1/currency
func (cApiV1 *CurrencyApiV1) get(ctx echo.Context) error {
	req := &grpc.GetCurrencyListRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	if req.Limit <= 0 {
		req.Limit = LimitDefault
	}

	if req.Offset <= 0 {
		req.Offset = OffsetDefault
	}

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, getFirstValidationError(err))
	}

	res, err := cApiV1.billingService.GetCurrencyList(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Currency list error")
	}

	return ctx.JSON(http.StatusOK, res)
}

// getByName return currency by alpha code
// GET /api/v1/currency/name
func (cApiV1 *CurrencyApiV1) getByName(ctx echo.Context) error {
	req := &billing.GetCurrencyRequest{
		A3: ctx.QueryParam("name"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, getFirstValidationError(err))
	}

	res, err := cApiV1.billingService.GetCurrency(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Currency not found")
	}

	return ctx.JSON(http.StatusOK, res)
}

// getById return currency by numeric ISO 4217 code
// GET /api/v1/currency/{id}
func (cApiV1 *CurrencyApiV1) getById(ctx echo.Context) error {
	i, err := strconv.ParseInt(ctx.QueryParam("id"), 10, 32)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	req := &billing.GetCurrencyRequest{Int: int32(i)}
	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, getFirstValidationError(err))
	}

	res, err := cApiV1.billingService.GetCurrency(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Currency not found")
	}

	return ctx.JSON(http.StatusOK, res)
}
