package api

import (
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	reporterProto "github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"net/http"
)

type reportFileRoute struct {
	*Api
}

type reportFileRequest struct {
	Id         bson.ObjectId          `bson:"_id"`
	FileType   string                 `bson:"file_type"`
	ReportType string                 `bson:"report_type"`
	Template   string                 `bson:"template"`
	Params     map[string]interface{} `bson:"params"`
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
	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("unable to find merchant by user", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageMerchantNotFound)
	}

	data := &reportFileRequest{}
	if err := ctx.Bind(data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	params, err := json.Marshal(data.Params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorRequestDataInvalid)
	}

	req := &reporterProto.ReportFile{
		Id:         data.Id.Hex(),
		ReportType: data.ReportType,
		FileType:   data.FileType,
		Template:   data.Template,
		Params:     params,
	}
	res, err := r.reporterService.CreateFile(ctx.Request().Context(), req)
	if err != nil {
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

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("unable to find merchant by user", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageMerchantNotFound)
	}

	req := &reporterProto.LoadFileRequest{
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
