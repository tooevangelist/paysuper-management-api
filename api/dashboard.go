package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
)

type dashboardRoute struct {
	*Api
}

func (api *Api) initDashboardRoutes() *Api {
	route := &dashboardRoute{Api: api}

	api.authUserRouteGroup.GET("/merchants/:id/dashboard/main", route.getMainReports)
	api.authUserRouteGroup.GET("/merchants/:id/dashboard/revenue_dynamics", route.getRevenueDynamicsReport)
	api.authUserRouteGroup.GET("/merchants/:id/dashboard/base", route.getBaseReports)

	return api
}

// @Description get main reports data for dashboard
// @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' \
//  https://api.paysuper.online/admin/api/v1/merchants/ffffffffffffffffffffffff/dashboard/main?period=previous_month
func (r *dashboardRoute) getMainReports(ctx echo.Context) error {
	req := &grpc.GetDashboardMainRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.MerchantId = ctx.Param(requestParameterId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.GetDashboardMainReport(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetDashboardMainReport"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

// @Description get revenue dynamics report data for dashboard
// @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' \
//  'http://127.0.0.1:3001/admin/api/v1/merchants/ffffffffffffffffffffffff/dashboard/revenue_dynamics?period=previous_month'
func (r *dashboardRoute) getRevenueDynamicsReport(ctx echo.Context) error {
	req := &grpc.GetDashboardMainRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.MerchantId = ctx.Param(requestParameterId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.GetDashboardRevenueDynamicsReport(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetDashboardMainReport"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

// @Description get base reports data for dashboard
// @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' \
//  https://api.paysuper.online/admin/api/v1/merchants/ffffffffffffffffffffffff/dashboard/base?period=previous_month
func (r *dashboardRoute) getBaseReports(ctx echo.Context) error {
	req := &grpc.GetDashboardBaseReportRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.MerchantId = ctx.Param(requestParameterId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.GetDashboardBaseReport(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetDashboardMainReport"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}
