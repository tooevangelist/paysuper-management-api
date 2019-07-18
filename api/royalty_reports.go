package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type royaltyReportsRoute struct {
	*Api
}

func (api *Api) initRoyaltyReportsRoutes() *Api {
	cApiV1 := royaltyReportsRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/royalty_reports", cApiV1.getRoyaltyReportsList)
	api.authUserRouteGroup.GET("/royalty_reports/details/:id", cApiV1.listRoyaltyReportOrders)

	return api
}

// Get vat reports for country
// GET /admin/api/v1/royalty_reports
func (cApiV1 *royaltyReportsRoute) getRoyaltyReportsList(ctx echo.Context) error {
	req := &grpc.ListRoyaltyReportsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.ListRoyaltyReports(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}

// Get transactions for vat report
// GET /admin/api/v1/royalty_reports/details/5ced34d689fce60bf4440829
func (cApiV1 *royaltyReportsRoute) listRoyaltyReportOrders(ctx echo.Context) error {
	req := &grpc.ListRoyaltyReportOrdersRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	req.ReportId = ctx.Param(requestParameterId)

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.ListRoyaltyReportOrders(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}
