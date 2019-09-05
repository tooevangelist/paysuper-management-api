package api

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	reporterProto "github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"net/http"
)

type reportFileRoute struct {
	*Api
}

type reportFileRequest struct {
	Id         string                 `json:"id" form:"id" bson:"_id"`
	MerchantId string                 `json:"merchant_id" form:"merchant_id" bson:"merchant_id"`
	FileType   string                 `json:"file_type" form:"file_type" bson:"file_type"`
	ReportType string                 `json:"report_type" form:"report_type" bson:"report_type"`
	Template   string                 `json:"template" form:"template" bson:"template"`
	Params     map[string]interface{} `json:"params" form:"params" bson:"params"`
}

func (api *Api) initReportFileRoute() *Api {
	route := &reportFileRoute{
		Api: api,
	}

	api.authUserRouteGroup.POST("/report_file", route.create)
	api.authUserRouteGroup.GET("/report_file/:id", route.download)

	return api
}

// Send a request to create a report for download.
// POST /admin/api/v1/report_file
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"file_type": "pdf", "period_from": 1566727410, "period_to": "1566736763"}' \
//      https://api.paysuper.online/admin/api/v1/report_file
//
func (r *reportFileRoute) create(ctx echo.Context) error {
	req1 := &grpc.GetMerchantByRequest{UserId: r.authUser.Id}
	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), req1)
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.L().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.Error(err),
				zap.String(ErrorFieldService, pkg.ServiceName),
				zap.String(ErrorFieldMethod, "GetMerchantBy"),
				zap.Any(ErrorFieldRequest, req1),
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageMerchantNotFound)
	}

	data := &reportFileRequest{}
	if err := ctx.Bind(data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	params, err := json.Marshal(data.Params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorRequestDataInvalid)
	}

	err = r.validate.Struct(data)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req2 := &reporterProto.ReportFile{
		MerchantId: merchant.Item.Id,
		ReportType: data.ReportType,
		FileType:   data.FileType,
		Template:   data.Template,
		Params:     params,
	}

	res, err := r.reporterService.CreateFile(ctx.Request().Context(), req2)
	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "CreateFile"),
			zap.Any(ErrorFieldRequest, req1),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageCreateReportFile)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Send a request to create a report for download.
// GET /admin/api/v1/vat_reports/report/download/5ced34d689fce60bf4440829
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      https://api.paysuper.online/admin/api/v1/report_file/5ced34d689fce60bf4440829
//
func (r *reportFileRoute) download(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		zap.S().Error("unable to find the file id")
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req1 := &grpc.GetMerchantByRequest{UserId: r.authUser.Id}
	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), req1)
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.L().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.Error(err),
				zap.String(ErrorFieldService, pkg.ServiceName),
				zap.String(ErrorFieldMethod, "GetMerchantBy"),
				zap.Any(ErrorFieldRequest, req1),
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageMerchantNotFound)
	}

	req2 := &reporterProto.LoadFileRequest{
		Id:         id,
		MerchantId: merchant.Item.Id,
	}

	if err = r.validate.Struct(req2); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.reporterService.LoadFile(ctx.Request().Context(), req2)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageDownloadReportFile)
	}

	return ctx.Blob(http.StatusOK, res.ContentType, res.File.File)
}
