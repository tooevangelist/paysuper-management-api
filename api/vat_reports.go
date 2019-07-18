package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type vatReportsRoute struct {
	*Api
}

func (api *Api) initVatReportsRoutes() *Api {
	cApiV1 := vatReportsRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/vat_reports", cApiV1.getVatReportsDashboard)
	api.authUserRouteGroup.GET("/vat_reports/country/:country", cApiV1.getVatReportsForCountry)
	api.authUserRouteGroup.GET("/vat_reports/details/:id", cApiV1.getVatReportTransactions)
	api.authUserRouteGroup.POST("/vat_reports/status/:id", cApiV1.updateVatReportStatus)

	return api
}

// Get vat reports dashboard
// GET /admin/api/v1/vat_reports
func (cApiV1 *vatReportsRoute) getVatReportsDashboard(ctx echo.Context) error {

	res, err := cApiV1.billingService.GetVatReportsDashboard(ctx.Request().Context(), &grpc.EmptyRequest{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}

// Get vat reports for country
// GET /admin/api/v1/vat_reports/country/ru
func (cApiV1 *vatReportsRoute) getVatReportsForCountry(ctx echo.Context) error {
	req := &grpc.VatReportsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	req.Country = ctx.Param(requestParameterCountry)

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(getFirstValidationError(err)))
	}

	res, err := cApiV1.billingService.GetVatReportsForCountry(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}

// Get transactions for vat report
// GET /admin/api/v1/vat_reports/details/5ced34d689fce60bf4440829
func (cApiV1 *vatReportsRoute) getVatReportTransactions(ctx echo.Context) error {
	req := &grpc.VatTransactionsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	req.VatReportId = ctx.Param(requestParameterId)

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(getFirstValidationError(err)))
	}

	res, err := cApiV1.billingService.GetVatReportTransactions(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}

// Manually change status for vat report
// (only "paid" and "canceled" values are accepted as correct status)
// POST /admin/api/v1/vat_reports/status/5ced34d689fce60bf4440829
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"status": "paid"}' \
//      https://api.paysuper.online/admin/api/v1/vat_reports/status/5ced34d689fce60bf4440829
//
func (cApiV1 *vatReportsRoute) updateVatReportStatus(ctx echo.Context) error {
	req := &grpc.UpdateVatReportStatusRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	req.Id = ctx.Param(requestParameterId)

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(getFirstValidationError(err)))
	}

	res, err := cApiV1.billingService.UpdateVatReportStatus(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}
