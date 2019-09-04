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
	api.authUserRouteGroup.POST("/royalty_reports/:id/accept", cApiV1.MerchantReviewRoyaltyReport)
	api.authUserRouteGroup.POST("/royalty_reports/:id/decline", cApiV1.merchantDeclineRoyaltyReport)
	api.authUserRouteGroup.POST("/royalty_reports/:id/change", cApiV1.changeRoyaltyReport)

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

// Accept royalty report by merchant
// POST /admin/api/v1/royalty_reports/5ced34d689fce60bf4440829/accept
func (cApiV1 *royaltyReportsRoute) MerchantReviewRoyaltyReport(ctx echo.Context) error {

	req := &grpc.MerchantReviewRoyaltyReportRequest{
		IsAccepted: true,
		Ip:         ctx.RealIP(),
		ReportId:   ctx.Param(requestParameterId),
	}

	res, err := cApiV1.billingService.MerchantReviewRoyaltyReport(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}

// Decline royalty report by merchant and start a dispute
// POST /admin/api/v1/royalty_reports/5ced34d689fce60bf4440829/decline
func (cApiV1 *royaltyReportsRoute) merchantDeclineRoyaltyReport(ctx echo.Context) error {

	req := &grpc.MerchantReviewRoyaltyReportRequest{
		IsAccepted: false,
		Ip:         ctx.RealIP(),
		ReportId:   ctx.Param(requestParameterId),
	}

	res, err := cApiV1.billingService.MerchantReviewRoyaltyReport(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}

// Change royalty report by admin
// POST /admin/api/v1/royalty_reports/5ced34d689fce60bf4440829/change
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"status": "Accepted", "correction": {"amount": 100500, "reason": "just for fun :)"}, payout_id: "5bdc39a95d1e1100019fb7df"}' \
//      https://api.paysuper.online/admin/api/v1/royalty_reports/5ced34d689fce60bf4440829/change
func (cApiV1 *royaltyReportsRoute) changeRoyaltyReport(ctx echo.Context) error {
	req := &grpc.ChangeRoyaltyReportRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	req.ReportId = ctx.Param(requestParameterId)
	req.Ip = ctx.RealIP()

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.ChangeRoyaltyReport(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}
