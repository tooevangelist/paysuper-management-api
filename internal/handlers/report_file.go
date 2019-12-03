package handlers

import (
	"fmt"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	reporterPkg "github.com/paysuper/paysuper-reporter/pkg"
	"net/http"
	"os"
	"strings"
)

const (
	reportFileDownloadPath = "/report_file/download/:file"
)

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
	groups.AuthUser.GET(reportFileDownloadPath, h.download)
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
