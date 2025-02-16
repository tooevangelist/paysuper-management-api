package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	reporterPkg "github.com/paysuper/paysuper-reporter/pkg"
	reporterProto "github.com/paysuper/paysuper-reporter/pkg/proto"
	"net/http"
	"os"
	"strings"
)

const (
	reportFilePath         = "/report_file"
	reportFileDownloadPath = "/report_file/download/:file"
)

type reportFileRequest struct {
	MerchantId string                 `json:"merchant_id" form:"merchant_id" bson:"merchant_id" validate:"required,hexadecimal,len=24"`
	FileType   string                 `json:"file_type" form:"file_type" bson:"file_type" validate:"required"`
	ReportType string                 `json:"report_type" form:"report_type" bson:"report_type" validate:"required"`
	Template   string                 `json:"template" form:"template" bson:"template"`
	Params     map[string]interface{} `json:"params" form:"params" bson:"params"`
}

type ReportFileRoute struct {
	dispatch   common.HandlerSet
	awsManager awsWrapper.AwsManagerInterface
	cfg        common.Config
	provider.LMT
}

func NewReportFileRoute(set common.HandlerSet, awsManager awsWrapper.AwsManagerInterface, cfg *common.Config) *ReportFileRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "ReportFileRoute"})
	return &ReportFileRoute{
		dispatch:   set,
		LMT:        &set.AwareSet,
		cfg:        *cfg,
		awsManager: awsManager,
	}
}

func (h *ReportFileRoute) Route(groups *common.Groups) {
	groups.AuthUser.POST(reportFilePath, h.create)
	groups.AuthUser.GET(reportFileDownloadPath, h.download)
}

func (h *ReportFileRoute) create(ctx echo.Context) error {
	data := &reportFileRequest{}

	if err := ctx.Bind(data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	params, err := json.Marshal(data.Params)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorRequestDataInvalid)
	}

	if err = h.dispatch.Validate.Struct(data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	req := &reporterProto.ReportFile{
		UserId:           common.ExtractUserContext(ctx).Id,
		MerchantId:       data.MerchantId,
		ReportType:       data.ReportType,
		FileType:         data.FileType,
		Template:         data.Template,
		Params:           params,
		SendNotification: true,
	}

	res, err := h.dispatch.Services.Reporter.CreateFile(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, reporterPkg.ServiceName, "CreateFile", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageCreateReportFile)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *ReportFileRoute) download(ctx echo.Context) error {
	file := ctx.Param("file")

	if file == "" {
		h.L().Error("unable to find the file")
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	params := strings.Split(file, ".")

	if len(params) != 2 {
		h.L().Error("incorrect of file string")
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	fileName := fmt.Sprintf(reporterPkg.FileMask, common.ExtractUserContext(ctx).Id, params[0], params[1])
	filePath := os.TempDir() + string(os.PathSeparator) + fileName
	_, err := h.awsManager.Download(ctx.Request().Context(), filePath, &awsWrapper.DownloadInput{FileName: fileName})

	if err != nil {
		h.L().Error("unable to download the file " + fileName + " with message: " + err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageDownloadReportFile)
	}

	defer os.Remove(filePath)
	return ctx.File(filePath)
}
