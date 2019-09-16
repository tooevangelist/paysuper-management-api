package api

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	billingMocks "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
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
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.handler = &reportFileRoute{Api: suite.api}
}

func (suite *ReportFileTestSuite) Test_Routes() {
	shouldHaveRoutes := [][]string{
		{"/admin/api/v1/report_file", http.MethodPost},
		{"/admin/api/v1/report_file/:id", http.MethodGet},
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

	billingService := &billingMocks.BillingService{}
	billingService.
		On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Item: &billing.Merchant{Id: bson.NewObjectId().Hex()}}, nil)
	suite.api.billingService = billingService

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

	billingService := &billingMocks.BillingService{}
	billingService.
		On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Item: &billing.Merchant{Id: bson.NewObjectId().Hex()}}, nil)
	suite.api.billingService = billingService

	reporterService := &reporterMocks.ReporterService{}
	reporterService.
		On("CreateFile", mock2.Anything, mock2.Anything).
		Return(&reporterProto.CreateFileResponse{FileId: bson.NewObjectId().Hex()}, nil)
	suite.api.reporterService = reporterService

	err := suite.handler.create(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_EmptyId() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.handler.download(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), errorRequestParamsIncorrect.Message, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_ValidationId() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/taxes/report/download/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("string")

	billingService := &billingMocks.BillingService{}
	billingService.
		On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Item: &billing.Merchant{Id: bson.NewObjectId().Hex()}}, nil)
	suite.api.billingService = billingService

	err := suite.handler.download(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "validation failed", httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_ValidationMerchantId() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/report_file/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &billingMocks.BillingService{}
	billingService.
		On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Item: &billing.Merchant{Id: "string"}}, nil)
	suite.api.billingService = billingService

	err := suite.handler.download(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "validation failed", httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Error_DownloadReportFile() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/report_file/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &billingMocks.BillingService{}
	billingService.
		On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Item: &billing.Merchant{Id: bson.NewObjectId().Hex()}}, nil)
	suite.api.billingService = billingService

	reporterService := &reporterMocks.ReporterService{}
	reporterService.
		On("LoadFile", mock2.Anything, mock2.Anything).
		Return(nil, errors.New("error"))
	suite.api.reporterService = reporterService

	err := suite.handler.download(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Regexp(suite.T(), errorMessageDownloadReportFile.Message, httpErr.Message)
}

func (suite *ReportFileTestSuite) TestReportFile_download_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/report_file/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/taxes/report/download/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &billingMocks.BillingService{}
	billingService.
		On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Item: &billing.Merchant{Id: bson.NewObjectId().Hex()}}, nil)
	suite.api.billingService = billingService

	reporterService := &reporterMocks.ReporterService{}
	reporterService.
		On("LoadFile", mock2.Anything, mock2.Anything).
		Return(&reporterProto.LoadFileResponse{File: &reporterProto.File{}}, nil)
	suite.api.reporterService = reporterService

	err := suite.handler.download(ctx)
	assert.NoError(suite.T(), err)
}
