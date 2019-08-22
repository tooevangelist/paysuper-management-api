package api

import (
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	reporterPkg "github.com/paysuper/paysuper-reporter/pkg"
	reporterProto "github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/paysuper/paysuper-tax-service/proto"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type taxesRoute struct {
	*Api
	mClt *minio.Client
}

func (api *Api) initTaxesRoutes() *Api {
	route := &taxesRoute{Api: api}

	api.authUserRouteGroup.GET("/taxes", route.getTaxes)
	api.authUserRouteGroup.POST("/taxes", route.setTax)
	api.authUserRouteGroup.DELETE("/taxes/:id", route.deleteTax)

	api.authUserRouteGroup.POST("/taxes/report/download", route.createTaxReportFile)
	api.authUserRouteGroup.GET("/taxes/report/download/:id", route.downloadTaxReportFile)

	return api
}

func (r *taxesRoute) getTaxes(ctx echo.Context) error {
	req := r.bindGetTaxes(ctx)
	res, err := r.taxService.GetRates(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.JSON(http.StatusOK, res.Rates)
}

func (r *taxesRoute) bindGetTaxes(ctx echo.Context) *tax_service.GetRatesRequest {
	structure := &tax_service.GetRatesRequest{}

	params := ctx.QueryParams()

	if v, ok := params["country"]; ok {
		structure.Country = string(v[0])
	}

	if v, ok := params["city"]; ok {
		structure.City = string(v[0])
	}

	if v, ok := params["state"]; ok {
		structure.State = string(v[0])
	}

	if v, ok := params["zip"]; ok {
		structure.Zip = string(v[0])
	}

	if v, ok := params[requestParameterLimit]; ok {
		if i, err := strconv.ParseInt(v[0], 10, 32); err == nil {
			structure.Limit = int32(i)
		}
	} else {
		structure.Limit = LimitDefault
	}

	if v, ok := params[requestParameterOffset]; ok {
		if i, err := strconv.ParseInt(v[0], 10, 32); err == nil {
			structure.Offset = int32(i)
		}
	} else {
		structure.Offset = OffsetDefault
	}

	return structure
}

func (r *taxesRoute) setTax(ctx echo.Context) error {
	req := &tax_service.TaxRate{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	res, err := r.taxService.CreateOrUpdate(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (r *taxesRoute) deleteTax(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	value, err := strconv.Atoi(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	res, err := r.taxService.DeleteRateById(ctx.Request().Context(), &tax_service.DeleteRateRequest{Id: uint32(value)})
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (r *taxesRoute) createTaxReportFile(ctx echo.Context) error {
	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("unable to find merchant by user", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageMerchantNotFound)
	}

	req := &reporterProto.CreateFileRequest{
		ReportType: reporterPkg.ReportTypeTax,
		MerchantId: merchant.Item.Id,
	}

	if err = ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	res, err := r.reporterService.CreateFile(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageCreateReportFile)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (r *taxesRoute) downloadTaxReportFile(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		zap.S().Error("unable to find the file id")
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("unable to find merchant by user", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageMerchantNotFound)
	}

	req := &reporterProto.GetFileRequest{
		Id:         id,
		MerchantId: merchant.Item.Id,
	}

	if err = r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.reporterService.LoadFile(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageDownloadReportFile)
	}

	return ctx.Blob(http.StatusOK, res.ContentType, res.File.File)
}
