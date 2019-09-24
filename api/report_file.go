package api

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	"github.com/paysuper/paysuper-billing-server/pkg"
	reporterPkg "github.com/paysuper/paysuper-reporter/pkg"
	reporterProto "github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
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

	api.accessRouteGroup.POST("/report_file", route.create)
	api.accessRouteGroup.GET("/report_file/download/:file", route.download)

	return api
}

// Send a request to create a report for download.
// POST /api/v1/s/report_file
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"file_type": "pdf", "period_from": 1566727410, "period_to": "1566736763"}' \
//      https://api.paysuper.online/api/v1/s/report_file
//
func (r *reportFileRoute) create(ctx echo.Context) error {
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

	req := &reporterProto.ReportFile{
		UserId:           r.authUser.Id,
		MerchantId:       data.MerchantId,
		ReportType:       data.ReportType,
		FileType:         data.FileType,
		Template:         data.Template,
		Params:           params,
		SendNotification: true,
	}

	res, err := r.reporterService.CreateFile(ctx.Request().Context(), req)
	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, reporterPkg.ServiceName),
			zap.String(ErrorFieldMethod, "CreateFile"),
			zap.Any(ErrorFieldRequest, req),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageCreateReportFile)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Send a request to create a report for download.
// GET /api/v1/s/report_file/download/5ced34d689fce60bf4440829.csv
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      https://api.paysuper.online/api/v1/s/report_file/download/5ced34d689fce60bf4440829.csv
//
func (r *reportFileRoute) download(ctx echo.Context) error {
	file := ctx.Param("file")
	if file == "" {
		zap.S().Error("unable to find the file")
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	params := strings.Split(file, ".")

	if len(params) != 2 {
		zap.S().Error("incorrect of file string")
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	awsOptions := []awsWrapper.Option{
		awsWrapper.AccessKeyId(r.config.AwsAccessKeyIdReporter),
		awsWrapper.SecretAccessKey(r.config.AwsSecretAccessKeyReporter),
		awsWrapper.Region(r.config.AwsRegionReporter),
		awsWrapper.Bucket(r.config.AwsBucketReporter),
	}
	awsManager, err := awsWrapper.New(awsOptions...)

	if err != nil {
		zap.S().Error("unable to find the file id")
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageDownloadReportFile)
	}

	fileName := fmt.Sprintf(reporterPkg.FileMask, r.authUser.Id, params[0], params[1])
	filePath := os.TempDir() + string(os.PathSeparator) + fileName
	_, err = awsManager.Download(ctx.Request().Context(), filePath, &awsWrapper.DownloadInput{FileName: fileName})

	defer os.Remove(filePath)

	return ctx.File(filePath)
}
