package api

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/config"
	reporterMocks "github.com/paysuper/paysuper-reporter/pkg/mocks"
	reporterProto "github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ReportFileTestSuite struct {
	suite.Suite
	handler *reportFileRoute
	api     *Api
}

func Test_ReportFile(t *testing.T) {
	suite.Run(t, new(ReportFileTestSuite))
}

func (suite *ReportFileTestSuite) SetupTest() {
	suite.api = &Api{
		Http:     echo.New(),
		validate: validator.New(),
		authUser: &AuthUser{Id: "ffffffffffffffffffffffff"},
		config: &config.Config{
			AwsBucketReporter:          "test",
			AwsRegionReporter:          "test",
			AwsSecretAccessKeyReporter: "test",
			AwsAccessKeyIdReporter:     "test",
		},
	}

	suite.api.accessRouteGroup = suite.api.Http.Group("/api/v1/s")
	suite.handler = &reportFileRoute{Api: suite.api}
}

func (suite *ReportFileTestSuite) Test_Routes() {
	shouldHaveRoutes := [][]string{
		{"/api/v1/s/report_file", http.MethodPost},
		{"/api/v1/s/report_file/download/:file", http.MethodGet},
	}

	api := suite.api.initReportFileRoute()

	routeCount := 0

	routes := api.Http.Routes()
	for _, v := range shouldHaveRoutes {
		for _, r := range routes {
			if v[0] != r.Path || v[1] != r.Method {
				continue
			}

			routeCount++
		}
	}

	assert.Equal(suite.T(), len(shouldHaveRoutes), routeCount)
}

func (suite *ReportFileTestSuite) TestReportFile_create_Error_CreateFile() {
	data := `{"period_from": 1, "period_to": 2}`
	req := httptest.NewRequest(http.MethodPost, "/report_file", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	reporterService := &reporterMocks.ReporterService{}
	reporterService.
		On("CreateFile", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.api.reporterService = reporterService

	err := suite.handler.create(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Regexp(suite.T(), errorMessageCreateReportFile.Message, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_create_Ok() {
	data := `{"period_from": 1, "period_to": 2}`
	req := httptest.NewRequest(http.MethodPost, "/report_file", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	reporterService := &reporterMocks.ReporterService{}
	reporterService.
		On("CreateFile", mock2.Anything, mock2.Anything).
		Return(&reporterProto.CreateFileResponse{FileId: bson.NewObjectId().Hex()}, nil)
	suite.api.reporterService = reporterService

	err := suite.handler.create(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_EmptyId() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/download/:file", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.handler.download(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), errorRequestParamsIncorrect.Message, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_ValidationFileEmpty() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/download/:file", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/report_file/download/:file")
	ctx.SetParamNames(requestParameterFile)
	ctx.SetParamValues("")

	err := suite.handler.download(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), errorRequestParamsIncorrect, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_ValidationFileIncorrect() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/download/:file", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/report_file/download/:file")
	ctx.SetParamNames(requestParameterFile)
	ctx.SetParamValues("test")

	err := suite.handler.download(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), errorRequestParamsIncorrect, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/download/:file", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/report_file/download/:file")
	ctx.SetParamNames(requestParameterFile)
	ctx.SetParamValues("string.csv")

	err := suite.handler.download(ctx)
	assert.NoError(suite.T(), err)
}
